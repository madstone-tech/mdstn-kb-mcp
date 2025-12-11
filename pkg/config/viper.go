package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/spf13/viper"
)

// ViperManager manages configuration using Viper with profile support
type ViperManager struct {
	// Global Viper instance for global configuration
	global *viper.Viper

	// Profile-specific Viper instances
	profiles map[string]*viper.Viper

	// Current active profile
	activeProfile string

	// Configuration directory paths
	globalConfigDir   string
	profilesConfigDir string
}

// NewViperManager creates a new Viper-based configuration manager
func NewViperManager() (*ViperManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	globalConfigDir := filepath.Join(homeDir, ".kbvault")
	profilesConfigDir := filepath.Join(globalConfigDir, "profiles")

	vm := &ViperManager{
		global:            viper.New(),
		profiles:          make(map[string]*viper.Viper),
		globalConfigDir:   globalConfigDir,
		profilesConfigDir: profilesConfigDir,
	}

	// Initialize global configuration
	if err := vm.initGlobalConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize global config: %w", err)
	}

	// Load active profile
	if err := vm.loadActiveProfile(); err != nil {
		return nil, fmt.Errorf("failed to load active profile: %w", err)
	}

	return vm, nil
}

// initGlobalConfig initializes the global Viper configuration
func (vm *ViperManager) initGlobalConfig() error {
	vm.global.SetConfigName("config")
	vm.global.SetConfigType("toml")
	vm.global.AddConfigPath(vm.globalConfigDir)
	vm.global.AddConfigPath(".")
	vm.global.AddConfigPath("./config")

	// Set environment variable prefix
	vm.global.SetEnvPrefix("KBVAULT")
	vm.global.AutomaticEnv()
	vm.global.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Set default values
	vm.setDefaultValues(vm.global)

	// Try to read global config file
	if err := vm.global.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading global config file: %w", err)
		}
		// Config file not found is okay, we'll use defaults
	}

	return nil
}

// loadActiveProfile loads the currently active profile
func (vm *ViperManager) loadActiveProfile() error {
	activeProfileFile := filepath.Join(vm.globalConfigDir, "active_profile")

	data, err := os.ReadFile(activeProfileFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No active profile set, use default
			vm.activeProfile = "default"
			return nil
		}
		return fmt.Errorf("failed to read active profile file: %w", err)
	}

	profileName := strings.TrimSpace(string(data))
	if profileName == "" {
		vm.activeProfile = "default"
	} else {
		vm.activeProfile = profileName
	}

	return nil
}

// GetConfig returns the configuration for the specified profile
// If profile is empty, uses the active profile
func (vm *ViperManager) GetConfig(profile string) (*types.Config, error) {
	if profile == "" {
		profile = vm.activeProfile
	}

	// Load profile if not already loaded
	if err := vm.loadProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to load profile %s: %w", profile, err)
	}

	// Get profile-specific Viper
	profileViper := vm.profiles[profile]
	if profileViper == nil {
		return nil, fmt.Errorf("profile %s not found", profile)
	}

	// Start with default configuration to ensure all fields have values
	config := types.DefaultConfig()

	// Then, unmarshal global settings (overwrites defaults)
	if err := vm.global.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal global config: %w", err)
	}

	// Finally, unmarshal profile-specific settings (overwrites global)
	if err := profileViper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile config: %w", err)
	}

	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// loadProfile loads a specific profile configuration
func (vm *ViperManager) loadProfile(profile string) error {
	// Skip if already loaded
	if _, exists := vm.profiles[profile]; exists {
		return nil
	}

	profileViper := viper.New()
	profileViper.SetConfigName(profile)
	profileViper.SetConfigType("toml")
	profileViper.AddConfigPath(vm.profilesConfigDir)

	// Set environment variable prefix with profile
	envPrefix := fmt.Sprintf("KBVAULT_%s", strings.ToUpper(profile))
	profileViper.SetEnvPrefix(envPrefix)
	profileViper.AutomaticEnv()
	profileViper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Try to read profile config file
	if err := profileViper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading profile config file: %w", err)
		}
		// Profile config file not found is okay for new profiles
	}

	vm.profiles[profile] = profileViper
	return nil
}

// CreateProfile creates a new profile with the given configuration
func (vm *ViperManager) CreateProfile(name string, config *types.Config) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Validate the configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("profile configuration validation failed: %w", err)
	}

	// Ensure profiles directory exists
	if err := os.MkdirAll(vm.profilesConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Create profile Viper instance
	profileViper := viper.New()

	// Set all config values
	vm.setConfigValues(profileViper, config)

	// Write profile config file
	profilePath := filepath.Join(vm.profilesConfigDir, name+".toml")
	if err := profileViper.WriteConfigAs(profilePath); err != nil {
		return fmt.Errorf("failed to write profile config: %w", err)
	}

	// Add to loaded profiles
	vm.profiles[name] = profileViper

	return nil
}

// DeleteProfile removes a profile
func (vm *ViperManager) DeleteProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	// Remove config file
	profilePath := filepath.Join(vm.profilesConfigDir, name+".toml")
	if err := os.Remove(profilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove profile config file: %w", err)
	}

	// Remove from loaded profiles
	delete(vm.profiles, name)

	// If this was the active profile, switch to default
	if vm.activeProfile == name {
		if err := vm.SetActiveProfile("default"); err != nil {
			return fmt.Errorf("failed to switch to default profile: %w", err)
		}
	}

	return nil
}

// ListProfiles returns a list of available profiles
func (vm *ViperManager) ListProfiles() ([]string, error) {
	profiles := []string{"default"} // default is always available

	// Check profiles directory
	if _, err := os.Stat(vm.profilesConfigDir); os.IsNotExist(err) {
		return profiles, nil
	}

	entries, err := os.ReadDir(vm.profilesConfigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".toml") {
			profileName := strings.TrimSuffix(name, ".toml")
			if profileName != "default" { // avoid duplicates
				profiles = append(profiles, profileName)
			}
		}
	}

	return profiles, nil
}

// GetActiveProfile returns the name of the currently active profile
func (vm *ViperManager) GetActiveProfile() string {
	return vm.activeProfile
}

// SetActiveProfile sets the active profile
func (vm *ViperManager) SetActiveProfile(profile string) error {
	if profile == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Verify profile exists
	profiles, err := vm.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	found := false
	for _, p := range profiles {
		if p == profile {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile %s does not exist", profile)
	}

	// Update active profile
	vm.activeProfile = profile

	// Save to file
	activeProfileFile := filepath.Join(vm.globalConfigDir, "active_profile")
	if err := os.MkdirAll(vm.globalConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(activeProfileFile, []byte(profile), 0644); err != nil {
		return fmt.Errorf("failed to write active profile file: %w", err)
	}

	return nil
}

// SaveProfile saves the configuration for a specific profile
func (vm *ViperManager) SaveProfile(name string, config *types.Config) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Load profile to ensure it exists in memory
	if err := vm.loadProfile(name); err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	profileViper := vm.profiles[name]
	if profileViper == nil {
		return fmt.Errorf("profile %s not found", name)
	}

	// Set all config values
	vm.setConfigValues(profileViper, config)

	// Write to file
	profilePath := filepath.Join(vm.profilesConfigDir, name+".toml")
	if err := os.MkdirAll(vm.profilesConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	if err := profileViper.WriteConfigAs(profilePath); err != nil {
		return fmt.Errorf("failed to write profile config: %w", err)
	}

	return nil
}

// GetGlobalConfig returns the global configuration (without profile overrides)
func (vm *ViperManager) GetGlobalConfig() (*types.Config, error) {
	config := &types.Config{}
	if err := vm.global.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal global config: %w", err)
	}
	return config, nil
}

// SaveGlobalConfig saves the global configuration
func (vm *ViperManager) SaveGlobalConfig(config *types.Config) error {
	vm.setConfigValues(vm.global, config)

	globalConfigPath := filepath.Join(vm.globalConfigDir, "config.toml")
	if err := os.MkdirAll(vm.globalConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := vm.global.WriteConfigAs(globalConfigPath); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	return nil
}

// setDefaultValues sets default configuration values
func (vm *ViperManager) setDefaultValues(v *viper.Viper) {
	defaultConfig := types.DefaultConfig()
	vm.setConfigValues(v, defaultConfig)
}

// setConfigValues sets configuration values in a Viper instance
func (vm *ViperManager) setConfigValues(v *viper.Viper, config *types.Config) {
	// Vault configuration
	v.Set("vault.name", config.Vault.Name)
	v.Set("vault.notes_dir", config.Vault.NotesDir)
	v.Set("vault.daily_dir", config.Vault.DailyDir)
	v.Set("vault.templates_dir", config.Vault.TemplatesDir)
	v.Set("vault.default_template", config.Vault.DefaultTemplate)
	v.Set("vault.max_file_size", config.Vault.MaxFileSize)
	v.Set("vault.date_format", config.Vault.DateFormat)
	v.Set("vault.time_format", config.Vault.TimeFormat)
	v.Set("vault.auto_save", config.Vault.AutoSave)
	v.Set("vault.auto_sync", config.Vault.AutoSync)

	// Storage configuration
	v.Set("storage.type", config.Storage.Type)

	// Local storage
	v.Set("storage.local.path", config.Storage.Local.Path)
	v.Set("storage.local.create_dirs", config.Storage.Local.CreateDirs)
	v.Set("storage.local.dir_perms", config.Storage.Local.DirPerms)
	v.Set("storage.local.file_perms", config.Storage.Local.FilePerms)
	v.Set("storage.local.enable_locking", config.Storage.Local.EnableLocking)
	v.Set("storage.local.lock_timeout", config.Storage.Local.LockTimeout)

	// S3 storage
	v.Set("storage.s3.bucket", config.Storage.S3.Bucket)
	v.Set("storage.s3.region", config.Storage.S3.Region)
	v.Set("storage.s3.endpoint", config.Storage.S3.Endpoint)
	v.Set("storage.s3.access_key_id", config.Storage.S3.AccessKeyID)
	v.Set("storage.s3.secret_access_key", config.Storage.S3.SecretAccessKey)
	v.Set("storage.s3.session_token", config.Storage.S3.SessionToken)
	v.Set("storage.s3.use_ssl", config.Storage.S3.UseSSL)
	v.Set("storage.s3.path_style", config.Storage.S3.PathStyle)
	v.Set("storage.s3.prefix", config.Storage.S3.Prefix)
	v.Set("storage.s3.storage_class", config.Storage.S3.StorageClass)
	v.Set("storage.s3.server_side_encryption", config.Storage.S3.ServerSideEncryption)
	v.Set("storage.s3.kms_key_id", config.Storage.S3.KMSKeyID)
	v.Set("storage.s3.retry_attempts", config.Storage.S3.RetryAttempts)
	v.Set("storage.s3.retry_delay", config.Storage.S3.RetryDelay)
	v.Set("storage.s3.request_timeout", config.Storage.S3.RequestTimeout)
	v.Set("storage.s3.enable_versioning", config.Storage.S3.EnableVersioning)

	// Cache configuration
	v.Set("storage.cache.enabled", config.Storage.Cache.Enabled)
	v.Set("storage.cache.auto_enable_for_remote", config.Storage.Cache.AutoEnable)
	v.Set("storage.cache.memory.enabled", config.Storage.Cache.Memory.Enabled)
	v.Set("storage.cache.memory.max_size_mb", config.Storage.Cache.Memory.MaxSizeMB)
	v.Set("storage.cache.memory.max_items", config.Storage.Cache.Memory.MaxItems)
	v.Set("storage.cache.memory.ttl_minutes", config.Storage.Cache.Memory.TTLMinutes)
	v.Set("storage.cache.disk.enabled", config.Storage.Cache.Disk.Enabled)
	v.Set("storage.cache.disk.path", config.Storage.Cache.Disk.Path)
	v.Set("storage.cache.disk.max_size_mb", config.Storage.Cache.Disk.MaxSizeMB)
	v.Set("storage.cache.disk.ttl_hours", config.Storage.Cache.Disk.TTLHours)
	v.Set("storage.cache.disk.cleanup_interval_hours", config.Storage.Cache.Disk.CleanupIntervalHours)

	// Server configuration
	v.Set("server.http.enabled", config.Server.HTTP.Enabled)
	v.Set("server.http.host", config.Server.HTTP.Host)
	v.Set("server.http.port", config.Server.HTTP.Port)
	v.Set("server.http.enable_cors", config.Server.HTTP.EnableCORS)
	v.Set("server.http.cors_origins", config.Server.HTTP.CORSOrigins)
	v.Set("server.http.read_timeout", config.Server.HTTP.ReadTimeout)
	v.Set("server.http.write_timeout", config.Server.HTTP.WriteTimeout)
	v.Set("server.http.idle_timeout", config.Server.HTTP.IdleTimeout)
	v.Set("server.http.max_request_size", config.Server.HTTP.MaxRequestSize)

	// Logging configuration
	v.Set("logging.level", config.Logging.Level)
	v.Set("logging.output", config.Logging.Output)
	v.Set("logging.file_path", config.Logging.FilePath)
	v.Set("logging.format", config.Logging.Format)
	v.Set("logging.enable_colors", config.Logging.EnableColors)
	v.Set("logging.enable_timestamp", config.Logging.EnableTimestamp)
	v.Set("logging.enable_caller", config.Logging.EnableCaller)
	v.Set("logging.rotate_size", config.Logging.RotateSize)
	v.Set("logging.rotate_count", config.Logging.RotateCount)

	// TUI configuration
	v.Set("tui.theme", config.TUI.Theme)
	v.Set("tui.vim_mode", config.TUI.VimMode)
	v.Set("tui.show_help", config.TUI.ShowHelp)
	v.Set("tui.refresh_interval", config.TUI.RefreshInterval)
	v.Set("tui.page_size", config.TUI.PageSize)
	v.Set("tui.enable_mouse", config.TUI.EnableMouse)

	// MCP configuration
	v.Set("mcp.enabled", config.MCP.Enabled)
	v.Set("mcp.socket_path", config.MCP.SocketPath)
	v.Set("mcp.use_stdio", config.MCP.UseStdio)
	v.Set("mcp.max_request_size", config.MCP.MaxRequestSize)
	v.Set("mcp.response_timeout", config.MCP.ResponseTimeout)
	v.Set("mcp.enable_bulk_operations", config.MCP.EnableBulkOperations)
	v.Set("mcp.max_bulk_size", config.MCP.MaxBulkSize)

	// Vector search configuration
	v.Set("vector_search.enabled", config.VectorSearch.Enabled)
	v.Set("vector_search.type", config.VectorSearch.Type)

	// Embedding configuration
	v.Set("vector_search.embedding.provider", config.VectorSearch.Embedding.Provider)
	v.Set("vector_search.embedding.model", config.VectorSearch.Embedding.Model)
	v.Set("vector_search.embedding.dimensions", config.VectorSearch.Embedding.Dimensions)

	// OpenAI embedding configuration
	v.Set("vector_search.embedding.openai.api_key", config.VectorSearch.Embedding.OpenAI.APIKey)
	v.Set("vector_search.embedding.openai.model", config.VectorSearch.Embedding.OpenAI.Model)
	v.Set("vector_search.embedding.openai.base_url", config.VectorSearch.Embedding.OpenAI.BaseURL)

	// Local vector configuration
	v.Set("vector_search.local.database_path", config.VectorSearch.Local.DatabasePath)
	v.Set("vector_search.local.engine", config.VectorSearch.Local.Engine)
	v.Set("vector_search.local.index_type", config.VectorSearch.Local.IndexType)
	v.Set("vector_search.local.distance_metric", config.VectorSearch.Local.DistanceMetric)

	// Indexing configuration
	v.Set("vector_search.indexing.auto_index", config.VectorSearch.Indexing.AutoIndex)
	v.Set("vector_search.indexing.chunk_size", config.VectorSearch.Indexing.ChunkSize)
	v.Set("vector_search.indexing.chunk_overlap", config.VectorSearch.Indexing.ChunkOverlap)
	v.Set("vector_search.indexing.batch_size", config.VectorSearch.Indexing.BatchSize)

	// Search configuration
	v.Set("vector_search.search.hybrid_enabled", config.VectorSearch.Search.HybridEnabled)
	v.Set("vector_search.search.hybrid_weight", config.VectorSearch.Search.HybridWeight)
	v.Set("vector_search.search.default_limit", config.VectorSearch.Search.DefaultLimit)
	v.Set("vector_search.search.min_score", config.VectorSearch.Search.MinScore)
}
