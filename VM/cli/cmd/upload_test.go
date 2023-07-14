package cmd

import (
	"log"
	"testing"
	"encoding/hex"

	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	rawBytes, err := hex.DecodeString("0061736d0100000001070160027e7e017e03020100070a010661646454776f00000a09010700200020017c0b000a046e616d650203010000")
	if err != nil {
		log.Fatal(err)
	}

	dbHost = "http://0.0.0.0:5000/"
	gas = 1000

	assert.True(t, upload(rawBytes), "upload failed.")
}
