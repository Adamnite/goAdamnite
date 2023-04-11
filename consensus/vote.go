package dpos
//Note to self, account for equivocation votes (votes using same coins to two different candidates)
//Add logic for verifying 
import (
	"https://github.com/Adamnite/goAdamnite/crypto"
	"errors"
	"fmt"
	"https://github.com/Adamnite/goAdamnite/common/types"
	"time"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

)

//The vote header consists of the key structures in a normal vote
type VoteHeader struct {
	Sender    Common.Address
	Round     uint64
	Timestamp int64
	Amount    uint64
	Recipient Common.Address
}

func (vh *VoteHeader) Serialize() []byte {
	result := make([]byte, 0)
	result = append(result, vh.Sender...)
	result = append(result, Uint64ToBytes(vh.Round)...)
	result = append(result, Int64ToBytes(vh.Timestamp)...)
	result = append(result, Uint64ToBytes(vh.Amount)...)
	result = append(result, vh.Recipient...)
	return result
}


//The vote structure consists of the vote header and a signature associated with the vote header

type Vote struct {
	Header     VoteHeader
	Signature  crypto.NewOneTimeSignature
}



func (v *Vote) Verify(pubKey *ecdsa.PublicKey) error {
	// Fetch consensus parameters for the current round
	consensusParams, err := GetConsensusParams(v.Header.Round)
	if err != nil {
		return errors.New("failed to fetch consensus parameters")
	}

	// Check if the vote timestamp is within the consensus parameters
	if v.Header.Timestamp > consensusParams.MaxTimestamp {
		return errors.New("vote timestamp exceeds maximum timestamp for the current round")
	}

	// Check if the recipient address is valid and in the current voting pool
	if !IsValidAddress(v.Header.Recipient) || !IsInVotingPool(v.Header.Recipient, consensusParams.VotingPool) {
		return errors.New("invalid recipient address or not in the current voting pool")
	}

	// Verify the vote signature
	hashed := sha256.Sum256(v.Header.Serialize())
	var sig struct {
		R, S *big.Int
	}
	_, err = asn1.Unmarshal(v.Signature, &sig)
	if err != nil {
		return errors.New("failed to unmarshal signature")
	}
	if !ecdsa.Verify(pubKey, hashed[:], sig.R, sig.S) {
		return errors.New("failed to verify vote signature")
	}
	return nil
}


func IsValidAddress(addr Address) bool {
	// Implementation to check if an address is valid
	return true
}

func IsInVotingPool(addr Address, pool []Address) bool {
	// Implementation to check if an address is in the current voting pool
	return true
}
