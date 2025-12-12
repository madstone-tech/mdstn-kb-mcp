# BUG-001: Editor Changes Lost When Using `kbvault new --open`

**Status:** Open  
**Severity:** 🔴 CRITICAL  
**Priority:** P0  
**Component:** cmd/kbvault/new.go  
**Date Reported:** 2025-12-11  

## Problem Statement

When users run `kbvault new "Title" --open`, the following happens:

1. Note is created with empty content and YAML frontmatter
2. Note is saved to disk with frontmatter intact
3. Text editor opens showing the file
4. User edits the file in the editor
5. User saves and exits the editor
6. **BUG: The edited content is LOST**
7. File is overwritten with only the frontmatter, no user content

## Root Cause

In `cmd/kbvault/new.go` lines 77-80:

```go
// Open in editor if requested
if open {
    return openInEditor(note.FilePath)  // ← Returns immediately!
}
```

The `openInEditor()` function:
1. Opens the editor process
2. Waits for it to close
3. **Returns without reading the edited file back from disk**
4. The note object still has empty `Content`
5. When the function returns, nothing updates the persistent file

## Impact

- ❌ Users cannot write note content when using `--open` flag
- ❌ All edits made in the editor are discarded
- ❌ Users see success message but changes are lost
- ❌ Defeats the primary workflow: create note → edit immediately → save

## Expected Behavior

After the editor closes:
1. Read the file back from disk
2. Extract the edited content
3. Preserve the YAML frontmatter
4. Ensure both are saved together

## Test Case

```bash
$ cd /Users/andhi/code/code-notes
$ kbvault new "test-123" --open
✅ Created note: test-123
📝 ID: 01KC7WAMJZPQB6960Q8NAK7ZT4
Opening in editor: nvim notes/01KC7WAMJZPQB6960Q8NAK7ZT4.md

# In editor:
# Type: "this is test 123, let's see if i can find it"
# Save with :wq

$ cat notes/01KC7WAMJZPQB6960Q8NAK7ZT4.md
# Shows ONLY: "this is test 123, let's see if i can find it"
# Missing YAML frontmatter!
```

## Files Affected

- `cmd/kbvault/new.go` - `openInEditor()` function (lines 152-165)
- `cmd/kbvault/new.go` - `newNewCmd()` RunE (lines 77-80)
- `cmd/kbvault/new.go` - `saveNote()` function (lines 164-191)

## Solution Requirements

1. **After editor closes**, read the file from disk
2. **Parse the file** to extract content and frontmatter
3. **Preserve frontmatter** (ID, title, tags, timestamps)
4. **Merge edited content** with frontmatter
5. **Save updated note** back to disk
6. **Verify** file has both frontmatter and content

## Proposed Fix

```go
func openInEditor(filePath string) error {
    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = "nano"
    }

    cmd := exec.Command(editor, filePath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("editor failed: %w", err)
    }

    // ← NEW: Read file back and update note with edited content
    // ← NEW: Ensure frontmatter + content are preserved
    
    return nil
}
```

## Dependencies

- Must handle YAML frontmatter parsing
- Must preserve metadata (ID, timestamps, tags)
- Must work with different editors
- Must handle both markdown content with headers and plain text

## Testing Strategy

1. Create note with `--open`
2. Edit in editor
3. Verify file has frontmatter + content
4. Verify `kbvault search` finds the note
5. Verify `kbvault edit` can open it again
6. Verify `kbvault show` displays correct info
7. Test with multiple editors (vim, nano, emacs)

## Related Issues

- BUG-002: Notes can't be found by edit/delete
- BUG-003: Inconsistent path handling

## Notes

This is blocking the primary use case of creating and immediately editing notes. The fix is straightforward but critical for functionality.
