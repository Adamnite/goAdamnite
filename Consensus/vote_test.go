import (
	"https://github.com/Adamnite/goAdamnite/crypto"
	"https://github.com/Adamnite/goAdamnite/crypto/secp256k1"
	"errors"
	"fmt"
	"https://github.com/Adamnite/goAdamnite/common/types"
	"time"
	"crypto/rand"

)


func test(){
		privKey, _ := Create_Key()
		pubKey := &privKey.PublicKey
	
		// Example vote
		// Note: Change example addresses to Adamnite Addresses, these are just placeholders.
		sender := Address("0x123456789abcdef")
		recipient := Address("0x987654321fedcba")
		vote := Vote{
			Header: VoteHeader{
				Sender:    sender,
				Round:     1,
				Timestamp: 1617793738,
				Amount:    1000,
				Recipient: recipient,
			},
		}
	
		// Sign the vote
		hashed := sha256.Sum256(vote.Header.Serialize())
		vote.Signature = NewOneTimeSignature()
	
		// Verify the vote
		err := vote.Verify(pubKey)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Vote is valid")
		}
	}