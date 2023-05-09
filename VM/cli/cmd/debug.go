package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/adamnite/go-adamnite/core/VM"
	"github.com/spf13/cobra"
)

func init() {

	debugCmd.Flags().StringVarP(&bytes, "from-bytes", "", "", "Bytes to execute")
	debugCmd.Flags().StringVarP(&filePath, "from-file", "", "", "Path of file containing bytes to execute")
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
	Short: "Print debugging information about a compiled contract.",
	Run: func(cmd *cobra.Command, args []string) {

		spoofer := VM.NewDBSpoofer()
		var bytes2 []byte
		var err interface{}

		if bytes != "" {
			bytes2, _ = hex.DecodeString(bytes)
		} else if filePath != "" {
			bytes2, err = os.ReadFile(filePath)
			if err != nil {
				panic(err)
			}
			bytes2, _ = hex.DecodeString(string(bytes2))
		} else {
			panic("Bytes or file path should be specified")
		}

		decodedModule := VM.DecodeModule(bytes2)
		err, hashes := spoofer.AddModuleToSpoofedCode(decodedModule)

		if err != nil {
			panic(err)
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
