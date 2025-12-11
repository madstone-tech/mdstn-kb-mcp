# Getting Started with kbVault

Learn how to install kbVault and create your first knowledge vault.

## Installation

### macOS (Homebrew)

```bash
# Add the Madstone Tech tap
brew tap madstone-tech/tap

# Install kbvault
brew install kbvault

# Verify installation
kbvault --version
```

### Linux (Homebrew)

```bash
# Add the Madstone Tech tap
brew tap madstone-tech/tap

# Install kbvault
brew install kbvault

# Verify installation
kbvault --version
```

### Using Go

If you have Go 1.24+ installed:

```bash
go install github.com/madstone-tech/mdstn-kb-mcp/cmd/kbvault@latest

# Verify installation
kbvault --version
```

### From Binary

1. Download the latest release from [GitHub Releases](https://github.com/madstone-tech/mdstn-kb-mcp/releases)
2. Extract the binary for your platform
3. Add to your PATH or run directly: `./kbvault --version`

### Shell Completions

After installation, set up shell completions for a better experience:

**Bash:**
```bash
source <(kbvault completion bash)
# To load completions for each session, run once:
kbvault completion bash | sudo tee /etc/bash_completion.d/kbvault
```

**Zsh:**
```bash
kbvault completion zsh | sudo tee /usr/share/zsh/site-functions/_kbvault
# Restart your shell for completions to take effect
```

**Fish:**
```bash
kbvault completion fish | sudo tee /usr/share/fish/vendor_completions.d/kbvault.fish
```

## Creating Your First Vault

### Step 1: Initialize a Vault

```bash
kbvault init ~/my-knowledge-vault
```

This creates:
- A `.kbvault/` directory with configuration
- `config.toml` with default settings
- An empty `notes/` directory

### Step 2: Create Your First Note

```bash
kbvault new "Welcome to kbVault"
```

This opens your default editor where you can:
1. Add front matter (YAML format):
   ```yaml
   ---
   tags: [getting-started, welcome]
   ---
   ```
2. Write your note content in markdown
3. Save and close the editor

Your note is now saved with a unique ID!

### Step 3: List Your Notes

```bash
kbvault list
```

Output:
```
ID                        Title                     Tags              Created
01ARZ3NDEKTSV4RRFFQ69G5FAV  Welcome to kbVault       getting-started   2025-12-11 10:30:00
```

### Step 4: View a Note

```bash
# View by ID
kbvault show 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Or by title (using a substring)
kbvault show "Welcome"
```

### Step 5: Search Your Notes

```bash
kbvault search "welcome"
```

### Step 6: Edit a Note

```bash
# Edit by ID
kbvault edit 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Or by title
kbvault edit "Welcome"
```

## Understanding the Vault Structure

After initialization, your vault contains:

```
~/my-knowledge-vault/
├── .kbvault/
│   ├── config.toml           # Main configuration
│   └── index/               # Search index (created on demand)
├── notes/                    # Your notes directory
│   ├── 01ARZ3NDEK.toml      # Note files (one per note)
│   └── path/
│       └── 01ARZ3NDET.toml  # Organized notes
└── README.md                 # Optional vault documentation
```

### Note File Format

Notes are stored as TOML files with the following structure:

```toml
# 01ARZ3NDEKTSV4RRFFQ69G5FAV.toml

id = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
title = "Welcome to kbVault"
created = "2025-12-11T10:30:00Z"
modified = "2025-12-11T10:30:00Z"
tags = ["getting-started", "welcome"]

[content]
body = """
# Welcome to kbVault

This is your first note. You can use markdown here...

## Features

- Create notes with unique IDs
- Link notes together: [[Other Note]]
- Search across all notes
- Organize with tags
"""
```

## Key Concepts

### Note IDs

Each note gets a unique ID (ULID - Universally Unique Lexicographically Sortable Identifier):
- Timestamp-based (can sort chronologically)
- 26 characters long
- Example: `01ARZ3NDEKTSV4RRFFQ69G5FAV`

### Links

Create connections between notes using wiki-style syntax:

```markdown
# My Note

This connects to [[Another Note]].

You can also link to specific paths:
- [[docs/architecture]]
- [[research/papers/important-paper]]
```

kbVault automatically tracks these links.

### Tags

Organize notes with tags in the front matter:

```toml
tags = ["python", "tutorial", "beginner"]
```

Search by tag:
```bash
kbvault search --field tags "python"
```

### Profiles

You can manage multiple vaults with different profiles:

```bash
# Create a new profile
kbvault profile create work --storage-type local --storage-path ~/work-vault

# Use the profile
kbvault --profile work new "Work Note"

# Set as active
kbvault profile set-active work
```

See [Profiles & Multi-Vault Guide](profiles.md) for details.

## Next Steps

- **[CLI Reference](cli-reference.md)** - Explore all available commands
- **[Configuration Guide](configuration.md)** - Customize your vault
- **[Profiles Guide](profiles.md)** - Manage multiple vaults
- **[Architecture Overview](../architecture/overview.md)** - Understand how kbVault works

## Troubleshooting

### Command Not Found

Make sure kbVault is installed and in your PATH:
```bash
which kbvault
kbvault --version
```

### Permission Denied

If you get permission errors:
```bash
# Make sure you have write access to the vault directory
ls -ld ~/my-knowledge-vault
chmod 755 ~/my-knowledge-vault
```

### Editor Not Opening

kbVault uses your default editor. Set it if needed:
```bash
export EDITOR=vim
# or
export EDITOR=nano
# or
export EDITOR=code
```

### Configuration Issues

Check your configuration:
```bash
kbvault config show
```

Reset to defaults:
```bash
kbvault configure --reset
```

---

**Ready to explore more?** Check out the [CLI Reference](cli-reference.md) to see all available commands.
