package vector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHealthCheckBackend is a mock Backend for health checking tests
type MockHealthCheckBackend struct {
	shouldFail    bool
	returnResults []SearchResult
}

func (m *MockHealthCheckBackend) IndexVector(ctx context.Context, id string, vector []float64, metadata map[string]interface{}) error {
	if m.shouldFail {
		return errors.New("backend error")
	}
	return nil
}

func (m *MockHealthCheckBackend) IndexVectors(ctx context.Context, vectors []IndexRequest) error {
	if m.shouldFail {
		return errors.New("backend error")
	}
	return nil
}

func (m *MockHealthCheckBackend) SearchVectors(ctx context.Context, vector []float64, limit int, threshold float64) ([]SearchResult, error) {
	if m.shouldFail {
		return nil, errors.New("search failed")
	}
	return m.returnResults, nil
}

func (m *MockHealthCheckBackend) GetVector(ctx context.Context, id string) ([]float64, error) {
	if m.shouldFail {
		return nil, errors.New("get failed")
	}
	return []float64{}, nil
}

func (m *MockHealthCheckBackend) DeleteVector(ctx context.Context, id string) error {
	if m.shouldFail {
		return errors.New("delete failed")
	}
	return nil
}

func (m *MockHealthCheckBackend) DeleteVectors(ctx context.Context, ids []string) error {
	if m.shouldFail {
		return errors.New("delete failed")
	}
	return nil
}

func (m *MockHealthCheckBackend) Close() error {
	return nil
}

func TestNewSimpleHealthChecker(t *testing.T) {
	backend := &MockHealthCheckBackend{}
	checker := NewSimpleHealthChecker("http://localhost:11434", backend)

	assert.NotNil(t, checker)
	assert.Equal(t, "http://localhost:11434", checker.endpoint)
	assert.Equal(t, 5*time.Second, checker.timeout)
}

func TestSimpleHealthChecker_Health_Healthy(t *testing.T) {
	backend := &MockHealthCheckBackend{
		shouldFail: false,
		returnResults: []SearchResult{
			{
				ID:    "test-id",
				Title: "Test",
				Score: 0.95,
			},
		},
	}

	checker := NewSimpleHealthChecker("http://localhost:11434", backend)
	result, err := checker.Health(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Equal(t, "Backend is healthy", result.Message)
	assert.Greater(t, result.ResponseTime, time.Duration(0))
	assert.Equal(t, 1, result.Details["results_returned"])
}

func TestSimpleHealthChecker_Health_Unhealthy(t *testing.T) {
	backend := &MockHealthCheckBackend{
		shouldFail: true,
	}

	checker := NewSimpleHealthChecker("http://localhost:11434", backend)
	result, err := checker.Health(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, HealthStatusUnhealthy, result.Status)
	assert.Equal(t, "Backend health check failed", result.Message)
	assert.NotEmpty(t, result.LastError)
}

func TestSimpleHealthChecker_IsHealthy(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
		expected   bool
	}{
		{"healthy", false, true},
		{"unhealthy", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := &MockHealthCheckBackend{shouldFail: tt.shouldFail}
			checker := NewSimpleHealthChecker("http://localhost:11434", backend)

			result := checker.IsHealthy(context.Background())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSimpleHealthChecker_SetTimeout(t *testing.T) {
	backend := &MockHealthCheckBackend{}
	checker := NewSimpleHealthChecker("http://localhost:11434", backend)

	assert.Equal(t, 5*time.Second, checker.timeout)

	checker.SetTimeout(10 * time.Second)
	assert.Equal(t, 10*time.Second, checker.timeout)
}

func TestNewMultiBackendHealthChecker(t *testing.T) {
	checker := NewMultiBackendHealthChecker()

	assert.NotNil(t, checker)
	assert.Empty(t, checker.checkers)
}

func TestMultiBackendHealthChecker_RegisterChecker(t *testing.T) {
	checker := NewMultiBackendHealthChecker()
	backend := &MockHealthCheckBackend{}
	healthChecker := NewSimpleHealthChecker("http://localhost:11434", backend)

	checker.RegisterChecker("primary", healthChecker)

	assert.Equal(t, 1, len(checker.checkers))
	assert.NotNil(t, checker.checkers["primary"])
}

func TestMultiBackendHealthChecker_CheckAll(t *testing.T) {
	checker := NewMultiBackendHealthChecker()

	backend1 := &MockHealthCheckBackend{shouldFail: false}
	backend2 := &MockHealthCheckBackend{shouldFail: true}

	checker.RegisterChecker("primary", NewSimpleHealthChecker("http://localhost:11434", backend1))
	checker.RegisterChecker("secondary", NewSimpleHealthChecker("http://localhost:11435", backend2))

	results := checker.CheckAll(context.Background())

	assert.Equal(t, 2, len(results))
	assert.Equal(t, HealthStatusHealthy, results["primary"].Status)
	assert.Equal(t, HealthStatusUnhealthy, results["secondary"].Status)
}

func TestMultiBackendHealthChecker_HasHealthyBackend(t *testing.T) {
	tests := []struct {
		name       string
		backend1OK bool
		backend2OK bool
		expected   bool
	}{
		{"both_healthy", true, true, true},
		{"primary_healthy", true, false, true},
		{"secondary_healthy", false, true, true},
		{"both_unhealthy", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewMultiBackendHealthChecker()

			backend1 := &MockHealthCheckBackend{shouldFail: !tt.backend1OK}
			backend2 := &MockHealthCheckBackend{shouldFail: !tt.backend2OK}

			checker.RegisterChecker("primary", NewSimpleHealthChecker("http://localhost:11434", backend1))
			checker.RegisterChecker("secondary", NewSimpleHealthChecker("http://localhost:11435", backend2))

			result := checker.HasHealthyBackend(context.Background())
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewHealthCheckWithFallback(t *testing.T) {
	primaryBackend := &MockHealthCheckBackend{}
	fallbackBackend := &MockHealthCheckBackend{}

	primary := NewSimpleHealthChecker("http://localhost:11434", primaryBackend)
	fallback := NewSimpleHealthChecker("http://localhost:11435", fallbackBackend)

	checker := NewHealthCheckWithFallback("primary", primary, "fallback", fallback)

	assert.NotNil(t, checker)
	assert.NotNil(t, checker.primary)
	assert.NotNil(t, checker.fallback)
}

func TestHealthCheckWithFallback_Health_PrimaryHealthy(t *testing.T) {
	primaryBackend := &MockHealthCheckBackend{shouldFail: false}
	fallbackBackend := &MockHealthCheckBackend{shouldFail: true}

	primary := NewSimpleHealthChecker("http://localhost:11434", primaryBackend)
	fallback := NewSimpleHealthChecker("http://localhost:11435", fallbackBackend)

	checker := NewHealthCheckWithFallback("primary", primary, "fallback", fallback)
	result, err := checker.Health(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, HealthStatusHealthy, result.Status)
}

func TestHealthCheckWithFallback_Health_PrimaryUnhealthy_FallbackHealthy(t *testing.T) {
	primaryBackend := &MockHealthCheckBackend{shouldFail: true}
	fallbackBackend := &MockHealthCheckBackend{shouldFail: false}

	primary := NewSimpleHealthChecker("http://localhost:11434", primaryBackend)
	fallback := NewSimpleHealthChecker("http://localhost:11435", fallbackBackend)

	checker := NewHealthCheckWithFallback("primary", primary, "fallback", fallback)
	result, err := checker.Health(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, HealthStatusHealthy, result.Status)
}

func TestHealthCheckWithFallback_Health_BothUnhealthy(t *testing.T) {
	primaryBackend := &MockHealthCheckBackend{shouldFail: true}
	fallbackBackend := &MockHealthCheckBackend{shouldFail: true}

	primary := NewSimpleHealthChecker("http://localhost:11434", primaryBackend)
	fallback := NewSimpleHealthChecker("http://localhost:11435", fallbackBackend)

	checker := NewHealthCheckWithFallback("primary", primary, "fallback", fallback)
	result, err := checker.Health(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, HealthStatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "unavailable")
}

func TestHealthCheckWithFallback_IsHealthy(t *testing.T) {
	tests := []struct {
		name       string
		primaryOK  bool
		fallbackOK bool
		expected   bool
	}{
		{"primary_healthy", true, false, true},
		{"fallback_healthy", false, true, true},
		{"both_healthy", true, true, true},
		{"both_unhealthy", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primaryBackend := &MockHealthCheckBackend{shouldFail: !tt.primaryOK}
			fallbackBackend := &MockHealthCheckBackend{shouldFail: !tt.fallbackOK}

			primary := NewSimpleHealthChecker("http://localhost:11434", primaryBackend)
			fallback := NewSimpleHealthChecker("http://localhost:11435", fallbackBackend)

			checker := NewHealthCheckWithFallback("primary", primary, "fallback", fallback)
			result := checker.IsHealthy(context.Background())

			assert.Equal(t, tt.expected, result)
		})
	}
}
