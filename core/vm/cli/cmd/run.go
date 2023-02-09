package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
	
  root.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
  Use:   "run",
  Short: "Compile and execute the specified function from A1 wasm bytes.",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Run")
  },
}