package binary

type PayloadType uint64

const (
	PayloadTypeSignedTransaction PayloadType = iota
	PayloadTypeMilestone
	PayloadTypeUnsignedData
	PayloadTypeSignedData
	PayloadTypeIndexation
)

type AddressAndSignatureType uint64

const (
	AddressAndSignatureTypeWOTS AddressAndSignatureType = iota
	AddressAndSignatureTypeEd25519
)

type UnlockBlockType uint64

const (
	UnlockBlockTypeSignatureUnlockBlock UnlockBlockType = iota
	UnlockBlockTypeReferenceUnlockBlock
)
