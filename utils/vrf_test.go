package utils

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
)

func TestVrf(t *testing.T) {
	witnesses := make([]Witness, 100)
	for i := range witnesses {
		witnesses[i] = &WitnessImpl{
			Address: common.Address{byte(i)},
			Voters:  []Voter{{common.Address{byte(i)}, big.NewInt(1)}},

			blocksApproved: 0,
			blocksReviewed: 0,
		}
	}
	witnesses = append(witnesses, &WitnessImpl{
		Address: common.Address{0, 0, 1},
		Voters:  []Voter{{common.Address{0, 0, 1}, big.NewInt(3)}},

		blocksApproved: 1,
		blocksReviewed: 1,
	})
	fmt.Println(SetVRFItems(witnesses))
}
