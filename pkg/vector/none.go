package vector

import (
	"context"
	"fmt"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// NoneBackend implements a no-op vector search backend
// This is used when vector search is disabled or not configured
type NoneBackend struct{}

// Type returns the vector search backend type
func (n *NoneBackend) Type() types.VectorSearchType {
	return types.VectorSearchTypeNone
}

// IndexDocument is a no-op for the none backend
func (n *NoneBackend) IndexDocument(ctx context.Context, doc *types.Document) error {
	// No-op: vector search is disabled
	return nil
}

// IndexDocuments is a no-op for the none backend
func (n *NoneBackend) IndexDocuments(ctx context.Context, docs []*types.Document) error {
	// No-op: vector search is disabled
	return nil
}

// DeleteDocument is a no-op for the none backend
func (n *NoneBackend) DeleteDocument(ctx context.Context, id string) error {
	// No-op: vector search is disabled
	return nil
}

// Search returns an error indicating vector search is not enabled
func (n *NoneBackend) Search(ctx context.Context, query *types.VectorQuery) (*types.VectorSearchResults, error) {
	return nil, fmt.Errorf("vector search is not enabled")
}

// GetEmbedding returns an error indicating vector search is not enabled
func (n *NoneBackend) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	return nil, fmt.Errorf("vector search is not enabled")
}

// GetEmbeddings returns an error indicating vector search is not enabled
func (n *NoneBackend) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return nil, fmt.Errorf("vector search is not enabled")
}

// Health always returns success for the none backend
func (n *NoneBackend) Health(ctx context.Context) error {
	return nil
}

// Close is a no-op for the none backend
func (n *NoneBackend) Close() error {
	return nil
}
