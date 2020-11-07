// TODO: obviously move all this into its separate pkg
package p2p

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/iotaledger/hive.go/configuration"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"go.uber.org/dig"

	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoreds"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"

	"github.com/gohornet/hornet/pkg/node"
	"github.com/iotaledger/hive.go/logger"

	p2ppkg "github.com/gohornet/hornet/pkg/p2p"
	"github.com/gohornet/hornet/pkg/shutdown"
	"github.com/gohornet/hornet/pkg/utils"
)

func init() {
	CorePlugin = &node.CorePlugin{
		Pluggable: node.Pluggable{
			Name:      "P2P",
			DepsFunc:  func(cDeps dependencies) { deps = cDeps },
			Params:    params,
			Provide:   provide,
			Configure: configure,
			Run:       run,
		},
	}
}

var (
	CorePlugin        *node.CorePlugin
	log               *logger.Logger
	deps              dependencies
	ErrNoPrivKeyFound = errors.New("no private key found")
)

const (
	pubKeyFileName = "key.pub"
)

type dependencies struct {
	dig.In
	Manager       *p2ppkg.Manager
	Host          host.Host
	NodeConfig    *configuration.Configuration `name:"nodeConfig"`
	PeeringConfig *configuration.Configuration `name:"peeringConfig"`
}

func provide(c *dig.Container) {
	log = logger.NewLogger(CorePlugin.Name)

	type hostdeps struct {
		dig.In

		NodeConfig *configuration.Configuration `name:"nodeConfig"`
	}

	if err := c.Provide(func(deps hostdeps) (host.Host, error) {

		ctx := context.Background()

		peerStorePath := deps.NodeConfig.String(CfgP2PPeerStorePath)
		_, statPeerStorePathErr := os.Stat(peerStorePath)

		// TODO: switch out with impl. using KVStore
		defaultOpts := badger.DefaultOptions
		// needed under Windows otherwise peer store is 'corrupted' after a restart
		defaultOpts.Truncate = runtime.GOOS == "windows"
		badgerStore, err := badger.NewDatastore(peerStorePath, &defaultOpts)
		if err != nil {
			panic(fmt.Sprintf("unable to initialize data store for peer store: %s", err))
		}

		// also takes care of this node's identity key pair
		peerStore, err := pstoreds.NewPeerstore(ctx, badgerStore, pstoreds.DefaultOpts())
		if err != nil {
			panic(fmt.Sprintf("unable to initialize peer store: %s", err))
		}

		// make sure nobody copies around the peer store since it contains the
		// private key of the node
		log.Infof("never share your %s folder as it contains your node's private key!", peerStorePath)

		// load up the previously generated identity or create a new one
		isPeerStoreNew := os.IsNotExist(statPeerStorePathErr)
		prvKey, err := loadOrCreateIdentity(deps.NodeConfig, isPeerStoreNew, peerStorePath, peerStore)
		if err != nil {
			panic(fmt.Sprintf("unable to load/create peer identity: %s", err))
		}

		bindAddrs := deps.NodeConfig.Strings(CfgP2PBindMultiAddresses)

		createdHost, err := libp2p.New(ctx,
			libp2p.Identity(prvKey),
			libp2p.ListenAddrStrings(bindAddrs...),
			libp2p.Peerstore(peerStore),
			libp2p.Transport(libp2pquic.NewTransport),
			libp2p.DefaultTransports,
			libp2p.ConnectionManager(connmgr.NewConnManager(
				deps.NodeConfig.Int(CfgP2PConnMngLowWatermark),
				deps.NodeConfig.Int(CfgP2PConnMngHighWatermark),
				time.Minute,
			)),
			libp2p.NATPortMap(),
		)
		createdHost.ID()

		if err != nil {
			return nil, fmt.Errorf("unable to initialize peer: %w", err)
		}

		return createdHost, nil
	}); err != nil {
		panic(err)
	}

	type mngdeps struct {
		dig.In

		Host   host.Host
		Config *configuration.Configuration `name:"nodeConfig"`
	}

	if err := c.Provide(func(deps mngdeps) *p2ppkg.Manager {
		return p2ppkg.NewManager(deps.Host,
			p2ppkg.WithManagerLogger(logger.NewLogger("P2P-Manager")),
			p2ppkg.WithManagerReconnectInterval(time.Duration(deps.Config.Int(CfgP2PReconnectIntervalSeconds))*time.Second, 1*time.Second),
		)
	}); err != nil {
		panic(err)
	}
}

func configure() {
	log.Infof("peer configured, ID: %s", deps.Host.ID())
}

func run() {

	// register a daemon to disconnect all peers up on shutdown
	_ = CorePlugin.Daemon().BackgroundWorker("Manager", func(shutdownSignal <-chan struct{}) {
		log.Infof("listening on: %s", deps.Host.Addrs())
		go deps.Manager.Start(shutdownSignal)
		connectConfigKnownPeers()
		<-shutdownSignal
		if err := deps.Host.Peerstore().Close(); err != nil {
			log.Error("unable to cleanly closing peer store: %s", err)
		}
	}, shutdown.PriorityP2PManager)
}

// connects to the peers defined in the config.
func connectConfigKnownPeers() {
	peerIDsStr := deps.PeeringConfig.Strings(CfgP2PPeers)
	peerAliases := deps.PeeringConfig.Strings(CfgP2PPeerAliases)

	applyAliases := true
	if len(peerIDsStr) != len(peerAliases) {
		log.Warnf("won't apply peer aliases: you must define aliases for all defined static peers (got %d aliases, %d peers).", len(peerAliases), len(peerIDsStr))
		applyAliases = false
	}

	for i, peerIDStr := range peerIDsStr {
		multiAddr, err := multiaddr.NewMultiaddr(peerIDStr)
		if err != nil {
			panic(fmt.Sprintf("invalid config peer address at pos %d: %s", i, err))
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
		if err != nil {
			panic(fmt.Sprintf("invalid config peer address info at pos %d: %s", i, err))
		}

		var alias string
		if applyAliases {
			alias = peerAliases[i]
		}
		_ = deps.Manager.ConnectPeer(addrInfo, p2ppkg.PeerRelationKnown, alias)
	}
}

// creates a new Ed25519 based key pair or loads up the existing identity
// by reading the public key file from disk.
func loadOrCreateIdentity(nodeConfig *configuration.Configuration, peerStoreIsNew bool, peerStorePath string, peerStore peerstore.Peerstore) (crypto.PrivKey, error) {
	pubKeyFilePath := path.Join(peerStorePath, pubKeyFileName)
	if peerStoreIsNew {
		return createIdentity(nodeConfig, pubKeyFilePath)
	}

	return loadExistingIdentity(nodeConfig, pubKeyFilePath, peerStore)
}

func loadIdentityFromConfig(nodeConfig *configuration.Configuration) (crypto.PrivKey, error) {
	if identityPrivKey := nodeConfig.String(CfgP2PIdentityPrivKey); identityPrivKey != "" {
		prvKey, err := utils.ParseEd25519PrivateKeyFromString(identityPrivKey)
		if err != nil {
			return nil, fmt.Errorf("config parameter '%s' contains an invalid private key", CfgP2PIdentityPrivKey)
		}

		sk, _, err := crypto.KeyPairFromStdKey(&prvKey)
		if err != nil {
			return nil, fmt.Errorf("unable to load Ed25519 key pair for peer identity: %w", err)
		}

		return sk, nil
	}

	return nil, ErrNoPrivKeyFound
}

// creates a new Ed25519 based identity and saves the public key
// as a separate file next to the peer store data.
func createIdentity(nodeConfig *configuration.Configuration, pubKeyFilePath string) (crypto.PrivKey, error) {

	log.Info("generating a new peer identity...")

	sk, err := loadIdentityFromConfig(nodeConfig)
	if err != nil {
		if err != ErrNoPrivKeyFound {
			return nil, err
		}

		sk, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			return nil, fmt.Errorf("unable to generate Ed25519 key pair for peer identity: %w", err)
		}
	}

	// even though the crypto.PrivKey is going to get stored
	// within the peer store, there is no way to retrieve the node's
	// identity via the peer store, so we must save the public key
	// separately to retrieve it later again
	// https://discuss.libp2p.io/t/generating-peer-id/111/2
	pubKeyPb, err := crypto.MarshalPublicKey(sk.GetPublic())
	if err != nil {
		return nil, fmt.Errorf("unable to marshal public key for public key identity file: %w", err)
	}

	if err := ioutil.WriteFile(pubKeyFilePath, pubKeyPb, 0666); err != nil {
		return nil, fmt.Errorf("unable to save public key identity file: %w", err)
	}

	log.Infof("stored public key under %s", pubKeyFilePath)
	return sk, nil
}

// loads an existing identity by reading in the public key from the public key identity file
// and then retrieving the associated private key from the given Peerstore.
func loadExistingIdentity(nodeConfig *configuration.Configuration, pubKeyFilePath string, peerStore peerstore.Peerstore) (crypto.PrivKey, error) {
	log.Infof("retrieving existing peer identity from %s", pubKeyFilePath)
	existingPubKeyBytes, err := ioutil.ReadFile(pubKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read public key identity file: %w", err)
	}

	pubKey, err := crypto.UnmarshalPublicKey(existingPubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal public key from public key identity file: %w", err)
	}
	peerID, err := peer.IDFromPublicKey(pubKey)

	// retrieve this node's private key from the peer store
	storedPrivKey := peerStore.PrivKey(peerID)

	// load an optional private key from the config and compare it to the stored private key
	sk, err := loadIdentityFromConfig(nodeConfig)
	if err != nil {
		if err != ErrNoPrivKeyFound {
			return nil, err
		}

		return storedPrivKey, nil
	}

	if !storedPrivKey.Equals(sk) {
		storedPrivKeyBytes, _ := storedPrivKey.Bytes()
		configPrivKeyBytes, _ := sk.Bytes()
		return nil, fmt.Errorf("stored Ed25519 private key (%s) for peer identity doesn't match private key in config (%s)", hex.EncodeToString(storedPrivKeyBytes[:]), hex.EncodeToString(configPrivKeyBytes[:]))
	}

	return storedPrivKey, nil
}
