package cmd

import (
  "fmt"

  "github.com/spf13/cobra"
)

func init() {
  root.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number of Adamnite command line tool",
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Adamnite Command Line Tool v0.1")
  },
}