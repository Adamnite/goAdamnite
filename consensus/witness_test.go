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
)

func TestWitnessSelection(t *testing.T) {
	witnessCount := 10

	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
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
	if len(selected) != witnessCount-1 {
		t.Fail()
	}
	for _, w := range witnesses {
		w.updateCandidate(pool)
		pool.AddCandidate(w.candidacy)
	}
	pool.witnessGoal = 2
	selected, _ = pool.SelectCurrentWitnesses()
	if len(selected) != 2 {
		t.Fail()
	}
}
func TestRoundSelections(t *testing.T) {
	pool, err := NewWitnessPool(0, networking.PrimaryTransactions, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	maxTimePerRound = time.Second * 1 //change the time between rounds for testing.

	candidates := generateTestWitnesses(15)

	//set the newRound caller to add the candidates again (as if people had reapplied in between rounds)
	pool.newRoundStartedCaller = func() {
		for _, can := range candidates {
			can.updateCandidate(pool)
			if err := pool.AddCandidate(can.candidacy); err != nil {
				t.Fatal(err)
			}
		}
	}
	pool.newRoundStartedCaller()
	pool.SelectCurrentWitnesses()
	if pool.currentRound != 1 {
		fmt.Println("round did not increment correctly")
		t.Fail()
	}

	if err := pool.StartAsyncTracking(); err != nil {
		t.Fatal(err)
	}
	if pool.currentRound != 1 {
		//check that it doesn't start the next round right away
		fmt.Println("new round started too soon.")
		t.Fail()
	}
	<-time.After(maxTimePerRound + 5)
	pool.StopAsyncTracker()

	if pool.currentRound != 2 {
		fmt.Println("new round hasn't started.")
		t.FailNow()
	}

	pool.newRoundStartedCaller()
	if err := pool.StartAsyncTracking(); err != nil {
		t.Fatal(err)
	}
	<-time.After(5*maxTimePerRound + 1)
	pool.StopAsyncTracker()

	if pool.currentRound != 7 {
		fmt.Println("timing for new round generation must be wrong.")
		fmt.Printf("current round is %v", pool.currentRound)
		t.Fail()
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
	tc.vrfPrivate, _ = crypto.GenerateVRFKey(rand.Reader)
	tc.spender, _ = accounts.GenerateAccount()
	tc.nodeAccount, _ = accounts.GenerateAccount()
	tc.candidacy, _ = utils.NewCandidate(0, seed, tc.vrfPrivate, 0, uint8(networking.PrimaryTransactions), "", tc.nodeAccount.PublicKey, *tc.spender, stakeAmount)
	return &tc
}
func (tc *testingCandidate) updateCandidate(pool *Witness_pool) {
	tc.candidacy, _ = tc.candidacy.UpdatedCandidate(pool.currentRound, pool.GetCurrentSeed(), tc.vrfPrivate, uint64(pool.GetCurrentRound().roundStartTime.Unix()), *tc.spender)
}

func generateTestWitnesses(count int) []*testingCandidate {
	candidates := []*testingCandidate{}
	for i := 0; i < count; i++ {
		candidates = append(candidates, newTestCandidate([]byte{}, big.NewInt(1)))
	}
	return candidates
}
