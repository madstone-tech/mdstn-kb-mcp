package config

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNewProfileManager(t *testing.T) {
	pm := setupTestProfileManager(t)
	assert.NotNil(t, pm)
	assert.Equal(t, "default", pm.GetActiveProfile())
}

func TestProfileManager_CreateProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	tests := []struct {
		name    string
		profile string
		options *CreateProfileOptions
		wantErr bool
		errMsg  string
	}{
		{
			name:    "create basic profile",
			profile: "test-profile",
			options: &CreateProfileOptions{
				VaultName: "test-vault",
			},
			wantErr: false,
		},
		{
			name:    "create S3 profile",
			profile: "s3-profile",
			options: &CreateProfileOptions{
				StorageType: types.StorageTypeS3,
				S3Bucket:    "my-bucket",
				S3Region:    "us-east-1",
				VaultName:   "s3-vault",
			},
			wantErr: false,
		},
		{
			name:    "create local profile",
			profile: "local-profile",
			options: &CreateProfileOptions{
				StorageType: types.StorageTypeLocal,
				LocalPath:   "/tmp/local-vault",
				VaultName:   "local-vault",
			},
			wantErr: false,
		},
		{
			name:    "empty profile name",
			profile: "",
			options: nil,
			wantErr: true,
			errMsg:  "profile name cannot be empty",
		},
		{
			name:    "invalid profile name",
			profile: "invalid/name",
			options: nil,
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "profile name too long",
			profile: "a-very-long-profile-name-that-exceeds-the-maximum-allowed-length-limit",
			options: nil,
			wantErr: true,
			errMsg:  "profile name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.CreateProfile(tt.profile, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				
				// Verify profile was created
				profiles, err := pm.ListProfiles()
				assert.NoError(t, err)
				found := false
				for _, p := range profiles {
					if p.Name == tt.profile {
						found = true
						if tt.options != nil {
							if tt.options.StorageType != "" {
								assert.Equal(t, string(tt.options.StorageType), p.StorageType)
							}
						}
						break
					}
				}
				assert.True(t, found, "Profile should be in the list")
			}
		})
	}
}

func TestProfileManager_CreateProfile_Duplicate(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create a profile
	err := pm.CreateProfile("duplicate", nil)
	require.NoError(t, err)

	// Try to create the same profile again
	err = pm.CreateProfile("duplicate", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestProfileManager_DeleteProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create a profile to delete
	err := pm.CreateProfile("delete-me", nil)
	require.NoError(t, err)

	// Delete it
	err = pm.DeleteProfile("delete-me")
	assert.NoError(t, err)

	// Verify it's gone
	profiles, err := pm.ListProfiles()
	assert.NoError(t, err)
	for _, p := range profiles {
		assert.NotEqual(t, "delete-me", p.Name)
	}
}

func TestProfileManager_DeleteProfile_Errors(t *testing.T) {
	pm := setupTestProfileManager(t)

	tests := []struct {
		name    string
		profile string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "delete default profile",
			profile: "default",
			wantErr: true,
			errMsg:  "cannot delete the default profile",
		},
		{
			name:    "empty profile name",
			profile: "",
			wantErr: true,
			errMsg:  "profile name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.DeleteProfile(tt.profile)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProfileManager_ListProfiles(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Initial state
	profiles, err := pm.ListProfiles()
	assert.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, "default", profiles[0].Name)
	assert.True(t, profiles[0].IsDefault)
	assert.True(t, profiles[0].IsActive)

	// Create some profiles
	err = pm.CreateProfile("work", &CreateProfileOptions{
		StorageType: types.StorageTypeS3,
		S3Bucket:    "work-bucket",
	})
	require.NoError(t, err)

	err = pm.CreateProfile("personal", &CreateProfileOptions{
		StorageType: types.StorageTypeLocal,
		LocalPath:   "/home/user/personal",
	})
	require.NoError(t, err)

	// List again
	profiles, err = pm.ListProfiles()
	assert.NoError(t, err)
	assert.Len(t, profiles, 3)

	// Check that default is first
	assert.Equal(t, "default", profiles[0].Name)
	assert.True(t, profiles[0].IsDefault)

	// Check storage types
	for _, p := range profiles {
		switch p.Name {
		case "work":
			assert.Equal(t, "s3", p.StorageType)
		case "personal":
			assert.Equal(t, "local", p.StorageType)
		}
	}
}

func TestProfileManager_SwitchProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create a profile to switch to
	err := pm.CreateProfile("work", nil)
	require.NoError(t, err)

	// Switch to it
	err = pm.SwitchProfile("work")
	assert.NoError(t, err)
	assert.Equal(t, "work", pm.GetActiveProfile())

	// Verify active status in list
	profiles, err := pm.ListProfiles()
	assert.NoError(t, err)
	for _, p := range profiles {
		if p.Name == "work" {
			assert.True(t, p.IsActive)
		} else {
			assert.False(t, p.IsActive)
		}
	}
}

func TestProfileManager_GetProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Get default profile
	profile, err := pm.GetProfile("default")
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "default", profile.Name)
	assert.True(t, profile.IsDefault)
	assert.True(t, profile.IsActive)

	// Get non-existent profile
	profile, err = pm.GetProfile("non-existent")
	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "not found")
}

func TestProfileManager_UpdateProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create a profile
	err := pm.CreateProfile("update-test", nil)
	require.NoError(t, err)

	// Get current config
	config, err := pm.GetConfig("update-test")
	require.NoError(t, err)

	// Modify config
	config.Vault.Name = "updated-vault"
	config.Storage.Type = types.StorageTypeS3
	config.Storage.S3.Bucket = "updated-bucket"

	// Update profile
	err = pm.UpdateProfile("update-test", config)
	assert.NoError(t, err)

	// Verify changes
	updatedConfig, err := pm.GetConfig("update-test")
	assert.NoError(t, err)
	assert.Equal(t, "updated-vault", updatedConfig.Vault.Name)
	assert.Equal(t, types.StorageTypeS3, updatedConfig.Storage.Type)
	assert.Equal(t, "updated-bucket", updatedConfig.Storage.S3.Bucket)
}

func TestProfileManager_CopyProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create source profile with custom config
	options := &CreateProfileOptions{
		StorageType: types.StorageTypeS3,
		S3Bucket:    "source-bucket",
		S3Region:    "us-west-2",
		VaultName:   "source-vault",
	}
	err := pm.CreateProfile("source", options)
	require.NoError(t, err)

	// Copy profile
	err = pm.CopyProfile("source", "copy")
	assert.NoError(t, err)

	// Verify copy has same config
	sourceConfig, err := pm.GetConfig("source")
	require.NoError(t, err)

	copyConfig, err := pm.GetConfig("copy")
	require.NoError(t, err)

	assert.Equal(t, sourceConfig.Vault.Name, copyConfig.Vault.Name)
	assert.Equal(t, sourceConfig.Storage.Type, copyConfig.Storage.Type)
	assert.Equal(t, sourceConfig.Storage.S3.Bucket, copyConfig.Storage.S3.Bucket)
	assert.Equal(t, sourceConfig.Storage.S3.Region, copyConfig.Storage.S3.Region)
}

func TestProfileManager_CopyProfile_Errors(t *testing.T) {
	pm := setupTestProfileManager(t)

	tests := []struct {
		name       string
		sourceName string
		targetName string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "invalid source name",
			sourceName: "",
			targetName: "target",
			wantErr:    true,
			errMsg:     "invalid source profile name",
		},
		{
			name:       "invalid target name",
			sourceName: "default",
			targetName: "",
			wantErr:    true,
			errMsg:     "invalid target profile name",
		},
		{
			name:       "source doesn't exist",
			sourceName: "non-existent",
			targetName: "target",
			wantErr:    true,
			errMsg:     "failed to get source profile config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.CopyProfile(tt.sourceName, tt.targetName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProfileManager_SetGetProfileValue(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create a test profile
	err := pm.CreateProfile("test-values", nil)
	require.NoError(t, err)

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "vault name",
			key:   "vault.name",
			value: "test-vault",
		},
		{
			name:  "storage type",
			key:   "storage.type",
			value: "s3",
		},
		{
			name:  "s3 bucket",
			key:   "storage.s3.bucket",
			value: "test-bucket",
		},
		{
			name:  "s3 region",
			key:   "storage.s3.region",
			value: "eu-west-1",
		},
		{
			name:  "local path",
			key:   "storage.local.path",
			value: "/tmp/test-vault",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set value
			err := pm.SetProfileValue("test-values", tt.key, tt.value)
			assert.NoError(t, err)

			// Get value
			retrievedValue, err := pm.GetProfileValue("test-values", tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.value, retrievedValue)
		})
	}
}

func TestProfileManager_SetProfileValue_Errors(t *testing.T) {
	pm := setupTestProfileManager(t)

	err := pm.CreateProfile("test-errors", nil)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "unsupported key",
			key:     "unsupported.key",
			value:   "value",
			wantErr: true,
			errMsg:  "unsupported configuration key",
		},
		{
			name:    "wrong type for vault.name",
			key:     "vault.name",
			value:   123,
			wantErr: true,
			errMsg:  "must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.SetProfileValue("test-errors", tt.key, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProfileManager_ValidateProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create a valid profile
	err := pm.CreateProfile("valid", &CreateProfileOptions{
		StorageType: types.StorageTypeS3,
		S3Bucket:    "valid-bucket",
		S3Region:    "us-east-1",
		VaultName:   "valid-vault",
	})
	require.NoError(t, err)

	// Validate it
	err = pm.ValidateProfile("valid")
	assert.NoError(t, err)

	// Create an invalid profile by setting invalid values
	err = pm.CreateProfile("invalid", nil)
	require.NoError(t, err)

	// Set invalid values directly (this bypasses validation during set)
	config, err := pm.GetConfig("invalid")
	require.NoError(t, err)
	config.Vault.Name = "" // Invalid: empty name
	err = pm.UpdateProfile("invalid", config)
	assert.Error(t, err) // Should fail validation
}

func TestProfileManager_ExportImportProfile(t *testing.T) {
	pm := setupTestProfileManager(t)

	// Create a profile with custom config
	options := &CreateProfileOptions{
		StorageType: types.StorageTypeS3,
		S3Bucket:    "export-bucket",
		S3Region:    "ap-southeast-1",
		VaultName:   "export-vault",
	}
	err := pm.CreateProfile("export-test", options)
	require.NoError(t, err)

	// Export profile
	exportedConfig, err := pm.ExportProfile("export-test")
	assert.NoError(t, err)
	assert.NotNil(t, exportedConfig)
	assert.Equal(t, "export-vault", exportedConfig.Vault.Name)
	assert.Equal(t, types.StorageTypeS3, exportedConfig.Storage.Type)
	assert.Equal(t, "export-bucket", exportedConfig.Storage.S3.Bucket)

	// Import as new profile
	err = pm.ImportProfile("import-test", exportedConfig)
	assert.NoError(t, err)

	// Verify imported profile
	importedConfig, err := pm.GetConfig("import-test")
	assert.NoError(t, err)
	assert.Equal(t, exportedConfig.Vault.Name, importedConfig.Vault.Name)
	assert.Equal(t, exportedConfig.Storage.Type, importedConfig.Storage.Type)
	assert.Equal(t, exportedConfig.Storage.S3.Bucket, importedConfig.Storage.S3.Bucket)
}

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "valid name",
			profile: "valid-name",
			wantErr: false,
		},
		{
			name:    "valid name with numbers",
			profile: "profile123",
			wantErr: false,
		},
		{
			name:    "empty name",
			profile: "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "name with slash",
			profile: "invalid/name",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "name with backslash",
			profile: "invalid\\name",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "name too long",
			profile: "this-is-a-very-long-profile-name-that-exceeds-the-maximum-allowed-length",
			wantErr: true,
			errMsg:  "too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProfileName(tt.profile)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to set up a test ProfileManager
func setupTestProfileManager(t *testing.T) *ProfileManager {
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

	return &ProfileManager{
		viperManager: vm,
	}
}

// Benchmark tests
func BenchmarkProfileManager_ListProfiles(b *testing.B) {
	pm := setupBenchmarkProfileManager(b)
	
	// Create some profiles
	for i := 0; i < 10; i++ {
		err := pm.CreateProfile(fmt.Sprintf("bench-%d", i), nil)
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.ListProfiles()
		require.NoError(b, err)
	}
}

func BenchmarkProfileManager_CreateProfile(b *testing.B) {
	pm := setupBenchmarkProfileManager(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profileName := fmt.Sprintf("bench-create-%d", i)
		err := pm.CreateProfile(profileName, nil)
		require.NoError(b, err)
	}
}

func setupBenchmarkProfileManager(b *testing.B) *ProfileManager {
	tmpDir := b.TempDir()
	
	vm := &ViperManager{
		profiles:          make(map[string]*viper.Viper),
		activeProfile:     "default",
		globalConfigDir:   filepath.Join(tmpDir, ".kbvault"),
		profilesConfigDir: filepath.Join(tmpDir, ".kbvault", "profiles"),
	}

	err := vm.initGlobalConfig()
	require.NoError(b, err)

	return &ProfileManager{
		viperManager: vm,
	}
}