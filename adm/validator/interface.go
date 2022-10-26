package validator

import (
	"github.com/adamnite/go-adamnite/core"
	"github.com/adamnite/go-adamnite/dpos"
)

type AdamniteImplInterface interface {
	Blockchain() *core.Blockchain
	TxPool() *core.TxPool
	WitnessPool() *dpos.WitnessPool
	WitnessCandidatePool() *dpos.WitnessCandidatePool
}
