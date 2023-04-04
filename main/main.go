// Copyright 2022 The Adamnite Authors
// This file is the main entry of the Adamnite engine. This is the sole main file, that boots everything.
//all CLI functions are within the CMD folder.

package main

import (
	"fmt"
	"os"

	"github.com/adamnite/go-adamnite/main/cmd/gnite/launcher"
)

func main() {
	if err := launcher.Launch(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
