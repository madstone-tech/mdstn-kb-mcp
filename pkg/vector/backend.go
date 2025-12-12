// Package vector provides vector search backend abstractions and implementations.
package vector

import (
	"context"
	"fmt"
)

// Backend defines the interface for vector search storage backends
type Backend interface {
	// IndexVector stores a vector embedding for a document
	IndexVector(ctx context.Context, id string, embedding []float64, metadata map[string]interface{}) error

	// IndexVectors stores multiple vector embeddings
	IndexVectors(ctx context.Context, documents []IndexRequest) error

	// SearchVectors searches for similar vectors
	SearchVectors(ctx context.Context, query []float64, limit int, threshold float64) ([]SearchResult, error)

	// GetVector retrieves a stored vector embedding
	GetVector(ctx context.Context, id string) ([]float64, error)

	// DeleteVector removes a vector from storage
	DeleteVector(ctx context.Context, id string) error

	// DeleteVectors removes multiple vectors from storage
	DeleteVectors(ctx context.Context, ids []string) error

	// Close closes the backend connection
	Close() error
}

// IndexRequest represents a document to be indexed
type IndexRequest struct {
	ID        string                 `json:"id"`
	Embedding []float64              `json:"embedding"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Title     string                 `json:"title,omitempty"`
	Content   string                 `json:"content,omitempty"`
}

// SearchResult represents a search result with score
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Distance float64                `json:"distance,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Title    string                 `json:"title,omitempty"`
	Content  string                 `json:"content,omitempty"`
}

// BackendError represents an error from a backend operation
type BackendError struct {
	Backend   string
	Operation string
	Err       error
}

func (e *BackendError) Error() string {
	return fmt.Sprintf("vector backend error [%s:%s]: %v", e.Backend, e.Operation, e.Err)
}

func (e *BackendError) Unwrap() error {
	return e.Err
}

// NewBackendError creates a new backend error
func NewBackendError(backend, operation string, err error) *BackendError {
	return &BackendError{
		Backend:   backend,
		Operation: operation,
		Err:       err,
	}
}
