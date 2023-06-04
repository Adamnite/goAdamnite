package consensus

import (
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
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

	validWitnessAddress := common.StringToAddress("23m3Ho7PwouaFzU8iXMLygwuXNW7")
	invalidWitnessAddress := common.StringToAddress("44m3Ho7PwouaFzU8iXMLygwuXN88")

	nextValidBlock := NewBlock(
		// parent block is genesis block
		common.Hash{},
		validWitnessAddress,
		// dummy values irrelevant for the test
		common.Hash{},
		common.Hash{},
		common.Hash{},
		big.NewInt(1),
		[]*Transaction{},
	)

	if ok, _ := n.ValidateBlock(nextValidBlock); !ok {
		t.Fatal("Block should be valid")
	}

	if _, ok := n.untrustworthyWitnesses[validWitnessAddress]; ok {
		t.Fatal("Trustworthy witness should not be reported")
	}

	nextInvalidBlock := NewBlock(
		// parent block is genesis block but we specify non-genesis hash as parent ID
		common.HexToHash("0x095af5a356d055ed095af5a356d055ed095af5a356d055ed095af5a356d055ed"),
		invalidWitnessAddress,
		// dummy values irrelevant for the test
		common.Hash{},
		common.Hash{},
		common.Hash{},
		big.NewInt(1),
		[]*Transaction{},
	)

	if ok, _ := n.ValidateBlock(nextInvalidBlock); ok {
		t.Fatal("Block should be invalid")
	}

	// TODO: Uncomment once chain reference is passed to each consensus node
	// if _, ok := n.untrustworthyWitnesses[invalidWitnessAddress]; !ok {
	// 	t.Fatal("Untrustworthy witness should be reported")
	// }
}