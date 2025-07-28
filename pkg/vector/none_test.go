package vector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNoneBackend_Type(t *testing.T) {
	backend := &NoneBackend{}
	assert.Equal(t, types.VectorSearchTypeNone, backend.Type())
}

func TestNoneBackend_IndexDocument(t *testing.T) {
	backend := &NoneBackend{}
	ctx := context.Background()
	
	doc := &types.Document{
		ID:      "test-doc",
		Content: "test content",
	}
	
	err := backend.IndexDocument(ctx, doc)
	assert.NoError(t, err) // Should be no-op
}

func TestNoneBackend_IndexDocuments(t *testing.T) {
	backend := &NoneBackend{}
	ctx := context.Background()
	
	docs := []*types.Document{
		{ID: "doc1", Content: "content1"},
		{ID: "doc2", Content: "content2"},
	}
	
	err := backend.IndexDocuments(ctx, docs)
	assert.NoError(t, err) // Should be no-op
}

func TestNoneBackend_DeleteDocument(t *testing.T) {
	backend := &NoneBackend{}
	ctx := context.Background()
	
	err := backend.DeleteDocument(ctx, "test-id")
	assert.NoError(t, err) // Should be no-op
}

func TestNoneBackend_Search(t *testing.T) {
	backend := &NoneBackend{}
	ctx := context.Background()
	
	query := &types.VectorQuery{
		Query: "test query",
		Limit: 10,
	}
	
	results, err := backend.Search(ctx, query)
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "vector search is not enabled")
}

func TestNoneBackend_GetEmbedding(t *testing.T) {
	backend := &NoneBackend{}
	ctx := context.Background()
	
	embedding, err := backend.GetEmbedding(ctx, "test text")
	assert.Error(t, err)
	assert.Nil(t, embedding)
	assert.Contains(t, err.Error(), "vector search is not enabled")
}

func TestNoneBackend_GetEmbeddings(t *testing.T) {
	backend := &NoneBackend{}
	ctx := context.Background()
	
	texts := []string{"text1", "text2"}
	embeddings, err := backend.GetEmbeddings(ctx, texts)
	assert.Error(t, err)
	assert.Nil(t, embeddings)
	assert.Contains(t, err.Error(), "vector search is not enabled")
}

func TestNoneBackend_Health(t *testing.T) {
	backend := &NoneBackend{}
	ctx := context.Background()
	
	err := backend.Health(ctx)
	assert.NoError(t, err) // Should always be healthy
}

func TestNoneBackend_Close(t *testing.T) {
	backend := &NoneBackend{}
	
	err := backend.Close()
	assert.NoError(t, err) // Should be no-op
}

func TestNoneBackend_Interface(t *testing.T) {
	// Test that NoneBackend implements VectorSearchBackend interface
	var _ types.VectorSearchBackend = &NoneBackend{}
}