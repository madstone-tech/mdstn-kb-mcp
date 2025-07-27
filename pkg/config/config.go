package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Manager handles configuration loading and validation
type Manager struct {
	config *types.Config
	paths  []string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		paths: getDefaultConfigPaths(),
	}
}

// Load loads configuration from multiple sources with precedence
func (m *Manager) Load() (*types.Config, error) {
	config := types.DefaultConfig()

	// Load from config files (lowest precedence)
	if err := m.loadFromFiles(config); err != nil {
		return nil, fmt.Errorf("failed to load config files: %w", err)
	}

	// Override with environment variables (highest precedence)
	if err := m.loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	m.config = config
	return config, nil
}

// LoadFromFile loads configuration from a specific file
func (m *Manager) LoadFromFile(path string) (*types.Config, error) {
	config := types.DefaultConfig()

	if err := m.loadFile(path, config); err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	if err := m.loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	m.config = config
	return config, nil
}

// GetConfig returns the currently loaded configuration
func (m *Manager) GetConfig() *types.Config {
	if m.config == nil {
		return types.DefaultConfig()
	}
	return m.config
}

// SaveToFile saves the given configuration to a file
func (m *Manager) SaveToFile(config *types.Config, path string) error {
	oldConfig := m.config
	m.config = config
	defer func() { m.config = oldConfig }()
	
	return m.WriteToFile(path)
}

// WriteToFile writes the current configuration to a file
func (m *Manager) WriteToFile(path string) error {
	if m.config == nil {
		return fmt.Errorf("no configuration loaded")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create temp file for atomic write
	tempPath := path + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}
	defer func() { _ = file.Close() }() // Ignore close error in defer

	// Encode configuration as TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(m.config); err != nil {
		_ = os.Remove(tempPath) // Ignore removal error when handling encode error
		return fmt.Errorf("failed to encode configuration: %w", err)
	}

	_ = file.Close() // Ignore close error before rename

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath) // Ignore removal error when handling rename error
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// loadFromFiles loads configuration from all discovered config files
func (m *Manager) loadFromFiles(config *types.Config) error {
	for _, path := range m.paths {
		if _, err := os.Stat(path); err == nil {
			if err := m.loadFile(path, config); err != nil {
				return fmt.Errorf("error loading %s: %w", path, err)
			}
		}
	}
	return nil
}

// loadFile loads configuration from a single file
func (m *Manager) loadFile(path string, config *types.Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := toml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse TOML: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func (m *Manager) loadFromEnv(config *types.Config) error {
	// Vault configuration
	if val := os.Getenv("KBVAULT_NAME"); val != "" {
		config.Vault.Name = val
	}
	if val := os.Getenv("KBVAULT_MAX_FILE_SIZE"); val != "" {
		if size, err := parseBytes(val); err == nil {
			config.Vault.MaxFileSize = size
		}
	}

	// Storage configuration
	if val := os.Getenv("KBVAULT_STORAGE_TYPE"); val != "" {
		config.Storage.Type = types.StorageType(val)
	}

	// Local storage
	if val := os.Getenv("KBVAULT_STORAGE_LOCAL_PATH"); val != "" {
		config.Storage.Local.Path = val
	}
	if val := os.Getenv("KBVAULT_STORAGE_LOCAL_CREATE_DIRS"); val != "" {
		config.Storage.Local.CreateDirs = parseBool(val)
	}
	if val := os.Getenv("KBVAULT_STORAGE_LOCAL_ENABLE_LOCKING"); val != "" {
		config.Storage.Local.EnableLocking = parseBool(val)
	}

	// S3 storage
	if val := os.Getenv("KBVAULT_STORAGE_S3_BUCKET"); val != "" {
		config.Storage.S3.Bucket = val
	}
	if val := os.Getenv("KBVAULT_STORAGE_S3_REGION"); val != "" {
		config.Storage.S3.Region = val
	}
	if val := os.Getenv("KBVAULT_STORAGE_S3_ENDPOINT"); val != "" {
		config.Storage.S3.Endpoint = val
	}
	if val := os.Getenv("AWS_ACCESS_KEY_ID"); val != "" {
		config.Storage.S3.AccessKeyID = val
	}
	if val := os.Getenv("AWS_SECRET_ACCESS_KEY"); val != "" {
		config.Storage.S3.SecretAccessKey = val
	}
	if val := os.Getenv("AWS_SESSION_TOKEN"); val != "" {
		config.Storage.S3.SessionToken = val
	}

	// Server configuration
	if val := os.Getenv("KBVAULT_HTTP_PORT"); val != "" {
		if port, err := parseInt(val); err == nil {
			config.Server.HTTP.Port = port
		}
	}
	if val := os.Getenv("KBVAULT_HTTP_HOST"); val != "" {
		config.Server.HTTP.Host = val
	}

	// Cache configuration
	if val := os.Getenv("KBVAULT_CACHE_ENABLED"); val != "" {
		config.Storage.Cache.Enabled = parseBool(val)
	}

	return nil
}

// getDefaultConfigPaths returns the default configuration file search paths
func getDefaultConfigPaths() []string {
	paths := []string{
		"./kbvault.toml",
		"./config/kbvault.toml",
	}

	// Add user config directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "kbvault", "config.toml"))
		paths = append(paths, filepath.Join(home, ".kbvault.toml"))
	}

	// Add system config directory
	paths = append(paths, "/etc/kbvault/config.toml")

	return paths
}

// GetConfigPaths returns the current config file search paths
func (m *Manager) GetConfigPaths() []string {
	return m.paths
}

// SetConfigPaths sets custom config file search paths
func (m *Manager) SetConfigPaths(paths []string) {
	m.paths = paths
}

// FindConfigFile returns the first existing config file in search paths
func (m *Manager) FindConfigFile() (string, error) {
	for _, path := range m.paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no config file found in search paths: %v", m.paths)
}

// Utility functions for parsing environment variables

func parseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func parseBytes(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	
	// Handle suffixes
	multiplier := int64(1)
	if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	}

	var size int64
	_, err := fmt.Sscanf(s, "%d", &size)
	return size * multiplier, err
}