package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func newInitCmd() *cobra.Command {
	var (
		vaultPath string
		vaultName string
		force     bool
	)

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new kbVault",
		Long: `Initialize a new kbVault in the specified directory.
Creates the necessary directory structure and configuration files.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine vault path
			if len(args) > 0 {
				vaultPath = args[0]
			} else if vaultPath == "" {
				var err error
				vaultPath, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(vaultPath)
			if err != nil {
				return fmt.Errorf("failed to resolve absolute path: %w", err)
			}
			vaultPath = absPath

			// Check if vault already exists
			configPath := filepath.Join(vaultPath, ".kbvault", "config.toml")
			if _, err := os.Stat(configPath); err == nil && !force {
				return fmt.Errorf("vault already exists at %s (use --force to overwrite)", vaultPath)
			}

			// Create vault directory structure
			if err := createVaultStructure(vaultPath); err != nil {
				return fmt.Errorf("failed to create vault structure: %w", err)
			}

			// Create default configuration
			if err := createDefaultConfig(vaultPath, vaultName); err != nil {
				return fmt.Errorf("failed to create configuration: %w", err)
			}

			fmt.Printf("✅ Initialized kbVault at: %s\n", vaultPath)
			fmt.Printf("📁 Configuration: %s\n", configPath)
			fmt.Printf("📝 Notes directory: %s\n", filepath.Join(vaultPath, "notes"))
			
			return nil
		},
	}

	cmd.Flags().StringVarP(&vaultPath, "path", "p", "", "Path to initialize the vault (default: current directory)")
	cmd.Flags().StringVarP(&vaultName, "name", "n", "", "Name for the vault (default: directory name)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force initialization even if vault exists")

	return cmd
}

func createVaultStructure(vaultPath string) error {
	// Create main directories
	dirs := []string{
		filepath.Join(vaultPath, ".kbvault"),
		filepath.Join(vaultPath, "notes"),
		filepath.Join(vaultPath, "templates"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create .gitignore if it doesn't exist
	gitignorePath := filepath.Join(vaultPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignoreContent := `# kbVault
.kbvault/cache/
.kbvault/locks/
*.log
`
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	return nil
}

func createDefaultConfig(vaultPath, vaultName string) error {
	// Use directory name if no vault name provided
	if vaultName == "" {
		vaultName = filepath.Base(vaultPath)
	}

	// Create default configuration
	cfg := types.DefaultConfig()
	cfg.Vault.Name = vaultName
	// Note: VaultConfig doesn't have a Path field

	// Save configuration
	manager := config.NewManager()
	configPath := filepath.Join(vaultPath, ".kbvault", "config.toml")
	
	return manager.SaveToFile(cfg, configPath)
}