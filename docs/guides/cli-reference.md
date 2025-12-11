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
- `--template <name>` - Use a specific template
- `--tags <tag1,tag2>` - Add tags to the note
- `--no-edit` - Create note without opening editor

**Examples:**
```bash
# Create a note with title
kbvault new "My New Note"

# Create and edit immediately
kbvault new "Research Topic"

# Create without opening editor
kbvault new "Quick Thought" --no-edit

# Create with tags
kbvault new "Python Tutorial" --tags python,tutorial,beginner

# Use a template
kbvault new "Meeting Notes" --template meeting
```

**Output:**
```
Created note: 01ARZ3NDEKTSV4RRFFQ69G5FAV
Title: My New Note
```

---

#### `show` - Display a note

Show the contents of a note.

```bash
kbvault show [id-or-title]
```

**Arguments:**
- `id-or-title` - Note ID or title (optional; lists if not provided)

**Options:**
- `--format <json|toml|text>` - Output format (default: text)
- `--raw` - Show raw content without formatting

**Examples:**
```bash
# Show by ID
kbvault show 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Show by title (substring match)
kbvault show "My New"

# Show in JSON format
kbvault show "My Note" --format json

# Show raw content
kbvault show "My Note" --raw

# Interactive selection if ambiguous
kbvault show
```

---

#### `list` - List all notes

List all notes in the vault with optional filtering.

```bash
kbvault list [options]
```

**Options:**
- `--tag <tag>` - Filter by tag
- `--search <query>` - Search notes
- `--sort <field>` - Sort by field (created, modified, title)
- `--format <json|table|csv>` - Output format (default: table)
- `--limit <n>` - Limit results

**Examples:**
```bash
# List all notes
kbvault list

# List notes with specific tag
kbvault list --tag python

# Search while listing
kbvault list --search "tutorial"

# Sort by modification time
kbvault list --sort modified

# Export as JSON
kbvault list --format json

# Limit results
kbvault list --limit 10
```

**Output:**
```
ID                        Title                     Tags              Created
01ARZ3NDEKTSV4RRFFQ69G5FAV  My New Note               -                 2025-12-11 10:30:00
01ARZ3NDEKTSV4RRDKJ50H1AB  Research Topic            -                 2025-12-11 10:35:00
```

---

#### `edit` - Edit an existing note

Edit a note in your default editor.

```bash
kbvault edit [id-or-title]
```

**Arguments:**
- `id-or-title` - Note ID or title

**Options:**
- `--delete-field <field>` - Remove a field from front matter

**Examples:**
```bash
# Edit by ID
kbvault edit 01ARZ3NDEKTSV4RRFFQ69G5FAV

# Edit by title
kbvault edit "My Note"

# Edit and remove a field
kbvault edit "My Note" --delete-field tags
```

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

Search for notes using full-text search.

```bash
kbvault search [query] [options]
```

**Arguments:**
- `query` - Search query

**Options:**
- `--field <field>` - Search specific field (title, content, tags)
- `--case-sensitive` - Case-sensitive search
- `--limit <n>` - Limit results
- `--format <json|table>` - Output format

**Examples:**
```bash
# Full-text search
kbvault search "python"

# Search in specific field
kbvault search --field title "tutorial"
kbvault search --field tags "python"
kbvault search --field content "async"

# Case-sensitive search
kbvault search --case-sensitive "Python"

# Limit results
kbvault search "note" --limit 5

# Export results
kbvault search "query" --format json
```

---

### Configuration Commands

#### `config` - Manage vault configuration

View and manage vault configuration.

```bash
kbvault config <subcommand> [options]
```

**Subcommands:**

**`config show`** - Display current configuration
```bash
kbvault config show
```

**`config set`** - Set a configuration value
```bash
kbvault config set <key> <value>
kbvault config set notes_dir ./my-notes
kbvault config set storage.type s3
```

**`config get`** - Get a specific configuration value
```bash
kbvault config get notes_dir
kbvault config get storage.path
```

**`config validate`** - Validate configuration
```bash
kbvault config validate
```

**Examples:**
```bash
# View all configuration
kbvault config show

# Change notes directory
kbvault config set notes_dir ./notes

# Change storage backend
kbvault config set storage.type s3

# Validate configuration
kbvault config validate
```

---

#### `configure` - Interactive configuration

Configure vault interactively.

```bash
kbvault configure [options]
```

**Options:**
- `--reset` - Reset to default configuration
- `--profile <name>` - Configure specific profile

**Examples:**
```bash
# Interactive configuration
kbvault configure

# Reset to defaults
kbvault configure --reset

# Configure profile
kbvault --profile work configure
```

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

# Search by tag
kbvault list --tag python

# Filter search results by tag
kbvault search --field tags "python"
```

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

### Search Multiple Fields

```bash
# Search title and tags
kbvault search --field title "python"
kbvault search --field tags "beginner"
```

### Export Notes

```bash
# Export as JSON
kbvault list --format json > notes.json

# Export search results
kbvault search "query" --format json > results.json
```

### Batch Operations

```bash
# Create multiple notes
for title in "Note 1" "Note 2" "Note 3"; do
  kbvault new "$title" --no-edit
done

# Delete all with tag
kbvault list --tag old --format json | jq -r '.[] | .id' | xargs kbvault delete --force
```

### Profile Switching

```bash
# Quick switch
kbvault --profile work list
kbvault --profile personal new "Personal Note"

# Set default
kbvault profile set-active work

# Then use without --profile flag
kbvault new "Work Note"
```

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
