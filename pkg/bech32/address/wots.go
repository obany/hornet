package address

import (
	"fmt"
	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
)

type wotsAddress hornet.Hash

func (wotsAddress) Version() Version {
	return WOTS
}

func (a wotsAddress) Bytes() []byte {
	return append([]byte{byte(WOTS)}, a...)
}

func (a wotsAddress) Length() int {
	return len(a)
}

func (a wotsAddress) String() string {
	return hornet.Hash(a).Trytes()
}

func validateHash(hash trinary.Hash) error {
	if err := address.ValidAddress(hash); err != nil {
		return err
	}
	// a valid addresses must have the last trit set to zero
	lastTrits := trinary.MustTrytesToTrits(string(hash[consts.HashTrytesSize-1]))
	if lastTrits[consts.TritsPerTryte-1] != 0 {
		return fmt.Errorf("%w: non-zero last trit", consts.ErrInvalidAddress)
	}
	return nil
}

// WOTSAddress creates an Address from the provided W-OTS hash.
func WOTSAddress(hash hornet.Hash) (Address, error) {
	err := validateHash(hash.Trytes())
	if err != nil {
		return nil, err
	}
	return wotsAddress(hash[:49]), nil
}
