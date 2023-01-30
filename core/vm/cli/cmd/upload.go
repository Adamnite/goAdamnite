package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
  root.AddCommand(uploadCmd)
}

var uploadCmd = &cobra.Command{
  Use:   "upload",
  Short: "Upload the compiled contract.",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Upload a function to chain")
  },
}