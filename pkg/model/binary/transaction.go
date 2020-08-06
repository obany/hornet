package binary

import (
	"io"
)

type UnsignedTransaction struct {
	TransactionType uint64
	Inputs          []UTXOInput
	Outputs         []SigLockedSingleDeposit
	Payload         Payload
}

type UnlockBlock interface{}

type SignedTransaction struct {
	Transaction  UnsignedTransaction
	UnlockBlocks []UnlockBlock
}

func ReadSignedTransaction(r io.Reader) (*SignedTransaction, error) {

	if err := readAndCheckPayloadType(r, PayloadTypeSignedTransaction); err != nil {
		return nil, err
	}

	// Read Unsigned Transaction

	// TODO: ask why this is needed

	// We need to read the type again
	if err := readAndCheckPayloadType(r, PayloadTypeSignedTransaction); err != nil {
		return nil, err
	}

	inputsCount, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	var inputs []UTXOInput
	for i := uint64(0); i < inputsCount; i++ {

	}

	outputsCount, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	var outputs []SigLockedSingleDeposit
	for i := uint64(0); i < outputsCount; i++ {

	}

	payload, err := readPayload(r)
	if err != nil && err != ErrEmptyPayload {
		return nil, err
	}

	if err != ErrEmptyPayload {
		// Verify the sub-payload type
		switch payload.GetType() {
		case PayloadTypeUnsignedData, PayloadTypeSignedData, PayloadTypeIndexation:
			break
		default:
			return nil, ErrInvalidSubPayload
		}
	}

	unsignedTransaction := UnsignedTransaction{
		TransactionType: 0,
		Inputs:          inputs,
		Outputs:         outputs,
		Payload:         payload,
	}

	// Unlock Blocks
	unlockBlocksCount, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	var unlockBlocks []UnlockBlock
	for i := uint64(0); i < unlockBlocksCount; i++ {

	}

	signedTransaction := &SignedTransaction{
		Transaction:  unsignedTransaction,
		UnlockBlocks: unlockBlocks,
	}

	return signedTransaction, nil
}
