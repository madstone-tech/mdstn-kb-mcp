package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long: `Manage kbVault configuration settings.
Allows viewing, setting, and validating configuration options.`,
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigSetCmd())
	cmd.AddCommand(newConfigValidateCmd())
	cmd.AddCommand(newConfigPathCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var (
		key    string
		format string
	)

	cmd := &cobra.Command{
		Use:   "show [key]",
		Short: "Show configuration values",
		Long: `Show configuration values. If no key is specified, shows all configuration.
If a key is specified, shows only that configuration value.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile-aware configuration
			cfg := getConfig()
			if cfg == nil {
				return fmt.Errorf("configuration not initialized")
			}

			// Show specific key or all config
			if len(args) > 0 {
				key = args[0]
				return showConfigKey(cfg, key)
			}

			return showAllConfig(cfg, format)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "Output format (yaml, json)")

	return cmd
}

func newConfigSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set configuration value",
		Long: `Set a configuration value and save it to the configuration file.
Supports nested keys using dot notation (e.g., server.http.port).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			// Get profile-aware configuration
			cfg := getConfig()
			if cfg == nil {
				return fmt.Errorf("configuration not initialized")
			}

			// Set the value
			if err := setConfigValue(cfg, key, value); err != nil {
				return fmt.Errorf("failed to set config value: %w", err)
			}

			// Save configuration to profile
			profileManager := getProfileManager()
			currentProfile := getProfile()
			
			if err := profileManager.UpdateProfile(currentProfile, cfg); err != nil {
				return fmt.Errorf("failed to save configuration to profile '%s': %w", currentProfile, err)
			}

			fmt.Printf("✅ Set %s = %s\n", key, value)
			return nil
		},
	}

	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Long:  `Validate the current configuration for correctness and completeness.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile-aware configuration
			cfg := getConfig()
			if cfg == nil {
				return fmt.Errorf("configuration not initialized")
			}

			// Validate configuration
			if err := cfg.Validate(); err != nil {
				fmt.Printf("❌ Configuration validation failed:\n%v\n", err)
				return err
			}

			fmt.Println("✅ Configuration is valid")
			return nil
		},
	}

	return cmd
}

func newConfigPathCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  `Show the path to the current configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			currentProfile := getProfile()
			
			// Check if we're using local vault configuration
			if currentProfile == "local" {
				// Find local vault config path
				vaultRoot, err := findVaultRoot()
				if err == nil {
					configPath := filepath.Join(vaultRoot, ".kbvault", "config.toml")
					fmt.Println(configPath)
					return nil
				}
			}
			
			// Show profile configuration path
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			configPath := filepath.Join(homeDir, ".kbvault", "profiles", currentProfile+".toml")
			fmt.Printf("Profile: %s\nConfiguration file: %s\n", currentProfile, configPath)
			return nil
		},
	}

	return cmd
}

func showAllConfig(config *types.Config, format string) error {
	switch format {
	case "json":
		// Simplified JSON output
		fmt.Printf(`{
  "vault": {
    "name": "%s",
    "notes_dir": "%s",
    "max_file_size": %d
  },
  "storage": {
    "type": "%s",
    "local": {
      "path": "%s",
      "lock_timeout": %d
    }
  },
  "server": {
    "http": {
      "host": "%s",
      "port": %d,
      "read_timeout": %d
    }
  },
  "logging": {
    "level": "%s",
    "output": "%s"
  }
}
`,
			config.Vault.Name,
			config.Vault.NotesDir,
			config.Vault.MaxFileSize,
			config.Storage.Type,
			config.Storage.Local.Path,
			config.Storage.Local.LockTimeout,
			config.Server.HTTP.Host,
			config.Server.HTTP.Port,
			config.Server.HTTP.ReadTimeout,
			config.Logging.Level,
			config.Logging.Output,
		)
	default:
		// YAML-like output
		fmt.Printf(`vault:
  name: %s
  notes_dir: %s
  max_file_size: %d

storage:
  type: %s
  local:
    path: %s
    lock_timeout: %d

server:
  http:
    host: %s
    port: %d
    read_timeout: %d

logging:
  level: %s
  output: %s
`,
			config.Vault.Name,
			config.Vault.NotesDir,
			config.Vault.MaxFileSize,
			config.Storage.Type,
			config.Storage.Local.Path,
			config.Storage.Local.LockTimeout,
			config.Server.HTTP.Host,
			config.Server.HTTP.Port,
			config.Server.HTTP.ReadTimeout,
			config.Logging.Level,
			config.Logging.Output,
		)
	}

	return nil
}

func showConfigKey(config *types.Config, key string) error {
	// Handle nested keys with dot notation
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "vault":
		if len(parts) == 1 {
			fmt.Printf("name: %s\nnotes_dir: %s\nmax_file_size: %d\n", config.Vault.Name, config.Vault.NotesDir, config.Vault.MaxFileSize)
		} else if parts[1] == "name" {
			fmt.Println(config.Vault.Name)
		} else if parts[1] == "notes_dir" {
			fmt.Println(config.Vault.NotesDir)
		} else if parts[1] == "max_file_size" {
			fmt.Println(config.Vault.MaxFileSize)
		} else {
			return fmt.Errorf("unknown vault key: %s", parts[1])
		}
	case "storage":
		if len(parts) == 1 {
			fmt.Printf("type: %s\n", config.Storage.Type)
		} else if parts[1] == "type" {
			fmt.Println(config.Storage.Type)
		} else if len(parts) >= 3 && parts[1] == "local" {
			switch parts[2] {
			case "path":
				fmt.Println(config.Storage.Local.Path)
			case "lock_timeout":
				fmt.Println(config.Storage.Local.LockTimeout)
			default:
				return fmt.Errorf("unknown storage.local key: %s", parts[2])
			}
		} else if len(parts) >= 3 && parts[1] == "s3" {
			switch parts[2] {
			case "bucket":
				fmt.Println(config.Storage.S3.Bucket)
			case "region":
				fmt.Println(config.Storage.S3.Region)
			case "endpoint":
				fmt.Println(config.Storage.S3.Endpoint)
			default:
				return fmt.Errorf("unknown storage.s3 key: %s", parts[2])
			}
		} else {
			return fmt.Errorf("unknown storage key: %s", parts[1])
		}
	case "server":
		if len(parts) >= 3 && parts[1] == "http" {
			switch parts[2] {
			case "host":
				fmt.Println(config.Server.HTTP.Host)
			case "port":
				fmt.Println(config.Server.HTTP.Port)
			case "read_timeout":
				fmt.Println(config.Server.HTTP.ReadTimeout)
			default:
				return fmt.Errorf("unknown server.http key: %s", parts[2])
			}
		} else {
			return fmt.Errorf("unknown server key: %s", strings.Join(parts[1:], "."))
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

func setConfigValue(config *types.Config, key, value string) error {
	// Handle nested keys with dot notation
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "vault":
		if len(parts) == 2 && parts[1] == "name" {
			config.Vault.Name = value
		} else if len(parts) == 2 && parts[1] == "notes_dir" {
			config.Vault.NotesDir = value
		} else {
			return fmt.Errorf("unknown vault key: %s", strings.Join(parts[1:], "."))
		}
	case "storage":
		if len(parts) == 2 && parts[1] == "type" {
			config.Storage.Type = types.StorageType(value)
		} else {
			return fmt.Errorf("unknown storage key: %s", strings.Join(parts[1:], "."))
		}
	case "server":
		if len(parts) >= 3 && parts[1] == "http" {
			switch parts[2] {
			case "host":
				config.Server.HTTP.Host = value
			default:
				return fmt.Errorf("setting %s not supported yet", key)
			}
		} else {
			return fmt.Errorf("unknown server key: %s", strings.Join(parts[1:], "."))
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}
