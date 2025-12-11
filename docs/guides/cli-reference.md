# CLI Reference

Complete reference for all kbVault command-line commands.

> ⚠️ **Note on Command Implementation Status**: Some commands listed below are partially implemented (placeholders). 
> - **Fully Working**: `init`, `new`, `search`, `delete`, `config`, `profile`, `configure`
> - **Partially Working**: `show` (returns placeholder), `list` (not yet implemented), `edit` (accepts input but limited), `completion`
> - For up-to-date implementation status, run `kbvault --help` or check the [Feature Status](../README.md#feature-status).

## Usage

```bash
kbvault [global-options] <command> [command-options]
```

## Global Options

All commands support these global options:

```
--profile <name>     Use a specific profile (default: active profile)
--help               Show help for a command
--version            Show kbVault version
```

## Commands

### Core Commands

#### `init` - Initialize a new vault

Initialize a new knowledge vault at the specified path.

```bash
kbvault init [path]
```

**Arguments:**
- `path` - Directory path for the vault (creates if doesn't exist)

**Examples:**
```bash
# Initialize in current directory
kbvault init

# Initialize at specific path
kbvault init ~/my-knowledge-vault

# Initialize with profile
kbvault --profile work init ~/work-vault
```

**Creates:**
- `.kbvault/config.toml` - Vault configuration
- `.kbvault/index/` - Search index directory
- `notes/` - Notes storage directory

---

#### `new` - Create a new note

Create a new note and open it in your default editor.

```bash
kbvault new [title]
```

**Arguments:**
- `title` - Note title (optional; prompted if not provided)

**Options:**
- `-o, --open` - Open the note in default editor after creation
- `--tags <tag1,tag2>` - Add tags to the note (comma-separated)
- `--template <name>` - Use a specific template (default: "default")
- `-t, --title <string>` - Note title (alternative to positional argument)

**Examples:**
```bash
# Create a note with title
kbvault new "My New Note"

# Create with tags
kbvault new "Python Tutorial" --tags python,tutorial,beginner

# Use a specific template
kbvault new "Meeting Notes" --template meeting

# Create without opening editor (just pass title without --open)
kbvault new "Quick Thought"

# Create with title flag instead of positional argument
kbvault new --title "My Note" --tags test
```

**Output:**
```
Created note: 01ARZ3NDEKTSV4RRFFQ69G5FAV
Title: My New Note
```

---

#### `show` - Display a note (Placeholder - Not Fully Implemented)

Display a note's content. This command is currently a placeholder and returns limited information.

```bash
kbvault show <note-id>
```

**Arguments:**
- `note-id` - Note ID (required)

**Options:**
- `-c, --content` - Show note content (default: true)
- `-f, --format <json|default|markdown>` - Output format (default: default)
- `-m, --metadata` - Show note metadata (default: true)

**Current Limitations:**
- Does not load actual note content
- Returns placeholder data
- Full implementation coming in future release

**Examples:**
```bash
# Show by ID (returns placeholder)
kbvault show 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Show in JSON format
kbvault show 01ARZ3NDEKTSV4RRFFQ69G5FAV --format json
```

**Workaround:** Use `search` to find notes until `show` is fully implemented.

---

#### `list` - List all notes (Placeholder - Not Fully Implemented)

List all notes in the vault. This command is currently a placeholder.

```bash
kbvault list [options]
```

**Options:**
- `-t, --tags <tag1,tag2>` - Filter by tags (comma-separated)
- `-s, --sort <field>` - Sort by field (title, created, updated, default: updated)
- `-r, --reverse` - Reverse sort order
- `-f, --format <format>` - Output format (default, compact, json, default: default)
- `-l, --limit <n>` - Limit number of results (0 = no limit)
- `-p, --paths` - Show file paths

**Current Limitations:**
- Returns "Note listing not yet implemented" placeholder message
- Filtering works partially (tags filter exists)
- Full implementation coming in future release

**Examples:**
```bash
# List all notes (returns placeholder)
kbvault list

# List with specific tags
kbvault list --tags python,tutorial

# Limit results
kbvault list --limit 10

# Show as JSON
kbvault list --format json
```

**Workaround:** Use `search` to find notes until `list` is fully implemented.

---

#### `edit` - Edit an existing note

Edit a note in your default editor.

```bash
kbvault edit <note-id-or-title>
```

**Arguments:**
- `note-id-or-title` - Note ID or title (required)

**Options:**
- `--editor <editor>` - Use specific editor (default: $EDITOR environment variable)

**Examples:**
```bash
# Edit by note ID
kbvault edit 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Edit by title
kbvault edit "My Note"

# Use specific editor
kbvault edit 01ARZ3NDEKTSV4RRFFQ69G5FAV --editor vim
```

**Note:** If multiple notes match the title, you'll be prompted to choose.

---

#### `delete` - Delete a note

Delete one or more notes.

```bash
kbvault delete [id-or-title...]
```

**Arguments:**
- `id-or-title` - One or more note IDs or titles

**Options:**
- `--force` - Skip confirmation prompt
- `--dry-run` - Show what would be deleted without deleting

**Examples:**
```bash
# Delete a note (with confirmation)
kbvault delete 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Delete by title
kbvault delete "Old Note"

# Delete multiple notes
kbvault delete "Note 1" "Note 2" 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Delete without confirmation
kbvault delete 01ARZ3NDEKTSV4RRFFQ69G5FAV --force

# Preview deletion
kbvault delete "Old Note" --dry-run
```

---

### Search Commands

#### `search` - Search notes

Search for notes using full-text search across all content.

```bash
kbvault search <query> [options]
```

**Arguments:**
- `query` - Search query (required)

**Options:**
- `--limit <n>` - Limit number of results
- `-f, --format <format>` - Output format (default: table, available: json)

**Current Limitations:**
- `--field` option exists but returns no results (partial implementation)
- `--case-sensitive` flag not available

**Examples:**
```bash
# Full-text search
kbvault search "python"

# Limit results
kbvault search "note" --limit 5

# Export results as JSON
kbvault search "query" --format json
```

**Note:** Search by field (title, tags, content) is not yet working reliably. Use general search for best results.

---

### Configuration Commands

#### `config` - Manage vault configuration

View and validate vault configuration.

```bash
kbvault config <subcommand> [options]
```

**Subcommands:**

**`config show`** - Display current configuration
```bash
kbvault config show
```

**`config path`** - Show configuration file path
```bash
kbvault config path
```

**`config validate`** - Validate configuration
```bash
kbvault config validate
```

**Current Limitations:**
- `config get` - Not available (use `config show` instead)
- `config set` - Causes crash (do not use)

**Examples:**
```bash
# View all configuration
kbvault config show

# Show configuration file path
kbvault config path

# Validate configuration
kbvault config validate
```

**Note:** To modify configuration, edit `.kbvault/config.toml` directly or use `kbvault configure` for interactive setup.

---

#### `configure` - Interactive configuration

Configure vault interactively, similar to AWS CLI.

```bash
kbvault configure [options]
```

**Options:**
- `--profile <name>` - Configure specific profile

**Current Limitations:**
- `--reset` flag not available (edit `.kbvault/config.toml` directly to reset)

**Examples:**
```bash
# Interactive configuration
kbvault configure

# Configure specific profile
kbvault --profile work configure
kbvault --profile personal configure
```

**Note:** You'll be guided through storage type, credentials, and other configuration options.

---

### Profile Commands

#### `profile` - Manage profiles

Manage multiple vault profiles.

```bash
kbvault profile <subcommand> [options]
```

**Subcommands:**

**`profile create`** - Create a new profile
```bash
kbvault profile create <name> [options]
```

Options:
- `--storage-type <local|s3>` - Storage backend (default: local)
- `--storage-path <path>` - Storage path
- `--s3-bucket <bucket>` - S3 bucket name (for S3 storage)
- `--s3-region <region>` - S3 region (for S3 storage)

**`profile list`** - List all profiles
```bash
kbvault profile list
```

**`profile delete`** - Delete a profile
```bash
kbvault profile delete <name> [--force]
```

**`profile set-active`** - Set active profile
```bash
kbvault profile set-active <name>
```

**`profile show`** - Show profile details
```bash
kbvault profile show [name]
```

**Examples:**
```bash
# Create a local profile
kbvault profile create personal --storage-path ~/personal-vault

# Create an S3 profile
kbvault profile create work \
  --storage-type s3 \
  --s3-bucket my-work-kb \
  --s3-region us-east-1

# List all profiles
kbvault profile list

# Set active profile
kbvault profile set-active work

# Use profile for command
kbvault --profile work new "Work Note"

# Show profile info
kbvault profile show work

# Delete profile
kbvault profile delete personal
```

---

### Utility Commands

#### `completion` - Generate shell completions

Generate shell completion scripts.

```bash
kbvault completion <bash|zsh|fish|powershell>
```

**Examples:**
```bash
# Generate bash completion
source <(kbvault completion bash)

# Generate zsh completion
kbvault completion zsh | sudo tee /usr/share/zsh/site-functions/_kbvault

# Generate fish completion
kbvault completion fish | sudo tee /usr/share/fish/vendor_completions.d/kbvault.fish

# Generate powershell completion
kbvault completion powershell | Out-String | Invoke-Expression
```

See [Getting Started](getting-started.md#shell-completions) for setup instructions.

---

## Note Organization

### Using Tags

Tags help organize notes:

```bash
# Create note with tags
kbvault new "Python Tips" --tags python,programming,tips

# Filter notes by tag (note: list command returns placeholder)
kbvault list --tags python

# Search for notes (general search works best)
kbvault search "python"
```

**Note:** The `--field tags` search option is not yet working. Use general search instead.

### Using Links

Create connections between notes:

```markdown
# My Note

This connects to [[Another Note]].
See also: [[docs/architecture]]
```

Links are wiki-style with `[[...]]` syntax.

### Directory Structure

Organize notes in subdirectories:

```bash
# Create note in subdirectory (automatic)
kbvault new "API Design" 
# -> notes/api/design-01AAAA.toml

# List notes from subdirectory
kbvault list  # all notes
kbvault list --search "api"
```

---

## Tips & Tricks

### Search Tips

```bash
# General search across all content
kbvault search "python"

# Search with result limit
kbvault search "tutorial" --limit 10
```

**Note:** Search-by-field is not yet working. General search performs best.

### Export Results

```bash
# Export search results as JSON
kbvault search "query" --format json > results.json
```

**Note:** Export via `list` is not yet available.

### Batch Operations

```bash
# Create multiple notes
for title in "Note 1" "Note 2" "Note 3"; do
  kbvault new "$title"
done

# Delete by ID (bulk)
# Note: kbvault delete accepts multiple IDs/titles
kbvault delete "Old Note 1" "Old Note 2" --force
```

**Note:** Batch operations via `list` are not available yet due to placeholder implementation.

### Profile Switching

```bash
# Quick switch
kbvault --profile work search "project"
kbvault --profile personal new "Personal Note"

# Set default
kbvault profile set-active work

# Then use without --profile flag
kbvault new "Work Note"

# View profile information
kbvault profile show work
```

**Note:** Profile list shows all available profiles.

---

## Environment Variables

- `EDITOR` - Default editor for `new` and `edit` commands
- `KBVAULT_CONFIG` - Path to configuration directory
- `KBVAULT_PROFILE` - Default profile name

---

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Command-line usage error
- `3` - Configuration error
- `4` - Storage error

---

## See Also

- [Getting Started](getting-started.md)
- [Configuration Guide](configuration.md)
- [Profiles & Multi-Vault](profiles.md)
- [Architecture Overview](../architecture/overview.md)
