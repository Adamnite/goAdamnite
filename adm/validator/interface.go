package validator

import "github.com/adamnite/go-adamnite/core"

type AdamniteImplInterface interface {
	Blockchain() *core.Blockchain
	TxPool() *core.TxPool
	WitnessPool() *core.WitnessPool
}
