package utils

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/adamnite/go-adamnite/node"
	"github.com/urfave/cli/v2"
)

func Fatalf(format string, args ...interface{}) {
	w := io.MultiWriter(os.Stdout, os.Stderr)
	if runtime.GOOS == "windows" {
		// The SameFile check below doesn't work on Windows.
		// stdout is unlikely to get redirected though, so just print there.
		w = os.Stdout
	} else {
		outf, _ := os.Stdout.Stat()
		errf, _ := os.Stderr.Stat()
		if outf != nil && errf != nil && os.SameFile(outf, errf) {
			w = os.Stderr
		}
	}
	fmt.Fprintf(w, "Fatal: "+format+"\n", args...)
	os.Exit(1)
}

func StartNode(ctx *cli.Context, node *node.Node) {
	if err := node.Start(); err != nil {
		Fatalf("Error starting protocol node: %v", err)
	}

	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

		defer signal.Stop(sigc)

		<-sigc
		log.Info("Adamnite engine shutting down...")
		go node.Close()

		// temp time
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Warn("Already shutting down, interrupt times", "times", i-1)
			}
		}

		// debug.LoudPanic("boom")
	}()
}
