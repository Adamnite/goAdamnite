 package adm

import (
	"math/big"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
)

type AdamniteAPI interface {
	DposEngine() dpos.DPOS
}

type txPool interface {
	Get(hash common.Hash) *types.Transaction
}