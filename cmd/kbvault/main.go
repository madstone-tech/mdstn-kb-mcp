package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Build information set via ldflags
	version    = "dev"
	commitHash = "unknown"
	buildTime  = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kbvault",
		Short: "High-performance Go knowledge management tool",
		Long: `kbVault is a high-performance knowledge management system built in Go.
It supports multiple storage backends (local, S3), provides full-text search,
and includes CLI, TUI, HTTP API, and MCP interfaces.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commitHash, buildTime),
		Run: func(cmd *cobra.Command, args []string) {
			// Show help if no subcommand is provided
			if err := cmd.Help(); err != nil {
				fmt.Fprintf(os.Stderr, "Error showing help: %v\n", err)
			}
		},
	}

	// Add subcommands
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newNewCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newConfigCmd())

	return cmd
}
