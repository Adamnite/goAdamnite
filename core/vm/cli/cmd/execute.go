package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"github.com/adamnite/go-adamnite/core/vm"
	"github.com/spf13/cobra"
)

var bytes string
var gas uint64
var funcHash string
var callArgs string
var filePath string
var stateless bool

func init() {
  executeCmd.Flags().StringVarP(&bytes, "from-bytes", "", "", "Bytes to execute")
  executeCmd.Flags().StringVarP(&filePath, "from-file", "", "", "Path of file containing bytes to execute")

  executeCmd.Flags().Uint64Var(&gas, "gas", 0, "Amount of gas to allocate for the execution")
  executeCmd.Flags().StringVarP(&funcHash, "function", "", "", "Hash of the function to execute")
  executeCmd.Flags().StringVarP(&callArgs, "call-args", "", "", "Wasm encoded arguments of the function")
  executeCmd.Flags().BoolVarP(&stateless, "stateless", "", true, "Whether to retrieve context from live blockchain. If true user has to provide block information")

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
	cfg.CodeGetter = spoofer.GetCode
	vMachine := VM.NewVirtualMachine([]byte{}, []uint64{0}, &cfg, gas)

	if (callArgs != "") {
		funcHash += callArgs
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
	if (stateless) {
		if (bytes != "") {
			executeStateless(bytes)
		} else if (filePath != "") {
			readBytes, err := os.ReadFile(filePath)
			if (err != nil) {
				panic(err)
			}
			executeStateless(string(readBytes))
		} else {
			panic("Bytes or file path should be specified")
		}
	} 
  },
}

