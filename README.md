# kbVault - Knowledge Base Vault

A high-performance Go tool for managing markdown-based knowledge vaults with multiple interaction interfaces (CLI, TUI, API, MCP), designed for both human users and LLM agents.

## Features

- **Multi-Interface Design**: CLI, TUI, HTTP API, and MCP support
- **Storage Flexibility**: Local filesystem and S3-compatible storage
- **Smart Caching**: Multi-level caching for optimal performance
- **Note ID System**: Unique timestamp-based note identification
- **Template System**: Customizable note templates
- **Link Management**: Bidirectional link tracking and validation
- **Search Engine**: Full-text and metadata-based search
- **Optional gRPC**: High-performance API for specialized use cases

## Quick Start

```bash
# Initialize a new vault
kbvault init ~/my-knowledge-vault

# Create your first note
kbvault new "My First Note"

# Search notes
kbvault search "first"

# Start the API server
kbvault server start

# Launch the TUI
kbvault tui
```

## Project Structure

```
├── cmd/kbvault/          # CLI application entry point
├── internal/             # Private application code
│   ├── core/            # Core business logic
│   ├── storage/         # Storage backends (local, S3)
│   ├── cache/           # Caching layer
│   ├── api/             # HTTP/gRPC API server
│   ├── mcp/             # MCP protocol implementation
│   └── tui/             # Terminal UI
├── pkg/                  # Public packages
│   ├── types/           # Shared types
│   └── utils/           # Utility functions
├── configs/              # Configuration templates
├── docs/                 # Documentation
├── scripts/              # Build and deployment scripts
└── test/                 # Test files
```

## Documentation

- [Product Requirements Document](docs/PRD.md) - Complete project specifications
- [Architecture Overview](docs/architecture.md) - System design and components
- [API Documentation](docs/api.md) - HTTP and gRPC API reference
- [MCP Integration](docs/mcp.md) - Model Context Protocol usage

## Development

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build

# Run development server
make dev
```

## Configuration

kbVault uses TOML configuration files. Example:

```toml
[vault]
name = "my-kb"
notes_dir = "notes"

[storage]
type = "local"
path = "/path/to/vault"

[server]
http_enabled = true
http_port = 8080
grpc_enabled = false
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.
