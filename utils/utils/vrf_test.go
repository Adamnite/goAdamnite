package utils

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/utils"
)

func TestVrf(t *testing.T) {
	witnesses := make([]Witness, 100)
	for i := range witnesses {
		witnesses[i] = &WitnessImpl{
			Address: utils.Address{byte(i)},
			Voters: []Voter{{
				To:            []byte{byte(i)},
				From:          []byte{byte(i)},
				StakingAmount: big.NewInt(1),
				Signature:     []byte{}}},

			blocksApproved: 0,
			blocksReviewed: 0,
		}
	}
	witnesses = append(witnesses, &WitnessImpl{
		Address: utils.Address{0, 0, 1},
		Voters: []Voter{{
			To:            []byte{0, 0, 1},
			From:          []byte{0, 0, 1},
			StakingAmount: big.NewInt(1),
			Signature:     []byte{}}},

		blocksApproved: 1,
		blocksReviewed: 1,
	})
	fmt.Println(SetVRFItems(witnesses))
}
