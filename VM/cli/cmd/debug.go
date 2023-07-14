package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/spf13/cobra"
)

func init() {
	debugCmd.Flags().StringVarP(&bytes, "from-hex", "", "", "bytes in hexadecimal representation to execute")
	debugCmd.Flags().StringVarP(&filePath, "from-file", "", "", "path to binary file to execute")
	root.AddCommand(debugCmd)
}

func TypeToString(t byte) string {
	switch t {
	case VM.Op_i32:
		return "i32"
	case VM.Op_i64:
		return "i64"
	case VM.Op_f32:
		return "f32"
	case VM.Op_f64:
		return "f64"
	}
	return "unknown"
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Output debug information about compiled contract.",
	Run: func(cmd *cobra.Command, args []string) {

		spoofer := VM.NewDBSpoofer()
		var rawBytes []byte
		var err error

		if bytes != "" {
			rawBytes, _ = hex.DecodeString(bytes)
		} else if filePath != "" {
			rawBytes, err = os.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("bytes or file path should be specified")
		}

		decodedModule := VM.DecodeModule(rawBytes)
		err, hashes := spoofer.AddModuleToSpoofedCode(decodedModule)

		if err != nil {
			log.Fatal(err)
		} else {
			for _, v := range hashes {
				hexV := hex.EncodeToString(v)
				types, _, _ := spoofer.GetCode(v)
				fmt.Print(hexV, " ===> ")
				for _, r := range types.Results() {
					fmt.Printf("%s", TypeToString(r))
				}

				fmt.Print(" (")
				for idx, p := range types.Params() {
					if idx == len(types.Params())-1 {
						fmt.Printf("%s)", TypeToString(p))
					} else {
						fmt.Printf("%s, ", TypeToString(p))
					}
				}
				fmt.Println()
			}
		}
	},
}
