package retry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Helper function for string containment check
func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

func TestExponentialBackoff_Duration(t *testing.T) {
	backoff := NewExponentialBackoff(100*time.Millisecond, 5*time.Second)

	testCases := []struct {
		attempt  int
		expected time.Duration
		maxDelta time.Duration
	}{
		{0, 100 * time.Millisecond, 50 * time.Millisecond},
		{1, 200 * time.Millisecond, 100 * time.Millisecond},
		{2, 400 * time.Millisecond, 200 * time.Millisecond},
		{3, 800 * time.Millisecond, 400 * time.Millisecond},
		{10, 5 * time.Second, 500 * time.Millisecond}, // Should cap at max delay
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("attempt_%d", tc.attempt), func(t *testing.T) {
			duration := backoff.Duration(tc.attempt)
			
			// Check if duration is within expected range (accounting for jitter)
			if duration < tc.expected-tc.maxDelta || duration > tc.expected+tc.maxDelta {
				t.Errorf("Duration for attempt %d: expected %v ± %v, got %v", 
					tc.attempt, tc.expected, tc.maxDelta, duration)
			}
		})
	}
}

func TestDefaultShouldRetry(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"regular error", errors.New("some error"), false},
		{"retryable storage error", types.NewStorageError(types.StorageTypeLocal, "write", "test", errors.New("temp error"), true), true},
		{"non-retryable storage error", types.NewStorageError(types.StorageTypeLocal, "read", "test", errors.New("not found"), false), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DefaultShouldRetry(tc.err)
			if result != tc.expected {
				t.Errorf("DefaultShouldRetry(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestStorageErrorShouldRetry(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"regular error", errors.New("some error"), false},
		{"retryable read error", types.NewStorageError(types.StorageTypeS3, "read", "test", errors.New("network error"), true), true},
		{"non-retryable read error", types.NewStorageError(types.StorageTypeLocal, "read", "test", errors.New("not found"), false), false},
		{"retryable write error", types.NewStorageError(types.StorageTypeS3, "write", "test", errors.New("temp error"), true), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := StorageErrorShouldRetry(tc.err)
			if result != tc.expected {
				t.Errorf("StorageErrorShouldRetry(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestRetry_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3

	attempts := 0
	err := Retry(ctx, config, func() error {
		attempts++
		if attempts < 2 {
			return types.NewStorageError(types.StorageTypeLocal, "write", "test", errors.New("temp error"), true)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3
	config.Backoff = NewExponentialBackoff(1*time.Millisecond, 10*time.Millisecond) // Fast for testing

	attempts := 0
	err := Retry(ctx, config, func() error {
		attempts++
		return types.NewStorageError(types.StorageTypeLocal, "write", "test", errors.New("persistent error"), true)
	})

	if err == nil {
		t.Fatal("Expected error, got success")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if !contains(err.Error(), "persistent error") {
		t.Errorf("Expected error to contain 'persistent error', got: %v", err)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3

	attempts := 0
	nonRetryableErr := types.NewStorageError(types.StorageTypeLocal, "read", "test", errors.New("not found"), false)
	
	err := Retry(ctx, config, func() error {
		attempts++
		return nonRetryableErr
	})

	if err != nonRetryableErr {
		t.Fatalf("Expected non-retryable error, got: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := DefaultConfig()
	config.MaxAttempts = 5
	config.Backoff = NewExponentialBackoff(100*time.Millisecond, 1*time.Second)

	attempts := 0
	
	// Cancel context after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Retry(ctx, config, func() error {
		attempts++
		return types.NewStorageError(types.StorageTypeLocal, "write", "test", errors.New("temp error"), true)
	})

	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled, got: %v", err)
	}

	if attempts == 0 {
		t.Error("Expected at least one attempt")
	}
}

func TestRetryWithResult_Success(t *testing.T) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3

	attempts := 0
	result, err := RetryWithResult(ctx, config, func() (string, error) {
		attempts++
		if attempts < 2 {
			return "", types.NewStorageError(types.StorageTypeLocal, "read", "test", errors.New("temp error"), true)
		}
		return "success", nil
	})

	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestCircuitBreaker_NormalOperation(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Normal successful operation
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if cb.State() != CircuitClosed {
		t.Errorf("Expected circuit to be closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_FailureThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Cause failures to reach threshold
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error {
			return testErr
		})
		if err != testErr {
			t.Errorf("Expected test error, got: %v", err)
		}
	}

	if cb.State() != CircuitOpen {
		t.Errorf("Expected circuit to be open, got %v", cb.State())
	}

	// Next call should fail fast
	err := cb.Execute(func() error {
		return nil
	})

	if err == nil {
		t.Error("Expected circuit breaker error, got success")
	}
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	cb := NewCircuitBreaker(2, 10*time.Millisecond) // Short timeout for testing

	// Cause failures to open circuit
	testErr := errors.New("test error")
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return testErr
		}) // Ignore error return in test - we're testing circuit state
	}

	if cb.State() != CircuitOpen {
		t.Errorf("Expected circuit to be open, got %v", cb.State())
	}

	// Wait for reset timeout
	time.Sleep(15 * time.Millisecond)

	// Next call should transition to half-open and succeed
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected success after recovery, got: %v", err)
	}

	if cb.State() != CircuitClosed {
		t.Errorf("Expected circuit to be closed after recovery, got %v", cb.State())
	}
}

// Mock storage backend for testing wrapper
type mockStorageBackend struct {
	readFunc   func(ctx context.Context, path string) ([]byte, error)
	writeFunc  func(ctx context.Context, path string, data []byte) error
	deleteFunc func(ctx context.Context, path string) error
	existsFunc func(ctx context.Context, path string) (bool, error)
	listFunc   func(ctx context.Context, prefix string) ([]string, error)
	statFunc   func(ctx context.Context, path string) (*types.FileInfo, error)
	healthFunc func(ctx context.Context) error
	closeFunc  func() error
	callCount  int
}

func (m *mockStorageBackend) Type() types.StorageType {
	return types.StorageTypeLocal
}

func (m *mockStorageBackend) Read(ctx context.Context, path string) ([]byte, error) {
	m.callCount++
	if m.readFunc != nil {
		return m.readFunc(ctx, path)
	}
	return []byte("test data"), nil
}

func (m *mockStorageBackend) Write(ctx context.Context, path string, data []byte) error {
	m.callCount++
	if m.writeFunc != nil {
		return m.writeFunc(ctx, path, data)
	}
	return nil
}

func (m *mockStorageBackend) Delete(ctx context.Context, path string) error {
	m.callCount++
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, path)
	}
	return nil
}

func (m *mockStorageBackend) Exists(ctx context.Context, path string) (bool, error) {
	m.callCount++
	if m.existsFunc != nil {
		return m.existsFunc(ctx, path)
	}
	return true, nil
}

func (m *mockStorageBackend) List(ctx context.Context, prefix string) ([]string, error) {
	m.callCount++
	if m.listFunc != nil {
		return m.listFunc(ctx, prefix)
	}
	return []string{"file1.md", "file2.md"}, nil
}

func (m *mockStorageBackend) Stat(ctx context.Context, path string) (*types.FileInfo, error) {
	m.callCount++
	if m.statFunc != nil {
		return m.statFunc(ctx, path)
	}
	return &types.FileInfo{Path: path, Size: 100}, nil
}

func (m *mockStorageBackend) ReadStream(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *mockStorageBackend) WriteStream(ctx context.Context, path string, reader io.Reader) error {
	return errors.New("not implemented")
}

func (m *mockStorageBackend) Copy(ctx context.Context, src, dst string) error {
	m.callCount++
	return nil
}

func (m *mockStorageBackend) Move(ctx context.Context, src, dst string) error {
	m.callCount++
	return nil
}

func (m *mockStorageBackend) Health(ctx context.Context) error {
	m.callCount++
	if m.healthFunc != nil {
		return m.healthFunc(ctx)
	}
	return nil
}

func (m *mockStorageBackend) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestStorageRetryWrapper_ReadSuccess(t *testing.T) {
	mock := &mockStorageBackend{}
	mock.readFunc = func(ctx context.Context, path string) ([]byte, error) {
		if mock.callCount == 1 {
			return nil, types.NewStorageError(types.StorageTypeLocal, "read", path, errors.New("temp error"), true)
		}
		return []byte("success"), nil
	}

	config := DefaultConfig()
	config.MaxAttempts = 3
	config.Backoff = NewExponentialBackoff(1*time.Millisecond, 10*time.Millisecond)
	config.ShouldRetry = StorageErrorShouldRetry

	wrapper := NewStorageRetryWrapper(mock, config, nil)

	data, err := wrapper.Read(context.Background(), "test.md")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if string(data) != "success" {
		t.Errorf("Expected 'success', got %s", string(data))
	}

	if mock.callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", mock.callCount)
	}
}

func TestStorageRetryWrapper_WithCircuitBreaker(t *testing.T) {
	mock := &mockStorageBackend{}
	mock.writeFunc = func(ctx context.Context, path string, data []byte) error {
		return types.NewStorageError(types.StorageTypeLocal, "write", path, errors.New("persistent error"), true)
	}

	config := DefaultConfig()
	config.MaxAttempts = 2
	config.Backoff = NewExponentialBackoff(1*time.Millisecond, 10*time.Millisecond)
	config.ShouldRetry = StorageErrorShouldRetry

	cb := NewCircuitBreaker(2, 100*time.Millisecond)
	wrapper := NewStorageRetryWrapper(mock, config, cb)

	// First call should trigger retries and then fail
	err := wrapper.Write(context.Background(), "test.md", []byte("data"))
	if err == nil {
		t.Error("Expected error, got success")
	}

	// Circuit breaker should now be open, second call should fail fast
	mock.callCount = 0 // Reset counter
	err = wrapper.Write(context.Background(), "test.md", []byte("data"))
	if err == nil {
		t.Error("Expected circuit breaker error, got success")
	}

	// Should not have called the backend due to circuit breaker
	if mock.callCount > 0 {
		t.Errorf("Expected 0 calls due to circuit breaker, got %d", mock.callCount)
	}
}

func TestStorageRetryWrapper_Type(t *testing.T) {
	mock := &mockStorageBackend{}
	wrapper := NewStorageRetryWrapper(mock, nil, nil)

	if wrapper.Type() != types.StorageTypeLocal {
		t.Errorf("Expected storage type 'local', got %s", wrapper.Type())
	}
}

// Benchmark tests
func BenchmarkRetry_Success(b *testing.B) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 1 // No retries for benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Retry(ctx, config, func() error {
			return nil
		}) // Ignore error in benchmark
	}
}

func BenchmarkRetry_WithRetries(b *testing.B) {
	ctx := context.Background()
	config := DefaultConfig()
	config.MaxAttempts = 3
	config.Backoff = NewExponentialBackoff(1*time.Microsecond, 1*time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempt := 0
		_ = Retry(ctx, config, func() error {
			attempt++
			if attempt < 2 {
				return types.NewStorageError(types.StorageTypeLocal, "test", "path", errors.New("temp"), true)
			}
			return nil
		}) // Ignore error in benchmark
	}
}

func BenchmarkCircuitBreaker_Closed(b *testing.B) {
	cb := NewCircuitBreaker(10, 1*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(func() error {
			return nil
		}) // Ignore error in benchmark
	}
}