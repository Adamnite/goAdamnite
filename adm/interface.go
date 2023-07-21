 package adm

import (
	"math/big"

	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
)

type AdamniteAPI interface {
	DposEngine() dpos.DPOS
}

type txPool interface {
	Get(hash utils.Hash) *types.Transaction
}

// CallMsg contains parameters for contract calls.
type CallMsg struct {
	From     utils.Address  
	To       *utils.Address 
	Ate      uint64        
	AteFee  *big.Int      
	Value    *big.Int      
	Data     []byte       
}
