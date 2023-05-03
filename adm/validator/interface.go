package validator

import (
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
	"github.com/adamnite/go-adamnite/common"
)

type AdamniteImplInterface interface {
	Blockchain() *core.Blockchain
	TxPool() *core.TxPool
	WitnessPool() *dpos.WitnessPool
	Witness() common.Address
	IsConnected() bool 
}
