package utils

import(
	"math/big"
)

// BigInt represents a big integer.
type BigInt struct {
	*big.Int
}

// NewBigInt creates a new BigInt with the provided value.
func NewBigInt(value int64) *BigInt {
	return &BigInt{
		Int: big.NewInt(value),
	}
}

// NewBigIntFromString creates a new BigInt from a string representation of a number.
func NewBigIntFromString(value string) (*BigInt, bool) {
	bi := new(big.Int)
	success, _ := bi.SetString(value, 10)
	if !success {
		return nil, false
	}
	return &BigInt{
		Int: bi,
	}, true
}

// BigFloat represents a big floating-point number.
type BigFloat struct {
	*big.Float
}

// NewBigFloat creates a new BigFloat with the provided value.
func NewBigFloat(value float64) *BigFloat {
	return &BigFloat{
		Float: big.NewFloat(value),
	}
}

// NewBigFloatFromString creates a new BigFloat from a string representation of a number.
func NewBigFloatFromString(value string) (*BigFloat, bool) {
	bf := new(big.Float)
	success, _ := bf.SetString(value)
	if !success {
		return nil, false
	}
	return &BigFloat{
		Float: bf,
	}, true
}

// BigRat represents a big rational number.
type BigRat struct {
	*big.Rat
}

// NewBigRat creates a new BigRat with the provided numerator and denominator.
func NewBigRat(numerator, denominator int64) *BigRat {
	return &BigRat{
		Rat: big.NewRat(numerator, denominator),
	}
}

// NewBigRatFromString creates a new BigRat from a string representation of a number.
func NewBigRatFromString(value string) (*BigRat, bool) {
	br := new(big.Rat)
	success, _ := br.SetString(value)
	if !success {
		return nil, false
	}
	return &BigRat{
		Rat: br,
	}, true
}

// PadBytesLeft pads the given byte slice on the left with zeros to the specified length.
func PadBytesLeft(b []byte, length int) []byte {
	if len(b) >= length {
		return b
	}

	padded := make([]byte, length)
	copy(padded[length-len(b):], b)
	return padded
}

// GetDecimalPercentage returns the percentage representation of a decimal value as a big integer.
// For example, for value 0.75 and scale 2, it returns 75.
func GetDecimalPercentage(value float64, scale int) *big.Int {
	percentage := new(big.Int)
	percentage.SetFloat64(value * 100)
	percentage.Mul(percentage, big.NewInt(int64(10*scale)))
	return percentage
}

// SliceBytes slices a big-endian byte representation into chunks of specified size.
func SliceBytes(b []byte, chunkSize int) [][]byte {
	if chunkSize <= 0 {
		return nil
	}

	numChunks := (len(b) + chunkSize - 1) / chunkSize
	chunks := make([][]byte, numChunks)

	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(b) {
			end = len(b)
		}
		chunks[i] = b[start:end]
	}

	return chunks
}

// BytesToBigEndianUint64 converts a byte slice to a big-endian uint64.
func BytesToBigEndianUint64(b []byte) uint64 {
	if len(b) > 8 {
		return 0
	}
	zeroPadded := PadBytesLeft(b, 8)
	return binary.BigEndian.Uint64(zeroPadded)
}

// BigEndianUint64ToBytes converts a big-endian uint64 to a byte slice.
func BigEndianUint64ToBytes(value uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, value)
	return b
}