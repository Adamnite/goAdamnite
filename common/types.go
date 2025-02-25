package common

import "encoding/hex"

const (
	// AddressLength is the expected length of Adamnite address
	AddressLength = 28

	// HashLength is the expected length of the hash
	HashLength = 32
)

type Address [AddressLength]byte
type Hash [HashLength]byte

type StorageSize float64

////TODO: Add encoding and decoding to different data types for both address and hash.
//Add Formatting

func (addr *Address) SetBytes(b []byte) {
	if len(b) > len(addr) {
		b = b[len(b)-AddressLength:]
	}
	copy(addr[AddressLength-len(b):], b)
}

func (addr *Address) Bytes() []byte {
	return addr[:]
}

func HexToAddress(str string) Address {
	return BytesToAddress(FromHex(str))
}

func StringToAddress(str string) Address {
	return BytesToAddress([]byte(str))
}

func BytesToAddress(b []byte) Address {
	var addr Address
	addr.SetBytes(b)
	return addr
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	return hex.EncodeToString(a[:])
}

// String implements fmt.Stringer.
func (a Address) String() string {
	return string(a[:])
}

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func (h Hash) Bytes() []byte { return h[:] }

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func HexToHash(s string) Hash {
	return BytesToHash(FromHex(s))
}

func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}
