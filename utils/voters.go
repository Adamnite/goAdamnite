package utils

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
)

type Voter struct {
	Address       common.Address
	StakingAmount *big.Int
}
