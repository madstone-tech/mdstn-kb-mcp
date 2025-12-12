package ollama

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// AvailabilityChecker checks if Ollama is available
type AvailabilityChecker struct {
	endpoint   string
	timeout    time.Duration
	httpClient *http.Client
}

// NewAvailabilityChecker creates a new availability checker
func NewAvailabilityChecker(endpoint string) *AvailabilityChecker {
	return &AvailabilityChecker{
		endpoint:   endpoint,
		timeout:    5 * time.Second,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Check checks if Ollama is available
func (ac *AvailabilityChecker) Check(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, ac.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ac.endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetAvailableModels retrieves the list of available models from Ollama
func (ac *AvailabilityChecker) GetAvailableModels(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, ac.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ac.endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// TODO: Parse the response and extract model names
	// For now, return empty list
	return []string{}, nil
}

// SetTimeout sets the timeout for availability checks
func (ac *AvailabilityChecker) SetTimeout(timeout time.Duration) {
	ac.timeout = timeout
	ac.httpClient.Timeout = timeout
}
