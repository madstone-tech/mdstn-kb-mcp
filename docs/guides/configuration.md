# Configuration Guide

Learn how to configure kbVault for your needs.

## Configuration Files

kbVault uses TOML configuration files located in `.kbvault/config.toml` within your vault directory.

### Finding Your Configuration

Your configuration is stored at:
- **Local vault**: `.kbvault/config.toml` in your vault directory
- **Profile-based**: `~/.config/kbvault/profiles/<profile-name>/config.toml`

View your current configuration:
```bash
kbvault config show
```

## Configuration Structure

### Basic Configuration

```toml
# Vault information
[vault]
name = "My Knowledge Vault"
version = "1.0"

# Storage configuration
[storage]
type = "local"
path = "./notes"

# Optional settings
[settings]
auto_index = true
case_sensitive_search = false
max_cache_size = "100mb"
```

### Complete Example

```toml
[vault]
name = "My Knowledge Vault"
description = "Personal knowledge base"
version = "1.0"

[storage]
type = "local"
path = "./notes"

[search]
engine = "built-in"
index_type = "inverted"

[vector]
enabled = false
engine = "none"

[settings]
auto_index = true
case_sensitive_search = false
max_cache_size = "100mb"
editor = "vim"
date_format = "2006-01-02"
```

## Storage Configuration

### Local Storage (Default)

Store notes in your local filesystem:

```toml
[storage]
type = "local"
path = "./notes"
```

**Options:**
- `type` - Must be `"local"`
- `path` - Directory path for notes (absolute or relative)

**Example:**
```bash
kbvault config set storage.type local
kbvault config set storage.path ~/my-notes
```

### S3 Storage

Store notes in S3-compatible storage:

```toml
[storage]
type = "s3"
bucket = "my-kb-bucket"
region = "us-east-1"
endpoint = "s3.amazonaws.com"  # Optional, for S3-compatible services

[storage.credentials]
access_key = "AKIAIOSFODNN7EXAMPLE"
secret_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
```

**Options:**
- `bucket` - S3 bucket name
- `region` - AWS region
- `endpoint` - Custom endpoint (optional, for MinIO, etc.)
- `credentials.access_key` - AWS access key
- `credentials.secret_key` - AWS secret key

**Using Environment Variables:**

Instead of hardcoding credentials, use environment variables:

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_REGION=us-east-1
```

Configuration:
```toml
[storage]
type = "s3"
bucket = "my-kb-bucket"
region = "${AWS_REGION}"  # Will use env var if set
```

**Example Setup:**
```bash
# Configure S3 storage
kbvault configure

# Or manually set:
kbvault config set storage.type s3
kbvault config set storage.bucket my-kb-bucket
kbvault config set storage.region us-east-1
```

### MinIO (S3-Compatible)

Use MinIO as your storage backend:

```toml
[storage]
type = "s3"
bucket = "kbvault"
region = "us-east-1"
endpoint = "http://localhost:9000"

[storage.credentials]
access_key = "minioadmin"
secret_key = "minioadmin"
```

## Search Configuration

### Built-in Search Engine

The default search engine is fast and requires no additional setup:

```toml
[search]
engine = "built-in"
index_type = "inverted"
```

**Options:**
- `engine` - Search engine type (`"built-in"`)
- `index_type` - Index type (`"inverted"` for full-text search)

## Vector Database Configuration

### Disabled (Default)

```toml
[vector]
enabled = false
engine = "none"
```

### Future: Vector Search

When semantic search is enabled:

```toml
[vector]
enabled = true
engine = "pinecone"  # or "local", "milvus", etc.

[vector.pinecone]
api_key = "${PINECONE_API_KEY}"
index_name = "kbvault"
environment = "us-west-1"
```

## Settings

### General Settings

```toml
[settings]
# Auto-index notes after changes
auto_index = true

# Case-sensitive search
case_sensitive_search = false

# Maximum cache size
max_cache_size = "100mb"

# Default editor for editing notes
editor = "vim"

# Date format (Go time format)
date_format = "2006-01-02"

# Time format (Go time format)
time_format = "15:04:05"
```

### Performance Settings

```toml
[settings]
# Cache settings
max_cache_size = "100mb"
cache_ttl = "1h"
cache_type = "memory"  # or "disk"

# Search settings
batch_size = 1000
max_results = 10000

# Indexing settings
auto_index = true
index_batch_size = 100
```

## Profiles

Use profiles to manage multiple vault configurations. See [Profiles & Multi-Vault Guide](profiles.md) for details.

## Configuration via CLI

### View Configuration

```bash
# Show all configuration
kbvault config show

# Get specific value
kbvault config get storage.type
kbvault config get vault.name
```

### Set Configuration

```bash
# Set top-level value
kbvault config set vault.name "My Vault"

# Set nested value
kbvault config set storage.path ~/my-notes
kbvault config set storage.bucket my-bucket

# Set array value
kbvault config set settings.tags "python,tutorial"
```

### Validate Configuration

```bash
kbvault config validate
```

This checks:
- Required fields are present
- File paths exist
- S3 credentials are valid
- Storage backend is accessible

## Interactive Configuration

Use the interactive setup wizard:

```bash
kbvault configure
```

This will prompt you for:
1. Vault name
2. Storage type (local or S3)
3. Storage path/bucket
4. Search engine preferences
5. Editor preference

### Reset Configuration

```bash
# Reset to defaults
kbvault configure --reset

# Reset specific profile
kbvault --profile work configure --reset
```

## Environment Variables

kbVault respects these environment variables:

```bash
# Configuration directory
export KBVAULT_CONFIG=~/.kbvault

# Default profile
export KBVAULT_PROFILE=work

# Editor
export EDITOR=vim

# AWS credentials (for S3)
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1
```

## Common Configurations

### Personal Knowledge Base (Local)

```toml
[vault]
name = "Personal KB"

[storage]
type = "local"
path = "~/personal-kb/notes"

[settings]
auto_index = true
```

### Work Knowledge Base (S3)

```toml
[vault]
name = "Work KB"

[storage]
type = "s3"
bucket = "company-kb"
region = "us-east-1"

[settings]
auto_index = true
case_sensitive_search = false
```

### Research Projects (Local with Performance)

```toml
[vault]
name = "Research"

[storage]
type = "local"
path = "~/research/notes"

[settings]
auto_index = true
max_cache_size = "500mb"
cache_ttl = "24h"
```

## Configuration Files Location

Default locations:

```
# Local vault config
<vault-directory>/.kbvault/config.toml

# Profile configs
~/.config/kbvault/profiles/<profile-name>/config.toml

# Global config (if supported)
~/.kbvault/config.toml
```

## Troubleshooting

### "Configuration not found"

Make sure you're in a vault directory or using `--profile`:

```bash
# Initialize vault
kbvault init

# Or use profile
kbvault --profile work list
```

### "Invalid configuration"

Validate your configuration:

```bash
kbvault config validate
```

Check for:
- Missing required fields
- Invalid TOML syntax
- Invalid file paths
- Invalid storage credentials

### "Can't reach storage"

For S3, verify:
- Credentials are correct
- Bucket exists and is accessible
- Endpoint is correct
- Region is correct

```bash
# Test S3 connection
aws s3 ls s3://your-bucket --region us-east-1
```

## Advanced Configuration

### Custom Configuration

Modify `.kbvault/config.toml` directly:

```bash
# Open in editor
vim .kbvault/config.toml

# Validate after editing
kbvault config validate
```

### Configuration Precedence

Settings are loaded in this order (later overrides earlier):

1. Default values
2. `.kbvault/config.toml`
3. Profile configuration
4. Command-line flags
5. Environment variables

### Backup Configuration

```bash
# Backup config
cp .kbvault/config.toml .kbvault/config.toml.backup

# Restore config
cp .kbvault/config.toml.backup .kbvault/config.toml
```

---

## See Also

- [Getting Started](getting-started.md)
- [Profiles & Multi-Vault](profiles.md)
- [CLI Reference](cli-reference.md)
