// Package vector provides vector search functionality.
package vector

import (
	"fmt"
	"math"
)

// CosineSimilarity computes the cosine similarity between two vectors.
// Returns a value between -1 and 1, where 1 means identical direction,
// 0 means perpendicular, and -1 means opposite direction.
// For embeddings, typically returns values between 0 and 1.
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector length mismatch: %d != %d", len(a), len(b))
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("empty vectors")
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	// Handle zero vectors
	if normA == 0 || normB == 0 {
		return 0, nil
	}

	return dotProduct / (normA * normB), nil
}

// NormalizeVector normalizes a vector to unit length
func NormalizeVector(v []float64) ([]float64, error) {
	if len(v) == 0 {
		return nil, fmt.Errorf("cannot normalize empty vector")
	}

	norm := 0.0
	for i := 0; i < len(v); i++ {
		norm += v[i] * v[i]
	}

	norm = math.Sqrt(norm)
	if norm == 0 {
		return make([]float64, len(v)), nil
	}

	normalized := make([]float64, len(v))
	for i := 0; i < len(v); i++ {
		normalized[i] = v[i] / norm
	}

	return normalized, nil
}

// DotProduct computes the dot product of two vectors
func DotProduct(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector length mismatch: %d != %d", len(a), len(b))
	}

	dotProduct := 0.0
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
	}

	return dotProduct, nil
}

// EuclideanDistance computes the Euclidean distance between two vectors
func EuclideanDistance(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector length mismatch: %d != %d", len(a), len(b))
	}

	sumSquares := 0.0
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sumSquares += diff * diff
	}

	return math.Sqrt(sumSquares), nil
}

// Magnitude computes the magnitude (L2 norm) of a vector
func Magnitude(v []float64) float64 {
	sumSquares := 0.0
	for i := 0; i < len(v); i++ {
		sumSquares += v[i] * v[i]
	}
	return math.Sqrt(sumSquares)
}
