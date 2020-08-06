package address

import "encoding/hex"

type ed25519Address [32]byte

func (ed25519Address) Version() Version {
	return Ed25519
}

func (a ed25519Address) Bytes() []byte {
	return append([]byte{byte(Ed25519)}, a[:]...)
}

func (a ed25519Address) String() string {
	return hex.EncodeToString(a[:])
}

func (a ed25519Address) Length() int {
	return len(a)
}

// Ed25519Address creates an address from a 32-byte hash.
func Ed25519Address(hash []byte) (Address, error) {
	var addr ed25519Address
	if len(hash) != len(addr) {
		return nil, ErrInvalidLength
	}
	copy(addr[:], hash)
	return addr, nil
}
