package cmd

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/adamnite/go-adamnite/adm/adamnitedb/rawdb"
	"github.com/adamnite/go-adamnite/adm/adamnitedb/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/core/VM"
	"github.com/spf13/cobra"
)

var serverUrl string

func init() {

	uploadCmd.Flags().StringVarP(&bytes, "from-bytes", "", "", "Bytes to execute")
	uploadCmd.Flags().StringVarP(&filePath, "from-file", "", "", "Path of file containing bytes to execute")
	uploadCmd.Flags().StringVarP(&serverUrl, "db-host", "", "http://localhost:5000", "The database server where to upload the byte code")
	uploadCmd.Flags().Uint64Var(&gas, "gas", 0, "Amount of gas to allocate for the execution")
	uploadCmd.Flags().BoolVarP(&stateless, "stateless", "", true, "Whether to retrieve context from live blockchain. If true user has to provide block information")

	root.AddCommand(uploadCmd)
}

func triggerUpload(codeBytes []byte) bool {
	// uploads a contract to the local DB, returns true if successful
	db := rawdb.NewMemoryDB()
	callerAddress := common.BytesToAddress([]byte{0x00})
	state, err := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	if err != nil {
		fmt.Println(err)
	}
	state.CreateAccount(callerAddress)
	state.AddBalance(callerAddress, big.NewInt(1000000))

	vmConfig := VM.GetDefaultConfig()
	vmConfig.Uri = serverUrl

	vMachine := VM.NewVM(state,
		VM.NewBlockContext(
			callerAddress,
			gas,
			big.NewInt(0),
			big.NewInt(0),
			big.NewInt(0),
			big.NewInt(0),
		),
		VM.TxContext{},
		&vmConfig,
		nil)
	_, _, err = vMachine.Create(callerAddress, codeBytes, gas, big.NewInt(1))
	if err != nil {
		panic(err)
	}

	// contract := VM.NewContract(common.Address{}, value, bytes, gas)
	// err := VM.UploadContract(serverUrl, *contract)
	err = vMachine.UploadContract(serverUrl)
	if err != nil {
		fmt.Println("Unable to upload specified contract")
		panic(err)
	} else {
		fmt.Print("Contract uploaded successfully")
	}
	return err == nil
}

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload the compiled contract.",
	Run: func(cmd *cobra.Command, args []string) {

		if stateless {
			if bytes != "" {
				bytes, _ := hex.DecodeString(bytes)
				triggerUpload(bytes)
			} else if filePath != "" {
				readBytes, err := os.ReadFile(filePath)
				if err != nil {
					panic(err)
				}
				triggerUpload(readBytes)
			} else {
				panic("Bytes or file path should be specified")
			}
		}
	},
}
