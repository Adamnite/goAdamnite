package math

import (
	"math/big"
	"testing"
)

func TestBigInt(t *testing.T) {
	a := big.NewInt(14454560000)
	b := big.NewInt(10000000000000)

	fDiv := GetPercent(a, b)
	if fDiv != 0.001445 {
		t.Error("Failed")
	}
}
