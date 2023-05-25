package utils

import (
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/utils/accounts"
)

type CaesarMessage struct {
	To          accounts.Account //neither of these have the private key sent during serialization.
	From        accounts.Account
	InitialTime time.Time
	Message     []byte
	Signature   []byte
}

func NewCaesarMessage(to accounts.Account, from accounts.Account, saying interface{}) (*CaesarMessage, error) {
	//format the message
	ansMessage := CaesarMessage{
		To:          to,
		From:        from,
		InitialTime: time.Now().UTC(),
	}

	//get the message from a variance of types
	messageBytes := []byte{}
	switch v := saying.(type) {
	case string:
		messageBytes = []byte(v)
	case []byte:
		messageBytes = v
	default:
		return nil, fmt.Errorf("i don't know how to handle the message type you sent") //TODO: replace with real error
	}
	//TODO: encrypt the message
	return &ansMessage, nil
}

// Hash the message
func (cm CaesarMessage) Hash() []byte {
	ans := append(cm.To.PublicKey, cm.From.PublicKey...)
	timeByte, err := cm.InitialTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	ans = append(ans, timeByte...)
	// return crypto.Sha512(ans)
	return ans
}

// Sign the message with the From Account
func (cm *CaesarMessage) Sign() error {
	newSignature, err := cm.From.Sign(cm)
	if err == nil {
		cm.Signature = newSignature
	}
	return err
}

func (cm CaesarMessage) Verify() bool {
	return cm.From.Verify(cm, cm.Signature)
}

func (cm CaesarMessage) GetMessageString() string {

	return string(cm.Message) //TODO: decrypt this
}
