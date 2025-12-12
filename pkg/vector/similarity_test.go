package vector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCosineSimilarityIdentical(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{1.0, 2.0, 3.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, sim, 0.0001, "identical vectors should have similarity 1.0")
}

func TestCosineSimilarityOpposite(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{-1.0, -2.0, -3.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.InDelta(t, -1.0, sim, 0.0001, "opposite vectors should have similarity -1.0")
}

func TestCosineSimilarityPerpendicular(t *testing.T) {
	a := []float64{1.0, 0.0}
	b := []float64{0.0, 1.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, sim, 0.0001, "perpendicular vectors should have similarity 0.0")
}

func TestCosineSimilarityPartial(t *testing.T) {
	a := []float64{1.0, 0.0, 0.0}
	b := []float64{1.0, 1.0, 0.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.Greater(t, sim, 0.5, "partially aligned vectors should have similarity > 0.5")
	assert.Less(t, sim, 1.0, "partially aligned vectors should have similarity < 1.0")
}

func TestCosineSimilarityZeroVectors(t *testing.T) {
	a := []float64{0.0, 0.0, 0.0}
	b := []float64{0.0, 0.0, 0.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.Equal(t, 0.0, sim, "zero vectors should have similarity 0.0")
}

func TestCosineSimilarityOneZeroVector(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{0.0, 0.0, 0.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.Equal(t, 0.0, sim, "similarity with zero vector should be 0.0")
}

func TestCosineSimilarityLengthMismatch(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{1.0, 2.0}

	_, err := CosineSimilarity(a, b)
	assert.Error(t, err, "mismatched vector lengths should return error")
}

func TestCosineSimilarityEmptyVectors(t *testing.T) {
	a := []float64{}
	b := []float64{}

	_, err := CosineSimilarity(a, b)
	assert.Error(t, err, "empty vectors should return error")
}

func TestCosineSimilarityNegativeValues(t *testing.T) {
	a := []float64{-1.0, -2.0, -3.0}
	b := []float64{-1.0, -2.0, -3.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, sim, 0.0001, "negative identical vectors should have similarity 1.0")
}

func TestNormalizeVector(t *testing.T) {
	v := []float64{3.0, 4.0}

	normalized, err := NormalizeVector(v)
	require.NoError(t, err)

	// Magnitude should be 5.0, so normalized should be [0.6, 0.8]
	assert.InDelta(t, 0.6, normalized[0], 0.0001)
	assert.InDelta(t, 0.8, normalized[1], 0.0001)

	// Verify magnitude is 1.0
	mag := Magnitude(normalized)
	assert.InDelta(t, 1.0, mag, 0.0001)
}

func TestNormalizeZeroVector(t *testing.T) {
	v := []float64{0.0, 0.0, 0.0}

	normalized, err := NormalizeVector(v)
	require.NoError(t, err)

	assert.Equal(t, v, normalized, "normalized zero vector should still be zero")
}

func TestNormalizeEmptyVector(t *testing.T) {
	v := []float64{}

	_, err := NormalizeVector(v)
	assert.Error(t, err, "normalizing empty vector should return error")
}

func TestDotProduct(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{4.0, 5.0, 6.0}

	dp, err := DotProduct(a, b)
	require.NoError(t, err)
	assert.Equal(t, 32.0, dp, "dot product should be 1*4 + 2*5 + 3*6 = 32")
}

func TestDotProductZero(t *testing.T) {
	a := []float64{1.0, 0.0, 0.0}
	b := []float64{0.0, 1.0, 0.0}

	dp, err := DotProduct(a, b)
	require.NoError(t, err)
	assert.Equal(t, 0.0, dp, "perpendicular vectors should have dot product 0")
}

func TestDotProductLengthMismatch(t *testing.T) {
	a := []float64{1.0, 2.0}
	b := []float64{1.0, 2.0, 3.0}

	_, err := DotProduct(a, b)
	assert.Error(t, err)
}

func TestEuclideanDistance(t *testing.T) {
	a := []float64{0.0, 0.0}
	b := []float64{3.0, 4.0}

	dist, err := EuclideanDistance(a, b)
	require.NoError(t, err)
	assert.Equal(t, 5.0, dist, "distance between [0,0] and [3,4] should be 5.0")
}

func TestEuclideanDistanceSame(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{1.0, 2.0, 3.0}

	dist, err := EuclideanDistance(a, b)
	require.NoError(t, err)
	assert.Equal(t, 0.0, dist, "distance between same points should be 0.0")
}

func TestMagnitude(t *testing.T) {
	v := []float64{3.0, 4.0}

	mag := Magnitude(v)
	assert.Equal(t, 5.0, mag, "magnitude of [3, 4] should be 5.0")
}

func TestMagnitudeZeroVector(t *testing.T) {
	v := []float64{0.0, 0.0, 0.0}

	mag := Magnitude(v)
	assert.Equal(t, 0.0, mag, "magnitude of zero vector should be 0.0")
}

func TestMagnitudeNegativeValues(t *testing.T) {
	v := []float64{-3.0, -4.0}

	mag := Magnitude(v)
	assert.Equal(t, 5.0, mag, "magnitude should be positive")
}

func TestCosineSimilarityLargeVectors(t *testing.T) {
	// Create 384-dimensional vectors (embedding size)
	a := make([]float64, 384)
	b := make([]float64, 384)

	for i := 0; i < 384; i++ {
		a[i] = float64(i) / 100.0
		b[i] = float64(i) / 100.0
	}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, sim, 0.0001, "identical large vectors should have similarity 1.0")
}

func TestNormalizeConsistency(t *testing.T) {
	v := []float64{3.0, 4.0}

	normalized, err := NormalizeVector(v)
	require.NoError(t, err)

	// Cosine similarity of normalized vector with itself should be 1.0
	sim, err := CosineSimilarity(normalized, normalized)
	require.NoError(t, err)
	assert.InDelta(t, 1.0, sim, 0.0001)
}

func BenchmarkCosineSimilarity(b *testing.B) {
	a := make([]float64, 384)
	b_vec := make([]float64, 384)

	for i := 0; i < 384; i++ {
		a[i] = float64(i) / 100.0
		b_vec[i] = float64(i) / 100.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CosineSimilarity(a, b_vec)
	}
}

func BenchmarkMagnitude(b *testing.B) {
	v := make([]float64, 384)
	for i := 0; i < 384; i++ {
		v[i] = float64(i) / 100.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Magnitude(v)
	}
}

func TestCosineSimilaritySmallDifferences(t *testing.T) {
	// Two very similar vectors
	a := []float64{0.1, 0.2, 0.3}
	b := []float64{0.10001, 0.20001, 0.30001}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.Greater(t, sim, 0.99999, "very similar vectors should have high similarity")
}

func TestCosineSimilarityInfinity(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	b := []float64{math.Inf(1), 0.0, 0.0}

	sim, err := CosineSimilarity(a, b)
	require.NoError(t, err)
	assert.True(t, math.IsNaN(sim) || math.IsInf(sim, 0), "should handle infinity gracefully")
}
