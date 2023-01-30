package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
  root.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
  Use:   "build",
  Short: "Build the specified A1 file. This will require to have the A1 compiler installed.",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Build ")
  },
}