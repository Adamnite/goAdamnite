package validator

import (
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/dpos"
)

type AdamniteImplInterface interface {
	Blockchain() *blockchain.Blockchain
	TxPool() *blockchain.TxPool
	WitnessPool() *dpos.WitnessPool
}
