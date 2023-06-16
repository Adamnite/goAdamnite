package gossip

import (
	"fmt"
	"math/big"
)

// Ring represents the ring structure used in the Koorde protocol.
type Ring struct {
	Size int        // Size of the ring
	Mod  *big.Int   // Modulus for ring arithmetic
	Log2 *big.Int   // Log base 2 of the ring size
	Mask *big.Int   // Bit mask for ring arithmetic
	One  *big.Int   // BigInt representing 1
	Two  *big.Int   // BigInt representing 2
}

// NewRing creates a new Ring structure with the given size.
func NewRing(size int) *Ring {
	mod := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(size)), nil)
	log2 := big.NewInt(int64(size))
	mask := big.NewInt(0).Sub(mod, big.NewInt(1))
	one := big.NewInt(1)
	two := big.NewInt(2)

	return &Ring{
		Size: size,
		Mod:  mod,
		Log2: log2,
		Mask: mask,
		One:  one,
		Two:  two,
	}
}

// Add performs ring addition between two big integers.
func (r *Ring) Add(a, b *big.Int) *big.Int {
	result := big.NewInt(0).Add(a, b)
	result.Mod(result, r.Mod)
	return result
}

// Sub performs ring subtraction between two big integers.
func (r *Ring) Sub(a, b *big.Int) *big.Int {
	result := big.NewInt(0).Sub(a, b)
	result.Mod(result, r.Mod)
	return result
}

// LeftShift performs a left shift operation on a big integer by the given shift amount.
func (r *Ring) LeftShift(a *big.Int, shift uint) *big.Int {
	result := big.NewInt(0).Lsh(a, shift)
	result.Mod(result, r.Mod)
	return result
}

// RightShift performs a right shift operation on a big integer by the given shift amount.
func (r *Ring) RightShift(a *big.Int, shift uint) *big.Int {
	result := big.NewInt(0).Rsh(a, shift)
	result.Mod(result, r.Mod)
	return result
}

// Log2 computes the log base 2 of a big integer.
func (r *Ring) Log2(a *big.Int) *big.Int {
	result := big.NewInt(0).Set(a)
	count := big.NewInt(0)

	for result.Cmp(r.One) > 0 {
		result.Rsh(result, 1)
		count.Add(count, r.One)
	}

	return count
}

// Contains checks if the ring contains a given key.
func (r *Ring) Contains(key *big.Int) bool {
	return key.Cmp(r.Mod) < 0
}
