package binary

import (
	"io"
)

type Payload interface {
	GetType() PayloadType

	GetLength() uint64
	Write(io.Writer) error
}

type UnsupportedPayload struct {
	payloadType PayloadType
}

func (u UnsupportedPayload) GetType() PayloadType {
	return u.payloadType
}

func (u UnsupportedPayload) GetLength() uint64 {
	return 0
}

func (u UnsupportedPayload) Write(io.Writer) error {
	return ErrUnsupportedPayloadType
}
