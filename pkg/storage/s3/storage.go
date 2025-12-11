package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Storage implements the StorageBackend interface for AWS S3
type Storage struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	config     types.S3StorageConfig
}

// NewStorage creates a new S3 storage backend
func NewStorage(cfg types.S3StorageConfig) (*Storage, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid S3 configuration: %w", err)
	}

	// Create AWS config
	awsConfig, err := createAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		if cfg.PathStyle {
			o.UsePathStyle = true
		}
	})

	// Create upload and download managers
	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	storage := &Storage{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		config:     cfg,
	}

	return storage, nil
}

// Type returns the storage backend type
func (s *Storage) Type() types.StorageType {
	return types.StorageTypeS3
}

// Read retrieves a file's content by path
func (s *Storage) Read(ctx context.Context, path string) ([]byte, error) {
	key := s.buildKey(path)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, s.handleError("read", path, err)
	}
	defer func() { _ = result.Body.Close() }()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, types.NewStorageError(types.StorageTypeS3, "read", path, err, false)
	}

	return data, nil
}

// Write stores content at the given path
func (s *Storage) Write(ctx context.Context, path string, data []byte) error {
	key := s.buildKey(path)

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(string(data)),
	}

	// Add server-side encryption if configured
	if s.config.ServerSideEncryption != "" {
		input.ServerSideEncryption = s3types.ServerSideEncryption(s.config.ServerSideEncryption)
		if s.config.KMSKeyID != "" {
			input.SSEKMSKeyId = aws.String(s.config.KMSKeyID)
		}
	}

	// Set storage class if configured
	if s.config.StorageClass != "" {
		input.StorageClass = s3types.StorageClass(s.config.StorageClass)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return s.handleError("write", path, err)
	}

	return nil
}

// Delete removes a file at the given path
func (s *Storage) Delete(ctx context.Context, path string) error {
	key := s.buildKey(path)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return s.handleError("delete", path, err)
	}

	return nil
}

// Exists checks if a file exists at the given path
func (s *Storage) Exists(ctx context.Context, path string) (bool, error) {
	key := s.buildKey(path)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var nsk *s3types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}
		return false, s.handleError("exists", path, err)
	}

	return true, nil
}

// List returns all files matching the given prefix
func (s *Storage) List(ctx context.Context, prefix string) ([]string, error) {
	key := s.buildKey(prefix)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.Bucket),
		Prefix: aws.String(key),
	}

	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, s.handleError("list", prefix, err)
		}

		for _, obj := range output.Contents {
			if obj.Key != nil {
				// Remove the prefix to get relative path
				relativePath := strings.TrimPrefix(*obj.Key, s.config.Prefix)
				if relativePath != "" {
					files = append(files, relativePath)
				}
			}
		}
	}

	return files, nil
}

// Stat returns metadata about a file
func (s *Storage) Stat(ctx context.Context, path string) (*types.FileInfo, error) {
	key := s.buildKey(path)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, s.handleError("stat", path, err)
	}

	info := &types.FileInfo{
		Path:    path,
		Size:    aws.ToInt64(result.ContentLength),
		ModTime: result.LastModified.Unix(),
	}

	// Add S3-specific metadata
	if result.ETag != nil {
		info.ETag = *result.ETag
	}
	if result.ContentType != nil {
		info.ContentType = *result.ContentType
	}
	if result.StorageClass != "" {
		info.StorageClass = string(result.StorageClass)
	}

	// Add custom metadata
	if len(result.Metadata) > 0 {
		info.Metadata = make(map[string]string)
		for k, v := range result.Metadata {
			info.Metadata[k] = v
		}
	}

	return info, nil
}

// ReadStream returns a reader for streaming large files
func (s *Storage) ReadStream(ctx context.Context, path string) (io.ReadCloser, error) {
	key := s.buildKey(path)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, s.handleError("read_stream", path, err)
	}

	return result.Body, nil
}

// WriteStream writes data from a reader to the given path
func (s *Storage) WriteStream(ctx context.Context, path string, reader io.Reader) error {
	key := s.buildKey(path)

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	// Add server-side encryption if configured
	if s.config.ServerSideEncryption != "" {
		input.ServerSideEncryption = s3types.ServerSideEncryption(s.config.ServerSideEncryption)
		if s.config.KMSKeyID != "" {
			input.SSEKMSKeyId = aws.String(s.config.KMSKeyID)
		}
	}

	// Set storage class if configured
	if s.config.StorageClass != "" {
		input.StorageClass = s3types.StorageClass(s.config.StorageClass)
	}

	_, err := s.uploader.Upload(ctx, input)
	if err != nil {
		return s.handleError("write_stream", path, err)
	}

	return nil
}

// Copy copies a file from src to dst within the same backend
func (s *Storage) Copy(ctx context.Context, src, dst string) error {
	srcKey := s.buildKey(src)
	dstKey := s.buildKey(dst)

	copySource := fmt.Sprintf("%s/%s", s.config.Bucket, srcKey)

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(s.config.Bucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(copySource),
	}

	// Add server-side encryption if configured
	if s.config.ServerSideEncryption != "" {
		input.ServerSideEncryption = s3types.ServerSideEncryption(s.config.ServerSideEncryption)
		if s.config.KMSKeyID != "" {
			input.SSEKMSKeyId = aws.String(s.config.KMSKeyID)
		}
	}

	_, err := s.client.CopyObject(ctx, input)
	if err != nil {
		return s.handleError("copy", fmt.Sprintf("%s->%s", src, dst), err)
	}

	return nil
}

// Move moves/renames a file from src to dst
func (s *Storage) Move(ctx context.Context, src, dst string) error {
	// Copy first
	if err := s.Copy(ctx, src, dst); err != nil {
		return err
	}

	// Then delete source
	if err := s.Delete(ctx, src); err != nil {
		// Try to cleanup the copy on failure
		_ = s.Delete(ctx, dst)
		return err
	}

	return nil
}

// Health performs a health check on the storage backend
func (s *Storage) Health(ctx context.Context) error {
	// Try to list objects with a limit to check connectivity
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.config.Bucket),
		MaxKeys: aws.Int32(1),
	}

	_, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return s.handleError("health", "", err)
	}

	return nil
}

// Close cleanly shuts down the storage backend
func (s *Storage) Close() error {
	// S3 client doesn't require explicit closing
	return nil
}

// buildKey constructs the full S3 key with prefix
func (s *Storage) buildKey(path string) string {
	if s.config.Prefix == "" {
		return path
	}
	return strings.TrimSuffix(s.config.Prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}

// handleError converts AWS S3 errors to storage errors
func (s *Storage) handleError(operation, path string, err error) error {
	// Check for retryable errors
	retryable := isRetryableS3Error(err)

	return types.NewStorageError(types.StorageTypeS3, operation, path, err, retryable)
}

// isRetryableS3Error determines if an S3 error is retryable
func isRetryableS3Error(err error) bool {
	// Check for smithy API errors
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		// Check error code
		switch apiErr.ErrorCode() {
		case "InternalError", "ServiceUnavailable", "SlowDown", "RequestTimeout":
			return true
		}

		// Check HTTP status code for retryable errors
		if httpErr, ok := apiErr.(interface{ HTTPStatusCode() int }); ok {
			code := httpErr.HTTPStatusCode()
			return code >= 500 || code == 429 || code == 408
		}
	}

	return false
}

// validateConfig validates the S3 configuration
func validateConfig(cfg types.S3StorageConfig) error {
	if cfg.Bucket == "" {
		return fmt.Errorf("bucket name cannot be empty")
	}

	if cfg.Region == "" {
		return fmt.Errorf("region cannot be empty")
	}

	// Validate retry configuration
	if cfg.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if cfg.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	if cfg.RequestTimeout < 0 {
		return fmt.Errorf("request timeout cannot be negative")
	}

	return nil
}

// createAWSConfig creates AWS SDK configuration from S3 config
func createAWSConfig(cfg types.S3StorageConfig) (aws.Config, error) {
	ctx := context.Background()

	var opts []func(*config.LoadOptions) error

	// Set region
	opts = append(opts, config.WithRegion(cfg.Region))

	// Set custom endpoint if provided
	if cfg.Endpoint != "" {
		// Note: Custom endpoint handling - using deprecated API for compatibility
		// TODO: Migrate to service-specific endpoint resolution in future version
		// nolint:staticcheck
		opts = append(opts, config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				// nolint:staticcheck
				return aws.Endpoint{
					URL:           cfg.Endpoint,
					SigningRegion: cfg.Region,
				}, nil
			})))
	}

	// Set credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		creds := credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			cfg.SessionToken,
		)
		opts = append(opts, config.WithCredentialsProvider(creds))
	}

	// Set retry configuration
	if cfg.RetryAttempts > 0 {
		retryMode := aws.RetryModeAdaptive
		opts = append(opts, config.WithRetryMode(retryMode))
		opts = append(opts, config.WithRetryMaxAttempts(cfg.RetryAttempts))
	}

	// Set HTTP client configuration
	if cfg.RequestTimeout > 0 {
		httpClient := &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		}
		opts = append(opts, config.WithHTTPClient(httpClient))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}
