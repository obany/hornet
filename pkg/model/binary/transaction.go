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
	GetType() UnlockBlockType
	GetLength() uint64
	Write(w io.Writer) error
}

type Signature interface {
	GetType() AddressAndSignatureType
	GetLength() uint64
	Write(w io.Writer) error
}

func (u *UnsignedTransaction) GetLength() uint64 {
	return varIntLength(u.TransactionType) +
		u.Inputs.GetLength() +
		u.Outputs.GetLength() +
		u.Payload.GetLength()
}

func (u *UnsignedTransaction) Write(w io.Writer) error {

	if err := writeVarint(w, u.TransactionType); err != nil {
		return err
	}

	if err := writeVarint(w, uint64(len(u.Inputs))); err != nil {
		return err
	}

	for _, input := range u.Inputs {
		if err := input.Write(w); err != nil {
			return err
		}
	}

	if err := writeVarint(w, uint64(len(u.Outputs))); err != nil {
		return err
	}

	for _, output := range u.Outputs {
		if err := output.Write(w); err != nil {
			return err
		}
	}

	if u.Payload != nil {
		if err := writeVarint(w, u.Payload.GetLength()); err != nil {
			return err
		}

		if err := u.Payload.Write(w); err != nil {
			return err
		}

	} else {
		// No payload, so set a length of 0
		if err := writeVarint(w, 0); err != nil {
			return err
		}
	}
	return nil
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

func (b UnlockBlocks) Write(w io.Writer) error {
	if err := writeVarint(w, uint64(len(b))); err != nil {
		return err
	}
	for _, blk := range b {
		if err := blk.Write(w); err != nil {
			return err
		}
	}
	return nil
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

	if err := writePayloadType(w, s); err != nil {
		return err
	}

	if err := s.Transaction.Write(w); err != nil {
		return err
	}

	if err := s.UnlockBlocks.Write(w); err != nil {
		return err
	}
	return nil
}

type SignatureUnlockBlock struct {
	Signature Signature
}

func (s *SignatureUnlockBlock) GetType() UnlockBlockType {
	return UnlockBlockTypeSignatureUnlockBlock
}

func (s *SignatureUnlockBlock) GetLength() uint64 {
	return varIntLength(uint64(s.GetType())) + s.Signature.GetLength()
}

func (s *SignatureUnlockBlock) Write(w io.Writer) error {
	if err := writeVarint(w, uint64(s.GetType())); err != nil {
		return err
	}
	if err := s.Signature.Write(w); err != nil {
		return err
	}
	return nil
}

type ReferenceUnlockBlock struct {
	Reference uint64
}

func (r *ReferenceUnlockBlock) GetType() UnlockBlockType {
	return UnlockBlockTypeReferenceUnlockBlock
}

func (r *ReferenceUnlockBlock) GetLength() uint64 {
	return varIntLength(uint64(r.GetType())) + varIntLength(r.Reference)
}

func (r *ReferenceUnlockBlock) Write(w io.Writer) error {
	if err := writeVarint(w, uint64(r.GetType())); err != nil {
		return err
	}
	if err := writeVarint(w, uint64(r.Reference)); err != nil {
		return err
	}
	return nil
}
