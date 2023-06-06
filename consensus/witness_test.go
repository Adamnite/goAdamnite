package consensus

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

func TestWitnessSelection(t *testing.T) {
	witnessCount := 10
	witVRFPrivs := []crypto.PrivateKey{}
	witAccounts := []*accounts.Account{}
	witCans := []*utils.Candidate{}
	seed := []byte{0, 1, 2, 3, 4, 5, 6}

	pool := newWitnessPool(0, networking.PrimaryTransactions)
	if err := pool.newRound(0, seed); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < witnessCount; i++ {
		sk, _ := crypto.GenerateVRFKey(rand.Reader)
		ac, _ := accounts.GenerateAccount()
		nodeAccounts, _ := accounts.GenerateAccount()
		witAccounts = append(witAccounts, ac)
		witVRFPrivs = append(witVRFPrivs, sk)
		can, err := utils.NewCandidate(0, seed, sk, 0, uint8(networking.PrimaryTransactions), "", nodeAccounts.PublicKey, *ac, big.NewInt(int64(seed[i%len(seed)])))
		if err != nil {
			t.Fatal(err)
		}
		witCans = append(witCans, can)
		pool.AddCandidate(can)
	}

	selected, nextSeed := pool.selectWitnesses(0, witnessCount-1)
	if len(selected) != witnessCount-1 {
		t.Fail()
	}
	pool.nextRound(nextSeed)
	for i := 0; i < witnessCount; i++ {
		oldCan := witCans[i]
		can, err := oldCan.UpdatedCandidate(1, nextSeed, witVRFPrivs[i], 1, *witAccounts[i])
		if err != nil {
			t.Fatal(err)
		}
		pool.AddCandidate(can)
	}

	selected, nextSeed = pool.selectWitnesses(1, 2)
	if len(selected) != 2 {
		t.Fail()
	}
}
