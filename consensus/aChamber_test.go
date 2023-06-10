package consensus

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

func TestVerifyBlock(t *testing.T) {
	account, err := accounts.GenerateAccount()
	if err != nil {
		t.Fatal(err)
	}

	n, err := NewAConsensus(*account)
	if err != nil {
		t.Fatal(err)
	}

	if !n.isANode() {
		t.Fatal("Failed to create A consensus node")
	}

	validWitness := accounts.AccountFromPubBytes([]byte{1, 2, 3})
	invalidWitness := accounts.AccountFromPubBytes([]byte{4, 5, 6})

	nextValidBlock := utils.NewBlock(
		// parent block is genesis block
		common.Hash{},
		validWitness.PublicKey,
		// dummy values irrelevant for the test
		common.Hash{},
		common.Hash{},
		common.Hash{},
		big.NewInt(1),
		[]*utils.Transaction{},
	)

	if ok, _ := n.ValidateBlock(nextValidBlock); !ok {
		t.Fatal("Block should be valid")
	}

	if _, ok := n.untrustworthyWitnesses[string(validWitness.PublicKey)]; ok {
		t.Fatal("Trustworthy witness should not be reported")
	}

	nextInvalidBlock := utils.NewBlock(
		// parent block is genesis block but we specify non-genesis hash as parent ID
		common.HexToHash("0x095af5a356d055ed095af5a356d055ed095af5a356d055ed095af5a356d055ed"),
		invalidWitness.PublicKey,
		// dummy values irrelevant for the test
		common.Hash{},
		common.Hash{},
		common.Hash{},
		big.NewInt(1),
		[]*utils.Transaction{},
	)

	if ok, _ := n.ValidateBlock(nextInvalidBlock); ok {
		t.Fatal("Block should be invalid")
	}

	// TODO: Uncomment once chain reference is passed to each consensus node
	// if _, ok := n.untrustworthyWitnesses[invalidWitnessAddress]; !ok {
	// 	t.Fatal("Untrustworthy witness should be reported")
	// }
}
