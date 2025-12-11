package vector

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestFactory_CreateVectorSearch(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name      string
		config    types.VectorSearchConfig
		wantType  types.VectorSearchType
		wantErr   bool
		errMsg    string
	}{
		{
			name: "disabled vector search",
			config: types.VectorSearchConfig{
				Enabled: false,
				Type:    types.VectorSearchTypeLocal,
			},
			wantType: types.VectorSearchTypeNone,
			wantErr:  false,
		},
		{
			name: "none type",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    types.VectorSearchTypeNone,
			},
			wantType: types.VectorSearchTypeNone,
			wantErr:  false,
		},
		{
			name: "local type - not implemented",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    types.VectorSearchTypeLocal,
			},
			wantErr: true,
			errMsg:  "not yet implemented",
		},
		{
			name: "pinecone type - not implemented",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    types.VectorSearchTypePinecone,
			},
			wantErr: true,
			errMsg:  "not yet implemented",
		},
		{
			name: "invalid type",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported vector search type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := factory.CreateVectorSearch(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, backend)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.Equal(t, tt.wantType, backend.Type())
			}
		})
	}
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  types.VectorSearchConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "disabled config",
			config: types.VectorSearchConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "none type",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    types.VectorSearchTypeNone,
			},
			wantErr: false,
		},
		{
			name: "valid local config",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    types.VectorSearchTypeLocal,
				Local: types.LocalVectorConfig{
					DatabasePath:   "/tmp/vector.db",
					Engine:         "sqlite",
					IndexType:      "flat",
					DistanceMetric: "cosine",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid local config - empty path",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    types.VectorSearchTypeLocal,
				Local: types.LocalVectorConfig{
					Engine:         "sqlite",
					IndexType:      "flat",
					DistanceMetric: "cosine",
				},
			},
			wantErr: true,
			errMsg:  "database path cannot be empty",
		},
		{
			name: "invalid local config - unsupported engine",
			config: types.VectorSearchConfig{
				Enabled: true,
				Type:    types.VectorSearchTypeLocal,
				Local: types.LocalVectorConfig{
					DatabasePath:   "/tmp/vector.db",
					Engine:         "invalid",
					IndexType:      "flat",
					DistanceMetric: "cosine",
				},
			},
			wantErr: true,
			errMsg:  "unsupported local vector engine",
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
	
	assert.Contains(t, supportedTypes, types.VectorSearchTypeNone)
	assert.Contains(t, supportedTypes, types.VectorSearchTypeLocal)
	assert.Contains(t, supportedTypes, types.VectorSearchTypePinecone)
	assert.Contains(t, supportedTypes, types.VectorSearchTypeWeaviate)
	assert.Contains(t, supportedTypes, types.VectorSearchTypeChroma)
	assert.Contains(t, supportedTypes, types.VectorSearchTypeQdrant)
}

func TestFactory_GetSupportedEmbeddingProviders(t *testing.T) {
	factory := NewFactory()
	providers := factory.GetSupportedEmbeddingProviders()
	
	assert.Contains(t, providers, types.EmbeddingProviderNone)
	assert.Contains(t, providers, types.EmbeddingProviderOpenAI)
	assert.Contains(t, providers, types.EmbeddingProviderAzure)
	assert.Contains(t, providers, types.EmbeddingProviderHugging)
	assert.Contains(t, providers, types.EmbeddingProviderCohere)
	assert.Contains(t, providers, types.EmbeddingProviderLocal)
}

func TestValidateLocalConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.LocalVectorConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: types.LocalVectorConfig{
				DatabasePath:   "/tmp/vector.db",
				Engine:         "sqlite",
				IndexType:      "flat",
				DistanceMetric: "cosine",
			},
			wantErr: false,
		},
		{
			name: "empty database path",
			config: types.LocalVectorConfig{
				Engine:         "sqlite",
				IndexType:      "flat",
				DistanceMetric: "cosine",
			},
			wantErr: true,
			errMsg:  "database path cannot be empty",
		},
		{
			name: "invalid engine",
			config: types.LocalVectorConfig{
				DatabasePath:   "/tmp/vector.db",
				Engine:         "invalid",
				IndexType:      "flat",
				DistanceMetric: "cosine",
			},
			wantErr: true,
			errMsg:  "unsupported local vector engine",
		},
		{
			name: "invalid index type",
			config: types.LocalVectorConfig{
				DatabasePath:   "/tmp/vector.db",
				Engine:         "sqlite",
				IndexType:      "invalid",
				DistanceMetric: "cosine",
			},
			wantErr: true,
			errMsg:  "unsupported index type",
		},
		{
			name: "invalid distance metric",
			config: types.LocalVectorConfig{
				DatabasePath:   "/tmp/vector.db",
				Engine:         "sqlite",
				IndexType:      "flat",
				DistanceMetric: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported distance metric",
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

func TestValidatePineconeConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.PineconeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: types.PineconeConfig{
				APIKey:      "test-key",
				Environment: "us-west1-gcp",
				IndexName:   "test-index",
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: types.PineconeConfig{
				Environment: "us-west1-gcp",
				IndexName:   "test-index",
			},
			wantErr: true,
			errMsg:  "API key cannot be empty",
		},
		{
			name: "empty environment",
			config: types.PineconeConfig{
				APIKey:    "test-key",
				IndexName: "test-index",
			},
			wantErr: true,
			errMsg:  "environment cannot be empty",
		},
		{
			name: "empty index name",
			config: types.PineconeConfig{
				APIKey:      "test-key",
				Environment: "us-west1-gcp",
			},
			wantErr: true,
			errMsg:  "index name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePineconeConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultFactory(t *testing.T) {
	assert.NotNil(t, DefaultFactory)
}

func TestPackageFunctions(t *testing.T) {
	// Test that package-level functions work
	config := types.VectorSearchConfig{
		Enabled: false,
		Type:    types.VectorSearchTypeNone,
	}

	// Test CreateVectorSearch
	backend, err := CreateVectorSearch(config)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.Equal(t, types.VectorSearchTypeNone, backend.Type())

	// Test ValidateConfig
	err = ValidateConfig(config)
	assert.NoError(t, err)

	// Test GetSupportedTypes
	supportedTypes := GetSupportedTypes()
	assert.Contains(t, supportedTypes, types.VectorSearchTypeNone)
	assert.Contains(t, supportedTypes, types.VectorSearchTypeLocal)

	// Test GetSupportedEmbeddingProviders
	supportedProviders := GetSupportedEmbeddingProviders()
	assert.Contains(t, supportedProviders, types.EmbeddingProviderNone)
	assert.Contains(t, supportedProviders, types.EmbeddingProviderOpenAI)
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	
	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.True(t, contains(slice, "c"))
	assert.False(t, contains(slice, "d"))
	assert.False(t, contains(slice, ""))
}