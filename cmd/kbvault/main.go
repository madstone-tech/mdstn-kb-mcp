package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

var (
	// Build information set via ldflags
	version    = "dev"
	commitHash = "unknown"
	buildTime  = "unknown"

	// Global configuration and profile management
	profileManager *config.ProfileManager
	currentConfig  *types.Config
	currentProfile string
)

// GlobalFlags contains flags that are available to all commands
type GlobalFlags struct {
	Profile string
}

var globalFlags = &GlobalFlags{}

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
and includes CLI, TUI, HTTP API, and MCP interfaces.

Profile Support:
  Use --profile to specify which configuration profile to use.
  Profiles allow you to maintain separate configurations for different
  environments (work, personal, research, etc.).

Examples:
  kbvault --profile work search "project planning"
  kbvault --profile personal new "Weekend Ideas"
  kbvault profile list
  kbvault profile create work --storage-type s3 --s3-bucket my-work-kb`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commitHash, buildTime),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Show help if no subcommand is provided
			if err := cmd.Help(); err != nil {
				fmt.Fprintf(os.Stderr, "Error showing help: %v\n", err)
			}
		},
	}

	// Add global flags
	cmd.PersistentFlags().StringVar(&globalFlags.Profile, "profile", "", 
		"Configuration profile to use (default: active profile)")

	// Add subcommands
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newNewCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newSearchCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newDeleteCmd())
	cmd.AddCommand(newProfileCmd())
	cmd.AddCommand(newConfigureCmd())

	return cmd
}

// initializeConfig initializes the profile manager and resolves the configuration
func initializeConfig() error {
	var err error
	
	// Initialize profile manager
	profileManager, err = config.NewProfileManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	// Determine which profile to use
	profile := globalFlags.Profile
	if profile == "" {
		// Use active profile if no profile specified
		profile = profileManager.GetActiveProfile()
	} else {
		// Validate that the specified profile exists
		profiles, err := profileManager.ListProfiles()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}
		
		found := false
		for _, p := range profiles {
			if p.Name == profile {
				found = true
				break
			}
		}
		
		if !found {
			return fmt.Errorf("profile '%s' does not exist. Use 'kbvault profile list' to see available profiles", profile)
		}
	}

	// Load configuration for the resolved profile
	currentConfig, err = profileManager.GetConfig(profile)
	if err != nil {
		return fmt.Errorf("failed to load configuration for profile '%s': %w", profile, err)
	}

	currentProfile = profile
	return nil
}

// getConfig returns the current configuration
func getConfig() *types.Config {
	return currentConfig
}

// getProfile returns the current profile name
func getProfile() string {
	return currentProfile
}

// getProfileManager returns the profile manager instance
func getProfileManager() *config.ProfileManager {
	return profileManager
}
