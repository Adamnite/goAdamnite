package types

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/utils"
)

var (
	voteTx = NewVoteTransaction(0, utils.HexToAddress(""), big.NewInt(5000), 21000)
)

func TestVoteTxSigHash(t *testing.T) {

}
