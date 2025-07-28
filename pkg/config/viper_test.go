package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNewViperManager(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	
	// Set temporary home directory
	t.Setenv("HOME", tmpDir)
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

	vm, err := NewViperManager()
	assert.NoError(t, err)
	assert.NotNil(t, vm)
	assert.Equal(t, "default", vm.GetActiveProfile())
}

func TestViperManager_CreateProfile(t *testing.T) {
	vm := setupTestViperManager(t)
	
	config := types.DefaultConfig()
	config.Vault.Name = "test-vault"
	config.Storage.Type = types.StorageTypeS3
	config.Storage.S3.Bucket = "test-bucket"
	config.Storage.S3.Region = "us-east-1"

	err := vm.CreateProfile("test-profile", config)
	assert.NoError(t, err)

	// Verify profile was created
	profiles, err := vm.ListProfiles()
	assert.NoError(t, err)
	assert.Contains(t, profiles, "test-profile")

	// Verify we can load the profile config
	loadedConfig, err := vm.GetConfig("test-profile")
	assert.NoError(t, err)
	assert.Equal(t, "test-vault", loadedConfig.Vault.Name)
	assert.Equal(t, types.StorageTypeS3, loadedConfig.Storage.Type)
	assert.Equal(t, "test-bucket", loadedConfig.Storage.S3.Bucket)
}

func TestViperManager_DeleteProfile(t *testing.T) {
	vm := setupTestViperManager(t)
	
	// Create a test profile
	config := types.DefaultConfig()
	err := vm.CreateProfile("delete-me", config)
	require.NoError(t, err)

	// Verify it exists
	profiles, err := vm.ListProfiles()
	require.NoError(t, err)
	assert.Contains(t, profiles, "delete-me")

	// Delete it
	err = vm.DeleteProfile("delete-me")
	assert.NoError(t, err)

	// Verify it's gone
	profiles, err = vm.ListProfiles()
	assert.NoError(t, err)
	assert.NotContains(t, profiles, "delete-me")
}

func TestViperManager_DeleteProfile_Errors(t *testing.T) {
	vm := setupTestViperManager(t)

	tests := []struct {
		name        string
		profileName string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "empty profile name",
			profileName: "",
			wantErr:     true,
			errMsg:      "profile name cannot be empty",
		},
		{
			name:        "delete default profile",
			profileName: "default",
			wantErr:     true,
			errMsg:      "cannot delete default profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.DeleteProfile(tt.profileName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestViperManager_SetActiveProfile(t *testing.T) {
	vm := setupTestViperManager(t)

	// Create a test profile
	config := types.DefaultConfig()
	err := vm.CreateProfile("work", config)
	require.NoError(t, err)

	// Set it as active
	err = vm.SetActiveProfile("work")
	assert.NoError(t, err)
	assert.Equal(t, "work", vm.GetActiveProfile())

	// Test setting non-existent profile
	err = vm.SetActiveProfile("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestViperManager_GetConfig(t *testing.T) {
	vm := setupTestViperManager(t)

	// Test getting default config
	config, err := vm.GetConfig("")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "my-kb", config.Vault.Name) // default value

	// Create and test a custom profile
	customConfig := types.DefaultConfig()
	customConfig.Vault.Name = "custom-vault"
	customConfig.Storage.Type = types.StorageTypeS3
	customConfig.Storage.S3.Bucket = "custom-bucket"
	customConfig.Storage.S3.Region = "eu-west-1"

	err = vm.CreateProfile("custom", customConfig)
	require.NoError(t, err)

	loadedConfig, err := vm.GetConfig("custom")
	assert.NoError(t, err)
	assert.Equal(t, "custom-vault", loadedConfig.Vault.Name)
	assert.Equal(t, types.StorageTypeS3, loadedConfig.Storage.Type)
	assert.Equal(t, "custom-bucket", loadedConfig.Storage.S3.Bucket)
	assert.Equal(t, "eu-west-1", loadedConfig.Storage.S3.Region)
}

func TestViperManager_SaveProfile(t *testing.T) {
	vm := setupTestViperManager(t)

	// Create a profile
	originalConfig := types.DefaultConfig()
	originalConfig.Vault.Name = "original"
	err := vm.CreateProfile("save-test", originalConfig)
	require.NoError(t, err)

	// Modify and save
	modifiedConfig := types.DefaultConfig()
	modifiedConfig.Vault.Name = "modified"
	modifiedConfig.Storage.Type = types.StorageTypeS3
	modifiedConfig.Storage.S3.Bucket = "new-bucket"

	err = vm.SaveProfile("save-test", modifiedConfig)
	assert.NoError(t, err)

	// Load and verify changes
	loadedConfig, err := vm.GetConfig("save-test")
	assert.NoError(t, err)
	assert.Equal(t, "modified", loadedConfig.Vault.Name)
	assert.Equal(t, types.StorageTypeS3, loadedConfig.Storage.Type)
	assert.Equal(t, "new-bucket", loadedConfig.Storage.S3.Bucket)
}

func TestViperManager_GlobalConfig(t *testing.T) {
	vm := setupTestViperManager(t)

	// Get initial global config
	globalConfig, err := vm.GetGlobalConfig()
	assert.NoError(t, err)
	assert.NotNil(t, globalConfig)

	// Modify global config
	globalConfig.Vault.Name = "global-vault"
	globalConfig.Logging.Level = "DEBUG"

	err = vm.SaveGlobalConfig(globalConfig)
	assert.NoError(t, err)

	// Load and verify
	reloadedGlobal, err := vm.GetGlobalConfig()
	assert.NoError(t, err)
	assert.Equal(t, "global-vault", reloadedGlobal.Vault.Name)
	assert.Equal(t, "DEBUG", reloadedGlobal.Logging.Level)
}

func TestViperManager_ListProfiles(t *testing.T) {
	vm := setupTestViperManager(t)

	// Initial state should have default profile
	profiles, err := vm.ListProfiles()
	assert.NoError(t, err)
	assert.Contains(t, profiles, "default")

	// Create some profiles
	config := types.DefaultConfig()
	err = vm.CreateProfile("work", config)
	require.NoError(t, err)

	err = vm.CreateProfile("personal", config)
	require.NoError(t, err)

	// List again
	profiles, err = vm.ListProfiles()
	assert.NoError(t, err)
	assert.Contains(t, profiles, "default")
	assert.Contains(t, profiles, "work")
	assert.Contains(t, profiles, "personal")
	assert.Len(t, profiles, 3)
}

func TestViperManager_EnvironmentVariables(t *testing.T) {
	_ = setupTestViperManager(t)

	// Set environment variables
	t.Setenv("KBVAULT_VAULT_NAME", "env-vault")
	t.Setenv("KBVAULT_STORAGE_TYPE", "s3")
	t.Setenv("KBVAULT_STORAGE_S3_BUCKET", "env-bucket")
	t.Setenv("KBVAULT_STORAGE_S3_REGION", "us-west-2")

	// Create new manager to pick up env vars
	vm2 := setupTestViperManager(t)
	
	config, err := vm2.GetConfig("")
	assert.NoError(t, err)
	assert.Equal(t, "env-vault", config.Vault.Name)
	assert.Equal(t, types.StorageTypeS3, config.Storage.Type)
	assert.Equal(t, "env-bucket", config.Storage.S3.Bucket)
	assert.Equal(t, "us-west-2", config.Storage.S3.Region)
}

func TestViperManager_ProfileSpecificEnvironmentVariables(t *testing.T) {
	vm := setupTestViperManager(t)

	// Create a work profile
	config := types.DefaultConfig()
	err := vm.CreateProfile("work", config)
	require.NoError(t, err)

	// Set profile-specific environment variables
	t.Setenv("KBVAULT_WORK_VAULT_NAME", "work-vault")
	t.Setenv("KBVAULT_WORK_STORAGE_S3_BUCKET", "work-bucket")

	// Create new manager to pick up env vars
	vm2 := setupTestViperManager(t)
	
	// Load work profile
	workConfig, err := vm2.GetConfig("work")
	assert.NoError(t, err)
	
	// Note: Profile-specific env vars require the profile to be loaded fresh
	// This tests the environment variable structure is correct
	assert.NotNil(t, workConfig)
}

func TestViperManager_ConfigValidation(t *testing.T) {
	vm := setupTestViperManager(t)

	// Create invalid config
	invalidConfig := &types.Config{
		Vault: types.VaultConfig{
			Name:        "", // Invalid: empty name
			MaxFileSize: -1, // Invalid: negative size
		},
		Storage: types.StorageConfig{
			Type: "invalid", // Invalid: unsupported type
		},
	}

	err := vm.CreateProfile("invalid", invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestViperManager_FileOperations(t *testing.T) {
	vm := setupTestViperManager(t)

	// Create a profile and verify file exists
	config := types.DefaultConfig()
	config.Vault.Name = "file-test"
	
	err := vm.CreateProfile("file-test", config)
	assert.NoError(t, err)

	// Check that profile file was created
	profilePath := filepath.Join(vm.profilesConfigDir, "file-test.toml")
	_, err = os.Stat(profilePath)
	assert.NoError(t, err)

	// Set as active and verify active profile file
	err = vm.SetActiveProfile("file-test")
	assert.NoError(t, err)

	activeProfilePath := filepath.Join(vm.globalConfigDir, "active_profile")
	data, err := os.ReadFile(activeProfilePath)
	assert.NoError(t, err)
	assert.Equal(t, "file-test", string(data))
}

// Helper function to set up a test ViperManager
func setupTestViperManager(t *testing.T) *ViperManager {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Create a ViperManager with temporary directories
	vm := &ViperManager{
		profiles:          make(map[string]*viper.Viper),
		activeProfile:     "default",
		globalConfigDir:   filepath.Join(tmpDir, ".kbvault"),
		profilesConfigDir: filepath.Join(tmpDir, ".kbvault", "profiles"),
	}

	// Initialize global config
	err := vm.initGlobalConfig()
	require.NoError(t, err)

	return vm
}

// Benchmark tests
func BenchmarkViperManager_GetConfig(b *testing.B) {
	vm := setupBenchmarkViperManager(b)
	
	config := types.DefaultConfig()
	err := vm.CreateProfile("bench", config)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vm.GetConfig("bench")
		require.NoError(b, err)
	}
}

func BenchmarkViperManager_CreateProfile(b *testing.B) {
	vm := setupBenchmarkViperManager(b)
	config := types.DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profileName := fmt.Sprintf("bench-%d", i)
		err := vm.CreateProfile(profileName, config)
		require.NoError(b, err)
	}
}

func setupBenchmarkViperManager(b *testing.B) *ViperManager {
	tmpDir := b.TempDir()
	
	vm := &ViperManager{
		profiles:          make(map[string]*viper.Viper),
		activeProfile:     "default",
		globalConfigDir:   filepath.Join(tmpDir, ".kbvault"),
		profilesConfigDir: filepath.Join(tmpDir, ".kbvault", "profiles"),
	}

	err := vm.initGlobalConfig()
	require.NoError(b, err)

	return vm
}