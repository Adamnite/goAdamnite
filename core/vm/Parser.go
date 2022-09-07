package vm

import (
	"encoding/hex"
	"reflect"
	"strings"
)

const wasmMagic = "0061736D01000000"

func parseString(opString string) []OperationCommon {
	//sanitize the input of all possible special characters. Mostly used for tests
	s := strings.ReplaceAll(opString, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	if len(s) <= len(wasmMagic) {
		s = wasmMagic + s
	}
	if s[0:16] != wasmMagic {
		s = wasmMagic + s //if it doesn't have the wasm magic, add it
	}

	ansBytes, err := hex.DecodeString(s)
	if err != nil {
		println("error parsing string to bytes")
		println(err.Error())
		panic(err)
	}
	return parseBytes(ansBytes)
}

func parseBytes(bytes []byte) []OperationCommon {
	ansOps := []OperationCommon{}
	pointInBytes := 8
	//should check that the first 8 bytes are indeed the "wasm binary magic" and version number
	// println(hex.EncodeToString(bytes))
	for pointInBytes < len(bytes) {
		switch bytes[pointInBytes] {
		case Op_i64_const:
			var op = i64Const{int64(bytes[pointInBytes+1])}

			ansOps = append(ansOps, op)
			pointInBytes += 2
		case Op_i64_add:
			ansOps = append(ansOps, i64Add{})
			pointInBytes += 1
		case Op_i64_sub:
			ansOps = append(ansOps, i64Sub{})
			pointInBytes += 1
		case Op_i64_mul:
			ansOps = append(ansOps, i64Mul{})
			pointInBytes += 1
		case Op_if:
			//has a param block type, im not sure what its used for, so lets ignore that...
			//the rest of the conditional statements code must be filled at the end
			ansOps = append(ansOps, opIf{0, 0})
			pointInBytes += 2
		case Op_else:
			//have to look back for the last if statement that does not yet have an end or an else statement
			foundIf := false
			for i := len(ansOps) - 1; i >= 0 && !foundIf; i-- {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opIf{}) {
					foo := ansOps[i].(opIf)
					if foo.elsePoint == 0 && foo.endPoint == 0 {
						foo.elsePoint = int64(len(ansOps))
						ansOps[i] = foo
						foundIf = true
					}
				}
			}
			pointInBytes += 1
		case Op_end:
			//look for last condition statement
			//TODO: add check for loop, and block statements
			foundConditional := false
			for i := len(ansOps) - 1; i >= 0 && !foundConditional; i-- {
				if reflect.TypeOf(ansOps[i]) == reflect.TypeOf(opIf{}) {
					foo := ansOps[i].(opIf)
					if foo.endPoint == 0 {
						foo.endPoint = int64(len(ansOps))
						ansOps[i] = foo
						foundConditional = true
					}
				}
			}
			pointInBytes += 1

		default:
			print("skipping over byte at: ")
			println(pointInBytes)
			print("with value: ")
			println(hex.EncodeToString(bytes[pointInBytes : pointInBytes+1]))
			pointInBytes += 1
		}

	}

	return ansOps
}
