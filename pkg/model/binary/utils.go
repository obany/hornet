package binary

import (
	"bufio"
	"encoding/binary"
	"io"
)

func readAndCheckPayloadType(r io.Reader, expectedType PayloadType) error {

	reader := bufio.NewReader(r)

	payloadType, err := binary.ReadUvarint(reader)
	if err != nil {
		return err
	}

	if PayloadType(payloadType) != expectedType {
		return ErrWrongPayloadType
	}

	return nil
}

func readVarintInRange(r io.Reader, maxValue uint64) (uint64, error) {
	reader := bufio.NewReader(r)

	value, err := binary.ReadUvarint(reader)
	if err != nil {
		return 0, err
	}

	if value > maxValue {
		return 0, ErrInvalidVarintRange
	}
	return value, nil
}

func writeVarint(w io.Writer, value uint64) error {
	var buf = make([]byte, 10)
	len := binary.PutUvarint(buf, value)
	if _, err := w.Write(buf[:len]); err != nil {
		return err
	}
	return nil
}

func readUint64(r io.Reader) (uint64, error) {
	var value uint64
	err := binary.Read(r, binary.LittleEndian, &value)
	return value, err
}

func writeUint64(w io.Writer, value uint64) error {
	return binary.Write(w, binary.LittleEndian, value)
}

func writeBytes(w io.Writer, bytes []byte) error {
	if _, err := w.Write(bytes); err != nil {
		return err
	}
	return nil
}

func readBytes(r io.Reader, size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func writePayloadType(w io.Writer, payload Payload) error {
	return writeVarint(w, uint64(payload.GetType()))
}

func varIntLength(value uint64) uint64 {
	var buf = make([]byte, 10)
	len := binary.PutUvarint(buf, value)
	return uint64(len)
}

func payloadTypeLength(payload Payload) uint64 {
	return varIntLength(uint64(payload.GetType()))
}

func writeByteArray(w io.Writer, data []byte) error {
	if err := writeVarint(w, uint64(len(data))); err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func byteArrayLength(data []byte) uint64 {

	dataLen := uint64(len(data))
	var buf = make([]byte, 10)

	len := binary.PutUvarint(buf, dataLen)

	return dataLen + uint64(len)
}

func readByteArray(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)

	length, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, err
	}
	return data, nil
}
