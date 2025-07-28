package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage configuration profiles",
		Long: `Manage configuration profiles for different environments.

Profiles allow you to maintain separate configurations for different
use cases such as work, personal, research, etc. Each profile can
have its own storage backend, vault settings, and other configurations.

Examples:
  kbvault profile list
  kbvault profile create work --storage-type s3 --s3-bucket my-work-kb
  kbvault profile switch work
  kbvault profile delete old-profile`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip initialization for profile commands to avoid circular dependency
			return nil
		},
	}

	cmd.AddCommand(newProfileListCmd())
	cmd.AddCommand(newProfileCreateCmd())
	cmd.AddCommand(newProfileDeleteCmd())
	cmd.AddCommand(newProfileSwitchCmd())
	cmd.AddCommand(newProfileShowCmd())
	cmd.AddCommand(newProfileCopyCmd())
	cmd.AddCommand(newProfileSetCmd())
	cmd.AddCommand(newProfileGetCmd())

	return cmd
}

func newProfileListCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available profiles",
		Long:  "List all available configuration profiles with their status and storage types.",
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			profiles, err := pm.ListProfiles()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			switch outputFormat {
			case "json":
				return printProfilesJSON(profiles)
			default:
				return printProfilesTable(profiles)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

func newProfileCreateCmd() *cobra.Command {
	var options config.CreateProfileOptions

	cmd := &cobra.Command{
		Use:   "create <profile-name>",
		Short: "Create a new profile",
		Long: `Create a new configuration profile.

You can specify storage type and related options. If no options are provided,
the profile will be created with default settings.

Examples:
  kbvault profile create work
  kbvault profile create work --storage-type s3 --s3-bucket my-work-kb --s3-region us-east-1
  kbvault profile create personal --storage-type local --local-path ~/personal-vault`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			if err := pm.CreateProfile(profileName, &options); err != nil {
				return fmt.Errorf("failed to create profile: %w", err)
			}

			fmt.Printf("Profile '%s' created successfully.\n", profileName)
			
			// Ask if user wants to switch to the new profile
			fmt.Printf("Switch to profile '%s'? (y/N): ", profileName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(strings.TrimSpace(response)) == "y" {
				if err := pm.SwitchProfile(profileName); err != nil {
					return fmt.Errorf("failed to switch to profile: %w", err)
				}
				fmt.Printf("Switched to profile '%s'.\n", profileName)
			}

			return nil
		},
	}

	// Storage type flags
	cmd.Flags().Var((*storageTypeValue)(&options.StorageType), "storage-type", "Storage backend type (local, s3)")
	
	// Local storage flags
	cmd.Flags().StringVar(&options.LocalPath, "local-path", "", "Path for local storage")
	
	// S3 storage flags
	cmd.Flags().StringVar(&options.S3Bucket, "s3-bucket", "", "S3 bucket name")
	cmd.Flags().StringVar(&options.S3Region, "s3-region", "", "S3 region")
	cmd.Flags().StringVar(&options.S3Endpoint, "s3-endpoint", "", "S3 endpoint URL (for S3-compatible services)")
	
	// Vault flags
	cmd.Flags().StringVar(&options.VaultName, "vault-name", "", "Name for the vault")
	cmd.Flags().StringVar(&options.Description, "description", "", "Description for the profile")

	return cmd
}

func newProfileDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <profile-name>",
		Short: "Delete a profile",
		Long:  "Delete a configuration profile. The default profile cannot be deleted.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			// Confirm deletion unless --force is used
			if !force {
				fmt.Printf("Are you sure you want to delete profile '%s'? (y/N): ", profileName)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(strings.TrimSpace(response)) != "y" {
					fmt.Println("Profile deletion cancelled.")
					return nil
				}
			}

			if err := pm.DeleteProfile(profileName); err != nil {
				return fmt.Errorf("failed to delete profile: %w", err)
			}

			fmt.Printf("Profile '%s' deleted successfully.\n", profileName)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete without confirmation")

	return cmd
}

func newProfileSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch <profile-name>",
		Short: "Switch to a different profile",
		Long:  "Switch the active profile to the specified profile.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			if err := pm.SwitchProfile(profileName); err != nil {
				return fmt.Errorf("failed to switch profile: %w", err)
			}

			fmt.Printf("Switched to profile '%s'.\n", profileName)
			return nil
		},
	}

	return cmd
}

func newProfileShowCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show [profile-name]",
		Short: "Show profile configuration",
		Long:  "Show the configuration for a specific profile. If no profile is specified, shows the active profile.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			profileName := ""
			if len(args) > 0 {
				profileName = args[0]
			} else {
				profileName = pm.GetActiveProfile()
			}

			config, err := pm.GetConfig(profileName)
			if err != nil {
				return fmt.Errorf("failed to get profile configuration: %w", err)
			}

			switch outputFormat {
			case "json":
				return printConfigJSON(config)
			default:
				return printConfigTable(profileName, config)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

func newProfileCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy <source-profile> <target-profile>",
		Short: "Copy a profile",
		Long:  "Create a new profile by copying the configuration from an existing profile.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceProfile := args[0]
			targetProfile := args[1]

			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			if err := pm.CopyProfile(sourceProfile, targetProfile); err != nil {
				return fmt.Errorf("failed to copy profile: %w", err)
			}

			fmt.Printf("Profile '%s' copied to '%s' successfully.\n", sourceProfile, targetProfile)
			return nil
		},
	}

	return cmd
}

func newProfileSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <profile-name> <key> <value>",
		Short: "Set a configuration value for a profile",
		Long: `Set a specific configuration value for a profile using dot notation.

Examples:
  kbvault profile set work vault.name "Work Vault"
  kbvault profile set work storage.type s3
  kbvault profile set work storage.s3.bucket my-work-bucket
  kbvault profile set work storage.s3.region us-east-1`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			key := args[1]
			value := args[2]

			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			if err := pm.SetProfileValue(profileName, key, value); err != nil {
				return fmt.Errorf("failed to set configuration value: %w", err)
			}

			fmt.Printf("Set %s.%s = %s\n", profileName, key, value)
			return nil
		},
	}

	return cmd
}

func newProfileGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <profile-name> <key>",
		Short: "Get a configuration value from a profile",
		Long: `Get a specific configuration value from a profile using dot notation.

Examples:
  kbvault profile get work vault.name
  kbvault profile get work storage.type
  kbvault profile get work storage.s3.bucket`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			key := args[1]

			pm, err := config.NewProfileManager()
			if err != nil {
				return fmt.Errorf("failed to initialize profile manager: %w", err)
			}

			value, err := pm.GetProfileValue(profileName, key)
			if err != nil {
				return fmt.Errorf("failed to get configuration value: %w", err)
			}

			fmt.Printf("%s\n", value)
			return nil
		},
	}

	return cmd
}

// Custom flag type for storage type
type storageTypeValue types.StorageType

func (s *storageTypeValue) String() string {
	return string(*s)
}

func (s *storageTypeValue) Set(v string) error {
	switch v {
	case "local", "s3":
		*s = storageTypeValue(v)
		return nil
	default:
		return fmt.Errorf("invalid storage type %q (must be local or s3)", v)
	}
}

func (s *storageTypeValue) Type() string {
	return "storage-type"
}

// Helper functions for output formatting

func printProfilesTable(profiles []config.ProfileInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "NAME\tACTIVE\tSTORAGE\tDEFAULT")
	for _, profile := range profiles {
		active := ""
		if profile.IsActive {
			active = "*"
		}
		
		defaultFlag := ""
		if profile.IsDefault {
			defaultFlag = "default"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", 
			profile.Name, active, profile.StorageType, defaultFlag)
	}

	return nil
}

func printProfilesJSON(profiles []config.ProfileInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(profiles)
}

func printConfigTable(profileName string, config *types.Config) error {
	fmt.Printf("Profile: %s\n\n", profileName)
	
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "SECTION\tKEY\tVALUE")
	
	// Vault configuration
	fmt.Fprintf(w, "vault\tname\t%s\n", config.Vault.Name)
	fmt.Fprintf(w, "vault\tnotes_dir\t%s\n", config.Vault.NotesDir)
	fmt.Fprintf(w, "vault\tdaily_dir\t%s\n", config.Vault.DailyDir)
	fmt.Fprintf(w, "vault\ttemplates_dir\t%s\n", config.Vault.TemplatesDir)
	
	// Storage configuration
	fmt.Fprintf(w, "storage\ttype\t%s\n", config.Storage.Type)
	
	switch config.Storage.Type {
	case types.StorageTypeLocal:
		fmt.Fprintf(w, "storage.local\tpath\t%s\n", config.Storage.Local.Path)
		fmt.Fprintf(w, "storage.local\tcreate_dirs\t%t\n", config.Storage.Local.CreateDirs)
	case types.StorageTypeS3:
		fmt.Fprintf(w, "storage.s3\tbucket\t%s\n", config.Storage.S3.Bucket)
		fmt.Fprintf(w, "storage.s3\tregion\t%s\n", config.Storage.S3.Region)
		if config.Storage.S3.Endpoint != "" {
			fmt.Fprintf(w, "storage.s3\tendpoint\t%s\n", config.Storage.S3.Endpoint)
		}
	}
	
	// Server configuration
	fmt.Fprintf(w, "server.http\tenabled\t%t\n", config.Server.HTTP.Enabled)
	fmt.Fprintf(w, "server.http\thost\t%s\n", config.Server.HTTP.Host)
	fmt.Fprintf(w, "server.http\tport\t%d\n", config.Server.HTTP.Port)

	return nil
}

func printConfigJSON(config *types.Config) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}