# Package Reference

Detailed reference for all public packages in kbVault.

## Overview

kbVault is organized into public packages (`pkg/`) and private packages (`internal/`). This document covers the public API.

## pkg/types

**Core data types used throughout kbVault.**

### Note

Represents a single knowledge base entry.

```go
type Note struct {
    ID       string    // Unique identifier (ULID)
    Title    string    // Note title
    Content  string    // Markdown content
    Tags     []string  // Categorical tags
    Created  time.Time // Creation timestamp
    Modified time.Time // Last modification timestamp
    Links    []Link    // References to other notes
    Path     string    // File path in vault
}
```

**Usage:**
```go
note := &types.Note{
    Title: "My Note",
    Content: "# Markdown content",
    Tags: []string{"python", "tutorial"},
}
```

### Config

Vault configuration.

```go
type Config struct {
    Vault    VaultConfig    // Vault metadata
    Storage  StorageConfig  // Storage settings
    Search   SearchConfig   // Search engine config
    Vector   VectorConfig   // Vector DB config (optional)
    Settings SettingsConfig // General settings
}
```

### StorageConfig

Storage backend configuration.

```go
type StorageConfig struct {
    Type      string // "local" or "s3"
    Path      string // Local path
    Bucket    string // S3 bucket
    Region    string // AWS region
    Endpoint  string // Custom S3 endpoint (optional)
    Credentials struct {
        AccessKey string
        SecretKey string
    }
}
```

### Error Types

Custom errors for proper error handling.

```go
type ConfigError struct {
    Field   string
    Message string
}

type StorageError struct {
    Operation string // "read", "write", "delete"
    Path      string
    Cause     error
}

type NotFoundError struct {
    Type string // "note", "profile", "config"
    ID   string
}

type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}
```

**Usage:**
```go
if err != nil {
    if _, ok := err.(*types.StorageError); ok {
        // Handle storage error
    }
}
```

## pkg/storage

**Pluggable storage backends for note persistence.**

### Backend Interface

All storage backends implement this interface:

```go
type Backend interface {
    // Read note content from path
    Read(ctx context.Context, path string) ([]byte, error)
    
    // Write note content to path
    Write(ctx context.Context, path string, data []byte) error
    
    // Delete note at path
    Delete(ctx context.Context, path string) error
    
    // List paths with given prefix
    List(ctx context.Context, prefix string) ([]string, error)
    
    // Check if path exists
    Exists(ctx context.Context, path string) (bool, error)
}
```

### Factory

Create storage backend based on configuration.

```go
backend, err := storage.CreateBackend(config.Storage)
// Returns appropriate backend implementation
```

### Local Storage (pkg/storage/local)

Store notes in local filesystem.

**Features:**
- Stores notes as TOML files
- Automatic directory creation
- No external dependencies
- Fast local access

**Configuration:**
```toml
[storage]
type = "local"
path = "./notes"
```

### S3 Storage (pkg/storage/s3)

Store notes in S3-compatible object storage.

**Features:**
- AWS S3 support
- MinIO compatibility
- Custom endpoints
- Credential management

**Configuration:**
```toml
[storage]
type = "s3"
bucket = "my-kb"
region = "us-east-1"

[storage.credentials]
access_key = "..."
secret_key = "..."
```

**Environment Variables:**
```bash
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1
```

## pkg/config

**Configuration management with profiles and Viper integration.**

### Manager

Load and manage vault configuration.

```go
manager := config.NewManager()
cfg, err := manager.LoadFromFile("path/to/config.toml")
```

### ProfileManager

Manage multiple vault profiles.

```go
pm, err := config.NewProfileManager()

// List profiles
profiles, err := pm.ListProfiles()

// Get specific profile config
cfg, err := pm.GetConfig("work")

// Set active profile
err := pm.SetActiveProfile("work")

// Get active profile
active := pm.GetActiveProfile()
```

**Profile Structure:**
```go
type Profile struct {
    Name     string
    Config   *Config
    Created  time.Time
    Modified time.Time
}
```

### Environment Variables

Configuration respects environment variables:

```bash
# Set default profile
export KBVAULT_PROFILE=work

# Configuration directory
export KBVAULT_CONFIG=~/.kbvault

# AWS credentials (for S3)
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
```

## pkg/types (continued)

### Link

Represents a reference between notes.

```go
type Link struct {
    ID           string // Unique identifier
    SourceID     string // Source note ID
    TargetID     string // Target note ID
    TargetTitle  string // Human-readable target
    LinkText     string // Text shown in markdown
    Type         string // "wiki", "markdown", etc.
    Position     int    // Character position in content
    Created      time.Time
}
```

### Vector

Vector database configuration (planned).

```go
type VectorConfig struct {
    Enabled  bool
    Engine   string // "none", "pinecone", "local"
    Pinecone PineconeConfig `toml:"pinecone"`
}
```

## pkg/ulid

**Unique identifier generation using ULID format.**

### Generator

Create unique IDs for notes.

```go
gen := ulid.NewGenerator()

// Generate new ID
id, err := gen.Generate()
// Returns: "01ARZ3NDEKTSV4RRFFQ69G5FAV"

// Generate ID from timestamp
id, err := gen.GenerateWithTime(time.Now())

// Validate ID
valid := ulid.IsValid("01ARZ3NDEKTSV4RRFFQ69G5FAV")
```

**Properties:**
- 26 characters
- Sortable (timestamp-based)
- No special characters
- 128-bit entropy

## pkg/retry

**Resilient operation retry logic.**

### Retrier

Retry operations with backoff.

```go
retrier := retry.NewRetrier(
    retry.WithMaxAttempts(3),
    retry.WithBackoff(100 * time.Millisecond),
)

err := retrier.Retry(ctx, func() error {
    return storage.Write(ctx, path, data)
})
```

**Options:**
- `WithMaxAttempts` - Maximum retry attempts
- `WithBackoff` - Backoff duration between retries
- `WithTimeout` - Overall operation timeout
- `WithJitter` - Add randomization to backoff

## pkg/vector

**Vector database integration for semantic search (planned).**

### Factory

Create vector backend based on configuration.

```go
backend, err := vector.CreateBackend(config.Vector)
```

### Supported Engines

- `none` - Disabled (default)
- `local` - Local vector DB (planned)
- `pinecone` - Pinecone cloud (planned)
- `milvus` - Milvus (planned)

## Internal Packages

While internal packages are not part of the public API, key ones are:

### internal/links

Link parsing and graph generation.

```go
// Parse links from note content
links := links.ParseLinksFromContent("[[Note 1]] and [[Note 2]]")

// Build link graph
graph, err := builder.BuildFromNotes(ctx, notes)
```

### internal/search

Full-text search engine.

```go
// Create search engine
engine, err := search.NewEngine(storage, indexPath)

// Index note
err := engine.IndexNote(note)

// Search
results, err := engine.Search(ctx, "query")
```

### internal/templates

Note template system.

```go
// Get template
template, err := templates.Get("default")

// Apply template to note
content := template.Apply(note)
```

## Usage Examples

### Creating and Saving a Note

```go
package main

import (
    "context"
    "github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
    "github.com/madstone-tech/mdstn-kb-mcp/pkg/storage"
    "github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
)

func main() {
    ctx := context.Background()
    
    // Load configuration
    manager := config.NewManager()
    cfg, err := manager.LoadFromFile(".kbvault/config.toml")
    if err != nil {
        panic(err)
    }
    
    // Create storage backend
    backend, err := storage.CreateBackend(cfg.Storage)
    if err != nil {
        panic(err)
    }
    
    // Create note
    note := &types.Note{
        Title: "My Note",
        Content: "# Content",
        Tags: []string{"tag1", "tag2"},
    }
    
    // Serialize to TOML (simplified)
    data := []byte(`
title = "My Note"
content = "# Content"
tags = ["tag1", "tag2"]
`)
    
    // Store note
    err = backend.Write(ctx, "notes/my-note.toml", data)
    if err != nil {
        panic(err)
    }
}
```

### Working with Profiles

```go
package main

import (
    "github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
)

func main() {
    pm, err := config.NewProfileManager()
    if err != nil {
        panic(err)
    }
    
    // List profiles
    profiles, err := pm.ListProfiles()
    for _, p := range profiles {
        println(p.Name)
    }
    
    // Get active profile
    activeName := pm.GetActiveProfile()
    
    // Load active profile config
    cfg, err := pm.GetConfig(activeName)
    if err != nil {
        panic(err)
    }
}
```

### Error Handling

```go
package main

import (
    "github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
    "github.com/madstone-tech/mdstn-kb-mcp/pkg/storage"
)

func readNote(backend storage.Backend) error {
    _, err := backend.Read(ctx, "notes/nonexistent.toml")
    
    if err != nil {
        if _, ok := err.(*types.NotFoundError); ok {
            println("Note not found")
            return nil
        }
        
        if se, ok := err.(*types.StorageError); ok {
            println("Storage error:", se.Message)
            return err
        }
    }
    
    return err
}
```

## Interface Documentation

### Package Organization

```
kbVault Package Structure

pkg/
├── config/           # Configuration & profiles
│   ├── Config         # Main config type
│   ├── ProfileManager # Multi-profile management
│   └── Manager        # Config loading/saving
│
├── types/            # Core types
│   ├── Note          # Note structure
│   ├── Config        # Config types
│   ├── Link          # Link references
│   ├── Vector        # Vector DB config
│   └── Errors        # Error types
│
├── storage/          # Storage abstraction
│   ├── Backend       # Storage interface
│   ├── Factory       # Backend creation
│   ├── local/        # Local filesystem
│   └── s3/           # S3-compatible
│
├── retry/            # Retry logic
│   └── Retrier       # Retry operations
│
├── ulid/             # ID generation
│   ├── Generator     # ULID generator
│   └── Validator     # ID validation
│
└── vector/           # Vector database (planned)
    └── Factory       # Vector backend creation
```

## Backward Compatibility

Public packages maintain backward compatibility:

- No breaking changes in public APIs without major version bump
- Deprecated functions kept for 2+ releases
- Migration guides provided

## Extensions

### Custom Storage Backend

Implement `storage.Backend` interface:

```go
type CustomBackend struct {
    // your fields
}

func (cb *CustomBackend) Read(ctx context.Context, path string) ([]byte, error) {
    // implementation
}

func (cb *CustomBackend) Write(ctx context.Context, path string, data []byte) error {
    // implementation
}

// ... implement other methods
```

---

## See Also

- [Architecture Overview](overview.md)
- [CLI Reference](../guides/cli-reference.md)
- [Contributing Guide](../../CONTRIBUTING.md)
