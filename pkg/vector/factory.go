package vector

import (
	"fmt"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Factory creates vector search backends based on configuration
type Factory struct{}

// NewFactory creates a new vector search factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateVectorSearch creates a vector search backend based on the provided configuration
func (f *Factory) CreateVectorSearch(config types.VectorSearchConfig) (types.VectorSearchBackend, error) {
	if !config.Enabled {
		return NewNoneBackend(), nil
	}

	switch config.Type {
	case types.VectorSearchTypeNone:
		return NewNoneBackend(), nil
	case types.VectorSearchTypeLocal:
		return NewLocalBackend(config)
	case types.VectorSearchTypePinecone:
		return NewPineconeBackend(config)
	case types.VectorSearchTypeWeaviate:
		return NewWeaviateBackend(config)
	case types.VectorSearchTypeChroma:
		return NewChromaBackend(config)
	case types.VectorSearchTypeQdrant:
		return NewQdrantBackend(config)
	default:
		return nil, fmt.Errorf("unsupported vector search type: %s", config.Type)
	}
}

// ValidateConfig validates a vector search configuration without creating the backend
func (f *Factory) ValidateConfig(config types.VectorSearchConfig) error {
	if !config.Enabled {
		return nil // No validation needed for disabled vector search
	}

	switch config.Type {
	case types.VectorSearchTypeNone:
		return nil
	case types.VectorSearchTypeLocal:
		return validateLocalConfig(config.Local)
	case types.VectorSearchTypePinecone:
		return validatePineconeConfig(config.Pinecone)
	case types.VectorSearchTypeWeaviate:
		return validateWeaviateConfig(config.Weaviate)
	case types.VectorSearchTypeChroma:
		return validateChromaConfig(config.Chroma)
	case types.VectorSearchTypeQdrant:
		return validateQdrantConfig(config.Qdrant)
	default:
		return fmt.Errorf("unsupported vector search type: %s", config.Type)
	}
}

// GetSupportedTypes returns a list of supported vector search types
func (f *Factory) GetSupportedTypes() []types.VectorSearchType {
	return []types.VectorSearchType{
		types.VectorSearchTypeNone,
		types.VectorSearchTypeLocal,
		types.VectorSearchTypePinecone,
		types.VectorSearchTypeWeaviate,
		types.VectorSearchTypeChroma,
		types.VectorSearchTypeQdrant,
	}
}

// GetSupportedEmbeddingProviders returns a list of supported embedding providers
func (f *Factory) GetSupportedEmbeddingProviders() []types.EmbeddingProvider {
	return []types.EmbeddingProvider{
		types.EmbeddingProviderNone,
		types.EmbeddingProviderOpenAI,
		types.EmbeddingProviderAzure,
		types.EmbeddingProviderHugging,
		types.EmbeddingProviderCohere,
		types.EmbeddingProviderLocal,
	}
}

// Validation functions for different backend types

func validateLocalConfig(config types.LocalVectorConfig) error {
	if config.DatabasePath == "" {
		return fmt.Errorf("local vector database path cannot be empty")
	}
	
	validEngines := []string{"sqlite", "duckdb"}
	if !contains(validEngines, config.Engine) {
		return fmt.Errorf("unsupported local vector engine: %s (supported: %v)", config.Engine, validEngines)
	}
	
	validIndexTypes := []string{"flat", "ivf", "hnsw"}
	if !contains(validIndexTypes, config.IndexType) {
		return fmt.Errorf("unsupported index type: %s (supported: %v)", config.IndexType, validIndexTypes)
	}
	
	validMetrics := []string{"cosine", "euclidean", "dot"}
	if !contains(validMetrics, config.DistanceMetric) {
		return fmt.Errorf("unsupported distance metric: %s (supported: %v)", config.DistanceMetric, validMetrics)
	}
	
	return nil
}

func validatePineconeConfig(config types.PineconeConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("Pinecone API key cannot be empty")
	}
	if config.Environment == "" {
		return fmt.Errorf("Pinecone environment cannot be empty")
	}
	if config.IndexName == "" {
		return fmt.Errorf("Pinecone index name cannot be empty")
	}
	return nil
}

func validateWeaviateConfig(config types.WeaviateConfig) error {
	if config.Host == "" {
		return fmt.Errorf("Weaviate host cannot be empty")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("Weaviate port must be between 1 and 65535")
	}
	if config.ClassName == "" {
		return fmt.Errorf("Weaviate class name cannot be empty")
	}
	return nil
}

func validateChromaConfig(config types.ChromaConfig) error {
	if config.Host == "" {
		return fmt.Errorf("Chroma host cannot be empty")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("Chroma port must be between 1 and 65535")
	}
	if config.CollectionName == "" {
		return fmt.Errorf("Chroma collection name cannot be empty")
	}
	return nil
}

func validateQdrantConfig(config types.QdrantConfig) error {
	if config.Host == "" {
		return fmt.Errorf("Qdrant host cannot be empty")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("Qdrant port must be between 1 and 65535")
	}
	if config.CollectionName == "" {
		return fmt.Errorf("Qdrant collection name cannot be empty")
	}
	return nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Backend creation functions (these will be implemented in Session 6)

// NewNoneBackend creates a no-op vector search backend
func NewNoneBackend() types.VectorSearchBackend {
	return &NoneBackend{}
}

// NewLocalBackend creates a local vector search backend
func NewLocalBackend(config types.VectorSearchConfig) (types.VectorSearchBackend, error) {
	// TODO: Implement in Session 6
	return nil, fmt.Errorf("local vector search backend not yet implemented")
}

// NewPineconeBackend creates a Pinecone vector search backend
func NewPineconeBackend(config types.VectorSearchConfig) (types.VectorSearchBackend, error) {
	// TODO: Implement in Session 6
	return nil, fmt.Errorf("Pinecone vector search backend not yet implemented")
}

// NewWeaviateBackend creates a Weaviate vector search backend
func NewWeaviateBackend(config types.VectorSearchConfig) (types.VectorSearchBackend, error) {
	// TODO: Implement in Session 6
	return nil, fmt.Errorf("Weaviate vector search backend not yet implemented")
}

// NewChromaBackend creates a Chroma vector search backend
func NewChromaBackend(config types.VectorSearchConfig) (types.VectorSearchBackend, error) {
	// TODO: Implement in Session 6
	return nil, fmt.Errorf("Chroma vector search backend not yet implemented")
}

// NewQdrantBackend creates a Qdrant vector search backend
func NewQdrantBackend(config types.VectorSearchConfig) (types.VectorSearchBackend, error) {
	// TODO: Implement in Session 6
	return nil, fmt.Errorf("Qdrant vector search backend not yet implemented")
}

// DefaultFactory is the default vector search factory instance
var DefaultFactory = NewFactory()

// CreateVectorSearch creates a vector search backend using the default factory
func CreateVectorSearch(config types.VectorSearchConfig) (types.VectorSearchBackend, error) {
	return DefaultFactory.CreateVectorSearch(config)
}

// ValidateConfig validates a vector search configuration using the default factory
func ValidateConfig(config types.VectorSearchConfig) error {
	return DefaultFactory.ValidateConfig(config)
}

// GetSupportedTypes returns supported vector search types using the default factory
func GetSupportedTypes() []types.VectorSearchType {
	return DefaultFactory.GetSupportedTypes()
}

// GetSupportedEmbeddingProviders returns supported embedding providers using the default factory
func GetSupportedEmbeddingProviders() []types.EmbeddingProvider {
	return DefaultFactory.GetSupportedEmbeddingProviders()
}