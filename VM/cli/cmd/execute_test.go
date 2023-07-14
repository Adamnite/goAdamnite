package cmd

import (
	"log"
	"testing"
	"encoding/hex"

	"github.com/stretchr/testify/assert"
)

func TestExecuteStateless(t *testing.T) {
	rawBytes, err := hex.DecodeString("0061736d0100000001070160027e7e017e03020100070a010661646454776f00000a09010700200020017c0b000a046e616d650203010000")
	if err != nil {
		log.Fatal(err)
	}

	testParams := []string{
		"1,2",
		"4, -1",
		"0x1, 0b10",
		"0x4, -0b01",
	}
	testResults := []string{
		"0 ::: 3\n",
		"0 ::: 3\n",
		"0 ::: 3\n",
		"0 ::: 3\n",
	}

	functionHash = "9703bdb17a160ed80486a83aa3c413c1"
	gas = 10000

	for i, args := range testParams {
		functionArgs = args
		assert.Equal(t, testResults[i], executeStateless(rawBytes), "error running tests with param:"+args)
	}
}
