package utils

import (
	"encoding/hex"
)


//void type
type Void struct{}

//Various operations to
// BytesToHex converts a byte slice to a hexadecimal string.
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

// HexToBytes converts a hexadecimal string to a byte slice.
func HexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// RemoveZeros removes leading zeros from a byte slice.
func RemoveZeros(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	firstNonZero := 0
	for i := range b {
		if b[i] != 0 {
			firstNonZero = i
			break
		}
	}

	return b[firstNonZero:]
}

// PrettyBytes returns a pretty representation of a byte slice.
// The byte slice is formatted as a hexadecimal string with each byte separated by a space.
func PrettyBytes(b []byte) string {
	hexStr := hex.EncodeToString(b)
	chunks := make([]string, len(b))

	for i := range b {
		chunks[i] = hexStr[i*2 : i*2+2]
	}

	return strings.Join(chunks, " ")
}

// CopyBytes creates a copy of a byte slice.
func CopyBytes(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

// ConcatBytes concatenates multiple byte slices into a single byte slice.
func ConcatBytes(slices ...[]byte) []byte {
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]byte, totalLen)
	pos := 0
	for _, s := range slices {
		copy(result[pos:], s)
		pos += len(s)
	}

	return result
}

// EqualBytes checks if two byte slices are equal.
func EqualBytes(a, b []byte) bool {
	return bytes.Equal(a, b)
}

func CheckHexPrefix(s string) bool {
	return len(s) >= 2 && s[:2] == "0x"
}

// FromHex converts a hexadecimal string to bytes. It handles '0x' prefixed strings.
func FromHex(hexStr string) ([]byte, error) {
	if CheckHexPrefix(hexStr) {
		hexStr = hexStr[2:]
	}
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hexadecimal string: %w", err)
	}
	return bytes, nil
}