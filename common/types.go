package common

const (
	// AddressLength is the expected length of Adamnite address
	AddressLength = 20

	// HashLength is the expected length of the hash
	HashLength = 32
)

type Address [AddressLength]byte
type Hash [HashLength]byte
