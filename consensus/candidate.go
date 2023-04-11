// A candidate is a new candidate/witness within the Delegated Proof of Stake protocol. 

//Note: Be sure to check for validity and fit with the rest of the code. 


package dpos

import (
	"https://github.com/adamnite/go-adamnite/crypto"
	"errors"
	"fmt"
	"math/big"
	"time"
)

type Candidate struct {
	Round           uint64
	StartTime       uint64
	IsActive        bool
	Stake           uint64
	Votes           uint64
	Reputation      uint64
	FastTimeout     uint64
	Deadline        uint64
	PublicKey       *ecdsa.PublicKey
	PrivateKey      *ecdsa.PrivateKey
	SeenMessages    map[string]bool // to prevent duplicate messages
	MessageChannel  chan Message
	QuitChannel     chan bool
	ConsensusParams ConsensusParams
}

type Message struct {
	Round    uint64
	Sender   string // sender's address
	Type     string // type of message: "activate", "vote"
	Data     []byte // encoded vote data or activation data
	Deadline uint64 // deadline for message to be processed
}

func (c *Candidate) HandleMessage(m Message) error {
	if m.Deadline < time.Now().Unix() {
		return errors.New("message expired")
	}
	if _, ok := c.SeenMessages[m.Sender]; ok {
		return errors.New("duplicate message")
	}
	c.SeenMessages[m.Sender] = true
	switch m.Type {
	case "activate":
		err := c.handleActivate(m.Data)
		if err != nil {
			return err
		}
	case "vote":
		err := c.handleVote(m.Data)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown message type")
	}
	return nil
}

func (c *Candidate) handleActivate(data []byte) error {
	// decode the activation data
	var activationData struct {
		Round     uint64
		StartTime uint64
		Stake     uint64
	}
	_, err := asn1.Unmarshal(data, &activationData)
	if err != nil {
		return err
	}
	// check if the round matches and the candidate is not already active
	if activationData.Round != c.Round {
		return errors.New("incorrect round for activation")
	}
	if c.IsActive {
		return errors.New("candidate already active")
	}
	// set candidate as active and update the stake
	c.IsActive = true
	c.Stake = activationData.Stake
	c.StartTime = activationData.StartTime
	c.Deadline = 0 // reset the deadline
	return nil
}

func (c *Candidate) handleVote(data []byte) error {
	// decode the vote data
	var voteData struct {
		Round     uint64
		Timestamp uint64
		Amount    uint64
		Recipient string
	}
	_, err := asn1.Unmarshal(data, &voteData)
	if err != nil {
		return err
	}
	// check if the round matches and the candidate is active
	if voteData.Round != c.Round {
		return errors.New("incorrect round for vote")
	}
	if !c.IsActive {
		return errors.New("candidate not active")
	}
	// verify the vote
	recipient := Address(voteData.Recipient)
	vote := Vote{
		Header: VoteHeader{
			Sender:    Address(""),
			Round:     voteData.Round,
			Timestamp: voteData.Timestamp,
			Amount:    voteData.Amount,
			Recipient: recipient,
		},
		Signature: nil,
	}
	err = vote.Verify(c.PublicKey)
	if err != nil {
		return err
	}
	// update the candidate's votes and reputation
	c.Votes += voteData.Amount
	c.Reputation += voteData.Amount
	return nil
}


func (c *Candidate) Run(ctx context.Context, gossip *GossipProtocol) error {
	// Start the ticker
	ticker := time.NewTicker(c.consensusParams.BlockTime)
	defer ticker.Stop()

	// Set the deadline to 0 when activating for next round
	if c.isActive {
		c.deadline = time.Time{}
	}

	for {
		select {
		case <-ctx.Done():
			// If context is done, return without error
			return nil

		case <-ticker.C:
			// Update the current time
			now := time.Now()

			// Check if the candidate should timeout
			if c.isActive && !c.deadline.IsZero() && now.After(c.deadline) {
				c.isActive = false
				c.stake = 0
				c.votes = 0
				c.reputation = 0
				c.deadline = time.Time{}
				c.logger.Printf("Candidate %s timed out\n", c.address)
			}

			// Check if candidate should declare themselves for the next round
			if !c.isActive && c.stake >= c.consensusParams.MinStake {
				c.isActive = true
				c.reputation = c.stake
				c.deadline = now.Add(c.consensusParams.RoundTimeout)
				c.logger.Printf("Candidate %s declared themselves for round %d with %d stake\n", c.address, c.currentRound+1, c.stake)
			}

		case msg := <-gossip.Events():
			// Handle incoming message events
			switch msg.Type {
			case MessageTypeVote:
				voteMsg := msg.Data.(VoteMessage)
				if voteMsg.Header.Recipient != c.address {
					// Ignore vote messages not meant for this candidate
					continue
				}

				// Verify the vote signature
				if !voteMsg.Vote.Verify(&c.pubKey) {
					c.logger.Printf("Invalid vote signature from %s\n", voteMsg.Vote.Header.Sender)
					continue
				}

				// Add the vote to the candidate's total votes and reputation
				c.votes += voteMsg.Vote.Header.Amount
				c.reputation += voteMsg.Vote.Header.Amount
				c.logger.Printf("Received %d votes from %s\n", voteMsg.Vote.Header.Amount, voteMsg.Vote.Header.Sender)

			default:
				// Ignore other message types
				continue
			}
		}
	}
}