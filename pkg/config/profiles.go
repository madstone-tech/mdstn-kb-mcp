package config

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// ProfileInfo contains metadata about a profile
type ProfileInfo struct {
	Name        string    `json:"name"`
	IsActive    bool      `json:"is_active"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	ModifiedAt  time.Time `json:"modified_at,omitempty"`
	StorageType string    `json:"storage_type"`
	Description string    `json:"description,omitempty"`
}

// ProfileManager provides high-level profile management operations
type ProfileManager struct {
	viperManager *ViperManager
}

// NewProfileManager creates a new profile manager
func NewProfileManager() (*ProfileManager, error) {
	vm, err := NewViperManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create viper manager: %w", err)
	}

	return &ProfileManager{
		viperManager: vm,
	}, nil
}

// CreateProfile creates a new profile with optional configuration
func (pm *ProfileManager) CreateProfile(name string, options *CreateProfileOptions) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	// Check if profile already exists
	profiles, err := pm.viperManager.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	for _, profile := range profiles {
		if profile == name {
			return fmt.Errorf("profile %s already exists", name)
		}
	}

	// Always start with default configuration to ensure all fields are populated
	config := types.DefaultConfig()

	if options != nil {
		if options.StorageType != "" {
			config.Storage.Type = options.StorageType
		}

		// Configure storage based on type
		switch options.StorageType {
		case types.StorageTypeS3:
			if options.S3Bucket != "" {
				config.Storage.S3.Bucket = options.S3Bucket
			}
			if options.S3Region != "" {
				config.Storage.S3.Region = options.S3Region
			}
			if options.S3Endpoint != "" {
				config.Storage.S3.Endpoint = options.S3Endpoint
			}
		case types.StorageTypeLocal:
			if options.LocalPath != "" {
				config.Storage.Local.Path = options.LocalPath
			}
		}

		// Configure vault settings
		if options.VaultName != "" {
			config.Vault.Name = options.VaultName
		}
	}

	// Create the profile with the complete configuration
	if err := pm.viperManager.CreateProfile(name, config); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

// DeleteProfile removes a profile
func (pm *ProfileManager) DeleteProfile(name string) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	if name == "default" {
		return fmt.Errorf("cannot delete the default profile")
	}

	return pm.viperManager.DeleteProfile(name)
}

// ListProfiles returns information about all profiles
func (pm *ProfileManager) ListProfiles() ([]ProfileInfo, error) {
	profileNames, err := pm.viperManager.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	activeProfile := pm.viperManager.GetActiveProfile()
	var profiles []ProfileInfo

	for _, name := range profileNames {
		info := ProfileInfo{
			Name:      name,
			IsActive:  name == activeProfile,
			IsDefault: name == "default",
		}

		// Get storage type from profile config
		if config, err := pm.viperManager.GetConfig(name); err == nil {
			info.StorageType = string(config.Storage.Type)
		}

		profiles = append(profiles, info)
	}

	// Sort profiles: default first, then active, then alphabetically
	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].IsDefault {
			return true
		}
		if profiles[j].IsDefault {
			return false
		}
		if profiles[i].IsActive {
			return true
		}
		if profiles[j].IsActive {
			return false
		}
		return profiles[i].Name < profiles[j].Name
	})

	return profiles, nil
}

// GetProfile returns information about a specific profile
func (pm *ProfileManager) GetProfile(name string) (*ProfileInfo, error) {
	profiles, err := pm.ListProfiles()
	if err != nil {
		return nil, err
	}

	for _, profile := range profiles {
		if profile.Name == name {
			return &profile, nil
		}
	}

	return nil, fmt.Errorf("profile %s not found", name)
}

// SwitchProfile changes the active profile
func (pm *ProfileManager) SwitchProfile(name string) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	return pm.viperManager.SetActiveProfile(name)
}

// GetActiveProfile returns the name of the currently active profile
func (pm *ProfileManager) GetActiveProfile() string {
	return pm.viperManager.GetActiveProfile()
}

// GetConfig returns the configuration for a specific profile
func (pm *ProfileManager) GetConfig(profile string) (*types.Config, error) {
	return pm.viperManager.GetConfig(profile)
}

// UpdateProfile updates the configuration for a specific profile
func (pm *ProfileManager) UpdateProfile(name string, config *types.Config) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	// Validate the configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid profile configuration: %w", err)
	}

	return pm.viperManager.SaveProfile(name, config)
}

// CopyProfile creates a new profile by copying an existing one
func (pm *ProfileManager) CopyProfile(sourceName, targetName string) error {
	if err := validateProfileName(sourceName); err != nil {
		return fmt.Errorf("invalid source profile name: %w", err)
	}
	if err := validateProfileName(targetName); err != nil {
		return fmt.Errorf("invalid target profile name: %w", err)
	}

	// Verify source profile exists
	profiles, err := pm.viperManager.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	sourceExists := false
	for _, p := range profiles {
		if p == sourceName {
			sourceExists = true
			break
		}
	}
	if !sourceExists {
		return fmt.Errorf("failed to get source profile config: source profile %s does not exist", sourceName)
	}

	// Check if target already exists
	if _, err := pm.GetProfile(targetName); err == nil {
		return fmt.Errorf("profile %s already exists", targetName)
	}

	// Get source configuration
	sourceConfig, err := pm.viperManager.GetConfig(sourceName)
	if err != nil {
		return fmt.Errorf("failed to get source profile config: %w", err)
	}

	// Create new profile with copied config
	if err := pm.viperManager.CreateProfile(targetName, sourceConfig); err != nil {
		return fmt.Errorf("failed to create target profile: %w", err)
	}

	return nil
}

// ConfigureProfile provides interactive configuration for a profile
func (pm *ProfileManager) ConfigureProfile(name string, interactive bool) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	// Get current configuration
	config, err := pm.viperManager.GetConfig(name)
	if err != nil {
		// If profile doesn't exist, create it with defaults
		config = types.DefaultConfig()
	}

	if interactive {
		// Interactive configuration would be implemented here
		// For now, this is a placeholder for future implementation
		return fmt.Errorf("interactive configuration not yet implemented")
	}

	// Save the configuration
	return pm.viperManager.SaveProfile(name, config)
}

// ValidateProfile validates a profile's configuration
func (pm *ProfileManager) ValidateProfile(name string) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	config, err := pm.viperManager.GetConfig(name)
	if err != nil {
		return fmt.Errorf("failed to get profile config: %w", err)
	}

	return config.Validate()
}

// ExportProfile exports a profile configuration to a file or returns it as a string
func (pm *ProfileManager) ExportProfile(name string) (*types.Config, error) {
	if err := validateProfileName(name); err != nil {
		return nil, err
	}

	return pm.viperManager.GetConfig(name)
}

// ImportProfile imports a profile configuration from a config object
func (pm *ProfileManager) ImportProfile(name string, config *types.Config) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	// Validate the imported configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return pm.viperManager.CreateProfile(name, config)
}

// CreateProfileOptions contains options for creating a new profile
type CreateProfileOptions struct {
	// Storage configuration
	StorageType types.StorageType `json:"storage_type,omitempty"`

	// Local storage options
	LocalPath string `json:"local_path,omitempty"`

	// S3 storage options
	S3Bucket   string `json:"s3_bucket,omitempty"`
	S3Region   string `json:"s3_region,omitempty"`
	S3Endpoint string `json:"s3_endpoint,omitempty"`

	// Vault options
	VaultName string `json:"vault_name,omitempty"`

	// Profile metadata
	Description string `json:"description,omitempty"`
}

// ProfileSetValue represents a configuration value to set
type ProfileSetValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// SetProfileValue sets a specific configuration value for a profile
func (pm *ProfileManager) SetProfileValue(profileName, key string, value interface{}) error {
	if err := validateProfileName(profileName); err != nil {
		return err
	}

	// Get current configuration
	config, err := pm.viperManager.GetConfig(profileName)
	if err != nil {
		return fmt.Errorf("failed to get profile config: %w", err)
	}

	// Set the value using dot notation
	if err := setConfigValue(config, key, value); err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	// Validate and save
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return pm.viperManager.SaveProfile(profileName, config)
}

// GetProfileValue gets a specific configuration value from a profile
func (pm *ProfileManager) GetProfileValue(profileName, key string) (interface{}, error) {
	if err := validateProfileName(profileName); err != nil {
		return nil, err
	}

	config, err := pm.viperManager.GetConfig(profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile config: %w", err)
	}

	return getConfigValue(config, key)
}

// validateProfileName validates a profile name
func validateProfileName(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return fmt.Errorf("profile name contains invalid characters")
	}

	if len(name) > 64 {
		return fmt.Errorf("profile name too long (max 64 characters)")
	}

	return nil
}

// setConfigValue sets a configuration value using dot notation
func setConfigValue(config *types.Config, key string, value interface{}) error {
	// This is a simplified implementation
	// In a full implementation, this would use reflection or a more sophisticated approach
	switch key {
	case "vault.name":
		if s, ok := value.(string); ok {
			config.Vault.Name = s
		} else {
			return fmt.Errorf("vault.name must be a string")
		}
	case "storage.type":
		if s, ok := value.(string); ok {
			config.Storage.Type = types.StorageType(s)
		} else {
			return fmt.Errorf("storage.type must be a string")
		}
	case "storage.s3.bucket":
		if s, ok := value.(string); ok {
			config.Storage.S3.Bucket = s
		} else {
			return fmt.Errorf("storage.s3.bucket must be a string")
		}
	case "storage.s3.region":
		if s, ok := value.(string); ok {
			config.Storage.S3.Region = s
		} else {
			return fmt.Errorf("storage.s3.region must be a string")
		}
	case "storage.local.path":
		if s, ok := value.(string); ok {
			config.Storage.Local.Path = s
		} else {
			return fmt.Errorf("storage.local.path must be a string")
		}
	default:
		return fmt.Errorf("unsupported configuration key: %s", key)
	}

	return nil
}

// getConfigValue gets a configuration value using dot notation
func getConfigValue(config *types.Config, key string) (interface{}, error) {
	// This is a simplified implementation
	switch key {
	case "vault.name":
		return config.Vault.Name, nil
	case "storage.type":
		return string(config.Storage.Type), nil
	case "storage.s3.bucket":
		return config.Storage.S3.Bucket, nil
	case "storage.s3.region":
		return config.Storage.S3.Region, nil
	case "storage.local.path":
		return config.Storage.Local.Path, nil
	default:
		return nil, fmt.Errorf("unsupported configuration key: %s", key)
	}
}
