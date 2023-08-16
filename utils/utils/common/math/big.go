package math

import (
	"math"
	"math/big"
)

const (
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

// ReadBits encodes the absolute value of bigint as big-endian bytes. Callers must ensure
// that buf has enough space. If buf is too short the result will be incomplete.
func ReadBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}

// PadBytesLeft encodes a big integer as a big-endian byte slice. The length
// of the slice is at least n bytes.
func PadBytesLeft(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	ReadBits(bigint, ret)
	return ret
}

func GetPercent(amount *big.Int, max *big.Int) float32 {
	aLen := GetDecimal(amount)
	bLen := GetDecimal(max)

	aExp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(bLen-aLen+3)), nil)
	aMul := new(big.Int).Mul(amount, aExp)

	div := new(big.Int).Div(aMul, max).Int64()

	dExp := math.Pow10(bLen - aLen + 3)
	fDiv := float32(div) / float32(dExp)
	return fDiv
}

func GetDecimal(a *big.Int) int {
	decimal := -1
	bInt := new(big.Int).Set(a)
	for bInt.Int64() != 0 {
		bInt = new(big.Int).Div(bInt, big.NewInt(10))
		decimal++
	}
	return decimal
}
