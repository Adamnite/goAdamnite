package utils

import (
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type CaesarMessage struct {
	To               accounts.Account //neither of these have the private key sent during serialization.
	From             accounts.Account
	InitialTime      time.Time
	Message          []byte
	Signature        []byte
	HasHostingServer bool //is this to a nodeID, who would have a server running directly (instead of needing to be shared to everyone)
}

// create a Caesar Message. "saying" can be of type byte or string
func NewCaesarMessage(to accounts.Account, from accounts.Account, saying interface{}) (*CaesarMessage, error) {
	ansMsg := CaesarMessage{
		To:               to,
		From:             from,
		InitialTime:      time.Now().UTC(),
		HasHostingServer: false,
	}

	//get the message from a variance of types
	var msgBytes []byte
	switch v := saying.(type) {
	case string:
		msgBytes = []byte(v)
	case []byte:
		msgBytes = v
	default:
		return nil, fmt.Errorf("i don't know how to handle the message type you sent") //TODO: replace with real error
	}

	ansBytes, err := to.Encrypt(msgBytes)
	if err != nil {
		return nil, err
	}
	ansMsg.Message = ansBytes

<<<<<<< Updated upstream
	err = ansMsg.Sign()
	return &ansMsg, err
=======
	err = c.Sign()
	return &c, err
}

// NewCaesarMessage creates a new Caesar message with signature set
func NewSignedCaesarMessage(to accounts.Account, from accounts.Account, message []byte, signature []byte) *CaesarMessage {
	return &CaesarMessage{
		To:               to,
		From:             from,
		InitialTime:      time.Now().UnixMicro(),
		HasHostingServer: false,
		Message:		  message,
		Signature: 		  signature,
	}
>>>>>>> Stashed changes
}

// Hash the message
func (cm CaesarMessage) Hash() []byte {
	ans := append(cm.To.PublicKey, cm.From.PublicKey...)
	timeByte, err := cm.InitialTime.MarshalBinary()
	if err != nil {
		panic(err)
	}
	ans = append(ans, timeByte...)
	return crypto.Sha512(ans)
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

// get the message contents by decrypting it
func (cm CaesarMessage) GetMessage(recipient accounts.Account) ([]byte, error) {
	return recipient.Decrypt(cm.Message)
}

// get the message contents by decrypting it
func (cm CaesarMessage) GetMessageString(recipient accounts.Account) (string, error) {
	msg, err := cm.GetMessage(recipient)
	return string(msg), err
}
