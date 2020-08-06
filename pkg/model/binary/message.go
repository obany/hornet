package binary

import (
	"bufio"
	"encoding/binary"
	"io"
)

type Message struct {
	Version uint64
	Parent1 [32]byte
	Parent2 [32]byte
	Payload Payload
	Nonce   uint64
}

func ReadMessage(r io.Reader) (*Message, error) {

	reader := bufio.NewReader(r)

	version, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}

	if version != 1 {
		return nil, ErrUnsupportedVersion
	}

	var data struct {
		Parent1 [32]byte
		Parent2 [32]byte
	}

	if err := binary.Read(reader, binary.LittleEndian, &data); err != nil {
		return nil, err
	}

	payload, err := readPayload(r)
	if err != nil {
		if err == ErrEmptyPayload {
			// We are expecting to have a payload here
			err = ErrInvalidPayloadLength
		}
		return nil, err
	}

	var nonce uint64
	if err := binary.Read(reader, binary.LittleEndian, &nonce); err != nil {
		return nil, err
	}

	message := &Message{
		Version: 1,
		Parent1: data.Parent1,
		Parent2: data.Parent2,
		Payload: payload,
		Nonce:   nonce,
	}

	return message, nil
}

func readPayload(r io.Reader) (Payload, error) {

	reader := bufio.NewReader(r)

	payloadLength, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}

	if payloadLength == 0 {
		return nil, ErrEmptyPayload
	}

	// Peek the payload type
	payloadTypeBytes, err := reader.Peek(10)

	payloadType, n := binary.Uvarint(payloadTypeBytes)
	if n < 0 {
		return nil, ErrWrongPayloadType
	}

	// Create a reader that reads at most the payload length,
	// so we can pass it over without the risk of it consuming the nonce
	payloadReader := io.LimitReader(reader, int64(payloadLength))

	switch PayloadType(payloadType) {

	case PayloadTypeSignedTransaction:
		return nil, nil

	case PayloadTypeMilestone:
		return ReadMilestone(payloadReader)

	case PayloadTypeUnsignedData:
		return ReadUnsignedData(payloadReader)

	case PayloadTypeSignedData:
		return ReadSignedData(payloadReader)

	case PayloadTypeIndexation:
		return ReadIndexation(payloadReader)

	default:
		// ignore the payload data but do not return error, we need to keep the message around
		reader.Discard(int(payloadLength))
		unsupported := &UnsupportedPayload{
			payloadType: PayloadType(payloadType),
		}
		return unsupported, nil
	}
}
