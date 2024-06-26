package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/adamnite/go-adamnite/databaseDeprecated/statedb"
	"github.com/adamnite/go-adamnite/params"
	"github.com/spf13/cobra"
)

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "parse A1 smart contract and execute the specified function",
	Run: func(cmd *cobra.Command, args []string) {
		if hexBytes == "" && filePath == "" {
			fmt.Println("Please, specify either hexadecimal bytes (--from-hex) or binary file path (--from-file)")
			return
		}

		if hexBytes != "" && filePath != "" {
			fmt.Println("Can't have both! Please, specify either hexadecimal bytes (--from-hex) or binary file path (--from-file)")
			return
		}

		if !testNet {
			return
		}

		var rawBytes []byte
		var err error

		if hexBytes != "" {
			rawBytes, err = hex.DecodeString(hexBytes)
			if err != nil {
				log.Fatal(err)
			}
		} else if filePath != "" {
			rawBytes, err = os.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
			}
		}

		executeStateless(rawBytes)
	},
}

func init() {
	executeCmd.Flags().StringVar(&hexBytes, "from-hex", "", "bytes in hexadecimal representation to execute")
	executeCmd.Flags().StringVar(&filePath, "from-file", "", "path to binary file to execute")

	executeCmd.Flags().Uint64VarP(&gas, "gas", "g", 0, "amount of gas to allocate for the execution")
	executeCmd.Flags().StringVarP(&functionHash, "function", "f", "", "hash of the function to be executed")
	executeCmd.Flags().StringVarP(&functionArgs, "args", "a", "", "comma separated function arguments")
	executeCmd.Flags().BoolVar(&testNet, "test", true, "use the test network (otherwise, main network will be used)")

	executeCmd.MarkFlagRequired("gas")
	executeCmd.MarkFlagRequired("function")

	rootCmd.AddCommand(executeCmd)
}

func executeStateless(bytes []byte) string {
	spoofer := VM.NewDBSpoofer()

	decodedModule := VM.DecodeModule(bytes)
	_, err := spoofer.AddModuleToSpoofedCode(decodedModule)
	if err != nil {
		log.Fatal(err)
	}

	var vmConfig VM.VMConfig
	vmConfig.CodeGetter = spoofer.GetCode

	vm := VM.NewVM(&statedb.StateDB{}, &vmConfig, &params.ChainConfig{})

	functionHashBytes, err := hex.DecodeString(functionHash)
	if err != nil {
		log.Fatal(err)
	}

	callHash := functionHash

	functionType, _, _ := vmConfig.CodeGetter(functionHashBytes)
	if functionArgs != "" {
		callHash += encodeFunctionArguments(functionArgs, functionType)
	}

	err = vm.Call2(callHash, gas)
	if err != nil {
		log.Fatal(err)
	}

	vm.DumpStack()
	return vm.OutputStack()
}

func encodeFunctionArguments(args string, functionType VM.FunctionType) string {
	// remove any extra characters, and sanitize the input arguments
	args = strings.ReplaceAll(args, " ", "")
	args = strings.ReplaceAll(args, "[", "")
	args = strings.ReplaceAll(args, "]", "")

	// split by comma separation
	paramsSplit := strings.Split(args, ",")
	params := []byte{}

	for i, paramType := range functionType.Params() {
		var param []byte

		switch paramType {
		case VM.Op_i32:
			//this will figure out the base
			value, err := strconv.ParseInt(paramsSplit[i], 0, 32)
			if err != nil {
				log.Fatal(err)
			}
			param = VM.EncodeInt32(int32(value))
		case VM.Op_i64:
			value, err := strconv.ParseInt(paramsSplit[i], 0, 64)
			if err != nil {
				log.Fatal(err)
			}
			param = VM.EncodeInt64(value)
		case VM.Op_f32:
			value, err := strconv.ParseFloat(paramsSplit[i], 32)
			if err != nil {
				log.Fatal(err)
			}
			param = VM.EncodeUint32(math.Float32bits(float32(value)))
		case VM.Op_f64:
			value, err := strconv.ParseFloat(paramsSplit[i], 64)
			if err != nil {
				log.Fatal(err)
			}
			// param = VM.LE.AppendUint64([]byte{}, math.Float64bits(value))
			VM.LE.PutUint64(param, math.Float64bits(value))
		default:
			log.Fatalf("Unknown parameter type: %v", paramType)
		}

		params = append(params, paramType)
		params = append(params, param...)
	}

	return strings.ReplaceAll(hex.EncodeToString(params), "0x", "")
}
