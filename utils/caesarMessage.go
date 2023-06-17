package utils

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/adamnite/go-adamnite/crypto"
	"github.com/adamnite/go-adamnite/utils/accounts"
)

type CaesarMessage struct {
	To               *accounts.Account //neither of these have the private key sent during serialization.
	From             *accounts.Account
	InitialTime      time.Time
	Message          []byte
	Signature        []byte
	HasHostingServer bool //is this to a nodeID, who would have a server running directly (instead of needing to be shared to everyone)
}

// create a Caesar Message. "saying" can be of type byte or string
func NewCaesarMessage(to *accounts.Account, from *accounts.Account, saying interface{}) (*CaesarMessage, error) {
	ansMsg := CaesarMessage{
		To:               to,
		From:             from,
		InitialTime:      time.Now().UnixMicro(),
		HasHostingServer: false,
	}

	// get the message from a variance of types
	var messageData []byte
	switch v := message.(type) {
	case string:
		messageData = []byte(v)
	case []byte:
		messageData = v
	default:
		return nil, fmt.Errorf("Unsupported message type: should be either byte array or string")
	}

	encryptedMessage, err := to.Encrypt(messageData)
	if err != nil {
		return nil, err
	}
	c.Message = encryptedMessage

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
}

// Hash the message
func (cm CaesarMessage) Hash() []byte {
	ans := append(cm.To.PublicKey, cm.From.PublicKey...)
	ans = binary.LittleEndian.AppendUint64(ans, uint64(cm.InitialTime))
	return crypto.Sha512(ans)
}

// Sign the message with the From Account
func (cm *CaesarMessage) Sign() error {
	newSignature, err := cm.From.Sign(cm.Message)
	if err == nil {
		cm.Signature = newSignature
	}
	return err
}

func (cm CaesarMessage) Verify() bool {
	return cm.From.Verify(cm.Message, cm.Signature)
}

// get the message contents by decrypting it
func (cm CaesarMessage) GetMessage(recipient *accounts.Account) ([]byte, error) {
	return recipient.Decrypt(cm.Message)
}

// get the message contents by decrypting it
func (cm CaesarMessage) GetMessageString(recipient *accounts.Account) (string, error) {
	msg, err := cm.GetMessage(recipient)
	return string(msg), err
}
func (cm CaesarMessage) GetTime() time.Time {
	return time.UnixMicro(cm.InitialTime)
}
