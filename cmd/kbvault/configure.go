package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func newConfigureCmd() *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure kbvault interactively",
		Long: `Configure kbvault profiles interactively, similar to AWS CLI.

This command will guide you through setting up a new profile or modifying
an existing one. It will prompt for storage type, credentials, and other
configuration options.

Examples:
  kbvault configure                    # Configure default profile
  kbvault configure --profile work     # Configure work profile
  kbvault configure --profile personal # Configure personal profile`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip initialization for configure command to avoid circular dependency
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigure(profileName)
		},
	}

	cmd.Flags().StringVar(&profileName, "profile", "", 
		"Profile name to configure (creates new if doesn't exist, default: active profile)")

	return cmd
}

func runConfigure(profileName string) error {
	pm, err := config.NewProfileManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	// Determine profile to configure
	if profileName == "" {
		profileName = pm.GetActiveProfile()
		fmt.Printf("Configuring profile: %s\n", profileName)
	} else {
		// Check if profile exists
		profiles, err := pm.ListProfiles()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		profileExists := false
		for _, p := range profiles {
			if p.Name == profileName {
				profileExists = true
				break
			}
		}

		if profileExists {
			fmt.Printf("Configuring existing profile: %s\n", profileName)
		} else {
			fmt.Printf("Creating new profile: %s\n", profileName)
		}
	}

	// Load existing configuration or create default
	var currentConfig *types.Config
	if profileExists(pm, profileName) {
		currentConfig, err = pm.GetConfig(profileName)
		if err != nil {
			return fmt.Errorf("failed to load current configuration: %w", err)
		}
	} else {
		currentConfig = types.DefaultConfig()
	}

	fmt.Println("\nPress Enter to accept default values shown in brackets.")
	fmt.Println()

	// Configure vault settings
	if err := configureVault(scanner, currentConfig); err != nil {
		return fmt.Errorf("failed to configure vault: %w", err)
	}

	// Configure storage
	if err := configureStorage(scanner, currentConfig); err != nil {
		return fmt.Errorf("failed to configure storage: %w", err)
	}

	// Configure server settings (optional)
	if confirmPrompt(scanner, "Configure HTTP server settings? (y/N)") {
		if err := configureServer(scanner, currentConfig); err != nil {
			return fmt.Errorf("failed to configure server: %w", err)
		}
	}

	// Save configuration
	if profileExists(pm, profileName) {
		err = pm.UpdateProfile(profileName, currentConfig)
	} else {
		err = pm.CreateProfile(profileName, &config.CreateProfileOptions{})
		if err == nil {
			err = pm.UpdateProfile(profileName, currentConfig)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("\nConfiguration saved for profile '%s'.\n", profileName)

	// Ask if user wants to make this the active profile
	if profileName != pm.GetActiveProfile() {
		if confirmPrompt(scanner, fmt.Sprintf("Set '%s' as the active profile? (y/N)", profileName)) {
			if err := pm.SwitchProfile(profileName); err != nil {
				return fmt.Errorf("failed to switch to profile: %w", err)
			}
			fmt.Printf("Switched to profile '%s'.\n", profileName)
		}
	}

	return nil
}

func configureVault(scanner *bufio.Scanner, config *types.Config) error {
	fmt.Println("Vault Configuration:")
	
	// Vault name
	config.Vault.Name = promptWithDefault(scanner, "Vault name", config.Vault.Name)
	
	// Notes directory
	config.Vault.NotesDir = promptWithDefault(scanner, "Notes directory", config.Vault.NotesDir)
	
	// Daily notes directory
	config.Vault.DailyDir = promptWithDefault(scanner, "Daily notes directory", config.Vault.DailyDir)
	
	// Templates directory
	config.Vault.TemplatesDir = promptWithDefault(scanner, "Templates directory", config.Vault.TemplatesDir)
	
	// Default template
	config.Vault.DefaultTemplate = promptWithDefault(scanner, "Default template", config.Vault.DefaultTemplate)

	fmt.Println()
	return nil
}

func configureStorage(scanner *bufio.Scanner, config *types.Config) error {
	fmt.Println("Storage Configuration:")
	
	// Storage type
	currentType := string(config.Storage.Type)
	for {
		storageType := promptWithDefault(scanner, "Storage type (local/s3)", currentType)
		if storageType == "local" || storageType == "s3" {
			config.Storage.Type = types.StorageType(storageType)
			break
		}
		fmt.Println("Please enter 'local' or 's3'")
	}

	// Configure based on storage type
	switch config.Storage.Type {
	case types.StorageTypeLocal:
		return configureLocalStorage(scanner, config)
	case types.StorageTypeS3:
		return configureS3Storage(scanner, config)
	}

	return nil
}

func configureLocalStorage(scanner *bufio.Scanner, config *types.Config) error {
	fmt.Println("\nLocal Storage Settings:")
	
	// Path
	config.Storage.Local.Path = promptWithDefault(scanner, "Vault path", config.Storage.Local.Path)
	
	// Create directories
	createDirs := promptBoolWithDefault(scanner, "Create directories automatically", config.Storage.Local.CreateDirs)
	config.Storage.Local.CreateDirs = createDirs
	
	// Enable locking
	enableLocking := promptBoolWithDefault(scanner, "Enable file locking", config.Storage.Local.EnableLocking)
	config.Storage.Local.EnableLocking = enableLocking

	fmt.Println()
	return nil
}

func configureS3Storage(scanner *bufio.Scanner, config *types.Config) error {
	fmt.Println("\nS3 Storage Settings:")
	
	// Bucket
	config.Storage.S3.Bucket = promptWithDefault(scanner, "S3 bucket name", config.Storage.S3.Bucket)
	
	// Region
	config.Storage.S3.Region = promptWithDefault(scanner, "S3 region", config.Storage.S3.Region)
	
	// Endpoint (optional)
	if confirmPrompt(scanner, "Use custom S3 endpoint? (y/N)") {
		config.Storage.S3.Endpoint = promptWithDefault(scanner, "S3 endpoint URL", config.Storage.S3.Endpoint)
	}
	
	// Prefix (optional)
	if confirmPrompt(scanner, "Use S3 key prefix? (y/N)") {
		config.Storage.S3.Prefix = promptWithDefault(scanner, "S3 key prefix", config.Storage.S3.Prefix)
	}

	// Encryption
	if confirmPrompt(scanner, "Enable server-side encryption? (y/N)") {
		return configureS3Encryption(scanner, config)
	}

	// Configure AWS credentials
	if confirmPrompt(scanner, "Configure AWS credentials? (y/N)") {
		return configureAWSCredentials(scanner, config)
	}

	fmt.Println()
	return nil
}

func configureS3Encryption(scanner *bufio.Scanner, config *types.Config) error {
	fmt.Println("\nS3 Encryption Settings:")
	
	// Encryption type
	fmt.Println("Server-side encryption options:")
	fmt.Println("  1. AES256")
	fmt.Println("  2. aws:kms")
	
	for {
		choice := promptWithDefault(scanner, "Encryption type (1/2)", "1")
		switch choice {
		case "1":
			config.Storage.S3.ServerSideEncryption = "AES256"
			return nil
		case "2":
			config.Storage.S3.ServerSideEncryption = "aws:kms"
			// Optional KMS key ID
			if confirmPrompt(scanner, "Specify KMS key ID? (y/N)") {
				config.Storage.S3.KMSKeyID = promptWithDefault(scanner, "KMS key ID", config.Storage.S3.KMSKeyID)
			}
			return nil
		default:
			fmt.Println("Please enter 1 or 2")
		}
	}
}

func configureAWSCredentials(scanner *bufio.Scanner, config *types.Config) error {
	fmt.Println("\nAWS Credentials:")
	fmt.Println("Note: It's recommended to use AWS credential files or IAM roles instead of hardcoding credentials.")
	
	if !confirmPrompt(scanner, "Store credentials in profile configuration? (y/N)") {
		fmt.Println("AWS credentials will be loaded from the standard credential chain (environment, files, IAM roles).")
		return nil
	}
	
	// Access Key ID
	config.Storage.S3.AccessKeyID = promptWithDefault(scanner, "AWS Access Key ID", config.Storage.S3.AccessKeyID)
	
	// Secret Access Key
	config.Storage.S3.SecretAccessKey = promptSensitive(scanner, "AWS Secret Access Key")
	
	// Session Token (optional)
	if confirmPrompt(scanner, "Using temporary credentials with session token? (y/N)") {
		config.Storage.S3.SessionToken = promptSensitive(scanner, "AWS Session Token")
	}

	fmt.Println()
	return nil
}

func configureServer(scanner *bufio.Scanner, config *types.Config) error {
	fmt.Println("\nHTTP Server Configuration:")
	
	// Enable server
	enabled := promptBoolWithDefault(scanner, "Enable HTTP server", config.Server.HTTP.Enabled)
	config.Server.HTTP.Enabled = enabled
	
	if !enabled {
		return nil
	}
	
	// Host
	config.Server.HTTP.Host = promptWithDefault(scanner, "Server host", config.Server.HTTP.Host)
	
	// Port
	portStr := promptWithDefault(scanner, "Server port", fmt.Sprintf("%d", config.Server.HTTP.Port))
	if port, err := strconv.Atoi(portStr); err == nil {
		config.Server.HTTP.Port = port
	}
	
	// CORS
	enableCORS := promptBoolWithDefault(scanner, "Enable CORS", config.Server.HTTP.EnableCORS)
	config.Server.HTTP.EnableCORS = enableCORS

	fmt.Println()
	return nil
}

// Helper functions

func profileExists(pm *config.ProfileManager, name string) bool {
	profiles, err := pm.ListProfiles()
	if err != nil {
		return false
	}
	
	for _, p := range profiles {
		if p.Name == name {
			return true
		}
	}
	return false
}

func promptWithDefault(scanner *bufio.Scanner, prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	
	if input == "" {
		return defaultValue
	}
	return input
}

func promptSensitive(scanner *bufio.Scanner, prompt string) string {
	fmt.Printf("%s: ", prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func promptBoolWithDefault(scanner *bufio.Scanner, prompt string, defaultValue bool) bool {
	defaultStr := "N"
	if defaultValue {
		defaultStr = "Y"
	}
	
	for {
		fmt.Printf("%s (y/N) [%s]: ", prompt, defaultStr)
		scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		
		if input == "" {
			return defaultValue
		}
		
		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter 'y' or 'n'")
		}
	}
}

func confirmPrompt(scanner *bufio.Scanner, prompt string) bool {
	fmt.Printf("%s ", prompt)
	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return input == "y" || input == "yes"
}