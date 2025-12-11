# kbVault - Knowledge Base Vault CLI

A high-performance, production-ready Go knowledge management system with multiple storage backends, full-text search, and powerful CLI interface. Designed for managing markdown-based knowledge vaults at scale.

## What is kbVault?

kbVault is a command-line tool for managing your knowledge base. Store notes in markdown, organize them with tags and links, search across your entire vault, and access your knowledge from anywhere—all with zero external dependencies for core functionality.

**Perfect for:**
- Personal knowledge management systems
- Team documentation
- Research note-taking
- Project-specific vaults
- Multiple concurrent knowledge bases with profiles

## Key Features

- **📝 Simple Note Management** - Create, edit, delete, and organize notes with ease
- **🔍 Full-Text Search** - Fast search across all your notes
- **🔗 Bidirectional Links** - Connect related notes automatically
- **📦 Storage Flexibility** - Local filesystem or S3-compatible storage
- **👥 Multi-Profile Support** - Manage multiple vaults with different configurations
- **⚙️ Zero Configuration** - Works out-of-the-box with sensible defaults
- **🚀 High Performance** - Optimized for vaults with thousands of notes
- **🐚 Shell Completions** - Tab completion for bash, zsh, and fish

## Quick Start

### Installation

**macOS & Linux (Homebrew):**
```bash
brew tap madstone-tech/tap
brew install kbvault
```

**Using Go:**
```bash
go install github.com/madstone-tech/mdstn-kb-mcp/cmd/kbvault@latest
```

**From Binary:**
Download from [GitHub Releases](https://github.com/madstone-tech/mdstn-kb-mcp/releases)

### Your First Vault

```bash
# Initialize a vault
kbvault init ~/my-vault

# Create your first note
kbvault new "Welcome to kbVault"

# List your notes
kbvault list

# Search notes
kbvault search "welcome"
```

See [Getting Started Guide](docs/guides/getting-started.md) for detailed setup.

## Usage Examples

```bash
# Create a note
kbvault new "Python Tips"

# Search your vault
kbvault search "async programming"

# List notes with filtering
kbvault list --tag python

# Edit a note
kbvault edit "Python Tips"

# Manage multiple vaults
kbvault --profile work new "Team Meeting"
kbvault --profile personal new "Personal Goal"

# View CLI help
kbvault --help
```

See [CLI Reference](docs/guides/cli-reference.md) for all commands.

## Documentation

### For Users

- **[Getting Started](docs/guides/getting-started.md)** - Installation and setup
- **[CLI Reference](docs/guides/cli-reference.md)** - Complete command documentation
- **[Configuration Guide](docs/guides/configuration.md)** - Configure your vault
- **[Profiles & Multi-Vault](docs/guides/profiles.md)** - Manage multiple vaults

### For Developers

- **[Documentation Index](docs/README.md)** - Central hub for all docs
- **[Architecture Overview](docs/architecture/overview.md)** - System design
- **[Package Reference](docs/architecture/packages.md)** - Public API documentation
- **[Building & Testing](docs/development/building.md)** - Development guide

### Project Information

- **[Product Requirements](docs/PRD.md)** - Complete specifications
- **[Implementation Plan](docs/implementation-sessions.md)** - Development progress
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute

## Project Structure

```
kbvault/
├── cmd/kbvault/              # CLI application
│   ├── main.go               # Entry point
│   ├── new.go, show.go, ...  # Commands
│   └── *_test.go             # Tests
│
├── pkg/                       # Public packages
│   ├── config/               # Configuration & profiles
│   ├── storage/              # Storage backends
│   │   ├── local/            # Filesystem storage
│   │   └── s3/               # S3-compatible storage
│   ├── types/                # Core types
│   ├── ulid/                 # ID generation
│   ├── retry/                # Retry logic
│   └── vector/               # Vector DB (planned)
│
├── internal/                 # Private packages
│   ├── links/                # Link management
│   ├── search/               # Search engine
│   ├── templates/            # Note templates
│   └── api/, mcp/, tui/      # Future interfaces
│
├── docs/                     # Documentation
│   ├── guides/               # User guides
│   ├── architecture/         # Architecture docs
│   ├── development/          # Development docs
│   └── README.md            # Docs index (MOC)
│
├── scripts/                  # Build & utility scripts
├── completions/              # Shell completions
├── configs/                  # Config templates
├── test/                     # Test data
└── Makefile                  # Build automation
```

## Supported Platforms

| Platform | Architecture | Status |
|----------|-------------|--------|
| macOS | Intel (amd64) | ✅ Supported |
| macOS | Apple Silicon (arm64) | ✅ Supported |
| Linux | x86_64 (amd64) | ✅ Supported |
| Linux | ARM (arm64) | ✅ Supported |
| Windows | x86_64 | 📋 Planned |

## Storage Options

### Local Storage (Default)
Store notes in your local filesystem as TOML files. Perfect for personal vaults and development.

```toml
[storage]
type = "local"
path = "./notes"
```

### S3-Compatible Storage
Store notes in AWS S3 or any S3-compatible service (MinIO, DigitalOcean Spaces, etc.). Ideal for team vaults and cloud backups.

```toml
[storage]
type = "s3"
bucket = "my-kb"
region = "us-east-1"
```

See [Configuration Guide](docs/guides/configuration.md) for setup details.

## Development

### Quick Build

```bash
# Build binary
make build

# Run tests
make test

# Format code
make fmt

# Full checks
make check
```

### Requirements

- Go 1.25 or later
- Make
- golangci-lint (for linting)

See [Building & Testing Guide](docs/development/building.md) for detailed setup.

## Testing

```bash
# Run full test suite
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Specific package
go test -v ./pkg/config
```

**Coverage:** 62.8% | **Target:** 70%+

## Performance

kbVault is optimized for performance:

- **⚡ Fast Search**: Indexed full-text search for sub-second results
- **💾 Minimal Memory**: Efficient indexing and caching
- **🚀 Scalable**: Supports vaults with thousands of notes
- **🔄 Incremental Updates**: Only updated notes are re-indexed

Benchmark results available in test output.

## Configuration

kbVault uses TOML configuration. Profiles allow you to manage multiple vaults:

```bash
# Create profiles
kbvault profile create work --storage-path ~/work-vault
kbvault profile create personal --storage-path ~/personal-vault

# Use specific profile
kbvault --profile work list
kbvault --profile personal new "Personal Note"

# Set default profile
kbvault profile set-active work
```

See [Profiles Guide](docs/guides/profiles.md) and [Configuration Guide](docs/guides/configuration.md).

## Feature Status

### Fully Implemented (v1.0.0+)
- ✅ Note management (CRUD operations)
- ✅ Full-text search with inverted indexing
- ✅ Local & S3-compatible storage
- ✅ Multi-profile support for multiple vaults
- ✅ Bidirectional link detection and management
- ✅ Shell completions (bash, zsh, fish)
- ✅ TOML-based configuration system
- ✅ Template system for note creation

### In Progress / Partial
- 🟡 MCP Protocol - Basic structure in place, not fully functional
- 🟡 HTTP Server - Configuration exists, API endpoints not yet implemented

### Planned (v1.1.0+)
- 📋 Vector-based semantic search
- 📋 HTTP REST API endpoints
- 📋 Terminal UI (TUI)
- 📋 gRPC API
- 📋 Windows support

See [Implementation Plan](docs/implementation-sessions.md) for details.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup
- Code style guidelines
- Testing requirements
- Pull request process

## License

MIT License - See [LICENSE](LICENSE) for details.

## Support

- 📖 **Documentation**: Start with [Getting Started](docs/guides/getting-started.md)
- 🐛 **Report Issues**: [GitHub Issues](https://github.com/madstone-tech/mdstn-kb-mcp/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/madstone-tech/mdstn-kb-mcp/discussions)
- 🤝 **Contribute**: See [CONTRIBUTING.md](CONTRIBUTING.md)

## Acknowledgments

Built with Go and inspired by modern knowledge management systems. Special thanks to all contributors and users.

---

**[📚 Documentation](docs/README.md) | [🚀 Quick Start](docs/guides/getting-started.md) | [💻 CLI Reference](docs/guides/cli-reference.md) | [🔧 API Docs](docs/architecture/packages.md)**
