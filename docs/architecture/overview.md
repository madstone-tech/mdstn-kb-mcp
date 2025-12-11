# Architecture Overview

Understanding the design and structure of kbVault.

## Design Philosophy

kbVault is built on these core principles:

1. **Modularity** - Separated concerns with clear interfaces
2. **Storage Agnostic** - Support multiple storage backends
3. **Performance** - Fast search and caching for large vaults
4. **Extensibility** - Easy to add new features and backends
5. **User-Friendly** - Intuitive CLI and configuration

## High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│                 User Interfaces                      │
│  CLI  │  TUI  │  HTTP API  │  MCP  │  gRPC         │
└─────────┬──────────────────────────────────────────┘
          │
┌─────────▼──────────────────────────────────────────┐
│           Command Layer (cmd/kbvault)              │
│  Parses CLI commands, handles I/O, user interactions│
└─────────┬──────────────────────────────────────────┘
          │
┌─────────▼──────────────────────────────────────────┐
│         Core Services (internal/)                  │
│  • Note Management  • Search  • Links  • Templates │
└─────────┬──────────────────────────────────────────┘
          │
┌─────────▼──────────────────────────────────────────┐
│        Storage Abstraction (pkg/storage)           │
│  Local  │  S3  │  Future backends                  │
└─────────┬──────────────────────────────────────────┘
          │
┌─────────▼──────────────────────────────────────────┐
│    Filesystem / Cloud Storage / Databases          │
│         (Where notes are actually stored)           │
└──────────────────────────────────────────────────────┘
```

## Core Components

### 1. Command Layer (cmd/kbvault/)

The entry point for all CLI interactions.

**Responsibilities:**
- Parse command-line arguments
- Handle user input/output
- Call appropriate services
- Format and display results
- Manage global flags (profile, config, etc.)

**Main Commands:**
- `new`, `show`, `list`, `edit`, `delete` - Note management
- `search` - Search functionality
- `config`, `configure` - Configuration
- `profile` - Profile management
- `init` - Vault initialization
- `completion` - Shell completion

**Architecture:**
```
cmd/kbvault/
├── main.go           # Entry point
├── init.go           # kbvault init
├── new.go            # kbvault new
├── show.go           # kbvault show
├── list.go           # kbvault list
├── edit.go           # kbvault edit
├── delete.go         # kbvault delete
├── search.go         # kbvault search
├── config.go         # kbvault config
├── configure.go      # kbvault configure
├── profile.go        # kbvault profile
└── completion.go     # kbvault completion
```

### 2. Core Services (internal/)

Business logic and feature implementation.

#### internal/links
**Note Link Management**

Handles wiki-style links in notes (`[[Note Title]]`).

**Key Functions:**
- Parse links from markdown content
- Detect bidirectional relationships
- Validate link targets
- Generate link graphs

**Example:**
```
Note: "Architecture"
Links: [[Overview]] -> [[Components]]
Backlinks: [[Introduction]] -> [[Architecture]]
```

#### internal/search
**Full-Text Search Engine**

Index and search notes efficiently.

**Key Components:**
- `engine.go` - Search execution
- `index.go` - Index management

**Features:**
- Inverted index for fast search
- Field-specific search (title, content, tags)
- Boolean operators (future)
- Ranking and relevance

#### internal/templates
**Note Templates**

Predefined templates for consistent note creation.

**Template System:**
- Front matter handling
- Dynamic field substitution
- Type-specific templates (meeting, research, etc.)

#### internal/api, internal/mcp, internal/tui
**Planned Interfaces**

- `api/` - HTTP API and gRPC (future)
- `mcp/` - Model Context Protocol
- `tui/` - Terminal UI (planned)

### 3. Storage Abstraction (pkg/storage)

Pluggable storage backends.

```
pkg/storage/
├── factory.go        # Backend selection
├── interface.go      # Storage interface
├── local/
│   └── storage.go   # Local filesystem
└── s3/
    └── storage.go   # S3-compatible storage
```

**Storage Interface:**
```go
type Backend interface {
    Read(ctx context.Context, path string) ([]byte, error)
    Write(ctx context.Context, path string, data []byte) error
    Delete(ctx context.Context, path string) error
    List(ctx context.Context, prefix string) ([]string, error)
    Exists(ctx context.Context, path string) (bool, error)
}
```

**Implementations:**
- **Local**: Notes stored as TOML files in filesystem
- **S3**: Notes stored in S3-compatible object storage

Adding a new backend:
1. Implement the `Backend` interface
2. Add factory method
3. Configure in `config.toml`

### 4. Configuration & Profiles (pkg/config)

Multi-profile configuration system.

```
pkg/config/
├── config.go         # Configuration types
├── profiles.go       # Profile management
├── viper.go         # TOML parsing
└── factory.go       # Configuration creation
```

**Features:**
- TOML-based configuration
- Profile isolation
- Environment variable support
- Viper for config management

### 5. Core Types (pkg/types)

Shared data types.

```
pkg/types/
├── note.go          # Note structure
├── config.go        # Configuration types
├── vector.go        # Vector database config
├── storage.go       # Storage types
└── errors.go        # Error types
```

**Key Types:**
- `Note` - Knowledge base entry with ID, title, content, metadata
- `Config` - Vault configuration
- `Link` - Inter-note connection

### 6. Utilities (pkg/)

Supporting packages.

- **pkg/ulid** - Unique ID generation (ULID format)
- **pkg/retry** - Retry logic for resilience
- **pkg/vector** - Vector database integration (future)

## Data Flow

### Creating a Note

```
User Input
    ↓
cmd/kbvault/new → Get title, open editor
    ↓
internal/templates → Apply template if specified
    ↓
pkg/types → Create Note object
    ↓
pkg/storage → Write to filesystem/S3
    ↓
internal/search → Index the note
    ↓
Display confirmation
```

### Searching Notes

```
User Query
    ↓
cmd/kbvault/search → Parse query
    ↓
internal/search → Query index
    ↓
pkg/storage → Fetch note details if needed
    ↓
Format Results
    ↓
Display to user
```

### Note Editing

```
User selects note
    ↓
cmd/kbvault/edit → Open in editor
    ↓
pkg/storage → Read current content
    ↓
User modifies → Editor saves changes
    ↓
pkg/storage → Write updated content
    ↓
internal/search → Re-index note
    ↓
internal/links → Update link graph
    ↓
Confirm changes
```

## Configuration System

```
Configuration Sources (precedence)
         ↓
┌─────────────────────┐
│ Default Values      │ (lowest precedence)
│ Profile Config      │
│ Global Config       │
│ CLI Flags           │
│ Environment Vars    │ (highest precedence)
└─────────────────────┘
         ↓
Merged Configuration
         ↓
Used for all operations
```

## Storage System

### Local Storage

Notes stored as TOML files:

```
vault/
├── .kbvault/
│   ├── config.toml
│   └── index/
├── notes/
│   ├── 01ARZ3NDEKTSV4RRFFQ69G5FAV.toml
│   ├── 01ARZ3NDEKTSV4RRFFQ69G5FBW.toml
│   └── docs/
│       └── 01ARZ3NDEKTSV4RRFFQ69G5FCX.toml
```

### S3 Storage

Same structure but in S3 bucket:

```
s3://bucket/
├── .kbvault/
│   ├── config.toml
│   └── index/
├── notes/
│   ├── 01ARZ3NDEKTSV4RRFFQ69G5FAV.toml
│   └── ...
```

## Caching Strategy

Future caching layer will provide:

1. **Memory Cache** - Fast access to recently used notes
2. **Index Cache** - Cached search indexes
3. **Configuration Cache** - Cached config values
4. **TTL Management** - Automatic expiration
5. **Invalidation** - Smart cache updates on changes

## Search Index Structure

```
Index File (.kbvault/index/)
    ↓
┌──────────────────────────────┐
│ Inverted Index               │
├──────────────────────────────┤
│ Word → [Note IDs]            │
│ "python" → [id1, id2, id5]   │
│ "tutorial" → [id2, id3, id7] │
│ "async" → [id1, id4]         │
└──────────────────────────────┘
    ↓
Field-specific Indexes
├─ Title Index
├─ Content Index
└─ Tag Index
```

## Error Handling

kbVault uses typed errors for clear error handling:

```go
// Custom error types
type ConfigError struct { ... }
type StorageError struct { ... }
type NotFoundError struct { ... }
type ValidationError struct { ... }

// Usage
if err != nil {
    if _, ok := err.(*StorageError); ok {
        // Handle storage issues
    }
}
```

## Concurrency

Current approach (Session 1-5):
- No concurrent operations
- Single-threaded CLI execution

Future improvements (Session 6+):
- Concurrent search indexing
- Parallel file operations
- Thread-safe caching

## Dependencies

**Key External Libraries:**
- `cobra` - CLI framework
- `viper` - Configuration management
- `oklog/ulid` - ID generation
- `aws/aws-sdk-go-v2` - S3 support
- `BurntSushi/toml` - TOML parsing
- `testify` - Testing framework

**No ORM/Database Framework:**
- Direct TOML file storage
- No complex database migrations
- Simple, portable storage format

## Extensibility Points

### Adding a New Backend

1. Implement `pkg/storage/Backend` interface
2. Add factory method in `pkg/storage/factory.go`
3. Add configuration option
4. Implement tests

### Adding a New Interface

1. Create package (e.g., `internal/http/`)
2. Use existing core services
3. Implement error handling
4. Add configuration

### Adding Search Features

1. Enhance `internal/search/engine.go`
2. Update index structure if needed
3. Update CLI in `cmd/kbvault/search.go`
4. Test with various queries

## Performance Considerations

### Current (Session 1-5)

- Notes stored as individual TOML files
- Search uses inverted index
- No external database required
- Suitable for vaults with 1000s of notes

### Future Optimizations

- Caching layer for frequently accessed notes
- Batch indexing for better performance
- Async operations for large vaults
- Database support for very large vaults

## Testing Architecture

```
Tests are located alongside code:
  cmd/kbvault/*_test.go
  internal/*_test.go
  pkg/*_test.go

Using testify framework:
  require.NoError(t, err)
  assert.Equal(t, expected, actual)

Test Coverage: 62.8%+ minimum
```

---

## See Also

- [Package Reference](packages.md)
- [Building & Testing](../development/building.md)
- [CLI Reference](../guides/cli-reference.md)
