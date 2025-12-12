# BUG-003: Inconsistent Note Storage and Path Handling

**Status:** Open  
**Severity:** 🟡 MEDIUM  
**Priority:** P1  
**Component:** Multiple files (config, storage, new, edit)  
**Date Reported:** 2025-12-11  
**Depends On:** BUG-001, BUG-002

## Problem Statement

The note storage system uses two different paths that sometimes conflict:

1. **Configuration path:** `storage.local.path = "./notes"`
2. **Vault path:** `vault.notes_dir = "notes"`

This creates confusion and inconsistent behavior:
- Sometimes notes are stored in `./notes/`
- Sometimes in `./notes/notes/`
- Path resolution logic is scattered across multiple files

## Root Cause Analysis

### Path Resolution Sources

**In `cmd/kbvault/new.go` createNewNote():**
```go
filePath := filepath.Join(config.Vault.NotesDir, filename)
// Results in: "notes/01KC7WAMJZPQB6960Q8NAK7ZT4.md"
```

**In `pkg/storage/local/storage.go`:**
```go
func (s *Storage) Write(ctx context.Context, path string, data []byte) error {
    fullPath := s.getFullPath(path)  // Prepends storage.local.path
    // Results in: "./notes/notes/01KC7WAMJZPQB6960Q8NAK7ZT4.md"
}
```

### Configuration Redundancy

Both are set to similar values:
```toml
[vault]
  notes_dir = "notes"

[storage.local]
  path = "./notes"
```

This means:
- `vault.notes_dir` is meant for vault-specific organization
- `storage.local.path` is meant for storage backend location
- **They are doing the same thing, creating duplication**

### Actual Behavior

When a note is created:
1. `createNewNote()` creates path: `"notes/filename.md"`
2. `storage.Write()` prefixes: `"./notes"` + `"notes/filename.md"` = `"./notes/notes/filename.md"`
3. **Nested directory structure created**
4. Some operations find files in `notes/`
5. Some operations look in `notes/notes/`
6. Inconsistent results

## Impact

- ⚠️ Nested `notes/notes/` directory created
- ⚠️ Inconsistent paths in different commands
- ⚠️ Potential for notes to be in wrong location
- ⚠️ Configuration is confusing (two paths doing the same thing)
- ⚠️ Hard to understand where notes actually are stored

## Evidence

```bash
$ ls -la /Users/andhi/code/code-notes/
drwxr-xr-x  notes/
drwxr-xr-x  vault/

$ ls -la /Users/andhi/code/code-notes/notes/
drwxr-xr-x  notes/  # ← Nested directory!
-rw-r--r--  01KC7VANT6YTN1TFZG8V6TRBS9.md

$ ls -la /Users/andhi/code/code-notes/notes/notes/
-rw-r--r--  01KC7VCA6HEHK4HR8XQQF5DV6R.md
# Files are in both locations!
```

## Files Affected

- `cmd/kbvault/new.go` - `createNewNote()` uses `config.Vault.NotesDir`
- `pkg/storage/local/storage.go` - Storage backend prepends `storage.local.path`
- `.kbvault/config.toml` - Has both `vault.notes_dir` and `storage.local.path`
- `cmd/kbvault/edit.go` - May look in wrong directory
- `cmd/kbvault/delete.go` - May look in wrong directory

## Solution Requirements

1. **Pick ONE source of truth** for note location
2. **Remove redundant configuration** (either vault.notes_dir OR storage.local.path)
3. **Use consistent path building** across all commands
4. **Don't combine paths** - use storage backend path ONLY or vault path ONLY, not both
5. **Update configuration** to remove ambiguity
6. **Migrate existing notes** if needed

## Options

### Option A: Use Storage Backend Only
- Keep `storage.local.path = "./notes"`
- Remove `vault.notes_dir`
- All paths resolved through storage backend
- Pros: Consistent, single source of truth
- Cons: Remove a config option

### Option B: Use Vault Config Only
- Keep `vault.notes_dir = "notes"`
- Remove `storage.local.path` (or use as base only)
- All paths resolved through vault config
- Pros: Clearer configuration
- Cons: Change storage backend interface

### Option C: Use Absolute Paths
- Store full paths in config
- No path combination
- Clear, explicit, no surprises
- Pros: Most explicit
- Cons: Less portable

## Proposed Fix

**Option A (Recommended):**

1. Remove `vault.notes_dir` from schema
2. Use only `storage.local.path` for all storage
3. In `createNewNote()`, use storage backend path directly
4. Update all path references to go through storage

```go
// Before
filePath := filepath.Join(config.Vault.NotesDir, filename)

// After
filePath := filepath.Join(filename)  // Storage backend handles base path
```

## Testing Strategy

1. Create new vault
2. Verify notes go to correct single location
3. No nested directories created
4. All commands (new, edit, delete, search, show) use same path
5. Migrate existing notes to new location

## Related Issues

- BUG-001: Editor changes lost
- BUG-002: Notes can't be found

## Notes

This is a design issue rather than a bug - the architecture conflates two different concerns (vault organization vs. storage backend). The fix is to clarify the architecture and use consistent path handling throughout.

Once this is fixed, it will be easier to diagnose and fix BUG-001 and BUG-002.
