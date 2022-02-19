package common

const (
	// AddressLength is the expected length of Adamnite address
	AddressLength = 20

	// HashLength is the expected length of the hash
	HashLength = 64
)

type Address [AddressLength]byte
type Hash [HashLength]byte


////TODO: Add encoding and decoding to different data types for both address and hash.
//Add Formatting
