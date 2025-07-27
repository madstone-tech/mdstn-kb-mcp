package types

import (
	"context"
	"fmt"
	"io"
)

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
)

// StorageBackend defines the interface for all storage implementations
type StorageBackend interface {
	// Type returns the storage backend type
	Type() StorageType

	// Read retrieves a file's content by path
	Read(ctx context.Context, path string) ([]byte, error)

	// Write stores content at the given path
	Write(ctx context.Context, path string, data []byte) error

	// Delete removes a file at the given path
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at the given path
	Exists(ctx context.Context, path string) (bool, error)

	// List returns all files matching the given prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// Stat returns metadata about a file
	Stat(ctx context.Context, path string) (*FileInfo, error)

	// ReadStream returns a reader for streaming large files
	ReadStream(ctx context.Context, path string) (io.ReadCloser, error)

	// WriteStream writes data from a reader to the given path
	WriteStream(ctx context.Context, path string, reader io.Reader) error

	// Copy copies a file from src to dst within the same backend
	Copy(ctx context.Context, src, dst string) error

	// Move moves/renames a file from src to dst
	Move(ctx context.Context, src, dst string) error

	// Health performs a health check on the storage backend
	Health(ctx context.Context) error

	// Close cleanly shuts down the storage backend
	Close() error
}

// FileInfo contains metadata about a stored file
type FileInfo struct {
	// Path is the full path to the file
	Path string `json:"path"`

	// Size is the file size in bytes
	Size int64 `json:"size"`

	// ModTime is when the file was last modified
	ModTime int64 `json:"mod_time"` // Unix timestamp

	// ETag is an entity tag for the file (for S3 compatibility)
	ETag string `json:"etag,omitempty"`

	// ContentType is the MIME type of the file
	ContentType string `json:"content_type,omitempty"`

	// StorageClass indicates the storage tier (S3 specific)
	StorageClass string `json:"storage_class,omitempty"`

	// Custom metadata from the storage backend
	Metadata map[string]string `json:"metadata,omitempty"`
}

// StorageConfig contains configuration for storage backends
type StorageConfig struct {
	// Type specifies which storage backend to use
	Type StorageType `toml:"type" json:"type"`

	// Local storage configuration
	Local LocalStorageConfig `toml:"local" json:"local"`

	// S3 storage configuration
	S3 S3StorageConfig `toml:"s3" json:"s3"`

	// Cache configuration
	Cache CacheConfig `toml:"cache" json:"cache"`
}

// LocalStorageConfig configures local filesystem storage
type LocalStorageConfig struct {
	// Path is the root directory for the vault
	Path string `toml:"path" json:"path"`

	// CreateDirs automatically creates directories if they don't exist
	CreateDirs bool `toml:"create_dirs" json:"create_dirs"`

	// Permissions for created directories (octal)
	DirPerms string `toml:"dir_perms" json:"dir_perms"`

	// Permissions for created files (octal)
	FilePerms string `toml:"file_perms" json:"file_perms"`

	// EnableLocking enables file locking for concurrent access
	EnableLocking bool `toml:"enable_locking" json:"enable_locking"`

	// LockTimeout is the maximum time to wait for a file lock (seconds)
	LockTimeout int `toml:"lock_timeout" json:"lock_timeout"`
}

// S3StorageConfig configures S3-compatible storage
type S3StorageConfig struct {
	// Bucket is the S3 bucket name
	Bucket string `toml:"bucket" json:"bucket"`

	// Region is the AWS region
	Region string `toml:"region" json:"region"`

	// Endpoint for S3-compatible services (optional)
	Endpoint string `toml:"endpoint" json:"endpoint"`

	// AccessKeyID for authentication
	AccessKeyID string `toml:"access_key_id" json:"access_key_id"`

	// SecretAccessKey for authentication
	SecretAccessKey string `toml:"secret_access_key" json:"secret_access_key"`

	// SessionToken for temporary credentials (optional)
	SessionToken string `toml:"session_token" json:"session_token"`

	// UseSSL enables HTTPS for API calls
	UseSSL bool `toml:"use_ssl" json:"use_ssl"`

	// PathStyle forces path-style addressing
	PathStyle bool `toml:"path_style" json:"path_style"`

	// Prefix for all objects in the bucket
	Prefix string `toml:"prefix" json:"prefix"`

	// StorageClass for uploaded objects
	StorageClass string `toml:"storage_class" json:"storage_class"`

	// ServerSideEncryption enables SSE
	ServerSideEncryption string `toml:"server_side_encryption" json:"server_side_encryption"`

	// KMSKeyID for SSE-KMS encryption
	KMSKeyID string `toml:"kms_key_id" json:"kms_key_id"`

	// RetryAttempts for failed operations
	RetryAttempts int `toml:"retry_attempts" json:"retry_attempts"`

	// RetryDelay base delay between retries (milliseconds)
	RetryDelay int `toml:"retry_delay" json:"retry_delay"`

	// RequestTimeout for individual requests (seconds)
	RequestTimeout int `toml:"request_timeout" json:"request_timeout"`

	// EnableVersioning enables S3 bucket versioning
	EnableVersioning bool `toml:"enable_versioning" json:"enable_versioning"`
}

// CacheConfig configures the caching layer
type CacheConfig struct {
	// Enabled turns on/off caching
	Enabled bool `toml:"enabled" json:"enabled"`

	// AutoEnable automatically enables caching for remote storage
	AutoEnable bool `toml:"auto_enable_for_remote" json:"auto_enable_for_remote"`

	// Memory cache configuration
	Memory MemoryCacheConfig `toml:"memory" json:"memory"`

	// Disk cache configuration
	Disk DiskCacheConfig `toml:"disk" json:"disk"`
}

// MemoryCacheConfig configures in-memory caching
type MemoryCacheConfig struct {
	// Enabled turns on/off memory caching
	Enabled bool `toml:"enabled" json:"enabled"`

	// MaxSizeMB is the maximum memory cache size in MB
	MaxSizeMB int `toml:"max_size_mb" json:"max_size_mb"`

	// MaxItems is the maximum number of items to cache
	MaxItems int `toml:"max_items" json:"max_items"`

	// TTLMinutes is the cache TTL in minutes
	TTLMinutes int `toml:"ttl_minutes" json:"ttl_minutes"`
}

// DiskCacheConfig configures disk-based caching
type DiskCacheConfig struct {
	// Enabled turns on/off disk caching
	Enabled bool `toml:"enabled" json:"enabled"`

	// Path is the cache directory
	Path string `toml:"path" json:"path"`

	// MaxSizeMB is the maximum disk cache size in MB
	MaxSizeMB int `toml:"max_size_mb" json:"max_size_mb"`

	// TTLHours is the cache TTL in hours
	TTLHours int `toml:"ttl_hours" json:"ttl_hours"`

	// CleanupIntervalHours is how often to clean expired entries
	CleanupIntervalHours int `toml:"cleanup_interval_hours" json:"cleanup_interval_hours"`
}

// StorageError represents an error from a storage backend
type StorageError struct {
	Backend   StorageType
	Operation string
	Path      string
	Err       error
	Retryable bool
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error [%s:%s] %s: %v",
		e.Backend, e.Operation, e.Path, e.Err)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

func (e *StorageError) IsRetryable() bool {
	return e.Retryable
}

// NewStorageError creates a new storage error
func NewStorageError(backend StorageType, operation, path string, err error, retryable bool) *StorageError {
	return &StorageError{
		Backend:   backend,
		Operation: operation,
		Path:      path,
		Err:       err,
		Retryable: retryable,
	}
}

// IsNotFound returns true if the error indicates a file was not found
func IsNotFound(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.Operation == "read" || storageErr.Operation == "stat"
	}
	return false
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.IsRetryable()
	}
	return false
}
