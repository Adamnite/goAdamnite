package dpos

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/common/math"
	"github.com/adamnite/go-adamnite/core/types"
	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/gossip"
	
)

type WitnessPool struct {
	AllWitnesses []*Witness
	VoteResults  map[string]*VoteResult
}

type LeaderSelection struct {
    WitnessPool        *WitnessPool
    CurrentLeader      *Witness
    LeaderSelectionC   chan *Witness
    blocksPerRound     int
    blocksInCurrentRound map[string]int
    roundCounter       int
}

func NewLeaderSelection(witnessPool *WitnessPool, blocksPerRound int) *LeaderSelection {
    return &LeaderSelection{
        WitnessPool:        witnessPool,
        CurrentLeader:      nil,
        LeaderSelectionC:   make(chan *Witness),
        blocksPerRound:     blocksPerRound,
        blocksInCurrentRound: make(map[string]int),
        roundCounter:       0,
    }
}

func (ls *LeaderSelection) Start() {
    for {
        // Select a leader randomly
        pool := ls.WitnessPool.SelectWitnesses(nil)

        // Find the next eligible witness
        var leader *Witness
        for i := 0; i < len(pool); i++ {
            index := (i + ls.roundCounter) % len(pool)
            if ls.blocksInCurrentRound[pool[index].Address] < ls.blocksPerRound {
                leader = pool[index]
                break
            }
        }

        if leader == nil {
            fmt.Println("Error: no eligible leader found")
            continue
        }

        // Send message to all nodes to confirm leader selection
        // (using placeholder messaging protocol: gossip)
        err := gossip.Broadcast([]byte(fmt.Sprintf("LEADER:%s", leader.Address)))
        if err != nil {
            fmt.Printf("Error broadcasting leader selection: %s", err)
            continue
        }

        // Wait for responses from all nodes to confirm leader selection
        // (using placeholder messaging protocol: gossip)
        responses, err := gossip.BroadcastAndWait([]byte(fmt.Sprintf("CONFIRM_LEADER:%s", leader.Address)), len(ls.WitnessPool.AllWitnesses)-1, 10*time.Second)
        if err != nil {
            fmt.Printf("Error waiting for leader confirmation: %s", err)
            continue
        }

        // Check if all nodes responded with confirmation message
        confirmed := true
        for _, response := range responses {
            if string(response) != "CONFIRMED" {
                confirmed = false
                break
            }
        }
        if !confirmed {
            fmt.Println("Error: not all nodes confirmed leader selection")
            continue
        }

        // Set current leader and send to channel
        ls.CurrentLeader = leader
        ls.LeaderSelectionC <- leader

        // Increment block counter for the selected witness
        ls.blocksInCurrentRound[leader.Address]++

        // If all witnesses have proposed the maximum number of blocks in the current round, reset the counter and start a new round
        if len(ls.blocksInCurrentRound) == len(ls.WitnessPool.AllWitnesses) && allBlocksProposed(ls.blocksInCurrentRound, ls.blocksPerRound) {
            ls.blocksInCurrentRound = make(map[string]int)
            ls.roundCounter++
        }

		// Wait for 6 blocks
		time.Sleep(6 * 500 * time.Millisecond)
	}
}

type StateTransition struct {
	BlockDigest    []byte
	LeaderAddress  string
	EncodingDigest []byte
}

type ValidatedTransition struct {
	StateTransition
	ValidatedBlock *Block
	Timestamp      time.Time
}

type Block struct {
	Transactions []Transaction
}

type Transaction struct {
	From   string
	To     string
	Amount uint64
}

func (ls *LeaderSelection) ProposeBlock(transactions []Transaction) {
	startTime := time.Now()

	// Check if current leader is set
	if ls.CurrentLeader == nil {
		fmt.Println("Error: no leader selected")
		return
	}

	// Check if block can be proposed in time
	if time.Since(startTime) > 500*time.Millisecond {
		// Propose empty block and replace current leader
		fmt.Println("Warning: block proposal took too long, proposing empty block")
		ls.CurrentLeader = nil
		return
	}

	// Create state transition
	blockDigest := hashTransactions(transactions)
	leaderAddress := ls.CurrentLeader.Address
	encodingDigest := []byte("msgpack")


	// Encode state transition
stateTransition := StateTransition{
	BlockDigest:    blockDigest,
	LeaderAddress:  leaderAddress,
	EncodingDigest: encodingDigest,
}
encodedStateTransition, err := msgpack.Marshal(stateTransition)
if err != nil {
	fmt.Printf("Error encoding state transition: %s", err)
	return
}

// Validate block and create validated transition
validatedBlock := &Block{Transactions: transactions}
validatedTransition := ValidatedTransition{
	StateTransition: stateTransition,
	ValidatedBlock:  validatedBlock,
	Timestamp:       time.Since(startTime),
}

// Add validated transition to ledger (using placeholder database: adamnitedb)
err = adamnitedb.Save(encodedStateTransition, validatedBlock)
if err != nil {
	fmt.Printf("Error saving validated transition to ledger: %s", err)
	return
}

// Broadcast validated transition to all nodes (using placeholder messaging protocol: gossip)
err = gossip.Broadcast(encodedStateTransition)
if err != nil {
	fmt.Printf("Error broadcasting validated transition: %s", err)
	return
}

}

func hashTransactions(transactions []Transaction) []byte {
// Placeholder implementation that returns a random byte array
b := make([]byte, 32)
rand.Read(b)
return b
}


