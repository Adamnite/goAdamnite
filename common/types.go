package common

import (
	"encoding/hex"
)

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

func (addr *Address) SetBytes(b []byte) {
	if len(b) > len(addr) {
		b = b[len(b)-AddressLength:]
	}
	copy(addr[AddressLength-len(b):], b)
}

func BytesToAddress(b []byte) Address {
	var addr Address
	addr.SetBytes(b)
	return addr
}

func (a Address) hex() []byte {
	var buf [len(a)*2 + 2]byte
	copy(buf[:2], "0x")
	hex.Encode(buf[2:], a[:])
	return buf[:]
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	return string(a.hex())
}

// String implements fmt.Stringer.
func (a Address) String() string {
	return a.Hex()
}
