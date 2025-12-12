package search

import (
	"fmt"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/vector/cache"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/vector/ollama"
)

// SemanticEngineFactory creates SemanticEngine instances with configured backends
type SemanticEngineFactory struct {
	// DefaultOllamaEndpoint is the default Ollama API endpoint
	DefaultOllamaEndpoint string

	// DefaultOllamaModel is the default embedding model
	DefaultOllamaModel string

	// DefaultDimensions is the default embedding dimension
	DefaultDimensions int

	// SupportedModels maps model names to their dimensions
	SupportedModels map[string]int
}

// NewSemanticEngineFactory creates a new semantic search factory with sensible defaults
func NewSemanticEngineFactory() *SemanticEngineFactory {
	return &SemanticEngineFactory{
		DefaultOllamaEndpoint: "http://localhost:11434",
		DefaultOllamaModel:    "nomic-embed-text",
		DefaultDimensions:     384,
		SupportedModels: map[string]int{
			"nomic-embed-text":     384,
			"bge-small":            384,
			"bge-base":             768,
			"bge-large":            1024,
			"all-minilm-l6-v2":     384,
			"all-mpnet-base-v2":    768,
			"e5-small-v2":          384,
			"e5-base-v2":           768,
			"e5-large-v2":          1024,
			"nomic-embed-text-v15": 384,
		},
	}
}

// CreateSemanticEngine creates a configured SemanticEngine with S3 Vectors backend
func (f *SemanticEngineFactory) CreateSemanticEngine(cfg *types.VectorSearchConfig) (*SemanticEngine, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("vector search is not enabled in configuration")
	}

	// Validate configuration
	if err := f.ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid semantic search configuration: %w", err)
	}

	// Determine endpoint and model from config
	endpoint := f.DefaultOllamaEndpoint
	model := f.DefaultOllamaModel

	if cfg.Embedding.Local.ModelPath != "" {
		// Use custom endpoint if specified
		endpoint = cfg.Embedding.Local.ModelPath
	}

	if cfg.Embedding.Model != "" {
		model = cfg.Embedding.Model
	}

	// Create Ollama embedder client
	embedder := ollama.NewClient(endpoint, model)

	// Validate model dimensions
	dimensions, ok := f.SupportedModels[model]
	if !ok {
		// Use config dimension if model not in registry, or default
		if cfg.Embedding.Dimensions > 0 {
			dimensions = cfg.Embedding.Dimensions
		} else {
			dimensions = f.DefaultDimensions
		}
	}

	// Verify dimension consistency
	if cfg.Embedding.Dimensions > 0 && cfg.Embedding.Dimensions != dimensions {
		return nil, fmt.Errorf("dimension mismatch: config specifies %d but model %s has %d dimensions",
			cfg.Embedding.Dimensions, model, dimensions)
	}

	// Create cache with configured or default settings
	cacheSize := 1000

	if cfg.Indexing.BatchSize > 0 {
		cacheSize = cfg.Indexing.BatchSize
	}

	cacheInstance := cache.NewCache(cacheSize, 0) // No TTL by default

	// Create semantic engine with embedder and cache
	// Note: S3 Vectors backend will be configured separately in a future session
	// For now, we instantiate with nil backend and cache for query caching
	engine := NewSemanticEngine(embedder, nil, cacheInstance)

	return engine, nil
}

// ValidateConfig validates the semantic search configuration
func (f *SemanticEngineFactory) ValidateConfig(cfg *types.VectorSearchConfig) error {
	if cfg == nil {
		return fmt.Errorf("configuration is nil")
	}

	if !cfg.Enabled {
		return fmt.Errorf("vector search is not enabled")
	}

	// Validate embedding configuration
	if err := f.ValidateEmbeddingConfig(&cfg.Embedding); err != nil {
		return fmt.Errorf("invalid embedding configuration: %w", err)
	}

	// Validate indexing configuration if present
	if cfg.Indexing.BatchSize > 0 && cfg.Indexing.BatchSize < 1 {
		return fmt.Errorf("batch size must be positive")
	}

	// Validate search configuration if present
	if cfg.Search.MinScore < 0 || cfg.Search.MinScore > 1 {
		return fmt.Errorf("min_score must be between 0 and 1")
	}

	if cfg.Search.DefaultLimit > cfg.Search.MaxLimit {
		return fmt.Errorf("default_limit cannot be greater than max_limit")
	}

	return nil
}

// ValidateEmbeddingConfig validates embedding configuration specifically for Ollama/Local
func (f *SemanticEngineFactory) ValidateEmbeddingConfig(cfg *types.EmbeddingConfig) error {
	if cfg == nil {
		return fmt.Errorf("embedding configuration is nil")
	}

	// For local embeddings (Ollama), check model is known or dimensions are specified
	if cfg.Provider == types.EmbeddingProviderLocal {
		if cfg.Model == "" {
			return fmt.Errorf("embedding model name is required for local provider")
		}

		// Check if model is supported
		if _, ok := f.SupportedModels[cfg.Model]; !ok {
			// Allow unknown models if dimensions are specified
			if cfg.Dimensions <= 0 {
				return fmt.Errorf("model %q not in supported registry; please specify dimensions explicitly", cfg.Model)
			}
		}

		// Validate dimensions if specified
		if cfg.Dimensions > 0 {
			if cfg.Dimensions < 64 || cfg.Dimensions > 4096 {
				return fmt.Errorf("dimensions must be between 64 and 4096, got %d", cfg.Dimensions)
			}
		}
	}

	return nil
}

// GetSupportedModels returns the list of supported embedding models
func (f *SemanticEngineFactory) GetSupportedModels() map[string]int {
	result := make(map[string]int)
	for k, v := range f.SupportedModels {
		result[k] = v
	}
	return result
}

// GetDefaultDimensions returns the default dimension for a model, or the factory default
func (f *SemanticEngineFactory) GetDefaultDimensions(model string) int {
	if dim, ok := f.SupportedModels[model]; ok {
		return dim
	}
	return f.DefaultDimensions
}

// SetOllamaEndpoint sets the default Ollama endpoint
func (f *SemanticEngineFactory) SetOllamaEndpoint(endpoint string) {
	if endpoint != "" {
		f.DefaultOllamaEndpoint = endpoint
	}
}

// SetOllamaModel sets the default embedding model
func (f *SemanticEngineFactory) SetOllamaModel(model string) {
	if model != "" {
		f.DefaultOllamaModel = model
	}
}

// DefaultSemanticEngineFactory is the default factory instance
var DefaultSemanticEngineFactory = NewSemanticEngineFactory()

// CreateSemanticEngine creates a semantic engine using the default factory
func CreateSemanticEngine(cfg *types.VectorSearchConfig) (*SemanticEngine, error) {
	return DefaultSemanticEngineFactory.CreateSemanticEngine(cfg)
}
