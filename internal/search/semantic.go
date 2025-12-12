// Package search provides search engine implementations.
package search

import (
	"context"
	"fmt"
	"sort"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/vector"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/vector/cache"
)

// OllamaEmbedder defines the interface for embedding generation
type OllamaEmbedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// SemanticEngine performs semantic search using embeddings
type SemanticEngine struct {
	embedder OllamaEmbedder
	backend  vector.Backend
	cache    *cache.Cache
}

// NewSemanticEngine creates a new semantic search engine
func NewSemanticEngine(embedder OllamaEmbedder, backend vector.Backend, c *cache.Cache) *SemanticEngine {
	return &SemanticEngine{
		embedder: embedder,
		backend:  backend,
		cache:    c,
	}
}

// Search performs a semantic search
func (se *SemanticEngine) Search(ctx context.Context, query string, limit int, threshold float64) ([]vector.SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if threshold < 0 || threshold > 1 {
		threshold = 0.7
	}

	// Check cache first
	var embedding []float64
	if cached, ok := se.cache.Get(query); ok {
		embedding = cached
	} else {
		// Generate embedding using embedder
		var err error
		embedding, err = se.embedder.Embed(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for query: %w", err)
		}

		// Cache the embedding
		se.cache.Set(query, embedding)
	}

	// Search using the embedding
	results, err := se.backend.SearchVectors(ctx, embedding, limit, threshold)
	if err != nil {
		return nil, fmt.Errorf("backend search failed: %w", err)
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// IndexNoteRequest represents a note to be indexed
type IndexNoteRequest struct {
	ID       string
	Title    string
	Content  string
	Metadata map[string]interface{}
}

// IndexNote indexes a note's content for semantic search
func (se *SemanticEngine) IndexNote(ctx context.Context, id string, content string, metadata map[string]interface{}) error {
	embedding, err := se.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding for note %s: %w", id, err)
	}

	if err := se.backend.IndexVector(ctx, id, embedding, metadata); err != nil {
		return fmt.Errorf("failed to index note %s: %w", id, err)
	}

	return nil
}

// IndexNotes indexes multiple notes for semantic search
func (se *SemanticEngine) IndexNotes(ctx context.Context, notes []IndexNoteRequest) error {
	if len(notes) == 0 {
		return nil
	}

	// Generate embeddings for all notes
	texts := make([]string, len(notes))
	for i, note := range notes {
		texts[i] = note.Content
	}

	embeddings, err := se.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings for batch: %w", err)
	}

	if len(embeddings) != len(notes) {
		return fmt.Errorf("embedding count mismatch: expected %d, got %d", len(notes), len(embeddings))
	}

	// Create index requests
	indexRequests := make([]vector.IndexRequest, len(notes))
	for i, note := range notes {
		indexRequests[i] = vector.IndexRequest{
			ID:        note.ID,
			Embedding: embeddings[i],
			Metadata:  note.Metadata,
			Title:     note.Title,
			Content:   note.Content,
		}
	}

	return se.backend.IndexVectors(ctx, indexRequests)
}

// DeleteNote deletes a note from the semantic index
func (se *SemanticEngine) DeleteNote(ctx context.Context, id string) error {
	return se.backend.DeleteVector(ctx, id)
}

// DeleteNotes deletes multiple notes from the semantic index
func (se *SemanticEngine) DeleteNotes(ctx context.Context, ids []string) error {
	return se.backend.DeleteVectors(ctx, ids)
}

// GetStats returns cache statistics
func (se *SemanticEngine) GetStats() (hits int64, misses int64, evicts int64) {
	return se.cache.Stats()
}
