package binary

import (
	"encoding/binary"
	"io"
	"math"
	"time"

	"github.com/gohornet/hornet/pkg/model/milestone"
)

type Milestone struct {
	Index       milestone.Index
	Timestamp   uint64
	MerkleProof [64]byte
	Signature   [64]byte
}

func NewMilestone(index milestone.Index, merkleProof [64]byte) *Milestone {
	return &Milestone{
		Index:       index,
		Timestamp:   uint64(time.Now().Unix()),
		MerkleProof: merkleProof,
		Signature:   [64]byte{},
	}
}

func ReadMilestone(r io.Reader) (*Milestone, error) {

	if err := readAndCheckPayloadType(r, PayloadTypeMilestone); err != nil {
		return nil
	}

	index, err := readVarintInRange(r, math.MaxUint32)
	if err != nil {
		return nil, err
	}

	var data struct {
		Timestamp   uint64
		MerkleProof [64]byte
		Signature   [64]byte
	}

	err = binary.Read(r, binary.LittleEndian, &data)
	if err != nil {
		return nil, err
	}

	milestone := &Milestone{
		Index:       milestone.Index(index),
		Timestamp:   data.Timestamp,
		MerkleProof: data.MerkleProof,
		Signature:   data.Signature,
	}

	return milestone, nil
}

func (m *Milestone) GetLength() uint64 {
	return payloadTypeLength(m) + varIntLength(uint64(m.Index)) + 8 + 64 + 64
}

func (m *Milestone) GetType() PayloadType {
	return PayloadTypeMilestone
}

func (m *Milestone) UpdateSignature(signature [64]byte) {
	m.Signature = signature
}

func (m Milestone) Write(w io.Writer) error {

	var encodedType = make([]byte, 1)
	binary.PutUvarint(encodedType, uint64(m.GetType()))
	if _, err := w.Write(encodedType); err != nil {
		return nil
	}

	var data = []interface{}{
		uint32(m.Index),
		m.Timestamp,
		m.MerkleProof,
		m.Signature,
	}

	for _, v := range data {
		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			return err
		}
	}
	return nil
}
