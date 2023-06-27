package utils

import (
	"fmt"
	"testing"

	"github.com/adamnite/go-adamnite/utils/accounts"
	"github.com/stretchr/testify/assert"
)

func TestMessaging(t *testing.T) {
	sender, _ := accounts.GenerateAccount()
	receiver, _ := accounts.GenerateAccount()

	//stripped receiver is the receiver account that everyone would see, and does not have a private key
	strippedReceiver := accounts.AccountFromPubBytes(receiver.PublicKey)
	testMessage := "Hello World!"
	msg, err := NewCaesarMessage(strippedReceiver, sender, testMessage)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, msg.Verify(), "could not verify to be true")
	assert.NotEqual(
		t,
		[]byte(testMessage),
		msg.Message,
		"looks like it isn't encrypting",
	)
	badMsg, err := msg.GetMessageString(strippedReceiver) //the stripped receiver, who doesn't have a private key, should not be able to get the message
	if err == nil {
		fmt.Println("error with bad account was not nil")
		t.Fail()
	}
	assert.NotEqual(t, testMessage, badMsg, "message was decrypted by stripped receiver (aka, anyone could)")

	ansMsg, err := msg.GetMessageString(receiver)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testMessage, ansMsg, "message not fully answered as the same")
	fmt.Println(ansMsg)
}

func TestLongMessaging(t *testing.T) {
	sender, _ := accounts.GenerateAccount()
	receiver, _ := accounts.GenerateAccount()

	//stripped receiver is the receiver account that everyone would see, and does not have a private key
	strippedReceiver := accounts.AccountFromPubBytes(receiver.PublicKey)
	testMessage := "Hello World!"
	for i := 0; i < 10; i++ {
		testMessage = testMessage + testMessage
	} //just change the test message to be *really* long
	msg, err := NewCaesarMessage(strippedReceiver, sender, testMessage)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, msg.Verify(), "could not verify to be true")
	assert.NotEqual(
		t,
		[]byte(testMessage),
		msg.Message,
		"looks like it isn't encrypting",
	)
	badMsg, wantedErr := msg.GetMessageString(strippedReceiver) //the stripped receiver, who doesn't have a private key, should not be able to get the message
	if wantedErr == nil {
		fmt.Println("error with bad account was not nil")
		t.Fail()
	}
	assert.NotEqual(t, testMessage, badMsg, "message was decrypted by stripped receiver (aka, anyone could)")

	ansMsg, err := msg.GetMessageString(receiver)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testMessage, ansMsg, "message not fully answered as the same")
}
