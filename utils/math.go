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