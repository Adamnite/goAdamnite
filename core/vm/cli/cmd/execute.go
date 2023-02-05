package cmd

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	VM "github.com/adamnite/go-adamnite/core/vm"
	"github.com/adamnite/go-adamnite/params"
	"github.com/spf13/cobra"
)

var bytes string
var gas uint64
var funcHash string
var callArgs string
var filePath string
var testnet bool
var stateless bool

func init() {
	executeCmd.Flags().StringVarP(&bytes, "from-bytes", "", "", "Bytes to execute")
	executeCmd.Flags().StringVarP(&filePath, "from-file", "", "", "Path of file containing bytes to execute")

	executeCmd.Flags().Uint64Var(&gas, "gas", 0, "Amount of gas to allocate for the execution")
	executeCmd.Flags().StringVarP(&funcHash, "function", "", "", "Hash of the function to execute")
	executeCmd.Flags().StringVarP(&callArgs, "call-args", "", "", "comma separated arguments of the function")
	executeCmd.Flags().BoolVarP(&testnet, "testnet", "", true, "The network type to use (mainnet, testnet)")

	executeCmd.MarkFlagRequired("gas")
	executeCmd.MarkFlagRequired("function")
	root.AddCommand(executeCmd)
}

func executeStateless(inputbytes string) {
	spoofer := VM.NewDBSpoofer()
	bytes, _ := hex.DecodeString(inputbytes)
	decodedModule := VM.DecodeModule(bytes)
	spoofer.AddModuleToSpoofedCode(decodedModule)
	var cfg VM.VMConfig
	var chainCfg params.ChainConfig

	cfg.CodeGetter = spoofer.GetCode
	vMachine := VM.NewVM(&statedb.StateDB{}, VM.BlockContext{}, VM.TxContext{}, &cfg, &chainCfg)
	funcHashBytes, _ := hex.DecodeString(funcHash)
	funcTypes, _, _ := cfg.CodeGetter(funcHashBytes)

	if callArgs != "" {
		funcHash += userInputToFuncArgsStr(callArgs, funcTypes)
	}

	err := vMachine.Call2(funcHash, gas)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	vMachine.DumpStack()
	// valueOfParam, _ := strconv.ParseFloat("-1", 64)
	// fmt.Println("floatsBits is: ", math.Float64bits(valueOfParam))
}

func userInputToFuncArgsStr(passedArgs string, funcTypes VM.FunctionType) string {
	// remove any extra characters, and sanitize the inputs.
	passedArgs = strings.ReplaceAll(passedArgs, " ", "")
	passedArgs = strings.ReplaceAll(passedArgs, "[", "")
	passedArgs = strings.ReplaceAll(passedArgs, "]", "")

	// split by comma separation
	paramsSplit := strings.Split(passedArgs, ",")
	callParamsHex := []byte{}

	for i, indexedTypeValue := range funcTypes.Params() {
		var valueOfParam []byte
		var loopErr error = nil

		switch indexedTypeValue {
		case VM.Op_i32:
			//this will figure out the base
			if paramV, loopErr := strconv.ParseInt(paramsSplit[i], 0, 32); loopErr == nil {
				valueOfParam = VM.EncodeInt32(int32(paramV))
			}
		case VM.Op_i64:
			if paramV, loopErr := strconv.ParseInt(paramsSplit[i], 0, 64); loopErr == nil {
				valueOfParam = VM.EncodeInt64(paramV)
			}
		case VM.Op_f32:
			if paramV, loopErr := strconv.ParseFloat(paramsSplit[i], 32); loopErr == nil {
				valueOfParam = VM.EncodeUint32(math.Float32bits(float32(paramV)))
			}
		case VM.Op_f64:
			if paramV, loopErr := strconv.ParseFloat(paramsSplit[i], 64); loopErr == nil {
				valueOfParam = VM.LE.AppendUint64([]byte{}, math.Float64bits(paramV))
			}
		default:
			panic(fmt.Errorf("unrecognized type passed as func param type: %v", indexedTypeValue))
		}

		if loopErr != nil {
			panic(fmt.Errorf("error in parsing parameters, %w", loopErr))
		}
		callParamsHex = append(callParamsHex, indexedTypeValue)
		callParamsHex = append(callParamsHex, valueOfParam...)
	}

	return strings.ReplaceAll( //sanitizing out any possible hex encoding fluff left over
		hex.EncodeToString(callParamsHex),
		"0x",
		"") // if any further cleanup needs to be done, it should be done here
}

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Parse and execute the specified function from A1 wasm bytes.",
	Run: func(cmd *cobra.Command, args []string) {

		// The execution done here is stateless. For state depending execution,
		// the user has to provide a block from which the state will be retrieved from
		if testnet {
			if bytes != "" {
				executeStateless(bytes)
			} else if filePath != "" {
				readBytes, err := os.ReadFile(filePath)
				if err != nil {
					panic(err)
				}
				executeStateless(string(readBytes))
			} else {
				panic("Bytes or file path should be specified")
			}
		}
	},
}
