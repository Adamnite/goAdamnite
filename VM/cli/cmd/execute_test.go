package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteStateless(t *testing.T) {
	readBytes := "0061736d0100000001070160027e7e017e03020100070a010661646454776f00000a09010700200020017c0b000a046e616d650203010000"
	paramTests := []string{
		"1,2",
		"4, -1",
		"0x1, 0b10",
		"0x4, -0b01",
	}
	paramAnswers := []string{
		"0 ::: 3\n",
		"0 ::: 3\n",
		"0 ::: 3\n",
		"0 ::: 3\n",
	}
	for i, indexArgs := range paramTests {
		callArgs = indexArgs
		gas = 10000
		funcHash = "9703bdb17a160ed80486a83aa3c413c1"
		returnStr := executeStateless(readBytes)
		assert.Equal(t, paramAnswers[i], returnStr, "error running tests with param:"+indexArgs)

	}

}
