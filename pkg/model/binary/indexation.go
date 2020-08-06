package binary

import (
	"io"
)

type Indexation struct {
	Tag [16]byte
}

func (i Indexation) GetLength() uint64 {
	return payloadTypeLength(i) + 16
}

func (i Indexation) GetType() PayloadType {
	return PayloadTypeIndexation
}

func (i Indexation) Write(w io.Writer) error {

	writePayloadType(w, i)

	if _, err := w.Write(i.Tag[:]); err != nil {
		return err
	}

	return nil
}

func readIndexation(r io.Reader) (*Indexation, error) {

	if err := readAndCheckPayloadType(r, PayloadTypeIndexation); err != nil {
		return nil, err
	}

	tag := make([]byte, 16)
	if _, err := io.ReadFull(r, tag); err != nil {
		return nil, err
	}

	indexation := &Indexation{}
	copy(indexation.Tag[:], tag)

	return indexation, nil
}
