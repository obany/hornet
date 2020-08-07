package binary

import (
	"github.com/gohornet/hornet/pkg/bech32/address"
	"io"
)

type UTXOInput struct {
	InputType              uint64
	TransactionId          [32]byte
	TransactionOutputIndex uint64
}

type SigLockedSingleDeposit struct {
	OutputType uint64
	Address    address.Address
	Amount     uint64
}

func readUTXOInput(r io.Reader) (*UTXOInput, error) {

	inputType, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	if inputType != 0 {
		return nil, ErrInvalidPayloadValue
	}

	txId, err := readBytes(r, 32)
	if err != nil {
		return nil, err
	}

	outputIndex, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	utxo := &UTXOInput{
		InputType:              inputType,
		TransactionOutputIndex: outputIndex,
	}
	copy(utxo.TransactionId[:], txId)
	return utxo, nil
}

func (u *UTXOInput) GetLength() uint64 {
	return varIntLength(u.InputType) + 32 + varIntLength(u.TransactionOutputIndex)
}

func (u *UTXOInput) Write(w io.Writer) error {
	if err := writeVarint(w, u.InputType); err != nil {
		return err
	}
	if err := writeBytes(w, u.TransactionId[:]); err != nil {
		return err
	}
	if err := writeVarint(w, u.TransactionOutputIndex); err != nil {
		return err
	}
	return nil
}

func readSigLockedSingleDeposit(r io.Reader) (*SigLockedSingleDeposit, error) {

	outputType, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	if outputType != 0 {
		return nil, ErrInvalidPayloadValue
	}

	addressType, err := readVarintInRange(r, 127)
	if err != nil {
		return nil, err
	}

	var addr address.Address
	switch AddressAndSignatureType(addressType) {
	case AddressAndSignatureTypeWOTS:
		addrData, err := readBytes(r, 49)
		if err != nil {
			return nil, err
		}
		addr, err = address.WOTSAddress(addrData)
		if err != nil {
			return nil, err
		}
	case AddressAndSignatureTypeEd25519:
		addrData, err := readBytes(r, 49)
		if err != nil {
			return nil, err
		}
		addr, err = address.Ed25519Address(addrData)
		if err != nil {
			return nil, err
		}
	}

	amount, err := readUint64(r)
	if err != nil {
		return nil, err
	}

	deposit := &SigLockedSingleDeposit{
		OutputType: outputType,
		Address:    addr,
		Amount:     amount,
	}
	return deposit, nil
}

func (s *SigLockedSingleDeposit) GetLength() uint64 {
	addrLen := uint64(s.Address.Length())
	return varIntLength(s.OutputType) +
		varIntLength(addrLen) +
		addrLen +
		8
}

func (s *SigLockedSingleDeposit) Write(w io.Writer) error {
	if err := writeVarint(w, s.OutputType); err != nil {
		return err
	}

	if err := writeVarint(w, uint64(s.Address.Length())); err != nil {
		return err
	}

	if err := writeBytes(w, s.Address.Bytes()); err != nil {
		return err
	}

	if err := writeUint64(w, s.Amount); err != nil {
		return err
	}

	return nil
}
