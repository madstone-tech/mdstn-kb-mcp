package search

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/vector"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/vector/cache"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/vector/testutil"
)

// MockBackend is a mock vector backend for testing
type MockBackend struct {
	vectors map[string][]float64
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		vectors: make(map[string][]float64),
	}
}

func (m *MockBackend) IndexVector(ctx context.Context, id string, embedding []float64, metadata map[string]interface{}) error {
	m.vectors[id] = embedding
	return nil
}

func (m *MockBackend) IndexVectors(ctx context.Context, documents []vector.IndexRequest) error {
	for _, doc := range documents {
		m.vectors[doc.ID] = doc.Embedding
	}
	return nil
}

func (m *MockBackend) SearchVectors(ctx context.Context, query []float64, limit int, threshold float64) ([]vector.SearchResult, error) {
	// Simple mock: return dummy results
	results := make([]vector.SearchResult, 0)
	for id := range m.vectors {
		results = append(results, vector.SearchResult{
			ID:    id,
			Score: 0.8,
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (m *MockBackend) GetVector(ctx context.Context, id string) ([]float64, error) {
	if v, ok := m.vectors[id]; ok {
		return v, nil
	}
	return nil, nil
}

func (m *MockBackend) DeleteVector(ctx context.Context, id string) error {
	delete(m.vectors, id)
	return nil
}

func (m *MockBackend) DeleteVectors(ctx context.Context, ids []string) error {
	for _, id := range ids {
		delete(m.vectors, id)
	}
	return nil
}

func (m *MockBackend) Close() error {
	return nil
}

// MockOllamaClient is a mock Ollama client for testing
type MockOllamaClient struct {
	model string
}

func NewMockOllamaClient() *MockOllamaClient {
	return &MockOllamaClient{model: "nomic-embed-text"}
}

func (m *MockOllamaClient) Embed(ctx context.Context, text string) ([]float64, error) {
	return testutil.SampleVectorNonzero(), nil
}

func (m *MockOllamaClient) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	for i := range texts {
		results[i] = testutil.SampleVectorNonzero()
	}
	return results, nil
}

func (m *MockOllamaClient) Close() error {
	return nil
}

func (m *MockOllamaClient) SetTimeout(timeout time.Duration) {
}

func (m *MockOllamaClient) SetModel(model string) {
	m.model = model
}

func (m *MockOllamaClient) GetModel() string {
	return m.model
}

func (m *MockOllamaClient) SetEndpoint(endpoint string) {
}

func (m *MockOllamaClient) GetEndpoint() string {
	return "http://localhost:11434"
}

func TestNewSemanticEngine(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.embedder)
	assert.NotNil(t, engine.backend)
	assert.NotNil(t, engine.cache)
}

func TestSemanticEngineSearch(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	results, err := engine.Search(context.Background(), "test query", 10, 0.7)

	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestSemanticEngineSearchWithCaching(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	// First search
	results1, err := engine.Search(context.Background(), "test query", 10, 0.7)
	require.NoError(t, err)

	hits1, _, _ := engine.GetStats()
	assert.Equal(t, int64(0), hits1, "first search should be a miss")

	// Second search (should hit cache)
	results2, err := engine.Search(context.Background(), "test query", 10, 0.7)
	require.NoError(t, err)

	hits2, _, _ := engine.GetStats()
	assert.Equal(t, int64(1), hits2, "second search should be a hit")

	assert.NotNil(t, results1)
	assert.NotNil(t, results2)
}

func TestSemanticEngineIndexNote(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	err := engine.IndexNote(context.Background(), "note1", "Test content", nil)
	require.NoError(t, err)

	// Verify the note was indexed
	_, ok := backend.vectors["note1"]
	assert.True(t, ok, "note should be in backend")
}

func TestSemanticEngineIndexNotes(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	notes := []IndexNoteRequest{
		{ID: "note1", Title: "Title 1", Content: "Content 1"},
		{ID: "note2", Title: "Title 2", Content: "Content 2"},
		{ID: "note3", Title: "Title 3", Content: "Content 3"},
	}

	err := engine.IndexNotes(context.Background(), notes)
	require.NoError(t, err)

	// Verify all notes were indexed
	assert.Equal(t, 3, len(backend.vectors))
	_, ok1 := backend.vectors["note1"]
	_, ok2 := backend.vectors["note2"]
	_, ok3 := backend.vectors["note3"]
	assert.True(t, ok1 && ok2 && ok3)
}

func TestSemanticEngineDeleteNote(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	// Index a note
	err := engine.IndexNote(context.Background(), "note1", "Content", nil)
	require.NoError(t, err)

	// Delete it
	err = engine.DeleteNote(context.Background(), "note1")
	require.NoError(t, err)

	// Verify it's deleted
	_, ok := backend.vectors["note1"]
	assert.False(t, ok, "note should be deleted")
}

func TestSemanticEngineDeleteNotes(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	// Index notes
	notes := []IndexNoteRequest{
		{ID: "note1", Content: "Content 1"},
		{ID: "note2", Content: "Content 2"},
		{ID: "note3", Content: "Content 3"},
	}
	err := engine.IndexNotes(context.Background(), notes)
	require.NoError(t, err)

	// Delete multiple notes
	err = engine.DeleteNotes(context.Background(), []string{"note1", "note2"})
	require.NoError(t, err)

	// Verify
	_, ok1 := backend.vectors["note1"]
	_, ok2 := backend.vectors["note2"]
	_, ok3 := backend.vectors["note3"]

	assert.False(t, ok1, "note1 should be deleted")
	assert.False(t, ok2, "note2 should be deleted")
	assert.True(t, ok3, "note3 should still exist")
}

func TestSemanticEngineGetStats(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	hits, misses, evicts := engine.GetStats()

	assert.Equal(t, int64(0), hits)
	assert.Equal(t, int64(0), misses)
	assert.Equal(t, int64(0), evicts)
}

func TestSemanticEngineSearchWithDifferentLimits(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	// Test with limit 0 (should default to 10)
	results, err := engine.Search(context.Background(), "query", 0, 0.7)
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Test with negative limit (should default to 10)
	results, err = engine.Search(context.Background(), "query2", -1, 0.7)
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Test with specific limit
	results, err = engine.Search(context.Background(), "query3", 5, 0.7)
	require.NoError(t, err)
	assert.NotNil(t, results)
}

func TestSemanticEngineSearchWithThreshold(t *testing.T) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	// Test with invalid threshold (should default to 0.7)
	results, err := engine.Search(context.Background(), "query", 10, -0.5)
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Test with threshold > 1 (should default to 0.7)
	results, err = engine.Search(context.Background(), "query", 10, 1.5)
	require.NoError(t, err)
	assert.NotNil(t, results)

	// Test with valid threshold
	results, err = engine.Search(context.Background(), "query", 10, 0.5)
	require.NoError(t, err)
	assert.NotNil(t, results)
}

func BenchmarkSemanticEngineSearch(b *testing.B) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Search(context.Background(), "test query", 10, 0.7)
	}
}

func BenchmarkSemanticEngineIndexNotes(b *testing.B) {
	backend := NewMockBackend()
	ollamaClient := NewMockOllamaClient()
	c := cache.NewCache(100, 1*time.Hour)

	engine := NewSemanticEngine(ollamaClient, backend, c)

	notes := []IndexNoteRequest{
		{ID: "note1", Content: "Content 1"},
		{ID: "note2", Content: "Content 2"},
		{ID: "note3", Content: "Content 3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.IndexNotes(context.Background(), notes)
	}
}
