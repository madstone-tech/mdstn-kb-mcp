package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestManager_Load(t *testing.T) {
	manager := NewManager()
	
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Should have default values
	if config.Vault.Name != "my-kb" {
		t.Errorf("Expected default vault name 'my-kb', got %s", config.Vault.Name)
	}

	if config.Storage.Type != types.StorageTypeLocal {
		t.Errorf("Expected default storage type 'local', got %s", config.Storage.Type)
	}
}

func TestManager_LoadFromFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	configContent := `
[vault]
name = "test-vault"
max_file_size = 5242880  # 5MB

[storage]
type = "local"

[storage.local]
path = "/tmp/test-vault"
create_dirs = true
enable_locking = true
lock_timeout = 10

[server.http]
port = 8080
host = "localhost"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	manager := NewManager()
	config, err := manager.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	// Verify loaded values
	if config.Vault.Name != "test-vault" {
		t.Errorf("Expected vault name 'test-vault', got %s", config.Vault.Name)
	}

	if config.Vault.MaxFileSize != 5242880 {
		t.Errorf("Expected max file size 5242880, got %d", config.Vault.MaxFileSize)
	}

	if config.Storage.Local.Path != "/tmp/test-vault" {
		t.Errorf("Expected local path '/tmp/test-vault', got %s", config.Storage.Local.Path)
	}

	if !config.Storage.Local.CreateDirs {
		t.Error("Expected create_dirs to be true")
	}

	if config.Server.HTTP.Port != 8080 {
		t.Errorf("Expected HTTP port 8080, got %d", config.Server.HTTP.Port)
	}
}

func TestManager_LoadFromEnv(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"KBVAULT_NAME":                     "env-vault",
		"KBVAULT_STORAGE_TYPE":             "s3",
		"KBVAULT_STORAGE_S3_BUCKET":        "test-bucket",
		"KBVAULT_STORAGE_S3_REGION":        "us-west-2",
		"AWS_ACCESS_KEY_ID":                "test-key",
		"AWS_SECRET_ACCESS_KEY":            "test-secret",
		"KBVAULT_HTTP_PORT":                "9090",
		"KBVAULT_CACHE_ENABLED":            "true",
		"KBVAULT_MAX_FILE_SIZE":            "20MB",
		"KBVAULT_STORAGE_LOCAL_CREATE_DIRS": "false",
	}

	// Set environment variables
	for key, value := range envVars {
		t.Setenv(key, value)
	}

	manager := NewManager()
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	// Verify environment variable overrides
	if config.Vault.Name != "env-vault" {
		t.Errorf("Expected vault name 'env-vault', got %s", config.Vault.Name)
	}

	if config.Storage.Type != types.StorageTypeS3 {
		t.Errorf("Expected storage type 's3', got %s", config.Storage.Type)
	}

	if config.Storage.S3.Bucket != "test-bucket" {
		t.Errorf("Expected S3 bucket 'test-bucket', got %s", config.Storage.S3.Bucket)
	}

	if config.Storage.S3.AccessKeyID != "test-key" {
		t.Errorf("Expected access key 'test-key', got %s", config.Storage.S3.AccessKeyID)
	}

	if config.Server.HTTP.Port != 9090 {
		t.Errorf("Expected HTTP port 9090, got %d", config.Server.HTTP.Port)
	}

	if !config.Storage.Cache.Enabled {
		t.Error("Expected cache to be enabled")
	}

	if config.Vault.MaxFileSize != 20*1024*1024 {
		t.Errorf("Expected max file size 20MB, got %d", config.Vault.MaxFileSize)
	}

	if config.Storage.Local.CreateDirs {
		t.Error("Expected create_dirs to be false")
	}
}

func TestManager_WriteToFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "output-config.toml")

	manager := NewManager()
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Modify some values
	config.Vault.Name = "written-vault"
	config.Storage.Type = types.StorageTypeS3
	config.Storage.S3.Bucket = "test-bucket"

	// Write to file
	if err := manager.WriteToFile(configPath); err != nil {
		t.Fatalf("Failed to write config to file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file was not created: %v", err)
	}

	// Load it back and verify
	newManager := NewManager()
	loadedConfig, err := newManager.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load written config: %v", err)
	}

	if loadedConfig.Vault.Name != "written-vault" {
		t.Errorf("Expected vault name 'written-vault', got %s", loadedConfig.Vault.Name)
	}

	if loadedConfig.Storage.Type != types.StorageTypeS3 {
		t.Errorf("Expected storage type 's3', got %s", loadedConfig.Storage.Type)
	}
}

func TestManager_FindConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a config file in temp directory
	configPath := filepath.Join(tempDir, "kbvault.toml")
	if err := os.WriteFile(configPath, []byte("[vault]\nname = \"test\""), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	manager := NewManager()
	manager.SetConfigPaths([]string{configPath})

	foundPath, err := manager.FindConfigFile()
	if err != nil {
		t.Fatalf("Failed to find config file: %v", err)
	}

	if foundPath != configPath {
		t.Errorf("Expected to find %s, got %s", configPath, foundPath)
	}

	// Test when no config file exists
	manager.SetConfigPaths([]string{"/nonexistent/path.toml"})
	_, err = manager.FindConfigFile()
	if err == nil {
		t.Error("Expected error when no config file exists")
	}
}

func TestParseBytes(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"1024", 1024},
		{"1KB", 1024},
		{"1MB", 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024},
		{"5MB", 5 * 1024 * 1024},
		{"10gb", 10 * 1024 * 1024 * 1024}, // lowercase
		{" 2MB ", 2 * 1024 * 1024},        // with spaces
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseBytes(tc.input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", tc.input, err)
			}

			if result != tc.expected {
				t.Errorf("parseBytes(%s) = %d, expected %d", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := parseBool(tc.input)
			if result != tc.expected {
				t.Errorf("parseBool(%s) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestManager_GetConfigPaths(t *testing.T) {
	manager := NewManager()
	paths := manager.GetConfigPaths()

	if len(paths) == 0 {
		t.Error("Config paths should not be empty")
	}

	// Should include common paths
	found := false
	for _, path := range paths {
		if path == "./kbvault.toml" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Should include ./kbvault.toml in default paths")
	}
}

func TestManager_ConfigPrecedence(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create config file with one value
	configPath := filepath.Join(tempDir, "test.toml")
	configContent := `
[vault]
name = "file-vault"

[server.http]
port = 8080
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variable that should override
	t.Setenv("KBVAULT_NAME", "env-vault")
	t.Setenv("KBVAULT_HTTP_PORT", "9090")

	manager := NewManager()
	config, err := manager.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Environment should override file
	if config.Vault.Name != "env-vault" {
		t.Errorf("Environment should override file: expected 'env-vault', got %s", config.Vault.Name)
	}

	if config.Server.HTTP.Port != 9090 {
		t.Errorf("Environment should override file: expected port 9090, got %d", config.Server.HTTP.Port)
	}
}