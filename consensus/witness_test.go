package consensus

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
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
		len(selected),
		"wrong number of witnesses selected",
	)
	for _, w := range witnesses {
		w.updateCandidate(pool)
		pool.AddCandidate(w.candidacy)
	}
	pool.witnessGoal = 2
	selected, _ = pool.SelectCurrentWitnesses()
	if len(selected) != 2 {
		fmt.Println("wrong number selected")
		t.Fail()
	}
}
func TestRoundSelections(t *testing.T) {
	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	pool.StopAsyncTracker()
	maxTimePerRound = time.Second * 1 //change the time between rounds for testing.

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
	pool.newRoundStartedCaller[0]()
	pool.SelectCurrentWitnesses()
	pool.nextRound()
	if pool.currentWorkingRound != 1 {
		fmt.Println("round did not increment correctly")
		t.Fail()
	}
	nextRoundStartTime := pool.GetWorkingRound().roundStartTime.Add(maxTimePerRound).Add(5 * time.Millisecond) //give it a hair over the time
	if err := pool.StartAsyncTracking(); err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t,
		1,
		int(pool.currentWorkingRound),
		"new round started too soon",
	)
	<-time.After(time.Until(nextRoundStartTime))
	pool.StopAsyncTracker()
	assert.Equal(
		t,
		2,
		int(pool.currentWorkingRound),
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
	maxTimePerRound = time.Millisecond * 50 //change the time between rounds for testing.
	// maxTimePerRound = time.Second * 5 //change the time between rounds for debugging.
	maxTimePrecision = time.Millisecond * 5 //the error tolerance we can handle

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
	<-time.After(maxTimePerRound*time.Duration(goal) + maxTimePrecision)
	pool.StopAsyncTracker()
	if pool.GetWorkingRound().roundStartTime.After(time.Now()) {
		fmt.Println("round after now")
		t.Fail()
	}
	assert.Equal(
		t,
		goal+1, //plus one since we need to start it!
		int(pool.currentWorkingRound),
		"timing for new round generation must be wrong",
	)
}

func TestWitnessLeadSelection(t *testing.T) {
	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	pool.StopAsyncTracker()
	maxTimePerRound = time.Millisecond * 500 //change the time between rounds for testing.
	// maxTimePerRound = time.Second * 5 //change the time between rounds for debugging.
	maxTimePrecision = time.Millisecond * 5 //the error tolerance we can handle

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
		len(wrd.leadWitnessOrder),
		"incorrect witness leader list",
	)
	activeWit := wrd.leadWitnessOrder[0]
	assert.True(
		t,
		wrd.IsActiveWitnessLead(&activeWit.spendingPub),
		"leading witness is not actually lead",
	)
	//test two witnesses who acts truthfully
	var blocksFaked uint64 = 0
	for _, lead := range wrd.leadWitnessOrder {
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
	}
	assert.Equal(t,
		maxBlocksPerRound,
		blocksFaked,
		"blocks faked appears to not be accurate to the max block limit",
	)
	assert.Equal(
		t,
		3,
		int(pool.currentWorkingRound),
		"adding blocks till round was filled did not automatically increment round",
	)
	wrd = pool.GetWorkingRound()
	for i := 0; pool.currentWorkingRound == 3; i = (i + 1) % len(candidates) {
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
	tc.candidacy, _ = tc.candidacy.UpdatedCandidate(pool.currentWorkingRound+1, pool.GetCurrentSeed(), tc.vrfPrivate, uint64(pool.GetApplyingRound().roundStartTime.Unix()), *tc.spender)
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
