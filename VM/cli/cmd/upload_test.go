package cmd

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTriggerUpload(t *testing.T) {
	readBytes, _ := hex.DecodeString("0061736d0100000001070160027e7e017e03020100070a010661646454776f00000a09010700200020017c0b000a046e616d650203010000")

	serverUrl = "http://0.0.0.0:5000/"
	gas = 1000

	assert.True(t, triggerUpload(readBytes), "upload failed.")
	print() //idk why this is needed, but it was failing without it...
}
