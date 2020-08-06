package binary

import (
	"io"
)

type SignedTransaction struct {
	Transaction  *UnsignedTransaction
	UnlockBlocks UnlockBlocks
}

type UnsignedTransaction struct {
	TransactionType uint64
	Inputs          TransactionInputs
	Outputs         TransactionOutputs
	Payload         Payload
}

type TransactionInputs []*UTXOInput
type TransactionOutputs []*SigLockedSingleDeposit
type UnlockBlocks []UnlockBlock

type UnlockBlock interface {
	GetLength() uint64
}

func (u *UnsignedTransaction) GetLength() uint64 {
	return varIntLength(u.TransactionType) +
		u.Inputs.GetLength() +
		u.Outputs.GetLength() +
		u.Payload.GetLength()
}

func (i TransactionInputs) GetLength() uint64 {
	len := varIntLength(uint64(len(i)))
	for _, input := range i {
		len += input.GetLength()
	}
	return len
}

func (o TransactionOutputs) GetLength() uint64 {
	len := varIntLength(uint64(len(o)))
	for _, output := range o {
		len += output.GetLength()
	}
	return len
}

func (b UnlockBlocks) GetLength() uint64 {
	len := varIntLength(uint64(len(b)))
	for _, blk := range b {
		len += blk.GetLength()
	}
	return len
}

func (s *SignedTransaction) GetLength() uint64 {
	return payloadTypeLength(s) +
		s.Transaction.GetLength() +
		s.UnlockBlocks.GetLength()
}

func (s *SignedTransaction) GetType() PayloadType {
	return PayloadTypeSignedTransaction
}

func readSignedTransaction(r io.Reader) (*SignedTransaction, error) {

	if err := readAndCheckPayloadType(r, PayloadTypeSignedTransaction); err != nil {
		return nil, err
	}

	// Read Unsigned Transaction
	transactionType, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	if transactionType != 0 {
		return nil, ErrInvalidPayloadValue
	}

	inputsCount, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	var inputs TransactionInputs
	for i := uint64(0); i < inputsCount; i++ {
		input, err := readUTXOInput(r)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}

	outputsCount, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	var outputs TransactionOutputs
	for i := uint64(0); i < outputsCount; i++ {
		output, err := readSigLockedSingleDeposit(r)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
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

	unsignedTransaction := &UnsignedTransaction{
		TransactionType: transactionType,
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

func (s *SignedTransaction) Write(w io.Writer) error {

	if err := writePayloadType(w, s.GetType()); err != nil {
		return err
	}

}
