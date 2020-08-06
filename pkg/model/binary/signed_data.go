package binary

import (
	"encoding/binary"
	"io"
)

type SignedData struct {
	Data      []byte
	PublicKey [32]byte
	Signature [64]byte
}

func (s SignedData) GetLength() uint64 {
	return payloadTypeLength(s) + byteArrayLength(s.Data) + 32 + 64
}

func (s SignedData) GetType() PayloadType {
	return PayloadTypeSignedData
}

func (s SignedData) Write(w io.Writer) error {

	if err := writePayloadType(w, s); err != nil {
		return err
	}

	if err := writeByteArray(w, s.Data); err != nil {
		return err
	}

	if _, err := w.Write(s.PublicKey[:]); err != nil {
		return err
	}

	if _, err := w.Write(s.Signature[:]); err != nil {
		return err
	}

	return nil
}

func ReadSignedData(r io.Reader) (*SignedData, error) {

	if err := readAndCheckPayloadType(r, PayloadTypeSignedData); err != nil {
		return nil, err
	}

	data, err := readByteArray(r)
	if err != nil {
		return nil, err
	}

	var sign struct {
		PublicKey [32]byte
		Signature [64]byte
	}

	err = binary.Read(r, binary.LittleEndian, &sign)
	if err != nil {
		return nil, err
	}

	signedData := &SignedData{
		Data:      data,
		PublicKey: sign.PublicKey,
		Signature: sign.Signature,
	}

	return signedData, nil
}
