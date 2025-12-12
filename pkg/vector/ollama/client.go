// Package ollama provides an HTTP client for the Ollama embedding API.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// EmbeddingRequest is the request payload for Ollama embedding API
type EmbeddingRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt,omitempty"`
	Input  []string `json:"input,omitempty"`
}

// EmbeddingResponse is the response from Ollama embedding API
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
}

// BatchEmbeddingResponse is the response for batch embedding requests
type BatchEmbeddingResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Model      string      `json:"model"`
}

// Client provides an HTTP client for Ollama embeddings with connection pooling
type Client struct {
	endpoint   string
	model      string
	timeout    time.Duration
	httpClient *http.Client
	mu         sync.RWMutex
}

// NewClient creates a new Ollama client
func NewClient(endpoint, model string) *Client {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false,
	}

	return &Client{
		endpoint: endpoint,
		model:    model,
		timeout:  30 * time.Second,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

// SetTimeout sets the request timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// Embed generates embeddings for a single text
func (c *Client) Embed(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	c.mu.RLock()
	endpoint := c.endpoint
	model := c.model
	c.mu.RUnlock()

	req := EmbeddingRequest{
		Model: model,
		Input: texts,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint+"/api/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embedResp BatchEmbeddingResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return embedResp.Embeddings, nil
}

// Close closes the HTTP client
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpClient.CloseIdleConnections()
	return nil
}

// SetModel updates the embedding model
func (c *Client) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.model = model
}

// GetModel returns the current model
func (c *Client) GetModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.model
}

// SetEndpoint updates the Ollama endpoint
func (c *Client) SetEndpoint(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.endpoint = endpoint
}

// GetEndpoint returns the current endpoint
func (c *Client) GetEndpoint() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.endpoint
}
