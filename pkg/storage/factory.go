package storage

import (
	"fmt"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage/local"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage/s3"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Factory creates storage backends based on configuration
type Factory struct{}

// NewFactory creates a new storage factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateStorage creates a storage backend based on the provided configuration
func (f *Factory) CreateStorage(config types.StorageConfig) (types.StorageBackend, error) {
	switch config.Type {
	case types.StorageTypeLocal:
		return local.New(config.Local)
	case types.StorageTypeS3:
		return s3.NewStorage(config.S3)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// ValidateConfig validates a storage configuration without creating the backend
func (f *Factory) ValidateConfig(config types.StorageConfig) error {
	switch config.Type {
	case types.StorageTypeLocal:
		return validateLocalConfig(config.Local)
	case types.StorageTypeS3:
		return validateS3Config(config.S3)
	default:
		return fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// GetSupportedTypes returns a list of supported storage types
func (f *Factory) GetSupportedTypes() []types.StorageType {
	return []types.StorageType{
		types.StorageTypeLocal,
		types.StorageTypeS3,
	}
}

// validateLocalConfig validates local storage configuration
func validateLocalConfig(config types.LocalStorageConfig) error {
	if config.Path == "" {
		return fmt.Errorf("local storage path cannot be empty")
	}
	return nil
}

// validateS3Config validates S3 storage configuration
func validateS3Config(config types.S3StorageConfig) error {
	if config.Bucket == "" {
		return fmt.Errorf("S3 bucket name cannot be empty")
	}
	if config.Region == "" {
		return fmt.Errorf("S3 region cannot be empty")
	}
	return nil
}

// DefaultFactory is the default storage factory instance
var DefaultFactory = NewFactory()

// CreateStorage creates a storage backend using the default factory
func CreateStorage(config types.StorageConfig) (types.StorageBackend, error) {
	return DefaultFactory.CreateStorage(config)
}

// ValidateConfig validates a storage configuration using the default factory
func ValidateConfig(config types.StorageConfig) error {
	return DefaultFactory.ValidateConfig(config)
}

// GetSupportedTypes returns supported storage types using the default factory
func GetSupportedTypes() []types.StorageType {
	return DefaultFactory.GetSupportedTypes()
}