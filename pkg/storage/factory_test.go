package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestFactory_CreateStorage(t *testing.T) {
	factory := NewFactory()

	// Create temporary directory for local storage test
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  types.StorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid local storage",
			config: types.StorageConfig{
				Type: types.StorageTypeLocal,
				Local: types.LocalStorageConfig{
					Path: tmpDir,
				},
			},
			wantErr: false,
		},
		{
			name: "valid S3 storage",
			config: types.StorageConfig{
				Type: types.StorageTypeS3,
				S3: types.S3StorageConfig{
					Bucket: "test-bucket",
					Region: "us-east-1",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid storage type",
			config: types.StorageConfig{
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported storage type",
		},
		{
			name: "invalid local config",
			config: types.StorageConfig{
				Type: types.StorageTypeLocal,
				Local: types.LocalStorageConfig{
					Path: "", // empty path
				},
			},
			wantErr: true,
			errMsg:  "path cannot be empty",
		},
		{
			name: "invalid S3 config - missing bucket",
			config: types.StorageConfig{
				Type: types.StorageTypeS3,
				S3: types.S3StorageConfig{
					Region: "us-east-1",
					// Bucket missing
				},
			},
			wantErr: true,
			errMsg:  "bucket name cannot be empty",
		},
		{
			name: "invalid S3 config - missing region",
			config: types.StorageConfig{
				Type: types.StorageTypeS3,
				S3: types.S3StorageConfig{
					Bucket: "test-bucket",
					// Region missing
				},
			},
			wantErr: true,
			errMsg:  "region cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := factory.CreateStorage(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, storage)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, tt.config.Type, storage.Type())
			}
		})
	}
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  types.StorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid local config",
			config: types.StorageConfig{
				Type: types.StorageTypeLocal,
				Local: types.LocalStorageConfig{
					Path: "/tmp/test-vault",
				},
			},
			wantErr: false,
		},
		{
			name: "valid S3 config",
			config: types.StorageConfig{
				Type: types.StorageTypeS3,
				S3: types.S3StorageConfig{
					Bucket: "test-bucket",
					Region: "us-east-1",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid storage type",
			config: types.StorageConfig{
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported storage type",
		},
		{
			name: "invalid local config",
			config: types.StorageConfig{
				Type: types.StorageTypeLocal,
				Local: types.LocalStorageConfig{
					Path: "",
				},
			},
			wantErr: true,
			errMsg:  "local storage path cannot be empty",
		},
		{
			name: "invalid S3 config",
			config: types.StorageConfig{
				Type: types.StorageTypeS3,
				S3: types.S3StorageConfig{
					Region: "us-east-1",
					// Bucket missing
				},
			},
			wantErr: true,
			errMsg:  "S3 bucket name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	supportedTypes := factory.GetSupportedTypes()

	assert.Contains(t, supportedTypes, types.StorageTypeLocal)
	assert.Contains(t, supportedTypes, types.StorageTypeS3)
	assert.Len(t, supportedTypes, 2)
}

func TestDefaultFactory(t *testing.T) {
	assert.NotNil(t, DefaultFactory)
}

func TestPackageFunctions(t *testing.T) {
	// Create temporary directory for local storage test
	tmpDir := t.TempDir()

	// Test that package-level functions work
	config := types.StorageConfig{
		Type: types.StorageTypeLocal,
		Local: types.LocalStorageConfig{
			Path: tmpDir,
		},
	}

	// Test CreateStorage
	storage, err := CreateStorage(config)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
	assert.Equal(t, types.StorageTypeLocal, storage.Type())

	// Test ValidateConfig
	err = ValidateConfig(config)
	assert.NoError(t, err)

	// Test GetSupportedTypes
	supportedTypes := GetSupportedTypes()
	assert.Contains(t, supportedTypes, types.StorageTypeLocal)
	assert.Contains(t, supportedTypes, types.StorageTypeS3)
}

func TestValidateLocalConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.LocalStorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: types.LocalStorageConfig{
				Path: "/tmp/test-vault",
			},
			wantErr: false,
		},
		{
			name: "empty path",
			config: types.LocalStorageConfig{
				Path: "",
			},
			wantErr: true,
			errMsg:  "local storage path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLocalConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateS3Config(t *testing.T) {
	tests := []struct {
		name    string
		config  types.S3StorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: types.S3StorageConfig{
				Bucket: "test-bucket",
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "empty bucket",
			config: types.S3StorageConfig{
				Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "S3 bucket name cannot be empty",
		},
		{
			name: "empty region",
			config: types.S3StorageConfig{
				Bucket: "test-bucket",
			},
			wantErr: true,
			errMsg:  "S3 region cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateS3Config(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkCreateStorage(b *testing.B) {
	factory := NewFactory()
	config := types.StorageConfig{
		Type: types.StorageTypeLocal,
		Local: types.LocalStorageConfig{
			Path: "/tmp/test-vault",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage, err := factory.CreateStorage(config)
		require.NoError(b, err)
		require.NotNil(b, storage)
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	factory := NewFactory()
	config := types.StorageConfig{
		Type: types.StorageTypeS3,
		S3: types.S3StorageConfig{
			Bucket: "test-bucket",
			Region: "us-east-1",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := factory.ValidateConfig(config)
		require.NoError(b, err)
	}
}
