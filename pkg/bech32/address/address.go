// Package address provides utility functionality to encode and decode bech32 based addresses.
package address

import (
	"errors"
	"fmt"
	"github.com/gohornet/hornet/pkg/bech32"
)

// Errors returned during address parsing.
var (
	ErrInvalidPrefix  = errors.New("invalid prefix")
	ErrInvalidVersion = errors.New("invalid version")
	ErrInvalidLength  = errors.New("invalid length")
)

// Prefix denotes the different network prefixes.
type Prefix int

// Network prefix options
const (
	Mainnet Prefix = iota
	Devnet
)

func (p Prefix) String() string {
	return hrpStrings[p]
}

func ParsePrefix(s string) (Prefix, error) {
	for i := range hrpStrings {
		if s == hrpStrings[i] {
			return Prefix(i), nil
		}
	}
	return 0, ErrInvalidPrefix
}

var (
	hrpStrings = [...]string{"iot", "tio"}
)

// Version denotes the version of an address.
type Version byte

// Supported address versions
const (
	WOTS Version = iota
	Ed25519
)

func (v Version) String() string {
	return [...]string{"WOTS", "Ed25519"}[v]
}

// Bech32 encodes the provided addr as a bech32 string.
func Bech32(hrp Prefix, addr Address) (string, error) {
	return bech32.Encode(hrp.String(), addr.Bytes())
}

// ParseBech32 decodes a bech32 encoded string.
func ParseBech32(s string) (Prefix, Address, error) {
	hrp, addrData, err := bech32.Decode(s)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid bech32 encoding: %w", err)
	}
	prefix, err := ParsePrefix(hrp)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid human-readable prefix: %w", err)
	}
	if len(addrData) == 0 {
		return 0, nil, fmt.Errorf("%w: no version", ErrInvalidVersion)
	}
	version := Version(addrData[0])
	addrData = addrData[1:]
	switch version {
	case WOTS:
		addr, err := WOTSAddress(addrData)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid wotsAddress address: %w", err)
		}
		return prefix, addr, nil
	case Ed25519:
		addr, err := Ed25519Address(addrData)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid Ed25519 address: %w", err)
		}
		return prefix, addr, nil
	}
	return 0, nil, fmt.Errorf("%w: %d", ErrInvalidVersion, version)
}

// Address specifies a general address of different underlying types.
type Address interface {
	Version() Version
	Bytes() []byte
	Length() int
	String() string
}
