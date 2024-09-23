package types

import "github.com/adamnite/go-adamnite/internal/common"

type WitnessTx struct {
	WitnessAddr common.Address
}

func CreateNewWitnessTx(addr common.Address) *Transaction {
	return NewTx(&WitnessTx{
		WitnessAddr: addr,
	})
}

func (tx *WitnessTx) txType() byte {
	return WitnessTxType
}

func (tx *WitnessTx) copy() TxData {
	cpy := &WitnessTx{
		WitnessAddr: tx.WitnessAddr,
	}

	return cpy
}

func (tx *WitnessTx) getValidator() common.Address {
	return tx.WitnessAddr
}
