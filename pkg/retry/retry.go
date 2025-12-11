package retry

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Backoff defines the interface for backoff strategies
type Backoff interface {
	// Duration returns the duration to wait for the given attempt number (0-indexed)
	Duration(attempt int) time.Duration
	// Reset resets the backoff state
	Reset()
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(initialDelay, maxDelay time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   2.0,
		Jitter:       0.1,
	}
}

// Duration calculates the delay for the given attempt
func (b *ExponentialBackoff) Duration(attempt int) time.Duration {
	if attempt == 0 {
		return b.InitialDelay
	}

	// Calculate exponential delay
	delay := float64(b.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= b.Multiplier
		if delay > float64(b.MaxDelay) {
			delay = float64(b.MaxDelay)
			break
		}
	}

	// Add jitter to prevent thundering herd
	if b.Jitter > 0 {
		jitterAmount := delay * b.Jitter
		// Random jitter between -jitter and +jitter
		randomFactor := rand.Float64()*2 - 1 // Random value between -1 and 1
		delay += jitterAmount * randomFactor
	}

	return time.Duration(delay)
}

// Reset resets the backoff state
func (b *ExponentialBackoff) Reset() {
	// Nothing to reset for exponential backoff
}

// Config holds retry configuration
type Config struct {
	MaxAttempts int
	Backoff     Backoff
	ShouldRetry func(error) bool
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts: 5,
		Backoff:     NewExponentialBackoff(100*time.Millisecond, 10*time.Second),
		ShouldRetry: DefaultShouldRetry,
	}
}

// DefaultShouldRetry determines if an error should trigger a retry
func DefaultShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a storage error and if it's retryable
	if storageErr, ok := err.(*types.StorageError); ok {
		return storageErr.IsRetryable()
	}

	// For other errors, don't retry by default
	return false
}

// StorageErrorShouldRetry is a specialized retry function for storage errors
func StorageErrorShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	storageErr, ok := err.(*types.StorageError)
	if !ok {
		return false
	}

	// Don't retry certain operations
	if storageErr.Operation == "read" || storageErr.Operation == "stat" {
		// Only retry read/stat if it's a temporary network issue, not file not found
		return storageErr.IsRetryable()
	}

	return storageErr.IsRetryable()
}

// Retry executes a function with retry logic
func Retry(ctx context.Context, config *Config, fn func() error) error {
	if config == nil {
		config = DefaultConfig()
	}

	var lastErr error
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !config.ShouldRetry(err) {
			return err
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay
		delay := config.Backoff.Duration(attempt)

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded, last error: %w", config.MaxAttempts, lastErr)
}

// RetryWithResult executes a function that returns a result and error with retry logic
func RetryWithResult[T any](ctx context.Context, config *Config, fn func() (T, error)) (T, error) {
	var zero T
	var result T
	var lastErr error

	if config == nil {
		config = DefaultConfig()
	}

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		var err error
		result, err = fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if we should retry
		if !config.ShouldRetry(err) {
			return zero, err
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay
		delay := config.Backoff.Duration(attempt)

		// Wait for the delay or context cancellation
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return zero, fmt.Errorf("max retry attempts (%d) exceeded, last error: %w", config.MaxAttempts, lastErr)
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitState
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	// CircuitClosed - normal operation
	CircuitClosed CircuitState = iota
	// CircuitOpen - circuit is open, calls will fail fast
	CircuitOpen
	// CircuitHalfOpen - testing if service has recovered
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
	}
}

// Execute runs a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	switch cb.state {
	case CircuitOpen:
		// Check if we should try to recover
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	case CircuitHalfOpen:
		// Test if service has recovered
		break
	case CircuitClosed:
		// Normal operation
		break
	}

	err := fn()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}

	return err
}

// recordFailure records a failure and updates circuit state
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

// recordSuccess records a success and updates circuit state
func (cb *CircuitBreaker) recordSuccess() {
	cb.failureCount = 0
	cb.state = CircuitClosed
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() CircuitState {
	return cb.state
}

// FailureCount returns the current failure count
func (cb *CircuitBreaker) FailureCount() int {
	return cb.failureCount
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.failureCount = 0
	cb.state = CircuitClosed
}

// StorageRetryWrapper wraps a storage backend with retry logic
type StorageRetryWrapper struct {
	backend types.StorageBackend
	config  *Config
	breaker *CircuitBreaker
}

// NewStorageRetryWrapper creates a new storage wrapper with retry logic
func NewStorageRetryWrapper(backend types.StorageBackend, config *Config, breaker *CircuitBreaker) *StorageRetryWrapper {
	if config == nil {
		config = DefaultConfig()
		config.ShouldRetry = StorageErrorShouldRetry
	}

	return &StorageRetryWrapper{
		backend: backend,
		config:  config,
		breaker: breaker,
	}
}

// Type returns the storage backend type
func (w *StorageRetryWrapper) Type() types.StorageType {
	return w.backend.Type()
}

// Read with retry logic
func (w *StorageRetryWrapper) Read(ctx context.Context, path string) ([]byte, error) {
	return RetryWithResult(ctx, w.config, func() ([]byte, error) {
		if w.breaker != nil {
			var result []byte
			err := w.breaker.Execute(func() error {
				var err error
				result, err = w.backend.Read(ctx, path)
				return err
			})
			return result, err
		}
		return w.backend.Read(ctx, path)
	})
}

// Write with retry logic
func (w *StorageRetryWrapper) Write(ctx context.Context, path string, data []byte) error {
	return Retry(ctx, w.config, func() error {
		if w.breaker != nil {
			return w.breaker.Execute(func() error {
				return w.backend.Write(ctx, path, data)
			})
		}
		return w.backend.Write(ctx, path, data)
	})
}

// Delete with retry logic
func (w *StorageRetryWrapper) Delete(ctx context.Context, path string) error {
	return Retry(ctx, w.config, func() error {
		if w.breaker != nil {
			return w.breaker.Execute(func() error {
				return w.backend.Delete(ctx, path)
			})
		}
		return w.backend.Delete(ctx, path)
	})
}

// Exists with retry logic
func (w *StorageRetryWrapper) Exists(ctx context.Context, path string) (bool, error) {
	return RetryWithResult(ctx, w.config, func() (bool, error) {
		if w.breaker != nil {
			var result bool
			err := w.breaker.Execute(func() error {
				var err error
				result, err = w.backend.Exists(ctx, path)
				return err
			})
			return result, err
		}
		return w.backend.Exists(ctx, path)
	})
}

// List with retry logic
func (w *StorageRetryWrapper) List(ctx context.Context, prefix string) ([]string, error) {
	return RetryWithResult(ctx, w.config, func() ([]string, error) {
		if w.breaker != nil {
			var result []string
			err := w.breaker.Execute(func() error {
				var err error
				result, err = w.backend.List(ctx, prefix)
				return err
			})
			return result, err
		}
		return w.backend.List(ctx, prefix)
	})
}

// Stat with retry logic
func (w *StorageRetryWrapper) Stat(ctx context.Context, path string) (*types.FileInfo, error) {
	return RetryWithResult(ctx, w.config, func() (*types.FileInfo, error) {
		if w.breaker != nil {
			var result *types.FileInfo
			err := w.breaker.Execute(func() error {
				var err error
				result, err = w.backend.Stat(ctx, path)
				return err
			})
			return result, err
		}
		return w.backend.Stat(ctx, path)
	})
}

// ReadStream delegates to underlying backend (no retry for streams)
func (w *StorageRetryWrapper) ReadStream(ctx context.Context, path string) (io.ReadCloser, error) {
	return w.backend.ReadStream(ctx, path)
}

// WriteStream delegates to underlying backend (no retry for streams)
func (w *StorageRetryWrapper) WriteStream(ctx context.Context, path string, reader io.Reader) error {
	return w.backend.WriteStream(ctx, path, reader)
}

// Copy with retry logic
func (w *StorageRetryWrapper) Copy(ctx context.Context, src, dst string) error {
	return Retry(ctx, w.config, func() error {
		if w.breaker != nil {
			return w.breaker.Execute(func() error {
				return w.backend.Copy(ctx, src, dst)
			})
		}
		return w.backend.Copy(ctx, src, dst)
	})
}

// Move with retry logic
func (w *StorageRetryWrapper) Move(ctx context.Context, src, dst string) error {
	return Retry(ctx, w.config, func() error {
		if w.breaker != nil {
			return w.breaker.Execute(func() error {
				return w.backend.Move(ctx, src, dst)
			})
		}
		return w.backend.Move(ctx, src, dst)
	})
}

// Health with retry logic
func (w *StorageRetryWrapper) Health(ctx context.Context) error {
	return Retry(ctx, w.config, func() error {
		if w.breaker != nil {
			return w.breaker.Execute(func() error {
				return w.backend.Health(ctx)
			})
		}
		return w.backend.Health(ctx)
	})
}

// Close delegates to underlying backend
func (w *StorageRetryWrapper) Close() error {
	return w.backend.Close()
}
