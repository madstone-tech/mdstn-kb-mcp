# Profiles & Multi-Vault Guide

Manage multiple knowledge vaults with different configurations using profiles.

## What are Profiles?

Profiles allow you to maintain separate knowledge bases with different configurations. Each profile has:

- Independent configuration
- Separate storage (local or S3)
- Own search index
- Distinct note collection

**Common Use Cases:**
- Work and personal vaults
- Different projects
- Research and production
- Client-specific vaults
- Public and private notes

## Getting Started with Profiles

### Create Your First Profile

```bash
# Create a local profile
kbvault profile create personal --storage-path ~/personal-vault

# Create an S3-based profile
kbvault profile create work \
  --storage-type s3 \
  --s3-bucket company-kb \
  --s3-region us-east-1
```

### List Profiles

```bash
kbvault profile list
```

Output:
```
Name         Storage Type    Path/Bucket          Active
personal     local           ~/personal-vault     true
work         s3              company-kb           false
research     local           ~/research           false
```

### Use a Profile

#### Option 1: Use `--profile` Flag

```bash
# Create note in work profile
kbvault --profile work new "Meeting Notes"

# List notes in personal profile
kbvault --profile personal list

# Search in research profile
kbvault --profile research search "python"
```

#### Option 2: Set Active Profile

```bash
# Set work as default
kbvault profile set-active work

# Now use commands without --profile flag
kbvault new "Meeting Notes"
kbvault list
kbvault search "python"

# Check active profile
kbvault profile show

# Switch back to personal
kbvault profile set-active personal
```

## Profile Management

### View Profile Details

```bash
# Show active profile
kbvault profile show

# Show specific profile
kbvault profile show work
kbvault profile show personal
```

Output:
```
Profile: work
Storage Type: s3
Storage Path/Bucket: company-kb
Region: us-east-1
Status: Active
Created: 2025-12-11
Last Used: 2025-12-11 14:30
```

### Create Profiles

#### Local Storage Profile

```bash
kbvault profile create <name> --storage-path <path>
```

Examples:
```bash
# Create personal vault
kbvault profile create personal --storage-path ~/personal-vault

# Create project vault
kbvault profile create myproject --storage-path ~/projects/myproject/kb

# Create with path in specific location
kbvault profile create research --storage-path /mnt/research/vault
```

#### S3 Storage Profile

```bash
kbvault profile create <name> \
  --storage-type s3 \
  --s3-bucket <bucket> \
  --s3-region <region>
```

Examples:
```bash
# Create AWS S3 profile
kbvault profile create work \
  --storage-type s3 \
  --s3-bucket company-kb \
  --s3-region us-east-1

# Create MinIO profile
kbvault profile create minio \
  --storage-type s3 \
  --s3-bucket local-kb \
  --s3-endpoint http://localhost:9000

# Create with custom endpoint
kbvault profile create storage \
  --storage-type s3 \
  --s3-bucket kb \
  --s3-endpoint https://custom-s3.example.com
```

#### Interactive Profile Creation

```bash
# Guided setup
kbvault profile create myprofile
# Will prompt for storage type, path/bucket, region, etc.
```

### Edit Profile

```bash
# Edit specific profile's configuration
kbvault --profile work configure

# Update storage path
kbvault config set storage.path ~/new-path

# Update vault name
kbvault config set vault.name "My Vault"
```

### Delete Profile

```bash
# Delete with confirmation
kbvault profile delete personal

# Delete without confirmation
kbvault profile delete personal --force

# Note: This only removes the profile, not the vault data
```

### Set Active Profile

```bash
# Set which profile to use by default
kbvault profile set-active work

# Verify active profile
kbvault profile show
```

## Working with Multiple Profiles

### Quick Profile Switching

```bash
# Work profile
kbvault --profile work new "Work Task"

# Personal profile
kbvault --profile personal new "Personal Note"

# Research profile
kbvault --profile research search "algorithm"
```

### Check Profile Status

```bash
# List all profiles with status
kbvault profile list

# Show current active profile
kbvault profile show
```

### Copy Configuration Between Profiles

```bash
# Get config from one profile
kbvault --profile work config show > work-config.toml

# Apply to another profile
kbvault --profile personal configure < work-config.toml
```

## Profile Isolation

### Complete Isolation

Each profile maintains:
- Separate note collections
- Independent search indexes
- Distinct configurations
- Own cache

**Example:**
```bash
# Create note in work
kbvault --profile work new "Quarterly Review"

# NOT visible in personal
kbvault --profile personal list  # Won't show "Quarterly Review"

# Switch to work to see it
kbvault --profile work list      # Shows "Quarterly Review"
```

### Shared Storage (Advanced)

While not recommended, you can share storage:

```toml
# work profile config
[storage]
type = "local"
path = "/shared/vault"

# personal profile config (same path)
[storage]
type = "local"
path = "/shared/vault"
```

**Note:** This can cause conflicts. Use separate paths.

## Advanced Profile Usage

### Environment-Based Profiles

Create profiles for different environments:

```bash
# Development
kbvault profile create dev \
  --storage-path ~/dev-vault

# Testing
kbvault profile create test \
  --storage-path ~/test-vault

# Production
kbvault profile create prod \
  --storage-type s3 \
  --s3-bucket prod-kb
```

Use with environment variables:

```bash
# Set active profile based on environment
export KBVAULT_PROFILE=$ENVIRONMENT

kbvault list  # Uses dev, test, or prod based on env
```

### Team Profiles

Set up shared team vaults:

```bash
# Team research vault
kbvault profile create team-research \
  --storage-type s3 \
  --s3-bucket team-research-kb

# Team documentation
kbvault profile create team-docs \
  --storage-type s3 \
  --s3-bucket team-docs

# Team standards
kbvault profile create team-standards \
  --storage-type s3 \
  --s3-bucket team-standards
```

### Project-Specific Profiles

Organize by project:

```bash
kbvault profile create project-alpha \
  --storage-path ~/projects/alpha/kb

kbvault profile create project-beta \
  --storage-path ~/projects/beta/kb

kbvault profile create project-gamma \
  --storage-path ~/projects/gamma/kb
```

## Profile Configuration

### View Profile Configuration

```bash
# Show configuration for current profile
kbvault config show

# Show configuration for specific profile
kbvault --profile work config show
```

### Customize Profile Settings

```bash
# Set vault name
kbvault --profile work config set vault.name "Work Knowledge Base"

# Set editor
kbvault --profile work config set settings.editor "code"

# Set auto-index
kbvault --profile work config set settings.auto_index "true"
```

### Profile Configuration Files

Profiles store configuration in:

```
~/.config/kbvault/profiles/
├── work/
│   ├── config.toml
│   └── .kbvault/
├── personal/
│   ├── config.toml
│   └── .kbvault/
└── research/
    ├── config.toml
    └── .kbvault/
```

## Scripting with Profiles

### Batch Operations

```bash
#!/bin/bash

# Back up all profiles
for profile in $(kbvault profile list --format json | jq -r '.[].name'); do
  echo "Backing up $profile..."
  kbvault --profile "$profile" list --format json > "${profile}_backup.json"
done
```

### Profile Status Check

```bash
#!/bin/bash

# Check all profiles are accessible
for profile in work personal research; do
  if kbvault --profile "$profile" list &>/dev/null; then
    echo "✓ $profile is accessible"
  else
    echo "✗ $profile is not accessible"
  fi
done
```

### Profile-Based Workflow

```bash
#!/bin/bash

PROFILE=$1

if [ -z "$PROFILE" ]; then
  echo "Usage: $0 <profile>"
  exit 1
fi

# Show profile status
echo "Profile: $PROFILE"
kbvault --profile "$PROFILE" profile show

# List notes
echo ""
echo "Notes:"
kbvault --profile "$PROFILE" list | head -20
```

## Troubleshooting

### Profile Not Found

```bash
# List available profiles
kbvault profile list

# Create profile if missing
kbvault profile create myprofile --storage-path ~/my-vault
```

### Storage Not Accessible

```bash
# Check storage configuration
kbvault --profile work config show

# Validate storage
kbvault --profile work config validate

# For S3, check credentials
aws s3 ls s3://bucket-name --region us-east-1
```

### Wrong Profile Active

```bash
# Check active profile
kbvault profile show

# Set correct profile
kbvault profile set-active work

# Or use --profile flag
kbvault --profile work list
```

### Lost Notes in Profile

Each profile has its own note collection. If notes seem missing:

1. Check you're using the correct profile
2. Verify storage configuration
3. Check storage path/bucket exists
4. List notes with correct profile: `kbvault --profile correct list`

## Migration

### Move Notes Between Profiles

```bash
# Export from source profile
kbvault --profile old list --format json > export.json

# Import to target profile
# (import feature coming in future version)
```

### Merge Profiles

```bash
# List notes from source
kbvault --profile source list --format json | jq -r '.[].id'

# Copy to target (manual process or with custom script)
```

## Best Practices

1. **Use Meaningful Names**: `work`, `personal`, `research` rather than `profile1`
2. **Separate Concerns**: Different profiles for different purposes
3. **Consistent Paths**: Use consistent path structure for local profiles
4. **Regular Backups**: Back up important profiles regularly
5. **Document Purpose**: Add description to vault config
6. **Set Active Profile**: Choose a default profile to reduce flag usage

## Examples

### Personal Knowledge System

```bash
# Create profiles
kbvault profile create personal --storage-path ~/personal-kb
kbvault profile create work --storage-path ~/work-kb
kbvault profile create learning --storage-path ~/learning-kb

# Set personal as default
kbvault profile set-active personal

# Use throughout day
kbvault new "Morning Thoughts"
kbvault --profile work new "Team Update"
kbvault --profile learning new "JavaScript Concept"
```

### Team Knowledge Base

```bash
# Create team profiles
kbvault profile create team-shared \
  --storage-type s3 \
  --s3-bucket team-kb-prod

kbvault profile create team-staging \
  --storage-type s3 \
  --s3-bucket team-kb-staging

# Use in CI/CD
export KBVAULT_PROFILE=team-shared
kbvault list
```

---

## See Also

- [Getting Started](getting-started.md)
- [Configuration Guide](configuration.md)
- [CLI Reference](cli-reference.md#profile)
