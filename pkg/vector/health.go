package vector

import (
	"context"
	"time"
)

// HealthStatus represents the health status of a vector backend
type HealthStatus string

const (
	// HealthStatusHealthy indicates the backend is working normally
	HealthStatusHealthy HealthStatus = "healthy"

	// HealthStatusDegraded indicates the backend is working but with issues
	HealthStatusDegraded HealthStatus = "degraded"

	// HealthStatusUnhealthy indicates the backend is not working
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheckResult contains the result of a health check
type HealthCheckResult struct {
	// Status is the overall health status
	Status HealthStatus `json:"status"`

	// Message provides additional details
	Message string `json:"message"`

	// Endpoint is the backend endpoint being checked
	Endpoint string `json:"endpoint,omitempty"`

	// ResponseTime is how long the check took
	ResponseTime time.Duration `json:"response_time"`

	// Timestamp is when the check was performed
	Timestamp time.Time `json:"timestamp"`

	// LastError is the last error encountered (if any)
	LastError string `json:"last_error,omitempty"`

	// Details contains provider-specific details
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthChecker defines the interface for health checking
type HealthChecker interface {
	// Health performs a health check and returns the result
	Health(ctx context.Context) (*HealthCheckResult, error)

	// IsHealthy quickly checks if the backend is healthy
	IsHealthy(ctx context.Context) bool
}

// SimpleHealthChecker provides basic health checking for vector backends
type SimpleHealthChecker struct {
	endpoint     string
	timeout      time.Duration
	checkBackend Backend
}

// NewSimpleHealthChecker creates a new simple health checker
func NewSimpleHealthChecker(endpoint string, backend Backend) *SimpleHealthChecker {
	return &SimpleHealthChecker{
		endpoint:     endpoint,
		timeout:      5 * time.Second,
		checkBackend: backend,
	}
}

// Health performs a health check on the backend
func (shc *SimpleHealthChecker) Health(ctx context.Context) (*HealthCheckResult, error) {
	start := time.Now()

	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, shc.timeout)
	defer cancel()

	result := &HealthCheckResult{
		Endpoint:  shc.endpoint,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Try to perform a simple operation (e.g., search with empty query)
	emptyVec := make([]float64, 384) // Default dimension
	searchResults, err := shc.checkBackend.SearchVectors(checkCtx, emptyVec, 1, 0)

	result.ResponseTime = time.Since(start)

	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Backend health check failed"
		result.LastError = err.Error()
		return result, nil
	}

	// Check if we got results (indicates backend is working)
	result.Status = HealthStatusHealthy
	result.Message = "Backend is healthy"
	result.Details["results_returned"] = len(searchResults)

	return result, nil
}

// IsHealthy quickly checks if the backend is healthy
func (shc *SimpleHealthChecker) IsHealthy(ctx context.Context) bool {
	result, err := shc.Health(ctx)
	if err != nil {
		return false
	}
	return result.Status == HealthStatusHealthy
}

// SetTimeout sets the health check timeout
func (shc *SimpleHealthChecker) SetTimeout(timeout time.Duration) {
	shc.timeout = timeout
}

// MultiBackendHealthChecker checks multiple backends for health
type MultiBackendHealthChecker struct {
	checkers map[string]HealthChecker
}

// NewMultiBackendHealthChecker creates a checker for multiple backends
func NewMultiBackendHealthChecker() *MultiBackendHealthChecker {
	return &MultiBackendHealthChecker{
		checkers: make(map[string]HealthChecker),
	}
}

// RegisterChecker registers a health checker for a backend name
func (mbhc *MultiBackendHealthChecker) RegisterChecker(name string, checker HealthChecker) {
	mbhc.checkers[name] = checker
}

// CheckAll checks all registered backends and returns results
func (mbhc *MultiBackendHealthChecker) CheckAll(ctx context.Context) map[string]*HealthCheckResult {
	results := make(map[string]*HealthCheckResult)

	for name, checker := range mbhc.checkers {
		result, err := checker.Health(ctx)
		if err != nil {
			results[name] = &HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "Health check error",
				LastError: err.Error(),
				Timestamp: time.Now(),
				Details:   make(map[string]interface{}),
			}
		} else {
			results[name] = result
		}
	}

	return results
}

// HasHealthyBackend checks if at least one backend is healthy
func (mbhc *MultiBackendHealthChecker) HasHealthyBackend(ctx context.Context) bool {
	results := mbhc.CheckAll(ctx)

	for _, result := range results {
		if result.Status == HealthStatusHealthy {
			return true
		}
	}

	return false
}

// HealthCheckWithFallback performs a health check with automatic fallback
type HealthCheckWithFallback struct {
	primary     HealthChecker
	fallback    HealthChecker
	primaryKey  string
	fallbackKey string
}

// NewHealthCheckWithFallback creates a health checker with fallback support
func NewHealthCheckWithFallback(primaryKey string, primary HealthChecker, fallbackKey string, fallback HealthChecker) *HealthCheckWithFallback {
	return &HealthCheckWithFallback{
		primary:     primary,
		fallback:    fallback,
		primaryKey:  primaryKey,
		fallbackKey: fallbackKey,
	}
}

// Health performs a health check with automatic fallback
func (hcf *HealthCheckWithFallback) Health(ctx context.Context) (*HealthCheckResult, error) {
	// Try primary first
	primaryResult, err := hcf.primary.Health(ctx)
	if err == nil && primaryResult.Status == HealthStatusHealthy {
		return primaryResult, nil
	}

	// Fall back to fallback
	if hcf.fallback != nil {
		fallbackResult, err := hcf.fallback.Health(ctx)
		if err == nil && fallbackResult.Status == HealthStatusHealthy {
			return fallbackResult, nil
		}
	}

	// Both unhealthy - return primary result with note about fallback failure
	if primaryResult != nil {
		primaryResult.Message = "Primary unhealthy; fallback also unavailable"
		return primaryResult, nil
	}

	return &HealthCheckResult{
		Status:    HealthStatusUnhealthy,
		Message:   "All backends unavailable",
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}, nil
}

// IsHealthy checks if the primary or fallback is healthy
func (hcf *HealthCheckWithFallback) IsHealthy(ctx context.Context) bool {
	result, err := hcf.Health(ctx)
	if err != nil {
		return false
	}
	return result.Status == HealthStatusHealthy
}
