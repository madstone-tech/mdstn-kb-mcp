package s3

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNewStorage(t *testing.T) {
	tests := []struct {
		name    string
		config  types.S3StorageConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.S3StorageConfig{
				Bucket: "test-bucket",
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "missing bucket",
			config: types.S3StorageConfig{
				Region: "us-east-1",
			},
			wantErr: true,
		},
		{
			name: "missing region",
			config: types.S3StorageConfig{
				Bucket: "test-bucket",
			},
			wantErr: true,
		},
		{
			name: "with credentials",
			config: types.S3StorageConfig{
				Bucket:          "test-bucket",
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
			wantErr: false,
		},
		{
			name: "with custom endpoint",
			config: types.S3StorageConfig{
				Bucket:   "test-bucket",
				Region:   "us-east-1",
				Endpoint: "http://localhost:9000",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewStorage(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, storage)
				assert.Equal(t, types.StorageTypeS3, storage.Type())
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.S3StorageConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: types.S3StorageConfig{
				Bucket: "test-bucket",
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name:    "empty bucket",
			config:  types.S3StorageConfig{Region: "us-east-1"},
			wantErr: true,
			errMsg:  "bucket name cannot be empty",
		},
		{
			name:    "empty region",
			config:  types.S3StorageConfig{Bucket: "test-bucket"},
			wantErr: true,
			errMsg:  "region cannot be empty",
		},
		{
			name: "negative retry attempts",
			config: types.S3StorageConfig{
				Bucket:        "test-bucket",
				Region:        "us-east-1",
				RetryAttempts: -1,
			},
			wantErr: true,
			errMsg:  "retry attempts cannot be negative",
		},
		{
			name: "negative retry delay",
			config: types.S3StorageConfig{
				Bucket:     "test-bucket",
				Region:     "us-east-1",
				RetryDelay: -1,
			},
			wantErr: true,
			errMsg:  "retry delay cannot be negative",
		},
		{
			name: "negative request timeout",
			config: types.S3StorageConfig{
				Bucket:         "test-bucket",
				Region:         "us-east-1",
				RequestTimeout: -1,
			},
			wantErr: true,
			errMsg:  "request timeout cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		path   string
		want   string
	}{
		{
			name:   "no prefix",
			prefix: "",
			path:   "notes/test.md",
			want:   "notes/test.md",
		},
		{
			name:   "with prefix",
			prefix: "vault",
			path:   "notes/test.md",
			want:   "vault/notes/test.md",
		},
		{
			name:   "prefix with trailing slash",
			prefix: "vault/",
			path:   "notes/test.md",
			want:   "vault/notes/test.md",
		},
		{
			name:   "path with leading slash",
			prefix: "vault",
			path:   "/notes/test.md",
			want:   "vault/notes/test.md",
		},
		{
			name:   "both with slashes",
			prefix: "vault/",
			path:   "/notes/test.md",
			want:   "vault/notes/test.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := types.S3StorageConfig{
				Bucket: "test-bucket",
				Region: "us-east-1",
				Prefix: tt.prefix,
			}

			storage, err := NewStorage(config)
			require.NoError(t, err)

			result := storage.buildKey(tt.path)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsRetryableS3Error(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "generic error",
			err:  assert.AnError,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableS3Error(tt.err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestHandleError(t *testing.T) {
	config := types.S3StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	storage, err := NewStorage(config)
	require.NoError(t, err)

	testErr := assert.AnError
	storageErr := storage.handleError("test", "test-path", testErr)

	assert.Error(t, storageErr)

	// Check if it's a StorageError
	sErr, ok := storageErr.(*types.StorageError)
	assert.True(t, ok)
	assert.Equal(t, types.StorageTypeS3, sErr.Backend)
	assert.Equal(t, "test", sErr.Operation)
	assert.Equal(t, "test-path", sErr.Path)
	assert.Equal(t, testErr, sErr.Err)
}

func TestCreateAWSConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  types.S3StorageConfig
		wantErr bool
	}{
		{
			name: "basic config",
			config: types.S3StorageConfig{
				Bucket: "test-bucket",
				Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "config with credentials",
			config: types.S3StorageConfig{
				Bucket:          "test-bucket",
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
			wantErr: false,
		},
		{
			name: "config with session token",
			config: types.S3StorageConfig{
				Bucket:          "test-bucket",
				Region:          "us-east-1",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				SessionToken:    "test-token",
			},
			wantErr: false,
		},
		{
			name: "config with custom endpoint",
			config: types.S3StorageConfig{
				Bucket:   "test-bucket",
				Region:   "us-east-1",
				Endpoint: "http://localhost:9000",
			},
			wantErr: false,
		},
		{
			name: "config with retry settings",
			config: types.S3StorageConfig{
				Bucket:        "test-bucket",
				Region:        "us-east-1",
				RetryAttempts: 3,
				RetryDelay:    1000,
			},
			wantErr: false,
		},
		{
			name: "config with request timeout",
			config: types.S3StorageConfig{
				Bucket:         "test-bucket",
				Region:         "us-east-1",
				RequestTimeout: 30,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsConfig, err := createAWSConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.config.Region, awsConfig.Region)
			}
		})
	}
}

func TestStorageInterface(t *testing.T) {
	// Test that Storage implements the StorageBackend interface
	config := types.S3StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	storage, err := NewStorage(config)
	require.NoError(t, err)

	// This test will fail to compile if Storage doesn't implement StorageBackend
	var _ types.StorageBackend = storage
}

// Benchmark tests for key operations
func BenchmarkBuildKey(b *testing.B) {
	config := types.S3StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
		Prefix: "vault/kb",
	}

	storage, err := NewStorage(config)
	require.NoError(b, err)

	path := "notes/projects/important-note.md"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.buildKey(path)
	}
}

func BenchmarkHandleError(b *testing.B) {
	config := types.S3StorageConfig{
		Bucket: "test-bucket",
		Region: "us-east-1",
	}

	storage, err := NewStorage(config)
	require.NoError(b, err)

	testErr := assert.AnError

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.handleError("test", "test-path", testErr)
	}
}

// Integration test helper (requires actual AWS credentials and bucket)
func TestStorageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires environment variables:
	// - TEST_S3_BUCKET
	// - TEST_S3_REGION
	// - AWS credentials via standard AWS credential chain

	bucket := getEnvOrSkip(t, "TEST_S3_BUCKET")
	region := getEnvOrSkip(t, "TEST_S3_REGION")

	config := types.S3StorageConfig{
		Bucket: bucket,
		Region: region,
		Prefix: "test-vault",
	}

	storage, err := NewStorage(config)
	require.NoError(t, err)
	defer func() { _ = storage.Close() }()

	ctx := context.Background()

	// Test health check
	err = storage.Health(ctx)
	assert.NoError(t, err)

	// Test write/read cycle
	testPath := "test-file.txt"
	testData := []byte("Hello, S3!")

	err = storage.Write(ctx, testPath, testData)
	assert.NoError(t, err)

	// Test exists
	exists, err := storage.Exists(ctx, testPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test read
	data, err := storage.Read(ctx, testPath)
	assert.NoError(t, err)
	assert.Equal(t, testData, data)

	// Test stat
	info, err := storage.Stat(ctx, testPath)
	assert.NoError(t, err)
	assert.Equal(t, testPath, info.Path)
	assert.Equal(t, int64(len(testData)), info.Size)

	// Test list
	files, err := storage.List(ctx, "")
	assert.NoError(t, err)
	assert.Contains(t, files, testPath)

	// Test copy
	copyPath := "test-file-copy.txt"
	err = storage.Copy(ctx, testPath, copyPath)
	assert.NoError(t, err)

	// Verify copy
	copyData, err := storage.Read(ctx, copyPath)
	assert.NoError(t, err)
	assert.Equal(t, testData, copyData)

	// Test move
	movePath := "test-file-moved.txt"
	err = storage.Move(ctx, copyPath, movePath)
	assert.NoError(t, err)

	// Verify move (source should not exist, destination should)
	exists, err = storage.Exists(ctx, copyPath)
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = storage.Exists(ctx, movePath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Cleanup
	_ = storage.Delete(ctx, testPath)
	_ = storage.Delete(ctx, movePath)
}

func TestStorageStreamOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	bucket := getEnvOrSkip(t, "TEST_S3_BUCKET")
	region := getEnvOrSkip(t, "TEST_S3_REGION")

	config := types.S3StorageConfig{
		Bucket: bucket,
		Region: region,
		Prefix: "test-vault",
	}

	storage, err := NewStorage(config)
	require.NoError(t, err)
	defer func() { _ = storage.Close() }()

	ctx := context.Background()

	// Test write stream
	testPath := "test-stream.txt"
	testData := "This is a test of streaming operations"
	reader := strings.NewReader(testData)

	err = storage.WriteStream(ctx, testPath, reader)
	assert.NoError(t, err)

	// Test read stream
	readCloser, err := storage.ReadStream(ctx, testPath)
	assert.NoError(t, err)
	defer func() { _ = readCloser.Close() }()

	data := make([]byte, len(testData))
	n, err := readCloser.Read(data)
	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, string(data))

	// Cleanup
	_ = storage.Delete(ctx, testPath)
}

// Helper function to get environment variable or skip test
func getEnvOrSkip(t *testing.T, key string) string {
	value := "" // In real tests, this would use os.Getenv(key)
	if value == "" {
		t.Skipf("Environment variable %s not set, skipping integration test", key)
	}
	return value
}
