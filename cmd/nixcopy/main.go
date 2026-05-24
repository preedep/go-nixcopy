package main

import (
	"fmt"
	"os"

	"github.com/preedep/go-nixcopy/internal/interfaces/cli"
)

// Injected at build time by GoReleaser via -ldflags.
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	cli.SetBuildInfo(Version, BuildTime, GitCommit)
	if err := cli.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
