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
	executeCmd.Flags().StringVarP(&callArgs, "call-args", "", "", "Wasm encoded arguments of the function")
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
		//remove any extra characters, and sanitize the inputs.
		callArgs = strings.ReplaceAll(callArgs, " ", "")
		callArgs = strings.ReplaceAll(callArgs, "[", "")
		callArgs = strings.ReplaceAll(callArgs, "]", "")

		//split by comma separation
		paramsSplit := strings.Split(callArgs, ",")
		callParamsHex := []byte{}
		for i, indexedTypeValue := range funcTypes.Params() {
			switch indexedTypeValue {
			case VM.Op_i32:
				valueOfParam, err := strconv.ParseInt(paramsSplit[i], 0, 32) //this will figure out the base
				if err != nil {
					panic(fmt.Errorf("error in parsing parameters, %w", err))
				}
				callParamsHex = append(callParamsHex, indexedTypeValue)

				callParamsHex = append(callParamsHex, VM.EncodeInt32(int32(valueOfParam))...)
			case VM.Op_i64:
				valueOfParam, err := strconv.ParseInt(paramsSplit[i], 0, 64)
				if err != nil {
					panic(fmt.Errorf("error in parsing parameters, %w", err))
				}
				callParamsHex = append(callParamsHex, indexedTypeValue)
				callParamsHex = append(callParamsHex, VM.EncodeInt64(valueOfParam)...)
			case VM.Op_f32:

				valueOfParam, err := strconv.ParseFloat(paramsSplit[i], 32)
				if err != nil {
					panic(fmt.Errorf("error in parsing parameters, %w", err))
				}
				callParamsHex = append(callParamsHex, indexedTypeValue)
				callParamsHex = append(callParamsHex, VM.EncodeUint32(math.Float32bits(float32(valueOfParam)))...)
			case VM.Op_f64:

				valueOfParam, err := strconv.ParseFloat(paramsSplit[i], 64)
				if err != nil {
					panic(fmt.Errorf("error in parsing parameters, %w", err))
				}
				callParamsHex = append(callParamsHex, indexedTypeValue)
				callParamsHex = append(callParamsHex, VM.EncodeUint64(math.Float64bits(valueOfParam))...)
			}
		}
		callEncodedString := strings.ReplaceAll( //sanitizing out any possible hex encoding fluff left over
			hex.EncodeToString(callParamsHex),
			"0x",
			"")
		//if any further cleanup needs to be done, it should be done here
		funcHash += callEncodedString
	}

	err := vMachine.Call2(funcHash, gas)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	vMachine.DumpStack()
	return
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
