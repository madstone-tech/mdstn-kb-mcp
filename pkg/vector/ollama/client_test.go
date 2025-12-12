package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:11434", "nomic-embed-text")

	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:11434", client.GetEndpoint())
	assert.Equal(t, "nomic-embed-text", client.GetModel())
}

func TestClientEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req EmbeddingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		resp := BatchEmbeddingResponse{
			Embeddings: [][]float64{
				{0.1, 0.2, 0.3},
			},
			Model: req.Model,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	embedding, err := client.Embed(context.Background(), "test text")
	require.NoError(t, err)
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, embedding)
}

func TestClientEmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req EmbeddingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, 3, len(req.Input))

		resp := BatchEmbeddingResponse{
			Embeddings: [][]float64{
				{0.1, 0.2, 0.3},
				{0.4, 0.5, 0.6},
				{0.7, 0.8, 0.9},
			},
			Model: req.Model,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	texts := []string{"text1", "text2", "text3"}
	embeddings, err := client.EmbedBatch(context.Background(), texts)
	require.NoError(t, err)

	assert.Equal(t, 3, len(embeddings))
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, embeddings[0])
	assert.Equal(t, []float64{0.4, 0.5, 0.6}, embeddings[1])
	assert.Equal(t, []float64{0.7, 0.8, 0.9}, embeddings[2])
}

func TestClientEmbedEmpty(t *testing.T) {
	client := NewClient("http://localhost:11434", "test-model")

	embeddings, err := client.EmbedBatch(context.Background(), []string{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(embeddings))
}

func TestClientEmbedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	_, err := client.Embed(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ollama API error")
}

func TestClientSetTimeout(t *testing.T) {
	client := NewClient("http://localhost:11434", "test-model")

	original := client.timeout
	client.SetTimeout(5 * time.Second)

	assert.NotEqual(t, original, client.timeout)
	assert.Equal(t, 5*time.Second, client.timeout)
}

func TestClientSetModel(t *testing.T) {
	client := NewClient("http://localhost:11434", "model1")

	client.SetModel("model2")
	assert.Equal(t, "model2", client.GetModel())
}

func TestClientSetEndpoint(t *testing.T) {
	client := NewClient("http://localhost:11434", "test-model")

	client.SetEndpoint("http://localhost:11435")
	assert.Equal(t, "http://localhost:11435", client.GetEndpoint())
}

func TestClientClose(t *testing.T) {
	client := NewClient("http://localhost:11434", "test-model")

	err := client.Close()
	assert.NoError(t, err)
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte(`{"embeddings": [[0.1, 0.2]]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")
	client.SetTimeout(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Embed(ctx, "test")
	assert.Error(t, err)
}

func TestClientMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	_, err := client.Embed(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestClientNoEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BatchEmbeddingResponse{
			Embeddings: [][]float64{},
			Model:      "test-model",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	_, err := client.Embed(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embeddings")
}

func TestClientConcurrentRequests(t *testing.T) {
	counter := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++

		resp := BatchEmbeddingResponse{
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
			Model:      "test-model",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			_, err := client.Embed(context.Background(), "test")
			done <- err
		}(i)
	}

	for i := 0; i < 5; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	assert.Equal(t, 5, counter)
}

func TestClientConnectionPool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BatchEmbeddingResponse{
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
			Model:      "test-model",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	// Multiple requests should reuse connections
	for i := 0; i < 10; i++ {
		_, err := client.Embed(context.Background(), "test")
		assert.NoError(t, err)
	}

	// Verify connection pool is configured
	transport := client.httpClient.Transport.(*http.Transport)
	assert.Greater(t, transport.MaxIdleConnsPerHost, 0)
}

func TestEmbeddingRequestMarshalling(t *testing.T) {
	req := EmbeddingRequest{
		Model: "test-model",
		Input: []string{"text1", "text2"},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var unmarshalled EmbeddingRequest
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, req.Model, unmarshalled.Model)
	assert.Equal(t, req.Input, unmarshalled.Input)
}

func BenchmarkClientEmbed(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BatchEmbeddingResponse{
			Embeddings: [][]float64{{0.1, 0.2, 0.3}},
			Model:      "test-model",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Embed(context.Background(), "test text")
	}
}

func BenchmarkClientEmbedBatch(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req EmbeddingRequest
		json.NewDecoder(r.Body).Decode(&req)

		embeddings := make([][]float64, len(req.Input))
		for i := range embeddings {
			embeddings[i] = []float64{0.1, 0.2, 0.3}
		}

		resp := BatchEmbeddingResponse{
			Embeddings: embeddings,
			Model:      req.Model,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")
	texts := []string{"text1", "text2", "text3", "text4", "text5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.EmbedBatch(context.Background(), texts)
	}
}

func TestClientResponseBodyRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response that can be read but might fail in processing
		w.Header().Set("Content-Type", "application/json")

		resp := BatchEmbeddingResponse{
			Embeddings: [][]float64{{0.1, 0.2}},
			Model:      "test",
		}

		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	embeddings, err := client.EmbedBatch(context.Background(), []string{"test"})
	require.NoError(t, err)
	require.Len(t, embeddings, 1)
}

func TestClientReadBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Write partial response to trigger read error
		w.Write([]byte(`{"embeddings": [[0.1, 0.2]`))
		// Note: httptest will close the connection before we can trigger a read error
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-model")

	_, err := client.Embed(context.Background(), "test")
	// This might fail with unmarshal error instead of read error
	assert.Error(t, err)
}
