package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/adamnite/go-adamnite/databaseDeprecated/rawdb"
	"github.com/adamnite/go-adamnite/databaseDeprecated/statedb"
	"github.com/adamnite/go-adamnite/common"
	"github.com/adamnite/go-adamnite/VM"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload the compiled contract",
	Run: func(cmd *cobra.Command, args []string) {
		if hexBytes == "" && filePath == "" {
			fmt.Println("Please, specify either hexadecimal bytes (--from-hex) or binary file path (--from-file)")
			return
		}

		if hexBytes != "" && filePath != "" {
			fmt.Println("Can't have both! Please, specify either hexadecimal bytes (--from-hex) or binary file path (--from-file)")
			return
		}

		if !stateless {
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

		if upload(rawBytes) {
			fmt.Println("Compiled smart contract uploaded successfully!")
		} else {
			fmt.Println("Problem occurred while uploading compiled smart contract!")
		}
	},
}

func init() {
	uploadCmd.Flags().StringVar(&hexBytes, "from-hex", "", "bytes in hexadecimal representation")
	uploadCmd.Flags().StringVar(&filePath, "from-file", "", "path to binary file")

	uploadCmd.Flags().StringVarP(&dbHost, "db-host", "d", "http://localhost:5000", "database server where smart contract will be uploaded")
	uploadCmd.Flags().Uint64VarP(&gas, "gas", "g", 0, "amount of gas to allocate for the execution")
	uploadCmd.Flags().BoolVar(&stateless, "stateless", true, "whether to retrieve context from live blockchain (if true user has to provide block information)")

	rootCmd.AddCommand(uploadCmd)
}

func upload(bytes []byte) bool {
	callerAddress := common.BytesToAddress([]byte{0x00})

	db := rawdb.NewMemoryDB()
	stateDB, err := statedb.New(common.Hash{}, statedb.NewDatabase(db))
	if err != nil {
		log.Fatal(err)
	}
	stateDB.CreateAccount(callerAddress)
	stateDB.AddBalance(callerAddress, big.NewInt(1000000))

	config := VM.GetDefaultConfig()
	config.Uri = dbHost

	vm := VM.NewVM(stateDB, &config, nil)
	_, _, err = vm.Create(callerAddress, bytes, gas, big.NewInt(1))
	if err != nil {
		log.Fatal(err)
	}

	// contract := vm.NewContract(common.Address{}, value, bytes, gas)
	// err := vm.UploadContract(dbHost, *contract)

	err = vm.UploadContract(dbHost)
	if err != nil {
		log.Fatal(err)
	}
	return true
}
