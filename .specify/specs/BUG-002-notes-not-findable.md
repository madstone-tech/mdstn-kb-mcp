# BUG-002: Notes Can't Be Found by edit/delete/show Commands

**Status:** Open  
**Severity:** 🔴 CRITICAL  
**Priority:** P0  
**Component:** cmd/kbvault/edit.go, cmd/kbvault/delete.go  
**Date Reported:** 2025-12-11  
**Depends On:** BUG-001

## Problem Statement

After creating a note with `kbvault new --open`:

1. Note appears in `kbvault search` results
2. But `kbvault edit`, `kbvault delete`, and `kbvault show` cannot find it
3. Commands return "no notes found matching" errors
4. Only `search` works because it uses a different code path

## Root Cause Analysis

### Issue 1: Missing/Incomplete Frontmatter

When BUG-001 occurs (editor changes lost), the file ends up with only user content:
```markdown
this is test 123, let's see if i can find it
```

Instead of proper format with frontmatter:
```yaml
---
id: 01KC7WAMJZPQB6960Q8NAK7ZT4
title: test-123
type: note
created: 2025-12-11T17:36:00Z
---

this is test 123, let's see if i can find it
```

### Issue 2: Title Extraction Logic Flawed

In `cmd/kbvault/edit.go` `readNote()` function:

```go
// Extract title from first # heading
lines := strings.Split(content, "\n")
for _, line := range lines {
    if strings.HasPrefix(line, "# ") {
        note.Title = strings.TrimPrefix(line, "# ")
        break
    }
}

if note.Title == "" {
    note.Title = note.ID  // Falls back to ID, not actual title
}
```

**Problem:** This extracts title from MARKDOWN heading (`# Title`), not from YAML frontmatter where it's actually stored.

When a note has no markdown heading:
- Title is extracted as empty string
- Falls back to using the ID (e.g., `01KC7WAMJZPQB6960Q8NAK7ZT4`)
- User search for "test-123" fails because the title is the ID, not the user's title

### Issue 3: Title Mismatch in Search

When `listAllNotesGeneric()` calls `readNote()`:
1. YAML frontmatter has: `title: test-123`
2. Markdown extraction gets: `` (empty)
3. Falls back to ID: `01KC7WAMJZPQB6960Q8NAK7ZT4`
4. `findNotesByTitle("test-123")` searches for this ID
5. No match found

But `search` command uses a different code path (full-text search engine) that:
- Indexes the entire file content
- Finds the user's text directly
- Doesn't depend on title parsing

## Impact

- ❌ `kbvault edit "test-123"` - Error: no notes found
- ❌ `kbvault delete "test-123"` - Error: no notes found
- ❌ `kbvault show "test-123"` - Can't work (requires title match)
- ⚠️ `kbvault show <ID>` - Returns placeholder, not actual content
- ✅ `kbvault search "test-123"` - Works (uses full-text search)

## Test Case

```bash
$ kbvault new "test-123" --open
# Edit and save in editor

$ kbvault search "test-123"
Found 1 results:  # ✅ Works

$ kbvault edit "test-123"
Error: no notes found matching 'test-123'  # ❌ Fails

$ kbvault delete "test-123"
Error: no notes found matching 'test-123'  # ❌ Fails

# Reason: Title extraction failed, stored as ID instead
$ grep -n "title:" notes/01KC7WAMJZPQB6960Q8NAK7ZT4.md
# Returns nothing - title wasn't saved
```

## Files Affected

- `cmd/kbvault/edit.go` - `readNote()` function (lines ~150-185)
- `cmd/kbvault/edit.go` - `findNotesByTitle()` function (lines ~60-80)
- `cmd/kbvault/edit.go` - `loadNoteByID()` function (lines ~100-120)
- `cmd/kbvault/delete.go` - `findNotesToDelete()` function

## Solution Requirements

1. **Parse YAML frontmatter properly** to extract title
2. **Use frontmatter title** as primary title source
3. **Fall back to markdown heading** only if frontmatter title missing
4. **Extract ID from frontmatter**, not just filename
5. **Store all metadata** (title, tags, timestamps) in frontmatter
6. **Make edit/delete/show use consistent finding logic**

## Proposed Fix

```go
func readNote(storage types.StorageBackend, path string) (*types.Note, error) {
    data, err := storage.Read(context.TODO(), path)
    if err != nil {
        return nil, err
    }

    content := string(data)
    note := &types.Note{
        Content:  content,
        FilePath: path,
    }

    // ← NEW: Parse YAML frontmatter
    // - Extract id from frontmatter
    // - Extract title from frontmatter (PRIMARY source)
    // - Extract tags from frontmatter
    // - Extract timestamps
    
    // ← Keep: Extract title from # heading (FALLBACK only)
    
    return note, nil
}
```

## Dependencies

- Must handle valid YAML parsing
- Must be resilient to missing frontmatter
- Must work with notes created before and after fix
- Must not break search functionality

## Testing Strategy

1. Create note with content
2. Verify file has proper frontmatter
3. Search for note by title ✓
4. Find note by title for edit ✓
5. Find note by title for delete ✓
6. Verify ID extraction works ✓
7. Test with notes missing frontmatter (migration case)

## Related Issues

- BUG-001: Editor changes lost (causes this issue)
- BUG-003: Inconsistent path handling
- BUG-004: Inconsistent note storage format

## Notes

This bug is a cascading failure from BUG-001. Once frontmatter is properly saved (BUG-001 fixed), this needs the title extraction logic fixed to properly parse YAML rather than looking for markdown headings.

The fact that `search` works proves the content is saved correctly - it's only the title extraction that fails.
