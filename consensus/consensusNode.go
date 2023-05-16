package consensus

import (
	"log"
	"time"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/blockchain"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/networking"
	"github.com/adamnite/go-adamnite/utils"
)

// The node that can handle consensus systems.
// Is built on top of a networking node, using a netNode to handle the networking interactions
type ConsensusNode struct {
	thisCandidate  *Candidate
	candidatesSeen []*Candidate //should probably also have a mapping for efficiency.
	netLogic       *networking.NetNode
	handlingType   consensusHandlingTypes
	//TODO: VRF keys, aka, the keyset used for applying for votes and whatnot. should have its own type upon implementation here
	spendingKey crypto.PrivateKey // each consensus node is forced to have its own account to spend from.
	state       *statedb.StateDB
	chain       *blockchain.Blockchain //we need to keep the chain
	ocdbLink    string                 //off chain database, if running the VM verification, this should be local.
	vm          *VM.Machine
}

func NewAConsensus() (ConsensusNode, error) {
	//TODO: setup the chain data and whatnot
	conNode, err := newConsensus(networking.NewNetNode(common.Address{}), nil, nil)
	conNode.handlingType = SecondaryTransactions
	return conNode, err
}

func newConsensus(hostingNode *networking.NetNode, state *statedb.StateDB, chain *blockchain.Blockchain) (ConsensusNode, error) {
	con := ConsensusNode{
		netLogic:     hostingNode,
		handlingType: NetworkingOnly,
		state:        state,
		chain:        chain,
	}
	if err := hostingNode.AddFullServer(state, chain, con.ReviewTransaction); err != nil {
		log.Printf("error:%v", err)
		return ConsensusNode{}, err
	}
	return con, nil
}
func (con *ConsensusNode) ReviewTransaction(transaction *utils.Transaction) error {
	//TODO: give this a quick look over, review it, if its good, add it locally and propagate it out, otherwise, ignore it.
	return nil
}
func (con *ConsensusNode) ReviewCandidacy(proposed *Candidate) {
	//TODO: this is how you would vote for a candidate.
}

// propose this node as a witness for the network.
func (con *ConsensusNode) ProposeCandidacy() error {
	if con.thisCandidate == nil {
		con.thisCandidate = con.generateCandidacy()
	}
	//TODO: update candidacy for this rounds
	//TODO: propagate a proposal through the network
	return nil
}

// generate a mostly blank self candidate proposal
func (con *ConsensusNode) generateCandidacy() *Candidate {

	thisCandidate := Candidate{
		Round: 0, //TODO: have the round numbers
		// StartTime: nil,//TODO:
		IsActive:      false,
		Stake:         0, //how much do we want to stake
		Votes:         0, //to start no ones voted for us.
		Reputation:    0, //TODO: figure out how we calculate this
		FastTimeout:   0, //TODO: figure out what this is for exactly
		NetworkPoint:  con.netLogic.GetOwnContact(),
		ConsensusPool: con.handlingType,
	}
	public, _ := con.spendingKey.Public()
	thisCandidate.PublicKey = public
	return &thisCandidate
}

type Candidate struct {
	Round       uint64           //round number proposing for
	StartTime   uint64           //when that round should start
	IsActive    bool             //save the math, is this round still running
	Stake       uint64           //how much we stake on ourselves
	Votes       uint64           //how many votes are for us
	Reputation  uint64           //our reputation
	FastTimeout uint64           //?
	Deadline    time.Time        //when this candidacy proposal expires?
	PublicKey   crypto.PublicKey //our spending public key
	// PrivateKey   *ecdsa.PrivateKey  //WHAT???
	NetworkPoint  networking.Contact     //where this node is
	ConsensusPool consensusHandlingTypes //do we want to offer support for VM
	SeenMessages  map[string]bool        // to prevent duplicate messages (I don't quite get this yet)
	// MessageChannel   chan Message //not sure why we would have a message channel?
	// QuitChannel      chan bool	//what are these for?
	// ConsensusParams  ConsensusParams//what parameters do we use?
	ParticipationKey string // key used to sign activation certificate. VRF key
}
