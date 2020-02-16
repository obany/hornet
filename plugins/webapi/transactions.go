package webapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"

	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/trinary"

	"github.com/gohornet/hornet/packages/model/tangle"
	"github.com/gohornet/hornet/packages/parameter"
	"github.com/gohornet/hornet/plugins/gossip"
)

func init() {
	addEndpoint("broadcastTransactions", broadcastTransactions, implementedAPIcalls)
	addEndpoint("findTransactions", findTransactions, implementedAPIcalls)
	addEndpoint("storeTransactions", storeTransactions, implementedAPIcalls)
}

func broadcastTransactions(i interface{}, c *gin.Context, abortSignal <-chan struct{}) {

	bt := &BroadcastTransactions{}
	e := ErrorReturn{}

	err := mapstructure.Decode(i, bt)
	if err != nil {
		e.Error = "Internal error"
		c.JSON(http.StatusInternalServerError, e)
		return
	}

	if len(bt.Trytes) == 0 {
		e.Error = "No trytes provided"
		c.JSON(http.StatusBadRequest, e)
		return
	}

	for _, trytes := range bt.Trytes {
		if err := trinary.ValidTrytes(trytes); err != nil {
			e.Error = "Trytes invalid"
			c.JSON(http.StatusBadRequest, e)
			return
		}
	}

	for _, trytes := range bt.Trytes {
		err = gossip.BroadcastTransactionFromAPI(trytes)
		if err != nil {
			e.Error = err.Error()
			c.JSON(http.StatusBadRequest, e)
			return
		}
	}
	c.JSON(http.StatusOK, BradcastTransactionsReturn{})
}

func findTransactions(i interface{}, c *gin.Context, abortSignal <-chan struct{}) {
	ft := &FindTransactions{}
	e := ErrorReturn{}

	maxFindTransactions := parameter.NodeConfig.GetInt("api.maxFindTransactions")

	err := mapstructure.Decode(i, ft)
	if err != nil {
		e.Error = "Internal error"
		c.JSON(http.StatusInternalServerError, e)
		return
	}

	if len(ft.Bundles) > maxFindTransactions || len(ft.Addresses) > maxFindTransactions {
		e.Error = "Too many transaction or bundle hashes. Max. allowed: " + strconv.Itoa(maxFindTransactions)
		c.JSON(http.StatusBadRequest, e)
		return
	}

	txHashes := []string{}

	if len(ft.Bundles) == 0 && len(ft.Addresses) == 0 {
		c.JSON(http.StatusOK, FindTransactionsReturn{Hashes: []string{}})
		return
	}

	// Searching for transactions that contains the given bundle hash
	for _, bdl := range ft.Bundles {
		if err := trinary.ValidTrytes(bdl); err != nil {
			e.Error = fmt.Sprintf("Bundle hash invalid: %s", bdl)
			c.JSON(http.StatusBadRequest, e)
			return
		}

		txHashes = append(txHashes, tangle.GetTransactionHashes(bdl, maxFindTransactions)...)
	}

	// Searching for transactions that contains the given address
	for _, addr := range ft.Addresses {
		err := address.ValidAddress(addr)
		if err == nil {
			if len(addr) == 90 {
				addr = addr[:81]
			}
			tx, err := tangle.ReadTransactionHashesForAddressFromDatabase(addr, maxFindTransactions)
			if err != nil {
				e.Error = "Internal error"
				c.JSON(http.StatusInternalServerError, e)
				return
			}
			txHashes = append(txHashes, tx...)
		}
	}

	// Searching for all approovers of the given transactions
	for _, approveeHash := range ft.Approvees {
		if guards.IsTransactionHash(approveeHash) {
			cachedTxApprovers := tangle.GetCachedApprovers(approveeHash, maxFindTransactions) // approvers +1
			for _, cachedTxApprover := range cachedTxApprovers {
				if !cachedTxApprover.Exists() {
					continue
				}

				txHashes = append(txHashes, cachedTxApprover.GetApprover().GetApproverHash())
			}
			cachedTxApprovers.Release() // approvers -1
		}
	}

	// Searching for transactions that contain the given tag
	for _, tag := range ft.Tags {
		err := trinary.ValidTrytes(tag)
		if err == nil {
			cachedTags := tangle.GetCachedTags(tag, maxFindTransactions) // tags +1
			for _, cachedTag := range cachedTags {
				if !cachedTag.Exists() {
					continue
				}
				txHashes = append(txHashes, cachedTag.GetTag().GetTransactionHash())
			}
			cachedTags.Release() // tags -1
		}
	}

	c.JSON(http.StatusOK, FindTransactionsReturn{Hashes: txHashes})
}

// redirect to broadcastTransactions
func storeTransactions(i interface{}, c *gin.Context, abortSignal <-chan struct{}) {
	broadcastTransactions(i, c, abortSignal)
}
