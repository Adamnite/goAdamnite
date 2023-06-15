package consensus

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/params"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
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

	if ok, _ := n.ValidateChamberABlock(nextValidBlock); !ok {
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

	if ok, _ := n.ValidateChamberABlock(nextInvalidBlock); ok {
		t.Fatal("Block should be invalid")
	}

	// TODO: Uncomment once chain reference is passed to each consensus node
	// if _, ok := n.untrustworthyWitnesses[invalidWitnessAddress]; !ok {
	// 	t.Fatal("Untrustworthy witness should be reported")
	// }
}

func TestTransactions(t *testing.T) {
	testNodeCount := 5
	seed := networking.NewNetNode(common.Address{0})
	// seed.AddServer()
	transactionsSeen := []*utils.Transaction{}
	blocksSeen := []utils.Block{}
	seed.AddFullServer(
		nil,
		nil,
		func(pt *utils.Transaction) error {
			//the transactions seen logger
			transactionsSeen = append(transactionsSeen, pt)
			return nil
		}, func(b utils.Block) error {
			blocksSeen = append(blocksSeen, b)
			//the blocks seen logger
			return nil
		},
		nil,
		nil,
	)

	seedContact := seed.GetOwnContact()

	conAccounts := []*accounts.Account{}
	conNodes := []*ConsensusNode{}
	maxTimePerRound = time.Second * 5
	for i := 0; i < testNodeCount; i++ {
		if ac, err := accounts.GenerateAccount(); err != nil {
			i -= 1
			continue
		} else {
			conAccounts = append(conAccounts, ac)
			cn, err := NewAConsensus(*ac)
			if err != nil {
				t.Fatal(err)
			}
			conNodes = append(conNodes, cn)
			if err := cn.netLogic.ConnectToContact(&seedContact); err != nil {
				t.Fatal(err)
			}
			//now we need to add the statedb
			db := rawdb.NewMemoryDB()
			stateDB, _ := statedb.New(common.Hash{}, statedb.NewDatabase(db))

			rootHash := stateDB.IntermediateRoot(false)
			stateDB.Database().TrieDB().Commit(rootHash, false, nil)
			blockchain, _ := blockchain.NewBlockchain(
				db,
				params.TestnetChainConfig,
			)
			stateDB.AddBalance(conAccounts[0].Address, big.NewInt(int64(testNodeCount)*5))
			//add the balance so that this 0th account can send all the test transaction
			_, err = stateDB.Commit(false)
			if err != nil {
				t.Fatal(err)
			}
			cn.state = stateDB
			cn.chain = blockchain

			cn.netLogic.ResetConnections()
			cn.autoStakeAmount = big.NewInt(1)

		}
	}
	round0StartTime := time.Now().UTC().Add(time.Second)

	for _, cn := range conNodes {
		cn.netLogic.SprawlConnections(3, 0)
		cn.netLogic.ResetConnections()
		if err := cn.ProposeCandidacy(0); err != nil {
			t.Fatal(err)
		}
		cn.poolsA.GetApplyingRound().roundStartTime = round0StartTime
	}

	if len(conNodes[1].poolsA.totalCandidates) < testNodeCount-1 {
		fmt.Println("nodes arent talking to each other")
		fmt.Println(len(conNodes[0].poolsA.totalCandidates))
		t.Fail()
	}
	//we now have x candidates
	maxTransactionsPerBlock = 5
	maxBlocksPerRound = uint64(testNodeCount)
	seed.FillOpenConnections()
	transactions := []*utils.Transaction{}
	<-time.After(maxTimePerRound + time.Second)
	assert.Equal(
		t,
		uint64(1),
		conNodes[0].poolsA.currentWorkingRoundID,
		"round is not correct",
	)
	for i := 0; i < 25; i++ {
		testTransaction, err := utils.NewTransaction(conAccounts[0], conAccounts[i%testNodeCount].Address, big.NewInt(1), big.NewInt(int64(i)))
		if err != nil {
			t.Fatal(err)
		}
		if err := conNodes[0].netLogic.Propagate(testTransaction); err != nil {
			t.Fatal(err)
		}
		transactions = append(transactions, testTransaction)
	}
	<-time.After(maxTimePerRound * 5)
	// maxTimePerRound = time.Second * 100

	//everything *should* be reviewed by now.
	assert.Equal(
		t,
		len(transactions),
		len(transactionsSeen),
		"wrong number of unique transactions passed the seed node",
	)
	assert.Equal(
		t,
		int(len(transactionsSeen)/maxTransactionsPerBlock),
		len(blocksSeen),
		"wrong number of blocks went past this node",
	)

}
