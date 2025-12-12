package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNewSemanticEngineFactory(t *testing.T) {
	factory := NewSemanticEngineFactory()

	assert.NotNil(t, factory)
	assert.Equal(t, "http://localhost:11434", factory.DefaultOllamaEndpoint)
	assert.Equal(t, "nomic-embed-text", factory.DefaultOllamaModel)
	assert.Equal(t, 384, factory.DefaultDimensions)
	assert.NotEmpty(t, factory.SupportedModels)
}

func TestSemanticEngineFactory_ValidateConfig_Enabled(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.VectorSearchConfig{
		Enabled: true,
		Type:    types.VectorSearchTypeLocal,
		Embedding: types.EmbeddingConfig{
			Provider:   types.EmbeddingProviderLocal,
			Model:      "nomic-embed-text",
			Dimensions: 384,
		},
	}

	err := factory.ValidateConfig(config)
	assert.NoError(t, err)
}

func TestSemanticEngineFactory_ValidateConfig_Disabled(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.VectorSearchConfig{
		Enabled: false,
	}

	err := factory.ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestSemanticEngineFactory_ValidateConfig_NilConfig(t *testing.T) {
	factory := NewSemanticEngineFactory()

	err := factory.ValidateConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestSemanticEngineFactory_ValidateEmbeddingConfig_KnownModel(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.EmbeddingConfig{
		Provider:   types.EmbeddingProviderLocal,
		Model:      "nomic-embed-text",
		Dimensions: 384,
	}

	err := factory.ValidateEmbeddingConfig(config)
	assert.NoError(t, err)
}

func TestSemanticEngineFactory_ValidateEmbeddingConfig_UnknownModelNoDimensions(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.EmbeddingConfig{
		Provider: types.EmbeddingProviderLocal,
		Model:    "unknown-model",
		// No dimensions specified
	}

	err := factory.ValidateEmbeddingConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in supported registry")
}

func TestSemanticEngineFactory_ValidateEmbeddingConfig_UnknownModelWithDimensions(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.EmbeddingConfig{
		Provider:   types.EmbeddingProviderLocal,
		Model:      "custom-model",
		Dimensions: 512,
	}

	err := factory.ValidateEmbeddingConfig(config)
	assert.NoError(t, err)
}

func TestSemanticEngineFactory_ValidateEmbeddingConfig_InvalidDimensions(t *testing.T) {
	tests := []struct {
		name       string
		dimensions int
		expectErr  bool
	}{
		{"too_small", 32, true},
		{"too_large", 5000, true},
		{"valid_low", 64, false},
		{"valid_mid", 384, false},
		{"valid_high", 4096, false},
	}

	factory := NewSemanticEngineFactory()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.EmbeddingConfig{
				Provider:   types.EmbeddingProviderLocal,
				Model:      "custom-model",
				Dimensions: tt.dimensions,
			}

			err := factory.ValidateEmbeddingConfig(config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSemanticEngineFactory_ValidateEmbeddingConfig_NoModel(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.EmbeddingConfig{
		Provider: types.EmbeddingProviderLocal,
		Model:    "", // Empty model
	}

	err := factory.ValidateEmbeddingConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model name is required")
}

func TestSemanticEngineFactory_ValidateConfig_InvalidSearchConfig(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.VectorSearchConfig{
		Enabled: true,
		Type:    types.VectorSearchTypeLocal,
		Embedding: types.EmbeddingConfig{
			Provider: types.EmbeddingProviderLocal,
			Model:    "nomic-embed-text",
		},
		Search: types.SearchConfig{
			MinScore: 1.5, // Invalid: > 1.0
		},
	}

	err := factory.ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min_score")
}

func TestSemanticEngineFactory_ValidateConfig_DefaultLimitGreaterThanMax(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.VectorSearchConfig{
		Enabled: true,
		Type:    types.VectorSearchTypeLocal,
		Embedding: types.EmbeddingConfig{
			Provider: types.EmbeddingProviderLocal,
			Model:    "nomic-embed-text",
		},
		Search: types.SearchConfig{
			DefaultLimit: 200,
			MaxLimit:     100,
		},
	}

	err := factory.ValidateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default_limit")
}

func TestSemanticEngineFactory_GetSupportedModels(t *testing.T) {
	factory := NewSemanticEngineFactory()

	models := factory.GetSupportedModels()

	assert.NotNil(t, models)
	assert.Greater(t, len(models), 0)
	assert.Equal(t, 384, models["nomic-embed-text"])
	assert.Equal(t, 768, models["bge-base"])
}

func TestSemanticEngineFactory_GetDefaultDimensions(t *testing.T) {
	factory := NewSemanticEngineFactory()

	tests := []struct {
		model    string
		expected int
	}{
		{"nomic-embed-text", 384},
		{"bge-base", 768},
		{"unknown-model", 384}, // Falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			dims := factory.GetDefaultDimensions(tt.model)
			assert.Equal(t, tt.expected, dims)
		})
	}
}

func TestSemanticEngineFactory_SetOllamaEndpoint(t *testing.T) {
	factory := NewSemanticEngineFactory()
	originalEndpoint := factory.DefaultOllamaEndpoint

	factory.SetOllamaEndpoint("http://custom:11434")
	assert.Equal(t, "http://custom:11434", factory.DefaultOllamaEndpoint)

	// Empty endpoint should not change
	factory.SetOllamaEndpoint("")
	assert.Equal(t, "http://custom:11434", factory.DefaultOllamaEndpoint)

	// Reset
	factory.DefaultOllamaEndpoint = originalEndpoint
}

func TestSemanticEngineFactory_SetOllamaModel(t *testing.T) {
	factory := NewSemanticEngineFactory()
	originalModel := factory.DefaultOllamaModel

	factory.SetOllamaModel("bge-large")
	assert.Equal(t, "bge-large", factory.DefaultOllamaModel)

	// Empty model should not change
	factory.SetOllamaModel("")
	assert.Equal(t, "bge-large", factory.DefaultOllamaModel)

	// Reset
	factory.DefaultOllamaModel = originalModel
}

func TestSemanticEngineFactory_CreateSemanticEngine_Success(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.VectorSearchConfig{
		Enabled: true,
		Type:    types.VectorSearchTypeLocal,
		Embedding: types.EmbeddingConfig{
			Provider:   types.EmbeddingProviderLocal,
			Model:      "nomic-embed-text",
			Dimensions: 384,
		},
		Indexing: types.IndexingConfig{
			BatchSize: 500,
		},
	}

	engine, err := factory.CreateSemanticEngine(config)
	require.NoError(t, err)
	assert.NotNil(t, engine)
}

func TestSemanticEngineFactory_CreateSemanticEngine_Disabled(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.VectorSearchConfig{
		Enabled: false,
	}

	engine, err := factory.CreateSemanticEngine(config)
	assert.Error(t, err)
	assert.Nil(t, engine)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestSemanticEngineFactory_CreateSemanticEngine_InvalidConfig(t *testing.T) {
	factory := NewSemanticEngineFactory()

	config := &types.VectorSearchConfig{
		Enabled: true,
		Type:    types.VectorSearchTypeLocal,
		Embedding: types.EmbeddingConfig{
			Provider: types.EmbeddingProviderLocal,
			Model:    "", // Invalid: no model specified
		},
	}

	engine, err := factory.CreateSemanticEngine(config)
	assert.Error(t, err)
	assert.Nil(t, engine)
}

func TestDefaultSemanticEngineFactory_CreateSemanticEngine(t *testing.T) {
	config := &types.VectorSearchConfig{
		Enabled: true,
		Type:    types.VectorSearchTypeLocal,
		Embedding: types.EmbeddingConfig{
			Provider:   types.EmbeddingProviderLocal,
			Model:      "nomic-embed-text",
			Dimensions: 384,
		},
	}

	engine, err := CreateSemanticEngine(config)
	require.NoError(t, err)
	assert.NotNil(t, engine)
}
