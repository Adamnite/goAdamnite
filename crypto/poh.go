package crypto

import (
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	DelayFactor   = 100000 // Delay factor for time-based delay
	HashOutputLen = 64     // Length of SHA-512 hash output in bytes
)

type ProofOfHistory struct {
	state  string
	proofs []string
}

func NewProofOfHistory() *ProofOfHistory {
	return &ProofOfHistory{}
}

func (poh *ProofOfHistory) ApplyChange(change string) {
	poh.state += change
	proof := poh.generateProof()
	poh.proofs = append(poh.proofs, proof)
}

func (poh *ProofOfHistory) generateProof() string {
	previousProof := ""
	if len(poh.proofs) > 0 {
		previousProof = poh.proofs[len(poh.proofs)-1]
	}

	proof := poh.computeProof(poh.state, previousProof)
	time.Sleep(time.Duration(len(poh.state)) * time.Millisecond * DelayFactor)
	return proof
}

func (poh *ProofOfHistory) computeProof(state string, previousProof string) string {
	hash := sha512.Sum512([]byte(state + previousProof))
	proof := hex.EncodeToString(hash[:])
	return proof
}

func main() {
	poh := NewProofOfHistory()

	// Apply a series of changes
	changes := []string{"Change 1", "Change 2", "Change 3", "Change 4"}
	for _, change := range changes {
		poh.ApplyChange(change)
	}

	// Print the final state and proofs
	fmt.Println("Final State:", poh.state)
	fmt.Println("Proofs:")
	for i, proof := range poh.proofs {
		fmt.Printf("Change %d: %s\n", i+1, proof)
	}
}
