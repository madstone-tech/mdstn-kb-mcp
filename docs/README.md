# kbVault Documentation

Welcome to the kbVault knowledge base vault documentation. This is your central hub for understanding how to use, extend, and contribute to kbVault.

## Quick Navigation

### For Users

- **[Getting Started](guides/getting-started.md)** - Installation, setup, and your first vault
- **[CLI Reference](guides/cli-reference.md)** - Complete command reference with examples
- **[Configuration Guide](guides/configuration.md)** - Setting up profiles, storage backends, and options
- **[Profiles & Multi-Vault](guides/profiles.md)** - Managing multiple knowledge vaults

### For Developers

- **[Architecture Overview](architecture/overview.md)** - System design and core concepts
- **[Package Reference](architecture/packages.md)** - Detailed breakdown of public packages
- **[Building & Testing](development/building.md)** - Build system, testing, and development workflow
- **[Contributing Guide](../CONTRIBUTING.md)** - Contribution guidelines and code style

### Project Documentation

- **[Product Requirements](PRD.md)** - Complete project specifications and roadmap
- **[Implementation Plan](implementation-sessions.md)** - Development sessions and progress tracking
- **[GitHub Issues Plan](github-issues-plan.md)** - Issue tracking and project planning

## Feature Overview

### Core Features

**Note Management**
- Create, read, update, and delete notes with unique IDs
- Full markdown support with front matter
- Bidirectional link tracking
- Template system for consistent note creation

**Storage & Caching**
- Local filesystem storage (default)
- S3-compatible storage backend
- Multi-level caching for performance
- Automatic cache invalidation

**Search Engine**
- Full-text search across all notes
- Field-specific search (title, content, tags, etc.)
- Vector-based semantic search (optional)
- Index management and optimization

**Link Management**
- Automatic bidirectional link detection
- Link validation and broken link detection
- Link graph visualization (planned)

**Profile Management**
- Multiple concurrent profiles
- Profile-specific configurations
- Easy switching between vaults
- Profile templates

### Interfaces

kbVault provides multiple ways to interact with your knowledge base:

- **CLI** ✅ - Command-line interface with shell completions
- **Configuration** ✅ - TOML-based configuration with profiles
- **HTTP API** 📋 - RESTful API endpoints (configuration system exists, implementation pending)
- **MCP** 🟡 - Model Context Protocol (basic structure, not fully functional)
- **TUI** 📋 - Terminal User Interface (planned)
- **gRPC** 📋 - High-performance API (planned)

## Common Tasks

### Setup & Administration
- [Install kbVault](guides/getting-started.md#installation)
- [Create your first vault](guides/getting-started.md#creating-your-first-vault)
- [Configure storage backends](guides/configuration.md#storage-backends)
- [Manage multiple profiles](guides/profiles.md)

### Working with Notes
- [Create a note](guides/cli-reference.md#new)
- [Search for notes](guides/cli-reference.md#search)
- [Edit existing notes](guides/cli-reference.md#edit)
- [Organize with tags and links](guides/cli-reference.md#note-organization)

### Troubleshooting
- Check the relevant guide for common issues
- Review [Building & Testing](development/building.md) for development problems
- See [Contributing Guide](../CONTRIBUTING.md) for reporting bugs

## Project Structure

```
kbvault/
├── cmd/kbvault/                    # CLI application
├── pkg/                             # Public packages
│   ├── config/                      # Configuration management (profiles, TOML)
│   ├── storage/                     # Storage abstraction (local, S3)
│   ├── types/                       # Core types (Note, Config, Vector)
│   ├── retry/                       # Retry logic for resilience
│   ├── vector/                      # Vector database integration
│   └── ulid/                        # Unique ID generation
├── internal/                        # Private application packages
│   ├── links/                       # Link parsing and validation
│   ├── search/                      # Full-text search engine
│   ├── templates/                   # Note template system
│   ├── api/                         # HTTP/gRPC API (planned)
│   ├── mcp/                         # MCP protocol implementation
│   ├── tui/                         # Terminal UI (planned)
│   └── cache/                       # Caching layer (planned)
├── completions/                     # Shell completions (bash, zsh, fish)
├── configs/                         # Configuration templates
├── docs/                            # Documentation (this directory)
├── scripts/                         # Build and utility scripts
└── test/                            # Test fixtures and data

```

## Key Concepts

### Notes
Notes are the core unit of kbVault. Each note:
- Has a unique timestamp-based ID (ULID)
- Contains markdown content with optional front matter
- Can link to other notes via wiki-style links
- Supports tags for categorization
- Automatically tracks creation and modification times

### Profiles
Profiles let you manage multiple knowledge vaults:
- Each profile has its own configuration
- Switch between profiles easily: `kbvault --profile work`
- Set an active profile to use by default
- Isolated storage per profile

### Links
kbVault automatically detects and manages links:
- Wiki-style links: `[[Note Title]]` or `[[path/to/note]]`
- Automatically tracks bidirectional relationships
- Validates links during operations
- Can detect and report broken links

### Storage Backends
Choose where your notes are stored:
- **Local** (default): Notes stored in local filesystem (TOML format)
- **S3**: Notes stored in S3-compatible storage
- Abstracted interface allows adding more backends

## Development Quick Links

- [Build Instructions](development/building.md)
- [Testing Guide](development/building.md#testing)
- [Architecture Details](architecture/overview.md)
- [Package Reference](architecture/packages.md)

## Support

- 📖 Read the documentation
- 🐛 [Report issues](https://github.com/madstone-tech/mdstn-kb-mcp/issues)
- 💬 Check [existing discussions](https://github.com/madstone-tech/mdstn-kb-mcp/discussions)
- 🤝 [Contribute](../CONTRIBUTING.md)

## License

kbVault is licensed under the MIT License. See [LICENSE](../LICENSE) for details.

---

**Last Updated**: December 2025  
**Status**: Active Development  
**Current Version**: 1.0.0+
