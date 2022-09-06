package vm

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_parseString(t *testing.T) {
	testStrings := []string{
		"",
		"42 00",
		"42 00 42 01",
	}
	expectedAnswers := [][]OperationCommon{
		{},
		{i64Const{00}},
		{i64Const{00}, i64Const{01}},
	}
	for i := 0; i < len(expectedAnswers); i++ {
		parse := parseString(testStrings[i])
		if len(parse) != len(expectedAnswers[i]) {
			println("wrong length of answer")
			print("got: ")
			print(len(parse))
			print(" expected: ")
			println(len(expectedAnswers[i]))

			fmt.Println(parse)
			fmt.Println(expectedAnswers[i])
			t.Fail()
		}
		if !expectedMatchResults(expectedAnswers[i], parse) {
			t.Fail()
		}
	}
}
func Test_parseBytes(t *testing.T) {
	wasmMagic := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}
	testBytes := [][]byte{
		{},
		append(wasmMagic, 0x42, 0x00),
		append(wasmMagic, 0x42, 0x00, 0x42, 0x01),
	}
	expectedAnswers := [][]OperationCommon{
		{},
		{i64Const{00}},
		{i64Const{00}, i64Const{01}},
	}
	for i := 0; i < len(expectedAnswers); i++ {
		parsed := parseBytes(testBytes[i])
		if len(parsed) != len(expectedAnswers[i]) {
			println("wrong length of answer")
			print("got: ")
			print(len(parsed))
			print(" expected: ")
			println(len(expectedAnswers[i]))
			t.Fail()
		}
		if !expectedMatchResults(expectedAnswers[i], parsed) {
			t.Fail()
		}
	}
}

func expectedMatchResults(expect []OperationCommon, ans []OperationCommon) bool {
	success := true
	for i, a := range ans {
		b := expect[i]
		if reflect.ValueOf(a.doOp).Pointer() != reflect.ValueOf(b.doOp).Pointer() {
			print("error occurred, values do not match at i:")
			println(i)
			success = false
		}
	}
	return success
}
