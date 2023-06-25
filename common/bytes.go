package common

import "encoding/hex"

// having a type to save memory space on hash tables
type Void struct{}

// FromHex converts hexadecimal string that may be prefixed with '0x' to bytes.
func FromHex(s string) []byte {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// ToHex converts bytes into a hexadecimal string prefixed with '0x'.
func ToHex(b []byte) string {
	h := make([]byte, len(b)*2+2)
	copy(h, "0x")
	hex.Encode(h[2:], b)
	return string(h)
}

// hasHexPrefix checks whether given string begins with '0x' or '0X'
func hasHexPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// CopyBytes returns an exact copy of the provided bytes.
func CopyBytes(b []byte) (copiedBytes []byte) {
	if b == nil {
		return nil
	}
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

// TrimLeftZeroes returns a subslice of s without leading zeroes
func TrimLeftZeroes(s []byte) []byte {
	idx := 0
	for ; idx < len(s); idx++ {
		if s[idx] != 0 {
			break
		}
	}
	return s[idx:]
}
