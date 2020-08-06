package binary

import (
	"io"
)

type UnsignedData struct {
	Data []byte
}

func (u UnsignedData) GetLength() uint64 {
	return payloadTypeLength(u) + byteArrayLength(u.Data)
}

func (u UnsignedData) GetType() PayloadType {
	return PayloadTypeUnsignedData
}

func (u UnsignedData) Write(w io.Writer) error {
	if err := writePayloadType(w, u); err != nil {
		return err
	}
	return writeByteArray(w, u.Data)
}

func readUnsignedData(r io.Reader) (*UnsignedData, error) {

	if err := readAndCheckPayloadType(r, PayloadTypeUnsignedData); err != nil {
		return nil, err
	}

	data, err := readByteArray(r)
	if err != nil {
		return nil, err
	}

	unsignedData := &UnsignedData{
		Data: data,
	}

	return unsignedData, nil
}
