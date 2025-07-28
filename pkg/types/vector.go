package types

import (
	"context"
	"fmt"
	"time"
)

// VectorSearchBackend defines the interface for vector search implementations
type VectorSearchBackend interface {
	// Type returns the vector search backend type
	Type() VectorSearchType

	// IndexDocument adds or updates a document in the vector index
	IndexDocument(ctx context.Context, doc *Document) error

	// IndexDocuments adds or updates multiple documents in the vector index
	IndexDocuments(ctx context.Context, docs []*Document) error

	// DeleteDocument removes a document from the vector index
	DeleteDocument(ctx context.Context, id string) error

	// Search performs semantic search and returns similar documents
	Search(ctx context.Context, query *VectorQuery) (*VectorSearchResults, error)

	// GetEmbedding generates an embedding for the given text
	GetEmbedding(ctx context.Context, text string) ([]float64, error)

	// GetEmbeddings generates embeddings for multiple texts
	GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error)

	// Health performs a health check on the vector search backend
	Health(ctx context.Context) error

	// Close cleanly shuts down the vector search backend
	Close() error
}

// VectorSearchType represents the type of vector search backend
type VectorSearchType string

const (
	VectorSearchTypeNone     VectorSearchType = "none"
	VectorSearchTypeLocal    VectorSearchType = "local"     // Local vector database (e.g., SQLite with vector extension)
	VectorSearchTypePinecone VectorSearchType = "pinecone"  // Pinecone vector database
	VectorSearchTypeWeaviate VectorSearchType = "weaviate"  // Weaviate vector database
	VectorSearchTypeChroma   VectorSearchType = "chroma"    // Chroma vector database
	VectorSearchTypeQdrant   VectorSearchType = "qdrant"    // Qdrant vector database
)

// EmbeddingProvider represents the type of embedding provider
type EmbeddingProvider string

const (
	EmbeddingProviderNone     EmbeddingProvider = "none"
	EmbeddingProviderOpenAI   EmbeddingProvider = "openai"
	EmbeddingProviderAzure    EmbeddingProvider = "azure"
	EmbeddingProviderHugging  EmbeddingProvider = "huggingface"
	EmbeddingProviderCohere   EmbeddingProvider = "cohere"
	EmbeddingProviderLocal    EmbeddingProvider = "local"    // Local embedding model
)

// Document represents a document for vector indexing
type Document struct {
	// ID is the unique identifier for the document
	ID string `json:"id"`

	// Content is the main text content to be indexed
	Content string `json:"content"`

	// Title is the document title (optional)
	Title string `json:"title,omitempty"`

	// Path is the file path or location
	Path string `json:"path,omitempty"`

	// Metadata contains additional document metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Tags associated with the document
	Tags []string `json:"tags,omitempty"`

	// CreatedAt is when the document was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the document was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Embedding is the vector representation (may be computed later)
	Embedding []float64 `json:"embedding,omitempty"`

	// ChunkSize indicates if this document was chunked
	ChunkSize int `json:"chunk_size,omitempty"`

	// ChunkIndex indicates which chunk this is (for chunked documents)
	ChunkIndex int `json:"chunk_index,omitempty"`

	// ParentID links to the original document (for chunks)
	ParentID string `json:"parent_id,omitempty"`
}

// VectorQuery represents a vector search query
type VectorQuery struct {
	// Query is the search text
	Query string `json:"query"`

	// QueryEmbedding is the pre-computed embedding (optional)
	QueryEmbedding []float64 `json:"query_embedding,omitempty"`

	// Limit is the maximum number of results to return
	Limit int `json:"limit"`

	// MinScore is the minimum similarity score (0.0 to 1.0)
	MinScore float64 `json:"min_score"`

	// Filters for metadata-based filtering
	Filters map[string]interface{} `json:"filters,omitempty"`

	// Tags for tag-based filtering
	Tags []string `json:"tags,omitempty"`

	// IncludeContent whether to include document content in results
	IncludeContent bool `json:"include_content"`

	// IncludeEmbeddings whether to include embeddings in results
	IncludeEmbeddings bool `json:"include_embeddings"`
}

// VectorSearchResults represents the results of a vector search
type VectorSearchResults struct {
	// Results are the matching documents with scores
	Results []*VectorSearchResult `json:"results"`

	// Total is the total number of matches (may be higher than len(Results))
	Total int `json:"total"`

	// QueryTime is how long the search took
	QueryTime time.Duration `json:"query_time"`

	// Query is the original query
	Query string `json:"query"`
}

// VectorSearchResult represents a single search result
type VectorSearchResult struct {
	// Document is the matching document
	Document *Document `json:"document"`

	// Score is the similarity score (0.0 to 1.0, higher is more similar)
	Score float64 `json:"score"`

	// Distance is the vector distance (lower is more similar)
	Distance float64 `json:"distance"`

	// Highlights contains text highlights (if supported)
	Highlights []string `json:"highlights,omitempty"`
}

// VectorSearchConfig contains configuration for vector search
type VectorSearchConfig struct {
	// Enabled turns on/off vector search capabilities
	Enabled bool `toml:"enabled" json:"enabled"`

	// Type specifies which vector search backend to use
	Type VectorSearchType `toml:"type" json:"type"`

	// Embedding configuration
	Embedding EmbeddingConfig `toml:"embedding" json:"embedding"`

	// Local vector search configuration
	Local LocalVectorConfig `toml:"local" json:"local"`

	// Pinecone configuration
	Pinecone PineconeConfig `toml:"pinecone" json:"pinecone"`

	// Weaviate configuration
	Weaviate WeaviateConfig `toml:"weaviate" json:"weaviate"`

	// Chroma configuration
	Chroma ChromaConfig `toml:"chroma" json:"chroma"`

	// Qdrant configuration
	Qdrant QdrantConfig `toml:"qdrant" json:"qdrant"`

	// Indexing configuration
	Indexing IndexingConfig `toml:"indexing" json:"indexing"`

	// Search configuration
	Search SearchConfig `toml:"search" json:"search"`
}

// EmbeddingConfig configures embedding generation
type EmbeddingConfig struct {
	// Provider specifies the embedding provider
	Provider EmbeddingProvider `toml:"provider" json:"provider"`

	// Model is the embedding model to use
	Model string `toml:"model" json:"model"`

	// Dimensions is the embedding vector dimensions
	Dimensions int `toml:"dimensions" json:"dimensions"`

	// OpenAI configuration
	OpenAI OpenAIEmbeddingConfig `toml:"openai" json:"openai"`

	// Azure configuration
	Azure AzureEmbeddingConfig `toml:"azure" json:"azure"`

	// Hugging Face configuration
	HuggingFace HuggingFaceEmbeddingConfig `toml:"huggingface" json:"huggingface"`

	// Cohere configuration
	Cohere CohereEmbeddingConfig `toml:"cohere" json:"cohere"`

	// Local embedding configuration
	Local LocalEmbeddingConfig `toml:"local" json:"local"`
}

// OpenAIEmbeddingConfig configures OpenAI embeddings
type OpenAIEmbeddingConfig struct {
	// APIKey for OpenAI API
	APIKey string `toml:"api_key" json:"api_key"`

	// Organization ID (optional)
	Organization string `toml:"organization" json:"organization"`

	// BaseURL for custom OpenAI-compatible endpoints
	BaseURL string `toml:"base_url" json:"base_url"`

	// Model name (e.g., "text-embedding-3-small")
	Model string `toml:"model" json:"model"`

	// RequestTimeout for API calls (seconds)
	RequestTimeout int `toml:"request_timeout" json:"request_timeout"`

	// MaxRetries for failed requests
	MaxRetries int `toml:"max_retries" json:"max_retries"`
}

// AzureEmbeddingConfig configures Azure OpenAI embeddings
type AzureEmbeddingConfig struct {
	// Endpoint is the Azure OpenAI endpoint
	Endpoint string `toml:"endpoint" json:"endpoint"`

	// APIKey for Azure OpenAI
	APIKey string `toml:"api_key" json:"api_key"`

	// DeploymentName is the model deployment name
	DeploymentName string `toml:"deployment_name" json:"deployment_name"`

	// APIVersion for Azure OpenAI API
	APIVersion string `toml:"api_version" json:"api_version"`
}

// HuggingFaceEmbeddingConfig configures Hugging Face embeddings
type HuggingFaceEmbeddingConfig struct {
	// APIKey for Hugging Face API
	APIKey string `toml:"api_key" json:"api_key"`

	// Model name on Hugging Face Hub
	Model string `toml:"model" json:"model"`

	// EndpointURL for custom endpoints
	EndpointURL string `toml:"endpoint_url" json:"endpoint_url"`
}

// CohereEmbeddingConfig configures Cohere embeddings
type CohereEmbeddingConfig struct {
	// APIKey for Cohere API
	APIKey string `toml:"api_key" json:"api_key"`

	// Model name
	Model string `toml:"model" json:"model"`

	// InputType for embeddings
	InputType string `toml:"input_type" json:"input_type"`
}

// LocalEmbeddingConfig configures local embedding models
type LocalEmbeddingConfig struct {
	// ModelPath to the local model files
	ModelPath string `toml:"model_path" json:"model_path"`

	// ModelType (e.g., "sentence-transformers", "onnx")
	ModelType string `toml:"model_type" json:"model_type"`

	// Device to run on ("cpu", "cuda", "mps")
	Device string `toml:"device" json:"device"`

	// BatchSize for processing multiple texts
	BatchSize int `toml:"batch_size" json:"batch_size"`
}

// LocalVectorConfig configures local vector search
type LocalVectorConfig struct {
	// DatabasePath for the local vector database
	DatabasePath string `toml:"database_path" json:"database_path"`

	// Engine type ("sqlite", "duckdb")
	Engine string `toml:"engine" json:"engine"`

	// IndexType for vector indexing ("flat", "ivf", "hnsw")
	IndexType string `toml:"index_type" json:"index_type"`

	// Distance metric ("cosine", "euclidean", "dot")
	DistanceMetric string `toml:"distance_metric" json:"distance_metric"`
}

// PineconeConfig configures Pinecone vector database
type PineconeConfig struct {
	// APIKey for Pinecone
	APIKey string `toml:"api_key" json:"api_key"`

	// Environment (e.g., "us-west1-gcp")
	Environment string `toml:"environment" json:"environment"`

	// IndexName in Pinecone
	IndexName string `toml:"index_name" json:"index_name"`

	// ProjectID in Pinecone
	ProjectID string `toml:"project_id" json:"project_id"`

	// Namespace for organizing vectors
	Namespace string `toml:"namespace" json:"namespace"`
}

// WeaviateConfig configures Weaviate vector database
type WeaviateConfig struct {
	// Host is the Weaviate server host
	Host string `toml:"host" json:"host"`

	// Port is the Weaviate server port
	Port int `toml:"port" json:"port"`

	// Scheme ("http" or "https")
	Scheme string `toml:"scheme" json:"scheme"`

	// APIKey for authentication
	APIKey string `toml:"api_key" json:"api_key"`

	// ClassName in Weaviate
	ClassName string `toml:"class_name" json:"class_name"`
}

// ChromaConfig configures Chroma vector database
type ChromaConfig struct {
	// Host is the Chroma server host
	Host string `toml:"host" json:"host"`

	// Port is the Chroma server port
	Port int `toml:"port" json:"port"`

	// UseSSL whether to use HTTPS
	UseSSL bool `toml:"use_ssl" json:"use_ssl"`

	// APIKey for authentication
	APIKey string `toml:"api_key" json:"api_key"`

	// CollectionName in Chroma
	CollectionName string `toml:"collection_name" json:"collection_name"`

	// PersistDirectory for local Chroma instances
	PersistDirectory string `toml:"persist_directory" json:"persist_directory"`
}

// QdrantConfig configures Qdrant vector database
type QdrantConfig struct {
	// Host is the Qdrant server host
	Host string `toml:"host" json:"host"`

	// Port is the Qdrant server port
	Port int `toml:"port" json:"port"`

	// UseSSL whether to use HTTPS
	UseSSL bool `toml:"use_ssl" json:"use_ssl"`

	// APIKey for authentication
	APIKey string `toml:"api_key" json:"api_key"`

	// CollectionName in Qdrant
	CollectionName string `toml:"collection_name" json:"collection_name"`
}

// IndexingConfig configures document indexing behavior
type IndexingConfig struct {
	// AutoIndex automatically indexes new documents
	AutoIndex bool `toml:"auto_index" json:"auto_index"`

	// ChunkSize for splitting large documents (0 = no chunking)
	ChunkSize int `toml:"chunk_size" json:"chunk_size"`

	// ChunkOverlap for overlapping chunks
	ChunkOverlap int `toml:"chunk_overlap" json:"chunk_overlap"`

	// MinChunkSize minimum size for chunks
	MinChunkSize int `toml:"min_chunk_size" json:"min_chunk_size"`

	// IncludeMetadata whether to index metadata
	IncludeMetadata bool `toml:"include_metadata" json:"include_metadata"`

	// BatchSize for bulk indexing operations
	BatchSize int `toml:"batch_size" json:"batch_size"`

	// RefreshInterval for index updates (seconds)
	RefreshInterval int `toml:"refresh_interval" json:"refresh_interval"`
}

// SearchConfig configures search behavior
type SearchConfig struct {
	// HybridEnabled combines text and vector search
	HybridEnabled bool `toml:"hybrid_enabled" json:"hybrid_enabled"`

	// HybridWeight balances text vs vector search (0.0 = all text, 1.0 = all vector)
	HybridWeight float64 `toml:"hybrid_weight" json:"hybrid_weight"`

	// DefaultLimit for search results
	DefaultLimit int `toml:"default_limit" json:"default_limit"`

	// MaxLimit maximum allowed results
	MaxLimit int `toml:"max_limit" json:"max_limit"`

	// MinScore minimum similarity score for results
	MinScore float64 `toml:"min_score" json:"min_score"`

	// EnableReranking whether to rerank results
	EnableReranking bool `toml:"enable_reranking" json:"enable_reranking"`

	// RerankingModel for reranking results
	RerankingModel string `toml:"reranking_model" json:"reranking_model"`
}

// DefaultVectorSearchConfig returns a configuration with sensible defaults
func DefaultVectorSearchConfig() *VectorSearchConfig {
	return &VectorSearchConfig{
		Enabled: false,
		Type:    VectorSearchTypeNone,
		Embedding: EmbeddingConfig{
			Provider:   EmbeddingProviderNone,
			Model:      "text-embedding-3-small",
			Dimensions: 1536,
			OpenAI: OpenAIEmbeddingConfig{
				Model:          "text-embedding-3-small",
				RequestTimeout: 30,
				MaxRetries:     3,
			},
		},
		Local: LocalVectorConfig{
			DatabasePath:   "./vector.db",
			Engine:         "sqlite",
			IndexType:      "flat",
			DistanceMetric: "cosine",
		},
		Indexing: IndexingConfig{
			AutoIndex:       false,
			ChunkSize:       1000,
			ChunkOverlap:    200,
			MinChunkSize:    100,
			IncludeMetadata: true,
			BatchSize:       100,
			RefreshInterval: 300, // 5 minutes
		},
		Search: SearchConfig{
			HybridEnabled:   false,
			HybridWeight:    0.7,
			DefaultLimit:    20,
			MaxLimit:        100,
			MinScore:        0.7,
			EnableReranking: false,
		},
	}
}

// VectorSearchError represents an error from a vector search backend
type VectorSearchError struct {
	Backend   VectorSearchType
	Operation string
	Query     string
	Err       error
	Retryable bool
}

func (e *VectorSearchError) Error() string {
	return fmt.Sprintf("vector search error [%s:%s] %s: %v",
		e.Backend, e.Operation, e.Query, e.Err)
}

func (e *VectorSearchError) Unwrap() error {
	return e.Err
}

func (e *VectorSearchError) IsRetryable() bool {
	return e.Retryable
}

// NewVectorSearchError creates a new vector search error
func NewVectorSearchError(backend VectorSearchType, operation, query string, err error, retryable bool) *VectorSearchError {
	return &VectorSearchError{
		Backend:   backend,
		Operation: operation,
		Query:     query,
		Err:       err,
		Retryable: retryable,
	}
}