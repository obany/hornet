package binary

import "github.com/gohornet/hornet/pkg/bech32/address"

type UTXOInput struct {
	InputType              uint64
	TransactionId          [32]byte
	TransactionOutPutIndex uint64
}

type SigLockedSingleDeposit struct {
	OutputType uint64
	Address    address.Address
	Amount     uint64
}
