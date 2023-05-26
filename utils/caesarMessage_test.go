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
	testMessage := "Hello World!"
	msg, err := NewCaesarMessage(*receiver, *sender, testMessage)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, msg.Sign(), "error signing!")
	assert.True(t, msg.Verify(), "could not verify to be true")
	// assert.NotEqual(//TODO: uncomment this once encryption is working
	// 	t,
	// 	[]byte(testMessage),
	// 	msg.Message,
	// 	"looks like it isn't encrypting",
	// )
	fmt.Println(msg.Message)

}
