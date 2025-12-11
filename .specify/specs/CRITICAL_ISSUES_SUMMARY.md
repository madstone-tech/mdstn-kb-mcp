# CRITICAL ISSUES SUMMARY - kbVault Note Management System

**Date:** 2025-12-11  
**Reporter:** User Testing Session  
**Status:** Open - Requires Team Discussion  

## Executive Summary

The kbVault note management system has **three critical and interconnected bugs** that severely impact core functionality:

1. **BUG-001**: Editor changes are lost when using `kbvault new --open`
2. **BUG-002**: Notes cannot be found by edit/delete/show commands
3. **BUG-003**: Inconsistent path handling and redundant configuration

These bugs create a **cascading failure** where:
- Users cannot create notes with content (BUG-001)
- Even if they could, the notes couldn't be edited/deleted (BUG-002)
- Path confusion makes debugging difficult (BUG-003)

## Severity Assessment

| Bug | Severity | Impact | Blockers |
|-----|----------|--------|----------|
| BUG-001 | 🔴 CRITICAL | Primary use case broken | Can't write notes with `--open` |
| BUG-002 | 🔴 CRITICAL | Can't manage notes | Can't edit/delete any notes |
| BUG-003 | 🟡 MEDIUM | Architecture confusion | Affects all path operations |

## User-Facing Symptoms

### Symptom 1: Edits Lost
```bash
$ kbvault new "test-123" --open
# User opens editor, types content, saves, exits
$ cat notes/01KC7WAMJZPQB6960Q8NAK7ZT4.md
# Shows only user's text, no frontmatter
# Original frontmatter was overwritten!
```

### Symptom 2: Can't Find Notes
```bash
$ kbvault search "test-123"
Found 1 results:  # ✅ Works

$ kbvault edit "test-123"
Error: no notes found matching 'test-123'  # ❌ Fails

$ kbvault delete "test-123"
Error: no notes found matching 'test-123'  # ❌ Fails
```

### Symptom 3: Path Confusion
```bash
$ ls notes/
drwxr-xr-x  notes/      # ← Nested directory
-rw-r--r--  *.md        # ← Some files here

$ ls notes/notes/
-rw-r--r--  *.md        # ← Other files here
```

## Technical Root Causes

### BUG-001: No Post-Edit File Reading

**File:** `cmd/kbvault/new.go` lines 77-80

The editor is opened but the function returns immediately without:
1. Reading the edited file from disk
2. Updating the note object with new content
3. Re-saving with frontmatter intact

**Result:** File overwritten with only user content, frontmatter lost.

### BUG-002: Title Extraction from Wrong Source

**File:** `cmd/kbvault/edit.go` readNote() function

Attempts to extract title from markdown heading (`# Title`) instead of YAML frontmatter where it's actually stored.

When note lacks markdown heading:
- Title becomes empty string
- Falls back to ID
- User's title lost
- Search by title fails

**Result:** Notes can't be found by edit/delete commands.

### BUG-003: Path Combination Issues

**Files:** Multiple (new.go, storage/local/storage.go, config)

Two paths are combined:
- `config.Vault.NotesDir = "notes"`
- `config.Storage.Local.Path = "./notes"`

Results in: `./notes/notes/filename.md`

**Result:** Inconsistent path handling, nested directories, confusion.

## Why These Are Interconnected

```
BUG-001 (Lost edits)
        ↓
    File has only content, no frontmatter
        ↓
    BUG-002 (Title extraction fails)
        ↓
    readNote() can't parse file
        ↓
    edit/delete/show can't find notes
        ↓
    BUG-003 (Path issues)
        ↓
    Makes debugging harder
```

## Affected Workflows

### ❌ Create Note with Editor (BROKEN)
```bash
kbvault new "Title" --open
# → Editor opens
# → User writes content
# → User saves and exits
# → Edits are LOST
```

### ❌ Edit Existing Note (BROKEN)
```bash
kbvault edit "Title"
# → Error: no notes found matching 'Title'
# Because title extraction failed
```

### ❌ Delete Note (BROKEN)
```bash
kbvault delete "Title"
# → Error: no notes found matching 'Title'
# Because title extraction failed
```

### ⚠️ View Note (PARTIALLY BROKEN)
```bash
kbvault show <ID>
# → Returns placeholder
# → No actual content
```

### ✅ Search Notes (WORKS)
```bash
kbvault search "keyword"
# → Works because uses full-text search
# → Doesn't depend on title parsing
# → Searches content directly
```

## Impact on Users

**Primary Workflows Blocked:**
1. Create note with content (`new --open`)
2. Edit existing note (`edit`)
3. Delete note (`delete`)

**Only Working Commands:**
- `search` - Full-text search
- `new` - Create empty note
- `config` - Configuration
- `profile` - Profile management

**Overall**: System is **non-functional for core use case** (creating and managing notes).

## What Was Tested

| Command | Status | Result |
|---------|--------|--------|
| `kbvault new "title"` | ✅ | Creates empty note |
| `kbvault new "title" --open` | ❌ | Edits lost |
| `kbvault edit "title"` | ❌ | Can't find note |
| `kbvault delete "title"` | ❌ | Can't find note |
| `kbvault show <ID>` | ⚠️ | Placeholder only |
| `kbvault search "keyword"` | ✅ | Works correctly |
| `kbvault config show` | ✅ | Works |
| `kbvault profile list` | ✅ | Works |

## Recommended Resolution Path

### Phase 1: BUG-003 (Architecture Clarity)
**Time:** 1-2 hours
- Decide on single path source
- Remove redundant configuration
- Update all path references

### Phase 2: BUG-001 (Editor Save)
**Time:** 2-3 hours
- Implement post-edit file reading
- Ensure frontmatter + content preservation
- Test with multiple editors

### Phase 3: BUG-002 (Title Extraction)
**Time:** 2-3 hours
- Implement proper YAML frontmatter parsing
- Extract title from frontmatter (primary)
- Fall back to markdown heading (secondary)
- Update all find functions

### Phase 4: Testing & Verification
**Time:** 1-2 hours
- Test all workflows
- Create/edit/delete notes
- Verify persistence
- Check path consistency

## Files to Review/Modify

```
cmd/kbvault/
  ├── new.go          (BUG-001, BUG-003)
  ├── edit.go         (BUG-002, BUG-003)
  ├── delete.go       (BUG-002)
  └── show.go         (BUG-002)

pkg/storage/local/
  └── storage.go      (BUG-003)

.kbvault/
  └── config.toml     (BUG-003)

pkg/types/
  └── note.go         (Title/metadata parsing)
```

## Questions for Team Discussion

1. **Path Resolution**: Should we use vault config path or storage backend path?
2. **Frontmatter Format**: Should we use YAML (current) or TOML (config style)?
3. **Title Storage**: Where should note title be stored - frontmatter only or also as markdown heading?
4. **File Format**: Should notes have structured frontmatter or just markdown?
5. **Migration**: How to handle existing notes during fixes?

## Next Steps

1. ✅ Create detailed bug specs (DONE - see BUG-001, BUG-002, BUG-003)
2. 📋 Team reviews and discusses findings
3. 📋 Decide on path forward (fix all, partial fixes, redesign?)
4. 📋 Assign implementation tasks
5. 📋 Execute fixes in recommended order
6. 📋 Comprehensive testing

## Documentation Files Created

All specifications available in `.specify/specs/`:
- `BUG-001-editor-changes-lost.md` - Editor workflow broken
- `BUG-002-notes-not-findable.md` - Note discovery broken
- `BUG-003-inconsistent-note-storage.md` - Architecture issues
- `CRITICAL_ISSUES_SUMMARY.md` - This file

## Conclusion

The system has fundamental issues in note persistence and retrieval that make it non-functional for core use cases. These are not simple bugs but rather architectural/design issues that require careful planning and coordinated fixes.

The good news: All issues are **solvable** and the root causes are **well-understood**. The codebase needs refactoring in the note management layer (create/read/find) but the underlying storage and search systems appear sound.

**Recommendation:** Schedule a team sync to discuss the issues and agree on a resolution approach before implementation begins.
