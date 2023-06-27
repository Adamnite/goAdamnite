package consensus

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/params"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/adamnite/go-adamnite/utils/safe"
	"github.com/stretchr/testify/assert"
)

var (
	apiEndpoint         = "http://127.0.0.1:5000/"
	addTwoFunctionCode  = "0061736d0100000001170460027e7e017e60017e017e60017e017f60027e7e017f032120000000000000000000000000000000010101010101020303030303030303030307eb0120036164640000037375620001036d756c0002056469765f730003056469765f7500040572656d5f7300050572656d5f75000603616e640007026f72000803786f7200090373686c000a057368725f73000b057368725f75000c04726f746c000d04726f7472000e03636c7a000f0363747a001006706f70636e74001109657874656e64385f7300120a657874656e6431365f7300130a657874656e6433325f7300140365717a00150265710016026e650017046c745f730018046c745f750019046c655f73001a046c655f75001b0467745f73001c0467745f75001d0467655f73001e0467655f75001f0af301200700200020017c0b0700200020017d0b0700200020017e0b0700200020017f0b070020002001800b070020002001810b070020002001820b070020002001830b070020002001840b070020002001850b070020002001860b070020002001870b070020002001880b070020002001890b0700200020018a0b05002000790b050020007a0b050020007b0b05002000c20b05002000c30b05002000c40b05002000500b070020002001510b070020002001520b070020002001530b070020002001540b070020002001570b070020002001580b070020002001550b070020002001560b070020002001590b0700200020015a0b00f401046e616d6502ec012000020001780101790102000178010179020200017801017903020001780101790402000178010179050200017801017906020001780101790702000178010179080200017801017909020001780101790a020001780101790b020001780101790c020001780101790d020001780101790e020001780101790f0100017810010001781101000178120100017813010001781401000178150100017816020001780101791702000178010179180200017801017919020001780101791a020001780101791b020001780101791c020001780101791d020001780101791e020001780101791f02000178010179"
	addTwoFunctionBytes = []byte{}
	addTwoCodeStored    VM.CodeStored
	addTwoFunctionHash  []byte
	testContract        VM.Contract
	testAccount         = common.Address{0, 1, 2}
)

func setup() error {
	addTwoFunctionBytes, _ = hex.DecodeString(addTwoFunctionCode)
	mod := VM.DecodeModule(addTwoFunctionBytes)
	stored, _, err := VM.UploadModuleFunctions(apiEndpoint, mod)
	if err != nil {
		return err
	}

	addTwoCodeStored = stored[0]
	addTwoFunctionHash, _ = addTwoCodeStored.Hash()
	testContract = VM.Contract{CallerAddress: common.Address{1}, Value: big.NewInt(0), Input: nil, Gas: 30000, CodeHashes: []string{hex.EncodeToString(addTwoFunctionHash)}}
	return nil
}
func TestProcessingRun(t *testing.T) {
	if err := setup(); err != nil {
		t.Log("server is most likely not running. Try again with the Offchain DB running")
		t.Skip(err)
	}

	bNode, err := NewBConsensus(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VM.UploadMethod(apiEndpoint, addTwoCodeStored); err != nil {
		t.Fatal(err)
	}
	if err := VM.UploadContract(apiEndpoint, testContract); err != nil {
		t.Fatal(err)
	}
	claim := utils.RuntimeChanges{
		Caller:            testAccount,
		CallTime:          time.Now().UTC(),
		ContractCalled:    testContract.Address,
		ParametersPassed:  append(addTwoFunctionHash, byte(VM.Op_i64)),
		GasLimit:          10000,
		ChangeStartPoints: []uint64{0},
		Changed:           [][]byte{{0, 1, 2, 3, 4, 5, 6, 7}},
	}
	//set the parameters to the hash, types, and values
	claim.ParametersPassed = append(claim.ParametersPassed, VM.EncodeUint64(1)...)
	claim.ParametersPassed = append(claim.ParametersPassed, VM.Op_i64)
	claim.ParametersPassed = append(claim.ParametersPassed, VM.EncodeUint64(2)...)
	didPass, _, err := bNode.VerifyRun(claim)
	if didPass {
		log.Println("should not pass, used incomplete claim")
		t.Fatal(err)
	}
	err = bNode.ProcessRun(&claim)
	if err != nil {
		t.Fatal(err)
	}
	didPass, _, err = bNode.VerifyRun(claim)
	if !didPass {
		log.Println("failed to get same results running same test twice")
		t.Fatal(err)
	}

}
func TestVMTransactions(t *testing.T) {
	rpc.USE_LOCAL_IP = true //use local IPs so we don't wait to get our IP, and don't need to deal with opening the firewall port
	testNodeCount := 5
	db := VM.NewSpoofedDBCache(nil, nil)
	methodsAsString := "0061736d01000000010a0260027f7f017f600000030302000105030100020614037f01419088040b7f004180080b7f004188080b073705066d656d6f727902000367657400001648656c6c6f576f726c645f44656661756c7443544f52000105626c6f636b0301036d736703020a0c020700200120006a0b02000b0b1301004180080b0c0000000000000000040400000045046e616d65011e020003676574011648656c6c6f576f726c645f44656661756c7443544f52071201000f5f5f737461636b5f706f696e746572090a0100072e726f64617461"
	//just has the add function.
	err, hashes := db.DB.AddModuleToSpoofedCode(methodsAsString)
	if err != nil {
		panic(err)
	}
	addTwoFunction := hashes[0]
	addTwoContract := VM.Contract{
		Address:    common.HexToAddress("0x123456"),
		CodeHashes: []string{hex.EncodeToString(addTwoFunction)},
	}
	db.DB.AddContract(addTwoContract.Address.Hex(), &addTwoContract)

	seed := networking.NewNetNode(common.Address{0})
	// seed.AddServer()
	transactionsSeen := safe.NewSafeSlice()
	blocksSeen := safe.NewSafeSlice()
	seed.AddFullServer(
		nil,
		nil,
		func(pt utils.TransactionType) error {
			//the transactions seen logger
			transactionsSeen.Append(pt)
			return nil
		}, func(b utils.BlockType) error {
			blocksSeen.Append(b.(*utils.VMBlock))
			//the blocks seen logger
			return nil
		},
		nil,
		nil,
	)

	seedContact := seed.GetOwnContact()

	conAccounts := []*accounts.Account{}
	conNodes := []*ConsensusNode{}

	maxTimePerRound.SetDuration(time.Millisecond * 1200)
	maxTimePrecision.SetDuration(time.Millisecond * 500)

	for i := 0; i < testNodeCount; i++ {
		if ac, err := accounts.GenerateAccount(); err != nil {
			i -= 1
			continue
		} else {
			conAccounts = append(conAccounts, ac)
			cn, err := NewAConsensus(ac)
			if err != nil {
				t.Fatal(err)
			}
			if err := cn.AddBConsensus(db); err != nil {
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
			stateDB.AddBalance(conAccounts[0].Address, big.NewInt(int64(testNodeCount)*50000))
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
		cn.poolsB.GetApplyingRound().SetStartTime(round0StartTime)
	}

	if len(conNodes[1].poolsB.totalCandidates) < testNodeCount-1 {
		fmt.Println("nodes aren't talking to each other")
		fmt.Println(len(conNodes[0].poolsB.totalCandidates))
		t.Fail()
	}
	//we now have x candidates
	maxTransactionsPerBlock = 10
	maxBlocksPerRound = uint64(testNodeCount)
	seed.FillOpenConnections()
	transactions := []*utils.VMCallTransaction{}
	<-time.After(maxTimePerRound.Duration())
	assert.EqualValues(
		t,
		1,
		conNodes[0].poolsB.currentWorkingRoundID.Get(),
		"round is not correct",
	)
	for i := 0; i < maxTransactionsPerBlock*int(maxBlocksPerRound)*testNodeCount; i++ {
		testTransaction, err := utils.NewBaseTransaction(conAccounts[0], addTwoContract.Address, big.NewInt(1), big.NewInt(100000))
		if err != nil {
			t.Fatal(err)
		}
		vmTransaction, err := utils.NewVMTransactionFrom(conAccounts[0], testTransaction, append(addTwoFunction, 0x7f, 0x1, 0x7f, 0x2))
		if err != nil {
			t.Fatal(err)
		}
		if err := conNodes[0].netLogic.Propagate(vmTransaction); err != nil {
			t.Fatal(err)
		}
		transactions = append(transactions, vmTransaction)
		<-time.After(maxTimePrecision.Duration())
	}
	<-time.After(maxTimePerRound.Duration())

	//everything *should* be reviewed by now.
	assert.Equal(
		t,
		len(transactions),
		transactionsSeen.Len(),
		"wrong number of unique transactions passed the seed node",
	) //if this returns, then its propagation of transactions that isn't working
	assert.Equal(
		t,
		int(transactionsSeen.Len()/maxTransactionsPerBlock),
		blocksSeen.Len(),
		"wrong number of blocks went past this node",
	)
	blockTransactions := []utils.TransactionType{}
	blockHashesSeen := [][]byte{}
	blocksSeen.ForEach(func(_ int, b any) bool {
		assert.Equal(
			t, maxTransactionsPerBlock, len(b.(*utils.VMBlock).Transactions),
			"god why. Wrong number of transactions per block",
		)
		blockTransactions = append(blockTransactions, b.(*utils.VMBlock).Transactions...)
		blockHashesSeen = append(blockHashesSeen, b.(*utils.VMBlock).Hash().Bytes())
		return true
	})
	for _, x := range blockHashesSeen {
		log.Printf("hashSeen: %x", x)
	}
	assert.Equal(t, len(transactions), len(blockTransactions), "wrong transaction count")
	log.Printf("current round is %v", conNodes[0].poolsB.currentWorkingRoundID.Get())
	//should be 2
}
