package consensus

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func TestWitnessSelection(t *testing.T) {
	witnessCount := 10

	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	pool.nextRound()
	if err != nil {
		t.Fatal(err)
	}
	witnesses := generateTestWitnesses(witnessCount)
	for _, w := range witnesses {
		w.updateCandidate(pool)
		if err := pool.AddCandidate(w.candidacy); err != nil {
			t.Fatal(err)
		}
	}
	pool.witnessGoal = witnessCount - 1
	selected, _ := pool.SelectCurrentWitnesses()
	assert.Equal(
		t,
		witnessCount-1,
		selected.Len(),
		"wrong number of witnesses selected",
	)
	for _, w := range witnesses {
		w.updateCandidate(pool)
		pool.AddCandidate(w.candidacy)
	}
	pool.witnessGoal = 2
	selected, _ = pool.SelectCurrentWitnesses()
	if selected.Len() != 2 {
		fmt.Println("wrong number selected")
		t.Fail()
	}
}
func TestRoundSelections(t *testing.T) {
	maxTimePerRound.SetDuration(time.Second * 1) //change the time between rounds for testing.
	maxTimePrecision.SetDuration(time.Millisecond * 10)
	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	pool.StopAsyncTracker()
	if pool.currentWorkingRoundID.Get() != 0 {
		fmt.Println("round incremented right away")
		t.Fail()
	}

	candidates := generateTestWitnesses(15)

	//set the newRound caller to add the candidates again (as if people had reapplied in between rounds)
	pool.newRoundStartedCaller = []func(){func() {
		for _, can := range candidates {
			can.updateCandidate(pool)
			if err := pool.AddCandidate(can.candidacy); err != nil {
				t.Fatal(err)
			}
		}
	}}
	if pool.currentWorkingRoundID.Get() != 0 {
		fmt.Println("round is incremented before call")
		t.Fail()
	}
	pool.newRoundStartedCaller[0]()
	pool.SelectCurrentWitnesses()
	pool.nextRound()
	if pool.currentWorkingRoundID.Get() != 1 {
		fmt.Println("round did not increment correctly")
		t.Fail()
	}
	nextRoundStartTime := pool.GetWorkingRound().roundStartTime.Add(maxTimePerRound.Duration()).Add(maxTimePrecision.Duration()) //give it a hair over the time
	if err := pool.StartAsyncTracking(); err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(
		t,
		1,
		pool.currentWorkingRoundID.Get(),
		"new round started too soon",
	)
	<-time.After(time.Until(nextRoundStartTime))
	pool.StopAsyncTracker()
	assert.EqualValues(
		t,
		2,
		pool.currentWorkingRoundID.Get(),
		"new round must have started too late(or like, *way* too early)",
	)

}

// test the longevity of the witness selection
func TestLongTermPoolCalculations(t *testing.T) {
	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	pool.StopAsyncTracker()
	maxTimePerRound.SetDuration(time.Millisecond * 50) //change the time between rounds for testing.
	// maxTimePerRound = time.Second * 5 //change the time between rounds for debugging.
	maxTimePrecision.SetDuration(time.Millisecond * 5) //the error tolerance we can handle

	candidates := generateTestWitnesses(15)

	//set the newRound caller to add the candidates again (as if people had reapplied in between rounds)
	pool.newRoundStartedCaller = []func(){func() {
		for _, can := range candidates {
			can.updateCandidate(pool)
			if err := pool.AddCandidate(can.candidacy); err != nil {
				t.Fatal(err)
			}
		}
	}}
	pool.nextRound()
	if err := pool.StartAsyncTracking(); err != nil {
		t.Fatal(err)
	}
	goal := 500
	<-time.After(maxTimePerRound.Duration()*time.Duration(goal) + maxTimePrecision.Duration())
	pool.StopAsyncTracker()
	if pool.GetWorkingRound().roundStartTime.After(time.Now()) {
		fmt.Println("round after now")
		t.Fail()
	}
	assert.EqualValues(
		t,
		goal+1, //plus one since we need to start it!
		pool.currentWorkingRoundID.Get(),
		"timing for new round generation must be wrong",
	)
}

func TestWitnessLeadSelection(t *testing.T) {
	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	pool.StopAsyncTracker()
	maxTimePerRound.SetDuration(time.Millisecond * 500) //change the time between rounds for testing.
	// maxTimePerRound = time.Second * 5 //change the time between rounds for debugging.
	maxTimePrecision.SetDuration(time.Millisecond * 5) //the error tolerance we can handle

	candidates := generateTestWitnesses(27)
	pool.witnessGoal = len(candidates)
	blocksPerLead := 5
	maxBlocksPerRound = uint64(blocksPerLead) * uint64(len(candidates))
	//set the newRound caller to add the candidates again (as if people had reapplied in between rounds)
	pool.newRoundStartedCaller = []func(){func() {
		for _, can := range candidates {
			can.updateCandidate(pool)
			if err := pool.AddCandidate(can.candidacy); err != nil {
				t.Fatal(err)
			}
		}
	}}
	pool.nextRound()
	pool.nextRound()
	wrd := pool.GetWorkingRound()
	assert.Equal(
		t,
		pool.witnessGoal,
		wrd.leadWitnessOrder.Len(),
		"incorrect witness leader list",
	)
	activeWit := wrd.leadWitnessOrder.Get(0).(*witness)
	assert.True(
		t,
		wrd.IsActiveWitnessLead(&activeWit.spendingPub),
		"leading witness is not actually lead",
	)
	//test two witnesses who acts truthfully
	var blocksFaked uint64 = 0
	wrd.leadWitnessOrder.ForEach(func(_ int, val interface{}) bool {
		lead := val.(*witness)
		blocksWhileLead := 0
		for wrd.IsActiveWitnessLead(&lead.spendingPub) {
			blocksFaked += 1
			blocksWhileLead += 1
			if err := pool.ActiveWitnessReviewed(&lead.spendingPub, true, blocksFaked); err != nil {
				t.Fatal(err)
			}
		}
		assert.Equal(t,
			blocksPerLead,
			blocksWhileLead,
			"lead had the wrong number of blocks to review",
		)
		return true
	})
	assert.Equal(t,
		maxBlocksPerRound,
		blocksFaked,
		"blocks faked appears to not be accurate to the max block limit",
	)
	assert.EqualValues(
		t,
		3,
		pool.currentWorkingRoundID.Get(),
		"adding blocks till round was filled did not automatically increment round",
	)
	wrd = pool.GetWorkingRound()
	for i := 0; pool.currentWorkingRoundID.Get() == 3; i = (i + 1) % len(candidates) {
		blocksFaked += 1
		testCan := candidates[i].candidacy
		if wrd.IsActiveWitnessLead(testCan.GetWitnessPub()) {
			err := pool.ActiveWitnessReviewed(testCan.GetWitnessPub(), true, blocksFaked)
			//then this should actually have worked
			assert.Equal(
				t,
				nil,
				err,
				"error was not nil when witness had active lead",
			)
		} else {
			err := pool.ActiveWitnessReviewed(testCan.GetWitnessPub(), true, blocksFaked)
			//they couldn't have reviewed that block!
			assert.Error(
				t,
				err,
				"this witness should've thrown an error!",
			)
		}
		//go through the witnesses until the round rollover

	}
}

// test the longevity of the witness selection
func TestLongTermLeadSelection(t *testing.T) {
	rpc.USE_LOCAL_IP = true           //use local IPs so we don't wait to get our IP, and don't need to deal with opening the firewall port
	maxTimePerRound.SetDuration(time.Second * 1) //change the time between rounds for testing.
	// maxTimePerRound = time.Second * 50       //change the time between rounds for debugging.
	maxTimePrecision.SetDuration(time.Millisecond * 50) //the error tolerance we can handle
	testCandidateCount := 5
	round0Start := time.Now().UTC().Add(maxTimePrecision.Duration())
	candidates := []*ConsensusNode{}
	for i := 0; i < testCandidateCount; i++ {
		ac, _ := accounts.GenerateAccount()
		newCan, err := NewAConsensus(*ac)
		if err != nil {
			t.Fatal(err)
		}
		if i != 0 {
			targetContact := candidates[i-1].netLogic.GetOwnContact()
			if err := newCan.netLogic.ConnectToContact(&targetContact); err != nil {
				t.Fatal(err)
			}
			if i == 1 {
				//also get 0 to connect to us
				targetContact = newCan.netLogic.GetOwnContact()
				candidates[0].netLogic.ConnectToContact(&targetContact)
			}
		}
		newCan.poolsA.GetWorkingRound().roundStartTime = round0Start
		newCan.autoStakeAmount = big.NewInt(1) //TODO: because of this, in future this test will need the chain data and states
		candidates = append(candidates, newCan)
	}
	//all nodes are now connected to the network, so next we'll sprawl our networks, and propose our candidacy
	for _, can := range candidates {
		can.netLogic.SprawlConnections(3, 0)
		if err := can.netLogic.FillOpenConnections(); err != nil {
			t.Fatal(err)
		}
		if err := can.ProposeCandidacy(0); err != nil {
			t.Fatal(err)
		}
	}
	//candidates should all be proposed, so now let's check everyone's candidates list to make sure they have all of them
	for i, can := range candidates {
		canPool := can.poolsA
		if !assert.Equal( //they have the right number of lines
			t,
			testCandidateCount,
			len(canPool.totalCandidates),
			"candidate does not have all other candidates listed",
		) {
			//this is helpful for seeing why something breaks.
			//if you don't have the networks spread enough, a candidate proposal wont spread to everyone right away
			log.Printf("candidate index %v did not have all it's candidates in order.", i)
			for j, candidate := range candidates {
				if _, exists := canPool.totalCandidates[string(candidate.spendingAccount.PublicKey)]; !exists {
					log.Printf("the missing candidate is index %v", j)
				}
			}
			t.FailNow() //if you don't fail now, the next step will give a LONG error list
		}
		assert.EqualValues(
			t,
			0,
			canPool.currentWorkingRoundID.Get(),
			"a candidate started their round too soon",
		)
		assert.Equal(
			t,
			0,
			canPool.GetWorkingRound().currentLeadIndex,
			"candidate lead index has moved unexpectedly",
		)
	}
	roundsToWait := 10
	<-time.After(maxTimePerRound.Duration() * time.Duration(roundsToWait))
	//wait 5 rounds
	for _, can := range candidates {
		cwr := can.poolsA.GetWorkingRound()
		assert.EqualValues(
			t,
			roundsToWait,
			can.poolsA.currentWorkingRoundID.Get(),
			"rounds are off",
		)
		assert.EqualValues(
			t,
			0,
			cwr.currentLeadIndex,
			"witness lead has started indexing",
		)
		for _, otherCan := range candidates {
			assert.Equal(
				t,
				cwr.leadWitnessOrder.GetItems(),
				otherCan.poolsA.GetWorkingRound().leadWitnessOrder.GetItems(),
				"witness lead order differed between candidates",
			)
			assert.Equal(
				t,
				cwr.roundStartTime.Round(maxTimePrecision.Duration()),
				otherCan.poolsA.GetWorkingRound().roundStartTime.Round(maxTimePrecision.Duration()),
				"round start times are our of sync by a noticeable amount",
			)
		}

	}
}

type testingCandidate struct {
	vrfPrivate  crypto.PrivateKey
	spender     *accounts.Account
	nodeAccount *accounts.Account
	candidacy   *utils.Candidate
}

func newTestCandidate(seed []byte, stakeAmount *big.Int) *testingCandidate {
	tc := testingCandidate{}
	var err error = fmt.Errorf("ignore this")
	for err != nil {
		tc.vrfPrivate, err = crypto.GenerateVRFKey(rand.Reader)
	}
	err = fmt.Errorf("ignore this")
	for err != nil {
		tc.spender, err = accounts.GenerateAccount()
	}
	err = fmt.Errorf("ignore this")
	for err != nil {
		tc.nodeAccount, err = accounts.GenerateAccount()
	}
	tc.candidacy, _ = utils.NewCandidate(0, seed, tc.vrfPrivate, 0, uint8(networking.PrimaryTransactions), "", tc.nodeAccount.PublicKey, *tc.spender, stakeAmount)
	return &tc
}
func (tc *testingCandidate) updateCandidate(pool *Witness_pool) {
	tc.candidacy, _ = tc.candidacy.UpdatedCandidate(uint64(pool.currentWorkingRoundID.Get()+1), pool.GetCurrentSeed(), tc.vrfPrivate, uint64(pool.GetApplyingRound().roundStartTime.Unix()), *tc.spender)
}

func generateTestWitnesses(count int) []*testingCandidate {
	candidates := []*testingCandidate{}
	for i := 0; len(candidates) <= count; i++ {
		newCandidate := newTestCandidate([]byte{}, big.NewInt(1))
		if newCandidate == nil {
			continue
		}
		candidates = append(candidates, newCandidate)
	}
	return candidates
}
