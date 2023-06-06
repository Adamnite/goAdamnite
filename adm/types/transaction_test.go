package types

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
)

var (
	voteTx = NewVoteTransaction(0, common.HexToAddress(""), big.NewInt(5000), 21000)
)

func TestVoteTxSigHash(t *testing.T) {

}
