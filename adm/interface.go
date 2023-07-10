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

// CallMsg contains parameters for contract calls.
type CallMsg struct {
	From     common.Address  
	To       *common.Address 
	Ate      uint64        
	AteFee  *big.Int      
	Value    *big.Int      
	Data     []byte       
}
