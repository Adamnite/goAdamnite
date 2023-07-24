 package adm

import (
	"math/big"

	"github.com/adamnite/go-adamnite/utils/bytes"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/dpos"
)

type AdamniteAPI interface {
	DposEngine() dpos.DPOS
}

type txPool interface {
	Get(hash bytes.Hash) *types.Transaction
}

// CallMsg contains parameters for contract calls.
type CallMsg struct {
	From     bytes.Address  
	To       *bytes.Address 
	Ate      uint64        
	AteFee  *big.Int      
	Value    *big.Int      
	Data     []byte       
}
