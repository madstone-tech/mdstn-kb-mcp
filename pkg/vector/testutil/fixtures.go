// Package testutil provides test fixtures and utilities for vector search tests.
package testutil

import (
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// MockOllamaEmbeddingResponse simulates Ollama API embedding response
type MockOllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

// SampleNote returns a test note
func SampleNote(id, title, content string) *types.Note {
	return &types.Note{
		ID:        id,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SampleVector384D returns a mock 384-dimensional vector (all zeros)
func SampleVector384D() []float64 {
	vec := make([]float64, 384)
	for i := range vec {
		vec[i] = 0.0
	}
	return vec
}

// SampleVector384DWithValue returns a mock 384-dimensional vector with initial value
func SampleVector384DWithValue(value float64) []float64 {
	vec := make([]float64, 384)
	for i := range vec {
		vec[i] = value
	}
	return vec
}

// SampleVectorNonzero returns a mock 384-dimensional vector with non-zero values
func SampleVectorNonzero() []float64 {
	vec := make([]float64, 384)
	for i := range vec {
		// Generate varied values between -1 and 1 to simulate embeddings
		vec[i] = float64((i%100)-50) / 100.0
	}
	return vec
}

// SampleDocument returns a test Document for vector indexing
func SampleDocument(id, title, content string) *types.Document {
	return &types.Document{
		ID:        id,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// MockOllamaResponses returns typical mock Ollama API responses
func MockOllamaResponses(texts []string) []MockOllamaEmbeddingResponse {
	responses := make([]MockOllamaEmbeddingResponse, len(texts))
	for i := range texts {
		responses[i] = MockOllamaEmbeddingResponse{
			Embedding: SampleVectorNonzero(),
			Model:     "nomic-embed-text",
		}
	}
	return responses
}

// NoiseVector returns a small 384-dimensional vector with pseudo-random noise
func NoiseVector(seed int) []float64 {
	vec := make([]float64, 384)
	for i := range vec {
		// Simple pseudo-random noise based on seed and index
		vec[i] = float64(((seed+i)*13)%256-128) / 1000.0
	}
	return vec
}
