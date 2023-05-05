// A candidate is a new candidate/witness within the Delegated Proof of Stake protocol. 

//Note: Be sure to check for validity and fit with the rest of the code. 


package dpos

import (

	//Replace with actual crypto imports (use the signature functions to sign messages)
	//I was having trouble getting it to play nice with the new crypto library.
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"time"
)

type Candidate struct {
	Round             uint64
	StartTime         uint64
	IsActive          bool
	Stake             uint64
	Votes             uint64
	Reputation        uint64
	FastTimeout       uint64
	Deadline          time.Time
	PublicKey         *ecdsa.PublicKey
	PrivateKey        *ecdsa.PrivateKey
	SeenMessages      map[string]bool // to prevent duplicate messages
	MessageChannel    chan Message
	QuitChannel       chan bool
	ConsensusParams   ConsensusParams
	ParticipationKey  string // key used to sign activation certificate
}

type Message struct {
	Round    uint64
	Sender   string // sender's address
	Type     string // type of message: "activate", "vote"
	Data     []byte // encoded vote data or activation data
	Deadline uint64 // deadline for message to be processed
}

type ActivationCertificate struct {
	Round            uint64
	StartTime        uint64
	Stake            uint64
	ParticipationKey string
	Signature        []byte
}

func (c *Candidate) HandleMessage(m Message) error {
	if m.Deadline < uint64(time.Now().Unix()) {
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

	// sign activation certificate
	cert := &ActivationCertificate{
		Round:            activationData.Round,
		StartTime:        activationData.StartTime,
		Stake:            activationData.Stake,
		ParticipationKey: c.ParticipationKey,
	}
	sig, err := c.signActivationCertificate(cert)
	if err != nil {
		return err
	}
	cert.Signature = sig

	// encode certificate and send through gossip protocol
	certBytes, err := asn1.Marshal(cert)
	if err != nil {
		return err
	}
	msg := Message{
		Round:    c.Round,
		Sender:   c.Address(),
		Type:     "activate",
		Data:     certBytes,
		Deadline:

c.ConsensusParams.ActivationDeadline + uint64(time.Now().Unix()),
}
for _, peer := range c.ConsensusParams.Peers {
go func(p string) {
err := c.ConsensusParams.SendMessage(p, msg)
if err != nil {
fmt.Println(err)
}
}(peer)
}
return nil
}

func (c *Candidate) handleVote(data []byte) error {
// decode vote data
var voteData struct {
Round uint64
Votes uint64
Hash []byte
SignR []byte
SignS []byte
Address string
}
_, err := asn1.Unmarshal(data, &voteData)
if err != nil {
return err
}


// check if the round matches and the vote hash is correct
if voteData.Round != c.Round {
	return errors.New("incorrect round for vote")
}
hash := sha256.Sum256(data)
if !bytes.Equal(hash[:], voteData.Hash) {
	return errors.New("invalid vote hash")
}

// verify the vote signature
pubKey := GetPublicKeyFromAddress(voteData.Address)
if pubKey == nil {
	return errors.New("invalid address")
}
signature := &ecdsa.Signature{
	R: new(big.Int).SetBytes(voteData.SignR),
	S: new(big.Int).SetBytes(voteData.SignS),
}
if !ecdsa.Verify(pubKey, hash[:], signature.R, signature.S) {
	return errors.New("invalid signature")
}

// update candidate's votes and reputation
c.Votes += voteData.Votes
c.Reputation += voteData.Votes

return nil
}

func (c *Candidate) signActivationCertificate(cert *ActivationCertificate) ([]byte, error) {
h := sha256.New()
_, err := h.Write(cert.ParticipationKey)
if err != nil {
return nil, err
}
_, err = h.Write([]byte(fmt.Sprintf("%d:%d:%d", cert.Round, cert.StartTime, cert.Stake)))
if err != nil {
return nil, err
}
hash := h.Sum(nil)
r, s, err := ecdsa.Sign(rand.Reader, c.PrivateKey, hash)
if err != nil {
return nil, err
}
signature, err := asn1.Marshal(struct{ R, S *big.Int }{r, s})
if err != nil {
return nil, err
}
return signature, nil
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
