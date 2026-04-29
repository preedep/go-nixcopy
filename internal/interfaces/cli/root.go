package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "nixcopy",
	Short: "Fast file transfer CLI tool",
	Long: `go-nixcopy is a high-speed CLI tool for file transfer.
Supports transfers between SFTP, FTPS, Azure Blob Storage, and AWS S3
using streaming to minimize memory usage.`,
}

// SetBuildInfo wires version metadata injected by GoReleaser into the CLI version string.
func SetBuildInfo(version, buildTime, gitCommit string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, gitCommit, buildTime)
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
