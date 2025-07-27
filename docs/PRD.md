# kbVault - Product Requirements Document (PRD) v3.0

## Product Overview

### Vision
Build a high-performance Go tool for managing markdown-based knowledge vaults with multiple interaction interfaces (CLI, TUI, API, MCP), designed for both human users and LLM agents. Support local and remote storage backends with flexible deployment options.

### Application Name: **kbVault** (Knowledge Base Vault)
- **Rationale**: Clear purpose (knowledge base), professional naming
- **Binary Name**: `kbvault`
- **Pronunciation**: "KB Vault" or "Knowledge Vault"

### Success Metrics
- Handle 2,000 notes with sub-second response times
- Support up to 10,000 notes with acceptable performance degradation
- Zero external dependencies beyond Go standard library
- Cross-platform compatibility (macOS, Linux, Windows)
- Seamless operation across multiple interaction modes

## Architecture Overview

### Multi-Interface Design
```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   CLI Interface │  │  TUI Interface  │  │  Web Interface  │  │  MCP Interface  │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ Setup & Mgmt    │  │ Setup & Mgmt    │  │ Setup & Mgmt    │  │ Regular Ops     │
│ Regular Ops     │  │ Regular Ops     │  │ Regular Ops     │  │ LLM Integration │
└─────────────────┘  └─────────────────┘  └─────────────────┘  └─────────────────┘
         │                    │                    │                    │
         └────────────────────┼────────────────────┼────────────────────┘
                              │                    │
         ┌─────────────────────────────────────────┼─────────────────────┐
         │                Core Engine              │                     │
         ├─────────────────────────────────────────┼─────────────────────┤
         │ • Note Management    • Search Engine    │ • gRPC Server       │
         │ • Template System    • Link Management  │ • HTTP/REST API     │
         │ • Storage Backends   • Plugin System    │ • WebSocket Support │
         └─────────────────────────────────────────┼─────────────────────┘
                              │                    │
         ┌─────────────────────────────────────────┼─────────────────────┐
         │              Storage Layer              │                     │
         ├─────────────────────────────────────────┼─────────────────────┤
         │ • Local Filesystem  • S3/MinIO         │                     │
         │ • Caching Layer     • Plugin Backends  │                     │
         └─────────────────────────────────────────┘
```

## Interface Specifications

### 1. CLI Interface (Primary)
**Purpose**: Setup, management, and regular operations
**Target Users**: Developers, power users, automation scripts

```bash
# Setup & Configuration
kbvault init                          # Initialize new vault
kbvault config show                   # Display current configuration
kbvault config edit                   # Open config in editor
kbvault storage setup                 # Configure storage backend

# Regular Operations
kbvault new "Note Title"              # Create new note
kbvault search "search term"          # Search notes
kbvault list                          # List all notes
kbvault show <note-id>                # Display note content
kbvault daily                         # Create/open daily note
kbvault sync                          # Sync with remote storage

# Management
kbvault validate                      # Check vault integrity
kbvault migrate --from local --to s3 # Migrate storage
kbvault server start                  # Start API server
kbvault server stop                   # Stop API server
```

### 2. TUI Interface (Interactive)
**Purpose**: User-friendly setup and management
**Target Users**: End users, initial setup, visual management
**Framework**: Bubble Tea (charm.sh/bubbletea)

```
┌─────────────────── kbVault TUI ─────────────────────┐
│                                                     │
│  [Setup] [Search] [Browse] [Settings] [Help]       │
│                                                     │
│  ┌─── Recent Notes ────────────────────────────┐    │
│  │ • 1737564123ABCD-recipe.md                  │    │
│  │ • 1737563890WXYZ-meeting-notes.md           │    │
│  │ • 1737563500DEFG-project-ideas.md           │    │
│  └─────────────────────────────────────────────┘    │
│                                                     │
│  ┌─── Quick Actions ───────────────────────────┐    │
│  │ [N] New Note    [S] Search    [D] Daily     │    │
│  │ [L] List All    [V] Validate  [Q] Quit      │    │
│  └─────────────────────────────────────────────┘    │
│                                                     │
│  Status: Connected to S3 | 1,247 notes | All OK   │
└─────────────────────────────────────────────────────┘
```

### 3. API Interface (HTTP/gRPC)
**Purpose**: Programmatic access, web integrations, microservices
**Protocols**: HTTP REST + gRPC (optional)
**Target Users**: Web applications, integrations, distributed systems

#### HTTP REST API
```yaml
# Endpoints
GET    /api/v1/notes                    # List notes
POST   /api/v1/notes                    # Create note
GET    /api/v1/notes/{id}               # Get note
PUT    /api/v1/notes/{id}               # Update note
DELETE /api/v1/notes/{id}               # Delete note
GET    /api/v1/search?q=term            # Search notes
GET    /api/v1/links/{id}/backlinks     # Get backlinks
POST   /api/v1/sync                     # Trigger sync
GET    /api/v1/health                   # Health check
GET    /api/v1/config                   # Get configuration
PUT    /api/v1/config                   # Update configuration
```

#### gRPC Service (Optional)
```protobuf
service KbVaultService {
  rpc CreateNote(CreateNoteRequest) returns (NoteResponse);
  rpc GetNote(GetNoteRequest) returns (NoteResponse);
  rpc UpdateNote(UpdateNoteRequest) returns (NoteResponse);
  rpc DeleteNote(DeleteNoteRequest) returns (DeleteNoteResponse);
  rpc SearchNotes(SearchRequest) returns (SearchResponse);
  rpc ListNotes(ListNotesRequest) returns (ListNotesResponse);
  rpc GetBacklinks(BacklinksRequest) returns (BacklinksResponse);
  rpc SyncVault(SyncRequest) returns (SyncResponse);
  rpc GetHealth(HealthRequest) returns (HealthResponse);
}
```

### 4. MCP Interface (Model Context Protocol)
**Purpose**: LLM agent operations, automated workflows
**Target Users**: LLM agents, AI assistants, automated systems

```typescript
// MCP Tools
{
  "search_notes": {
    "description": "Search vault notes by content or metadata",
    "parameters": { "query": "string", "limit": "number" }
  },
  "create_note": {
    "description": "Create a new note with generated ID",
    "parameters": { "title": "string", "content": "string", "template": "string" }
  },
  "get_note": {
    "description": "Retrieve note content by ID",
    "parameters": { "note_id": "string" }
  },
  "update_note": {
    "description": "Update existing note content",
    "parameters": { "note_id": "string", "content": "string" }
  },
  "list_notes": {
    "description": "List all notes with metadata",
    "parameters": { "limit": "number", "offset": "number" }
  },
  "get_backlinks": {
    "description": "Get notes linking to specified note",
    "parameters": { "note_id": "string" }
  }
}
```

## Core Features

### 1. Storage Backend Architecture

#### Local Storage
- Direct filesystem access
- File watching for real-time updates
- Atomic writes with backup

#### Remote Storage (S3/S3-Compatible)
- AWS S3, MinIO, DigitalOcean Spaces support
- Connection pooling and retry logic
- Local caching layer for performance

#### Storage Interface
```go
type StorageBackend interface {
    Read(path string) ([]byte, error)
    Write(path string, data []byte) error
    List(prefix string) ([]string, error)
    Delete(path string) error
    Exists(path string) (bool, error)
}
```

### 2. Note Management System

#### Note ID Generation
- **Format**: ULID (Universally Unique Lexicographically Sortable Identifier)
- **Length**: 26 characters, timestamp-based, lexicographically sortable
- **ASCII Only**: Portable across file systems (S3, Windows, Unix)
- **Example**: `01ARZ3NDEKTSV4RRFFQ69G5FAV.md`
- **Export Feature**: Mass rename files with human-readable names during export
- **Collision Handling**: ULID design prevents collisions

#### CRUD Operations
- **Create**: Generate unique ID, apply templates, create file
- **Read**: File content retrieval with optional frontmatter parsing
- **Update**: Content modification with timestamp tracking
- **Query**: Content and metadata search capabilities
- **Relate**: Link management using `[[link]]` syntax

### 3. Search & Discovery Engine

#### Search Capabilities
- **Content Search**: Full-text search across note bodies
- **Frontmatter Search**: Efficient metadata-based queries
- **Hybrid Search**: Combine content and metadata filters
- **Performance Tiers**:
  - Primary: Frontmatter-based search (high performance)
  - Secondary: Content search when metadata insufficient
- **Remote Storage Optimization**: Local index caching

#### Search Results Format
```json
{
  "query": "search term",
  "results": [
    {
      "file_path": "notes/1737564123ABCD-recipe.md",
      "title": "Note Title",
      "excerpt": "...context around match...",
      "score": 0.95,
      "match_type": "content|frontmatter|both",
      "storage_backend": "local|s3"
    }
  ],
  "summary": "Found 5 matches across 3 notes"
}
```

### 4. Template System

#### Core Variables
- `{{date}}` - Current date (YYYY-MM-DD)
- `{{time}}` - Current time (HH:MM:SS)
- `{{title}}` - Note title
- `{{id}}` - Generated note ID
- `{{timestamp}}` - Unix timestamp

#### Template Structure
```yaml
---
id: {{id}}
title: {{title}}
created: {{date}} {{time}}
tags: []
type: note
storage: {{storage_backend}}
---

# {{title}}

Content starts here...
```

### 5. Link Management

#### Link Types
- **Internal Links**: `[[note-title]]` or `[[note-id]]`
- **Link Resolution**: Support both title and ID-based linking
- **Backlink Tracking**: Maintain bidirectional link relationships
- **Cross-Storage Links**: Support links between local and remote notes

#### Link Operations
- **Real-time Validation**: Background process for link integrity
- **Batch Processing**: Periodic link validation and updates
- **Broken Link Detection**: Identify and report orphaned links

### 6. Plugin Architecture

#### Git Integration (Core Plugin)
- Auto-commit on note creation/modification
- Branch-based workflow support
- Conflict detection and resolution
- Remote storage sync integration

#### Community Plugin Framework
- Go plugin system using `plugin` package
- Standardized plugin interface
- Plugin discovery and management

## gRPC Integration Strategy

### gRPC Use Cases & Value Proposition

#### 1. **High-Performance Bulk Operations**
- Stream large imports/exports
- Batch processing multiple notes
- **Use Case**: Importing 1000+ notes from another system

#### 2. **Real-Time Collaboration**
- Real-time note watching
- Conflict resolution
- **Use Case**: Multiple users or agents working on same vault

#### 3. **Agent-to-Agent Communication**
- High-frequency operations between AI agents
- Distributed knowledge graph operations
- **Use Case**: Multiple AI agents collaborating on knowledge base

#### 4. **Microservices Architecture**
- Service-to-service communication
- Health checks and metrics
- **Use Case**: kbVault as part of larger microservices ecosystem

### gRPC Implementation: Optional & Modular

#### Configuration-Driven Approach
```toml
[server]
# HTTP always enabled (minimal overhead)
http_enabled = true
http_port = 8080

# gRPC optional (enable when needed)
grpc_enabled = false  # Default: disabled
grpc_port = 9090

# Enable specific gRPC services
[server.grpc]
enable_bulk_operations = false
enable_collaboration = false
enable_agent_service = false
enable_vault_service = true  # Always enabled if gRPC is on
```

#### When gRPC Makes Sense

**✅ Enable gRPC when:**
- Handling >1000 operations per minute
- Real-time collaboration requirements
- Multiple AI agents working together
- Microservices deployment
- Cross-language integrations
- Streaming large datasets
- Sub-millisecond latency requirements

**❌ Don't enable gRPC when:**
- Single user, local usage
- Simple CLI operations
- Web browser integrations
- Quick scripts and automation
- Learning/experimentation phase

## Caching Strategy

### Multi-Level Caching Architecture

```
┌─────────────────┐
│   Application   │
└─────────┬───────┘
          │
┌─────────▼───────┐  ← Level 1: In-Memory Cache (LRU)
│   Memory Cache  │    • Hot data: recently accessed notes
│   (LRU 100MB)   │    • TTL: 5-15 minutes
└─────────┬───────┘    • Use: Repeated access patterns
          │
┌─────────▼───────┐  ← Level 2: Disk Cache (Local files)
│   Disk Cache    │    • Warm data: frequently accessed
│   (1GB limit)   │    • TTL: 1-24 hours
└─────────┬───────┘    • Use: Offline access, large files
          │
┌─────────▼───────┐  ← Level 3: Remote Storage (S3/MinIO)
│ Remote Storage  │    • Cold data: original source
│ (Unlimited)     │    • TTL: Forever
└─────────────────┘    • Use: Backup, sharing, persistence
```

### Caching Configuration
```toml
[cache]
enabled = false  # Default: disabled for local storage
auto_enable_for_remote = true  # Auto-enable for S3/MinIO

[cache.memory]
enabled = true
max_size_mb = 100
max_items = 1000
ttl_minutes = 15

[cache.disk]
enabled = true     # Only if remote storage
path = "/tmp/kbvault-cache"
max_size_mb = 1000
ttl_hours = 24
cleanup_interval_hours = 6
```

### When to Enable Caching

**✅ Enable Caching:**
1. **Remote Storage**: Always for S3/MinIO
2. **Large Vaults**: >1,000 notes benefit from index caching
3. **Repeated Operations**: Search, link analysis, bulk operations
4. **Slow Storage**: Network storage, encrypted filesystems
5. **Multi-User**: Concurrent access patterns

**❌ Skip Caching:**
1. **Small Local Vaults**: <100 notes on fast local storage
2. **Write-Heavy Workloads**: Constantly changing data
3. **Memory Constrained**: Limited RAM environments
4. **Simple Operations**: Basic CRUD on small files
5. **Development/Testing**: Rapid iteration needs

## Technical Specifications

### Performance Requirements
- **2,000 notes**: Sub-second response times
- **10,000 notes**: <5 second response times for complex queries
- **API Throughput**: 1000+ requests/second (HTTP)
- **gRPC Performance**: 10,000+ requests/second
- **Memory Usage**: <100MB for 2,000 notes
- **Remote Storage**: Local caching to minimize API calls

### Configuration Management (TOML)

#### Simple Configuration
```toml
[vault]
name = "my-kb"
notes_dir = "notes"
daily_dir = "notes/dailies"
templates_dir = "templates"

[storage]
type = "local"  # local, s3
path = "/path/to/vault"

[storage.s3]
bucket = "my-vault"
region = "us-west-2"
endpoint = ""  # for MinIO
access_key_id = ""
secret_access_key = ""

[storage.cache]
enabled = true
local_path = "/tmp/kbvault-cache"
max_size_mb = 100
ttl_minutes = 60

[server]
http_enabled = true
http_port = 8080
grpc_enabled = false
grpc_port = 9090
host = "localhost"
enable_cors = true
enable_auth = false

[server.auth]
type = "none"  # jwt, apikey, none
jwt_secret = ""
api_keys = []

[logging]
level = "WARN"  # INFO, WARN, ERROR, DEBUG
output = "stdout"  # stdout, file, remote
file_path = ""

[tui]
theme = "default"  # default, dark, light
vim_mode = false
show_help = true

[mcp]
enabled = true
socket_path = "/tmp/kbvault.sock"
```

#### Environment Variable Support
```bash
# Override config values
KBVAULT_STORAGE_ACCESS_KEY_ID=your_key
KBVAULT_STORAGE_SECRET_ACCESS_KEY=your_secret
KBVAULT_STORAGE_ENDPOINT=https://minio.example.com
KBVAULT_LOG_LEVEL=INFO
```

### Logging System
- **Levels**: INFO, WARN, ERROR, DEBUG
- **Default**: WARN level
- **Outputs**: stdout (default), file, remote log server
- **Format**: Structured logging with contextual information
- **Remote Storage Context**: Include storage backend info in logs

## Command Interface

### CLI Commands
```bash
# Application Management
kbvault init [path]                   # Initialize vault
kbvault server start                  # Start API server
kbvault server stop                   # Stop API server
kbvault server status                 # Check server status
kbvault tui                          # Launch TUI interface

# Configuration
kbvault config init                   # Create default config
kbvault config show                   # Display config
kbvault config edit                   # Edit config
kbvault config validate              # Validate config

# Note Operations
kbvault new "Title"                   # Create note
kbvault show <id>                     # Show note
kbvault edit <id>                     # Edit note
kbvault delete <id>                   # Delete note
kbvault search "term"                 # Search notes
kbvault list                          # List notes

# Daily Notes
kbvault daily                         # Create/open today's daily note
kbvault daily --date 2025-01-20      # Create/open specific date

# Link Management
kbvault link <from> <to>              # Create link between notes
kbvault backlinks <note-id>           # Show notes linking to this note
kbvault orphans                       # List notes with no links

# Storage Management
kbvault sync                          # Sync with remote storage
kbvault cache clear                   # Clear local cache
kbvault cache status                  # Show cache statistics

# Vault Management
kbvault validate                      # Check vault integrity
kbvault backup <path>                 # Create backup
kbvault restore <path>                # Restore from backup
kbvault migrate --from local --to s3 # Migrate storage

# Development
kbvault version                       # Show version
kbvault health                        # Health check
kbvault debug                         # Debug information
```

### TUI Key Bindings
```
Global:
  q/Ctrl+C    - Quit
  h/?         - Help
  Tab         - Next panel
  Shift+Tab   - Previous panel

Navigation:
  j/↓         - Down
  k/↑         - Up
  g/Home      - Top
  G/End       - Bottom
  
Actions:
  n           - New note
  s           - Search
  d           - Daily note
  r           - Refresh
  Enter       - Select/Open
```

### MCP Integration
```bash
# JSON output for programmatic access
kbvault search "term" --format json
kbvault list --format json --storage all
kbvault show <id> --format json
```

## Implementation Phases

### Phase 1: Core Foundation & CLI
- Note ID generation and CRUD operations
- TOML configuration system
- Local storage backend
- Basic CLI interface
- Template system

### Phase 2: Remote Storage & API
- S3 client integration and caching
- HTTP REST API server
- Basic authentication
- Storage migration utilities

### Phase 3: gRPC & Advanced Features
- gRPC server implementation (optional)
- TUI interface with Bubble Tea
- Advanced search and indexing
- Link management system

### Phase 4: MCP Integration
- MCP protocol implementation
- Unix socket communication
- LLM agent optimization
- Plugin architecture foundation

### Phase 5: Polish & Ecosystem
- Performance optimization
- Comprehensive documentation
- Community plugin support
- Integration examples

## Success Criteria

### Must Have
- All four interfaces (CLI, TUI, API, MCP) functional
- TOML configuration with validation
- Local and S3 storage backends
- Sub-second performance for 2,000 notes
- Cross-platform compatibility
- Smart caching system

### Should Have
- Optional gRPC server for high-performance scenarios
- Authentication and authorization
- Real-time updates via WebSocket
- Comprehensive TUI with all features
- Plugin architecture foundation
- Storage migration tools
- Backup and restore functionality

### Could Have
- Advanced search algorithms
- Multi-vault support
- Web UI frontend
- Metrics and monitoring
- Distributed deployment support
- Advanced plugin ecosystem

## Risk Mitigation

### Performance Risks
- **Mitigation**: Implement tiered search + smart caching
- **Fallback**: Graceful degradation for large vaults

### Remote Storage Risks
- **Mitigation**: Robust caching, offline mode, retry logic
- **Monitoring**: Connection health checks, cache hit rates

### Platform Compatibility
- **Mitigation**: Use Go standard library exclusively
- **Testing**: Automated cross-platform testing

### Configuration Complexity
- **Mitigation**: Sensible defaults, clear documentation
- **Validation**: Config validation with helpful error messages

### gRPC Complexity
- **Mitigation**: Make gRPC optional, HTTP-first approach
- **Fallback**: All functionality available via HTTP REST

---

## Security Architecture

### Input Validation Strategy

#### Note IDs
- **Format**: ULID only (26 characters, alphanumeric)
- **Validation**: Strict ULID format validation on all inputs
- **ASCII Only**: No Unicode, portable across all file systems
- **Max Length**: 32 characters total (26 ULID + .md extension)

#### File Path Security
- **Filename Sanitization**: Strip special characters, use ULID as filename
- **Directory Traversal**: Prevent `../` attacks, sandbox to vault directory
- **Flat Structure**: No subdirectories, use tags for organization
- **Export Safety**: Human-readable filenames only during export operations

#### Content Security
- **Markdown Only**: GitHub Flavored Markdown with sanitization
- **HTML Stripping**: Remove dangerous HTML tags and scripts
- **Size Limits**: 10MB per note to prevent performance issues
- **Encoding**: UTF-8 with BOM detection and normalization

#### Authentication Model
- **Local Deployment**: No authentication required (trusted environment)
- **Remote Deployment**: API key-based authentication
- **Future**: JWT token support via external service/API gateway
- **Machine-to-Machine**: API keys for MCP and automated systems

### Trust Boundaries
- **Local**: Full trust, no authentication needed
- **Remote**: API gateway handles auth, throttling, and security
- **API Keys**: For machine-to-machine communication
- **Rate Limiting**: Delegated to API gateway in production

---

## Concurrency and Data Safety

### Multi-Process Coordination

#### Process Priority System
1. **Human-Led Interfaces** (highest priority): TUI, CLI with human interaction
2. **Headless Operations** (medium priority): CLI scripts, automation
3. **API/Machine Access** (lowest priority): HTTP API, MCP requests

#### File Locking Strategy
- **Local Storage**: OS-level advisory file locks per note
- **Lock Timeout**: 10 seconds maximum wait time
- **Lock Granularity**: Per-note file locking (not vault-wide)
- **Deadlock Prevention**: Sorted lock acquisition, timeout-based release

#### S3 Conflict Resolution
- **Optimistic Locking**: Use ETags for conditional operations
- **Last-Write-Wins**: Simpler conflict resolution, minimal data loss risk
- **S3 Versioning**: Enable bucket versioning for recovery
- **Conflict Detection**: Log when overwrite conflicts occur

### Cache Consistency
- **Cache TTL**: 5 minutes for all cached content
- **Local Storage**: Time-based expiration, no file watchers
- **S3 Storage**: Periodic polling, assume single-writer for most operations
- **Cache Invalidation**: Manual invalidation via API/CLI commands

---

## Error Handling and Resilience

### S3 Operations Resilience

#### Retry Strategy
- **Exponential Backoff**: With jitter to prevent thundering herd
- **Retry Count**: 5 attempts maximum (based on S3 SLA)
- **Auth Failures**: No retries, immediate fallback to local cache
- **Circuit Breaker**: Stop S3 attempts for 15 minutes after repeated failures

#### Fallback Behavior
- **S3 Unavailable**: Cache writes locally, sync when service returns
- **Auth Failures**: Switch to read-only mode with cached data
- **Health Checks**: Continuous S3 connectivity monitoring
- **Graceful Degradation**: Inform users of degraded performance

### Data Validation and Recovery
- **Markdown Validation**: Parse and validate structure on read operations
- **Corruption Detection**: Checksum validation for critical operations
- **Backup Strategy**: S3 versioning provides automated backup
- **Recovery**: Manual recovery tools for corrupted local files

### Service Resilience
- **Search Index**: Rebuild automatically on corruption, notify users of degradation
- **Cache Failures**: Operate without cache, log performance impact
- **HTTP Service**: Auto-restart on crashes, health check endpoints
- **Circuit Breakers**: Disable failing subsystems temporarily with health status

---

## Testing Strategy

### Unit and Integration Testing

#### S3 Testing
- **Mock S3**: Implement S3 interface mock for unit tests
- **Test Isolation**: Each test uses separate temporary directories
- **Error Simulation**: Mock S3 failures, timeouts, and auth errors
- **No Real S3**: Assume S3 SDK works correctly, focus on our logic

#### Concurrency Testing
- **Race Condition Tests**: Multiple goroutines accessing same notes
- **Lock Testing**: Verify file locking prevents corruption
- **Stress Tests**: High-load scenarios with multiple interfaces
- **Integration**: CLI + TUI + HTTP API running simultaneously

#### Platform Support
- **Target Platforms**: macOS and Linux (Windows via WSL)
- **Cross-Platform**: File path handling, case sensitivity
- **Test Automation**: GitHub Actions for cross-platform CI

### Performance Testing
- **Dynamic Test Data**: Generate realistic note content and structures
- **Baseline Metrics**: Establish performance baselines before optimization
- **Monitoring**: Track performance trends, identify regressions
- **SLI/SLO**: Define and monitor after initial implementation

### MCP Testing
- **Claude Code Integration**: Test with actual Claude Code/Desktop
- **Mock MCP Client**: Unit tests without full MCP stack
- **Protocol Compliance**: Validate MCP protocol implementation

---

## Documentation Plan

### Core Documentation (Create After Architecture)
- `/docs/architecture.md` - System design and data flow
- `/docs/api.md` - HTTP/gRPC API specifications
- `/docs/mcp.md` - MCP integration guide
- `LICENSE` - MIT License for maximum adoption
- `CONTRIBUTING.md` - Development guidelines

### Operational Documentation (Post-Implementation)
- Deployment guide for multiple environments
- Troubleshooting guide based on real usage
- Migration guide for importing from other tools
- Performance tuning guide

---

## Critical Architectural Decisions Made

### 1. Security Architecture ✅
- ULID-based note IDs for portability and uniqueness
- GitHub Flavored Markdown with HTML sanitization
- 10MB file size limit for performance
- Local trust model, remote API key authentication

### 2. Concurrency Model ✅
- Priority-based multi-process coordination
- Per-note file locking with 10-second timeout
- S3 last-write-wins with ETag optimistic locking
- 5-minute cache TTL with time-based invalidation

### 3. Error Handling ✅
- S3 exponential backoff with 5 retries and 15-minute circuit breaker
- Local cache fallback for S3 outages
- Automatic search index rebuild on corruption
- Service auto-restart with health monitoring

### 4. Testing Strategy ✅
- Mock S3 for unit tests, no real cloud costs
- Dynamic test data generation with isolated test environments
- macOS/Linux focus, Windows via WSL
- Claude Code integration for MCP testing

### 5. Operations ✅
- MIT License for maximum adoption and commercial flexibility
- Post-implementation documentation strategy
- Performance monitoring without premature optimization
- Health checks and degraded state management

### 1. Concurrency & Data Safety
- **Questions**: File locking strategy? Multi-process safety? S3 eventual consistency handling? Crash recovery?
- **Impact**: Core reliability and data integrity

### 2. Plugin Architecture Deep Dive
- **Questions**: Go plugins vs alternatives? Security model? API versioning? Discovery mechanism?
- **Impact**: Extensibility and ecosystem growth

### 3. Error Handling & Reliability
- **Questions**: Retry policies? Circuit breakers? Data validation? Backup/recovery strategies?
- **Impact**: Production readiness and user experience

### 4. Security Model
- **Questions**: Input validation? Path traversal protection? Rate limiting? Encryption at rest? Credential management?
- **Impact**: Enterprise adoption and data protection

### 5. Observability & Operations
- **Questions**: Metrics collection? Log aggregation? Performance monitoring? Debugging tools?
- **Impact**: Production deployment and maintenance

### 6. Data Migration & Compatibility
- **Questions**: Import from Obsidian/Notion/Roam? Export formats? ID migration? Link preservation?
- **Impact**: User adoption and vendor lock-in prevention

### 7. Performance Testing & Limits
- **Questions**: Load testing methodology? Memory/CPU limits? Performance regression detection?
- **Impact**: Scalability and reliability claims

### 8. Deployment & Distribution
- **Questions**: Installation methods? Auto-updates? Version compatibility? Container strategy?
- **Impact**: User onboarding and maintenance

### 9. Testing Strategy
- **Questions**: Integration testing across interfaces? Cross-platform automation? S3 mocking?
- **Impact**: Code quality and reliability

### 10. Community & Ecosystem
- **Questions**: Plugin marketplace? Documentation strategy? Contribution guidelines?
- **Impact**: Long-term project sustainability

### 11. Configuration Management
- **Questions**: Config validation? Templates? Environment-specific configs? Drift detection?
- **Impact**: Operational complexity and user experience

### 12. Backward Compatibility
- **Questions**: API versioning? Configuration evolution? Storage format migrations?
- **Impact**: Upgrade path and user retention