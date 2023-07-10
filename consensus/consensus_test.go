package consensus

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/rpc"
	"github.com/adamnite/go-adamnite/utils"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func TestBaseCandidacy(t *testing.T) {
	rpc.USE_LOCAL_IP = true //use local IPs so we don't wait to get our IP, and don't need to deal with opening the firewall port
	aAccount, _ := accounts.GenerateAccount()
	a, err := NewAConsensus(*aAccount)
	if err != nil {
		t.Fatal(err)
	}
	bAccount, _ := accounts.GenerateAccount()
	b, err := NewAConsensus(*bAccount)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("A Address: %v\nB Address: %v", aAccount.Address.Hex(), bAccount.Address.Hex())
	bcontact := b.netLogic.GetOwnContact()
	if err := a.netLogic.ConnectToContact(&bcontact); err != nil {
		t.Fatal(err)
	}
	a.autoStakeAmount = big.NewInt(1) //you need to have some stake amount
	fmt.Println(b.netLogic.FillOpenConnections())
	if err := a.ProposeCandidacy(0); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t,
		a.thisCandidateA,
		b.poolsA.GetCandidate((*crypto.PublicKey)(&a.spendingAccount.PublicKey)),
		"candidacy not properly sending",
	)
	assert.Equal(t,
		a.thisCandidateA,
		a.poolsA.GetCandidate((*crypto.PublicKey)(&a.spendingAccount.PublicKey)),
		"candidacy not saving on self properly sending",
	)
	if err := b.VoteFor(a.thisCandidateA, big.NewInt(1)); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t,
		1, a.poolsA.GetApplyingRound().eligibleWitnesses.Len(),
		"extra vote catagories compared to number of people running",
	)
	assert.Equal(t,
		2, len(a.poolsA.GetApplyingRound().GetVotesFor(a.spendingAccount.PublicKey)),
		"not enough votes correctly registered",
	)
	assert.Equal(t,
		1, b.poolsA.GetApplyingRound().eligibleWitnesses.Len(),
		"extra vote catagories compared to number of people running",
	)
	assert.Equal(t,
		2, len(b.poolsA.GetApplyingRound().GetVotesFor(a.spendingAccount.PublicKey)),
		"not enough votes correctly registered",
	)
}
func TestVoteForAllEqually(t *testing.T) {
	rpc.USE_LOCAL_IP = true //use local IPs so we don't wait to get our IP, and don't need to deal with opening the firewall port
	const (
		candidateTotal int = 5
		voterTotal     int = 50
	)
	seedNode := networking.NewNetNode(common.Address{0, 0, 0})
	if err := seedNode.AddServer(); err != nil {
		t.Fatal(err)
	}
	seedContact := seedNode.GetOwnContact()

	candidates := []*ConsensusNode{}

	voters := []*ConsensusNode{}
	//fill the voters and candidates
	for i := 0; i < candidateTotal || i < voterTotal; i++ {
		if i < candidateTotal {
			var account *accounts.Account
			//there's a slight chance of an error happening at key gen. (which only really comes up in bulk running tests)
			// To solve this we make sure to create an account until we, well, create one!
			for account == nil {
				account, _ = accounts.GenerateAccount()
			}
			node, err := NewAConsensus(*account)
			for err != nil {
				log.Println("error creating consensus node, trying again")
				node.netLogic.Close() //close and try again any time theres an error
				node, err = NewAConsensus(*account)
			}
			if err := node.netLogic.ConnectToContact(&seedContact); err != nil {
				t.Fatal(err)
			}
			node.autoStakeAmount = big.NewInt(1)
			candidates = append(candidates, node)
		}
		if i < voterTotal {
			var account *accounts.Account
			for account == nil {
				account, _ = accounts.GenerateAccount()
			}
			node, err := NewAConsensus(*account)
			for err != nil {
				log.Println("error creating consensus node, trying again")
				node.netLogic.Close() //close and try again any time theres an error
				node, err = NewAConsensus(*account)
			}
			if err := node.netLogic.ConnectToContact(&seedContact); err != nil {
				t.Fatal(err)
			}
			voters = append(voters, node)
		}
	}
	//get everyone to know each other
	for i, voter := range voters {
		if i < candidateTotal {
			if err := candidates[i].netLogic.SprawlConnections(3, 0); err != nil {
				t.Fatal(err)
			}
			if err := candidates[i].netLogic.ResetConnections(); err != nil {
				t.Fatal(err)
			}
		}
		if err := voter.netLogic.SprawlConnections(3, 0); err != nil {
			t.Fatal(err)
		}
		if err := voter.netLogic.ResetConnections(); err != nil {
			t.Fatal(err)
		}
	}
	log.Println("\n\n\tstart of candidate proposal ")
	//everyone's spun up and connected. Now the people who want to propose can.(in our case, all the candidates)
	for _, can := range candidates {
		if err := can.ProposeCandidacy(0); err != nil {
			t.Fatal(err)
		}
	}

	for i, v := range voters {
		candidateToVoteFor := candidates[i%candidateTotal]
		if err := v.VoteFor(candidateToVoteFor.thisCandidateA, big.NewInt(1)); err != nil {
			t.Fatal(err)
			t.FailNow()
		}
	}
	for _, c := range candidates {
		if len(c.poolsA.totalCandidates) != len(candidates) {
			log.Println("not recording all candidates")
			t.Fail()
		}
		crd := c.poolsA.GetWorkingRound()
		crd.votes.Range(func(key, value any) bool {
			val := value.([]*utils.Voter)

			if len(val) != (voterTotal/candidateTotal)+1 { //cant forget that you vote for yourself!
				log.Printf("incorrect vote tally recorded. Expected %v, got %v", voterTotal/candidateTotal, len(val))
				t.Fail()
			}
			valTotals, _ := crd.valueTotals.Load(key)
			if valTotals.(*big.Int).Cmp(big.NewInt(int64((voterTotal/candidateTotal)+1))) != 0 {
				log.Printf("not the right vote total value. Expected %v, got %v", (voterTotal / candidateTotal), valTotals.(*big.Int).Int64())
				t.Fail()
			}
			return true
		})

	}
}
