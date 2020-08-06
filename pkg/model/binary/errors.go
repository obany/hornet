package binary

import (
	"errors"
)

var (
	ErrUnsupportedVersion     = errors.New("message version not supported")
	ErrWrongPayloadType       = errors.New("payload type does not match")
	ErrUnsupportedPayloadType = errors.New("unsupported payload type")
	ErrInvalidPayloadLength   = errors.New("invalid payload length")
	ErrInvalidSubPayload      = errors.New("invalid sub-payload found")
	ErrInvalidVarintRange     = errors.New("invalid varint range")
	ErrEmptyPayload           = errors.New("empty payload")
	ErrInvalidPayloadValue    = errors.New("invalid payload value")
)
