package cmd

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/adamnite/go-adamnite/common"
	VM "github.com/adamnite/go-adamnite/core/vm"
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


func triggerUpload(bytes []byte) {
  value := big.NewInt(100)
  contract := VM.NewContract(common.Address{}, value, bytes, gas)
  err := VM.UploadContract(serverUrl, *contract)
  if (err != nil) {
    fmt.Println("Unable to upload specified contract")
    panic(err)
  } else {
    fmt.Print("Contract uploaded successfully")
  }
}

var uploadCmd = &cobra.Command{
  Use:   "upload",
  Short: "Upload the compiled contract.",
  Run: func(cmd *cobra.Command, args []string) {
    
    if (stateless) {
      if (bytes != "") {
        bytes, _ := hex.DecodeString(bytes)
        triggerUpload(bytes)
      } else if (filePath != "") {
        readBytes, err := os.ReadFile(filePath)
        if (err != nil) {
          panic(err)
        }
        triggerUpload(readBytes)
      } else {
        panic("Bytes or file path should be specified")
      }
    }
  },
}