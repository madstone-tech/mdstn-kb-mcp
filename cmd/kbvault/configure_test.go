package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNewConfigureCmd(t *testing.T) {
	cmd := newConfigureCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "configure", cmd.Use)
	assert.Contains(t, cmd.Long, "Configure kbvault profiles interactively")
}

func TestPromptWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{
			name:         "empty input uses default",
			input:        "\n",
			defaultValue: "default-value",
			expected:     "default-value",
		},
		{
			name:         "non-empty input overrides default",
			input:        "user-input\n",
			defaultValue: "default-value",
			expected:     "user-input",
		},
		{
			name:         "whitespace trimmed",
			input:        "  trimmed  \n",
			defaultValue: "default",
			expected:     "trimmed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			result := promptWithDefault(scanner, "test prompt", tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPromptSensitive(t *testing.T) {
	input := "sensitive-data\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	result := promptSensitive(scanner, "Enter sensitive data")
	assert.Equal(t, "sensitive-data", result)
}

func TestPromptBoolWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "empty input uses default true",
			input:        "\n",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "empty input uses default false",
			input:        "\n",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "y input returns true",
			input:        "y\n",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "yes input returns true",
			input:        "yes\n",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "n input returns false",
			input:        "n\n",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "no input returns false",
			input:        "no\n",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "case insensitive Y",
			input:        "Y\n",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "case insensitive NO",
			input:        "NO\n",
			defaultValue: true,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			result := promptBoolWithDefault(scanner, "test prompt", tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPromptBoolWithDefault_InvalidInput(t *testing.T) {
	// Test with invalid input followed by valid input
	input := "invalid\ny\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	result := promptBoolWithDefault(scanner, "test prompt", false)
	assert.True(t, result)
}

func TestConfirmPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "y returns true",
			input:    "y\n",
			expected: true,
		},
		{
			name:     "yes returns true",
			input:    "yes\n",
			expected: true,
		},
		{
			name:     "Y returns true",
			input:    "Y\n",
			expected: true,
		},
		{
			name:     "YES returns true",
			input:    "YES\n",
			expected: true,
		},
		{
			name:     "n returns false",
			input:    "n\n",
			expected: false,
		},
		{
			name:     "no returns false",
			input:    "no\n",
			expected: false,
		},
		{
			name:     "empty returns false",
			input:    "\n",
			expected: false,
		},
		{
			name:     "anything else returns false",
			input:    "maybe\n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			result := confirmPrompt(scanner, "test prompt")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProfileExists(t *testing.T) {
	// Set up temporary home directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	pm, err := config.NewProfileManager()
	require.NoError(t, err)

	// Test default profile exists
	assert.True(t, profileExists(pm, "default"))

	// Test non-existent profile
	assert.False(t, profileExists(pm, "non-existent"))

	// Create a profile and test it exists
	err = pm.CreateProfile("test-profile", nil)
	require.NoError(t, err)
	assert.True(t, profileExists(pm, "test-profile"))
}

func TestConfigureVault(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: accept all defaults
	input := "\n\n\n\n\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureVault(scanner, config)
	assert.NoError(t, err)

	// Should still have default values
	assert.Equal(t, "my-kb", config.Vault.Name)
}

func TestConfigureVault_CustomValues(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input with custom values
	input := "custom-vault\ncustom-notes\ncustom-daily\ncustom-templates\ncustom-template\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureVault(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "custom-vault", config.Vault.Name)
	assert.Equal(t, "custom-notes", config.Vault.NotesDir)
	assert.Equal(t, "custom-daily", config.Vault.DailyDir)
	assert.Equal(t, "custom-templates", config.Vault.TemplatesDir)
	assert.Equal(t, "custom-template", config.Vault.DefaultTemplate)
}

func TestConfigureLocalStorage(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: custom path, enable auto-create dirs, disable locking
	input := "/custom/path\ny\nn\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureLocalStorage(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "/custom/path", config.Storage.Local.Path)
	assert.True(t, config.Storage.Local.CreateDirs)
	assert.False(t, config.Storage.Local.EnableLocking)
}

func TestConfigureS3Storage_Basic(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: bucket, region, no custom endpoint, no prefix, no encryption, no credentials
	input := "my-bucket\nus-east-1\nn\nn\nn\nn\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureS3Storage(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "my-bucket", config.Storage.S3.Bucket)
	assert.Equal(t, "us-east-1", config.Storage.S3.Region)
	assert.Equal(t, "", config.Storage.S3.Endpoint)
	assert.Equal(t, "", config.Storage.S3.Prefix)
}

func TestConfigureS3Storage_WithEndpointAndPrefix(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: bucket, region, custom endpoint, prefix, no encryption, no credentials
	input := "my-bucket\nus-west-2\ny\nhttp://localhost:9000\ny\nvault-prefix\nn\nn\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureS3Storage(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "my-bucket", config.Storage.S3.Bucket)
	assert.Equal(t, "us-west-2", config.Storage.S3.Region)
	assert.Equal(t, "http://localhost:9000", config.Storage.S3.Endpoint)
	assert.Equal(t, "vault-prefix", config.Storage.S3.Prefix)
}

func TestConfigureS3Encryption_AES256(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: choose AES256 encryption
	input := "1\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureS3Encryption(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "AES256", config.Storage.S3.ServerSideEncryption)
	assert.Equal(t, "", config.Storage.S3.KMSKeyID)
}

func TestConfigureS3Encryption_KMS(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: choose KMS encryption without custom key
	input := "2\nn\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureS3Encryption(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "aws:kms", config.Storage.S3.ServerSideEncryption)
	assert.Equal(t, "", config.Storage.S3.KMSKeyID)
}

func TestConfigureS3Encryption_KMSWithKey(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: choose KMS encryption with custom key
	input := "2\ny\nmy-kms-key-id\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureS3Encryption(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "aws:kms", config.Storage.S3.ServerSideEncryption)
	assert.Equal(t, "my-kms-key-id", config.Storage.S3.KMSKeyID)
}

func TestConfigureAWSCredentials_Skip(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: don't store credentials
	input := "n\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureAWSCredentials(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "", config.Storage.S3.AccessKeyID)
	assert.Equal(t, "", config.Storage.S3.SecretAccessKey)
	assert.Equal(t, "", config.Storage.S3.SessionToken)
}

func TestConfigureAWSCredentials_Basic(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: store basic credentials
	input := "y\nmy-access-key\nmy-secret-key\nn\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureAWSCredentials(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "my-access-key", config.Storage.S3.AccessKeyID)
	assert.Equal(t, "my-secret-key", config.Storage.S3.SecretAccessKey)
	assert.Equal(t, "", config.Storage.S3.SessionToken)
}

func TestConfigureAWSCredentials_WithSessionToken(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: store credentials with session token
	input := "y\nmy-access-key\nmy-secret-key\ny\nmy-session-token\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureAWSCredentials(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, "my-access-key", config.Storage.S3.AccessKeyID)
	assert.Equal(t, "my-secret-key", config.Storage.S3.SecretAccessKey)
	assert.Equal(t, "my-session-token", config.Storage.S3.SessionToken)
}

func TestConfigureServer_Disabled(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: disable server
	input := "n\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureServer(scanner, config)
	assert.NoError(t, err)

	assert.False(t, config.Server.HTTP.Enabled)
}

func TestConfigureServer_Enabled(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: enable server with custom settings
	input := "y\n0.0.0.0\n9090\ny\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureServer(scanner, config)
	assert.NoError(t, err)

	assert.True(t, config.Server.HTTP.Enabled)
	assert.Equal(t, "0.0.0.0", config.Server.HTTP.Host)
	assert.Equal(t, 9090, config.Server.HTTP.Port)
	assert.True(t, config.Server.HTTP.EnableCORS)
}

func TestConfigureStorage_Local(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: choose local storage with custom path
	input := "local\n/custom/vault/path\ny\nn\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureStorage(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, types.StorageTypeLocal, config.Storage.Type)
	assert.Equal(t, "/custom/vault/path", config.Storage.Local.Path)
	assert.True(t, config.Storage.Local.CreateDirs)
	assert.False(t, config.Storage.Local.EnableLocking)
}

func TestConfigureStorage_S3(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: choose S3 storage with basic settings
	input := "s3\nmy-test-bucket\nus-west-1\nn\nn\nn\nn\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureStorage(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, types.StorageTypeS3, config.Storage.Type)
	assert.Equal(t, "my-test-bucket", config.Storage.S3.Bucket)
	assert.Equal(t, "us-west-1", config.Storage.S3.Region)
}

func TestConfigureStorage_InvalidType(t *testing.T) {
	config := types.DefaultConfig()

	// Simulate user input: invalid type followed by valid type
	input := "invalid\nlocal\n/tmp/vault\ny\ny\n"
	scanner := bufio.NewScanner(strings.NewReader(input))

	err := configureStorage(scanner, config)
	assert.NoError(t, err)

	assert.Equal(t, types.StorageTypeLocal, config.Storage.Type)
}
