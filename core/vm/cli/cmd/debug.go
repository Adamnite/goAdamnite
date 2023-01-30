package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
  root.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
  Use:   "debug",
  Short: "Print debugging information about a compiled contract.",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Print debugging info")
  },
}