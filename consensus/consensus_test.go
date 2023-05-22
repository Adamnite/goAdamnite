package consensus

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func TestBaseCandidacy(t *testing.T) {
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
	fmt.Println(b.netLogic.FillOpenConnections())
	if err := a.ProposeCandidacy(); err != nil {
		t.Fatal(err)
	}

	// fmt.Println(b.candidates)
	assert.Equal(t,
		a.thisCandidate,
		b.candidates[string(a.thisCandidate.NodeID)],
		"candidacy not properly sending")
	if err := b.VoteFor(a.thisCandidate.NodeID, big.NewInt(1)); err != nil {
		t.Fatal(err)
	}
}
func TestVoteForAllEqually(t *testing.T) {
	const (
		candidateTotal int = 5
		voterTotal     int = 5
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
			account, _ := accounts.GenerateAccount()
			node, _ := NewAConsensus(*account)
			if err := node.netLogic.ConnectToContact(&seedContact); err != nil {
				t.Fatal(err)
			}
			candidates = append(candidates, node)
		}
		if i < voterTotal {
			account, _ := accounts.GenerateAccount()
			node, _ := NewAConsensus(*account)
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
		if err := can.ProposeCandidacy(); err != nil {
			t.Fatal(err)
		}
	}

	for i, v := range voters {
		candidateToVoteFor := candidates[i%candidateTotal].thisCandidate.NodeID
		if err := v.VoteFor(candidateToVoteFor, big.NewInt(1)); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range candidates {
		if len(c.candidates)+1 != len(candidates) {
			log.Println("not recording all candidates")
			t.Fail()
		}
		for _, val := range c.votesSeen {
			if len(val) != voterTotal/candidateTotal {
				log.Println("incorrect vote tally recorded")
				t.Fail()
			}
		}
	}
	// fmt.Println("hi")
}
