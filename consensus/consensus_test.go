package consensus

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/adamnite/go-adamnite/utils/accounts"
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
	bcontact := b.netLogic.GetOwnContact()
	if err := a.netLogic.ConnectToContact(&bcontact); err != nil {
		t.Fatal(err)
	}
	fmt.Println(b.netLogic.FillOpenConnections())
	if err := a.ProposeCandidacy(); err != nil {
		t.Fatal(err)
	}
	if err := b.VoteFor(a.thisCandidate.NodeID, big.NewInt(1)); err != nil {
		t.Fatal(err)
	}

	fmt.Println(b.candidates)
	// assert.Equal(t,
	// 	a.thisCandidate,
	// 	b.candidates[a.thisCandidate.NodeID],
	// 	"candidacy not properly sending")

}
