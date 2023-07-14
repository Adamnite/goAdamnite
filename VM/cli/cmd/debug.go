package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/adamnite/go-adamnite/VM"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "output debug information about compiled contract",
	Run: func(cmd *cobra.Command, args []string) {
		if hexBytes == "" && filePath == "" {
			fmt.Println("Please, specify either hexadecimal bytes (--from-hex) or binary file path (--from-file)")
			return
		}

		if hexBytes != "" && filePath != "" {
			fmt.Println("Can't have both! Please, specify either hexadecimal bytes (--from-hex) or binary file path (--from-file)")
			return
		}

		spoofer := VM.NewDBSpoofer()

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

		decodedModule := VM.DecodeModule(rawBytes)
		hashes, err := spoofer.AddModuleToSpoofedCode(decodedModule)
		if err != nil {
			log.Fatal(err)
		}

		for _, h := range hashes {
			hexHash := hex.EncodeToString(h)
			types, _, _ := spoofer.GetCode(h)
			fmt.Print(hexHash, " ===> ")

			for _, r := range types.Results() {
				fmt.Printf("%s", TypeToString(r))
			}

			fmt.Print(" (")
			for idx, p := range types.Params() {
				fmt.Printf("%s", TypeToString(p))

				if idx < len(types.Params())-1 {
					fmt.Printf(", ")
				}
			}
			fmt.Print(")")
			fmt.Println()
		}
	},
}

func init() {
	debugCmd.Flags().StringVar(&hexBytes, "from-hex", "", "bytes in hexadecimal representation to execute")
	debugCmd.Flags().StringVar(&filePath, "from-file", "", "path to binary file to execute")

	rootCmd.AddCommand(debugCmd)
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
