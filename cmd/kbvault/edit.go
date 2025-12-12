package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	var (
		editor    string
		createNew bool
	)

	cmd := &cobra.Command{
		Use:   "edit [note-id-or-title]",
		Short: "Edit an existing note",
		Long: `Edit an existing note in your configured editor.

You can specify the note by its ID or title. If multiple notes match
the title, you'll be prompted to choose.

Examples:
  # Edit by note ID
  kbvault edit note-123

  # Edit by partial title
  kbvault edit "meeting notes"

  # Use specific editor
  kbvault edit note-123 --editor vim

  # Create new note if not found
  kbvault edit "new topic" --create`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile-aware configuration
			cfg := getConfig()
			if cfg == nil {
				return fmt.Errorf("configuration not initialized")
			}

			// Initialize storage backend
			storageBackend, err := storage.CreateStorage(cfg.Storage)
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			defer func() {
				if closeErr := storageBackend.Close(); closeErr != nil {
					// Log error but don't fail the command
					fmt.Printf("Warning: failed to close storage: %v\n", closeErr)
				}
			}()

			query := args[0]

			// Find the note
			note, err := findNoteByQuery(storageBackend, query)
			if err != nil {
				if createNew {
					return createAndEditNote(storageBackend, query, editor)
				}
				return fmt.Errorf("note not found: %w", err)
			}

			// Edit the note
			return editNote(storageBackend, note, editor)
		},
	}

	cmd.Flags().StringVar(&editor, "editor", "", "Editor to use (overrides EDITOR env var)")
	cmd.Flags().BoolVar(&createNew, "create", false, "Create new note if not found")

	return cmd
}

// findNoteByQuery searches for a note by ID or title
func findNoteByQuery(storage types.StorageBackend, query string) (*types.Note, error) {
	// First try as exact ID
	if note, err := loadNoteByID(storage, query); err == nil {
		return note, nil
	}

	// Then search by title
	notes, err := findNotesByTitle(storage, query)
	if err != nil {
		return nil, err
	}

	if len(notes) == 0 {
		return nil, fmt.Errorf("no notes found matching '%s'", query)
	}

	if len(notes) == 1 {
		return notes[0], nil
	}

	// Multiple matches - let user choose
	return selectFromMultipleNotes(notes, query)
}

// findNotesByTitle searches for notes with matching titles
func findNotesByTitle(storage types.StorageBackend, titleQuery string) ([]*types.Note, error) {
	allNotes, err := listAllNotesGeneric(storage)
	if err != nil {
		return nil, err
	}

	var matches []*types.Note
	titleQuery = strings.ToLower(titleQuery)

	for _, metadata := range allNotes {
		// Check for exact match first
		if strings.ToLower(metadata.Title) == titleQuery {
			note, err := loadNoteByID(storage, metadata.ID)
			if err == nil {
				return []*types.Note{note}, nil
			}
		}

		// Check for partial match
		if strings.Contains(strings.ToLower(metadata.Title), titleQuery) {
			note, err := loadNoteByID(storage, metadata.ID)
			if err == nil {
				matches = append(matches, note)
			}
		}
	}

	return matches, nil
}

// selectFromMultipleNotes prompts user to choose from multiple matches
func selectFromMultipleNotes(notes []*types.Note, query string) (*types.Note, error) {
	fmt.Printf("Multiple notes found matching '%s':\n\n", query)

	for i, note := range notes {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, note.Title, note.ID)
		if len(note.Frontmatter.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(note.Frontmatter.Tags, ", "))
		}
		fmt.Printf("   Path: %s\n", note.FilePath)
		fmt.Println()
	}

	fmt.Print("Select note number (1-" + fmt.Sprintf("%d", len(notes)) + "): ")

	var choice int
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return nil, fmt.Errorf("invalid selection")
	}

	if choice < 1 || choice > len(notes) {
		return nil, fmt.Errorf("invalid selection: must be between 1 and %d", len(notes))
	}

	return notes[choice-1], nil
}

// loadNoteByID loads a complete note by its ID
func loadNoteByID(storage types.StorageBackend, noteID string) (*types.Note, error) {
	// Try common file extensions and patterns
	possiblePaths := []string{
		noteID + ".md",
		noteID,
		"notes/" + noteID + ".md",
		"daily/" + noteID + ".md",
	}

	for _, path := range possiblePaths {
		if note, err := readNote(storage, path); err == nil {
			return note, nil
		}
	}

	return nil, fmt.Errorf("note not found: %s", noteID)
}

// readNote reads and parses a note from storage
func readNote(storage types.StorageBackend, path string) (*types.Note, error) {
	data, err := storage.Read(context.TODO(), path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	note := &types.Note{
		FilePath: path,
	}

	// Extract ID from filename
	base := filepath.Base(path)
	if strings.HasSuffix(base, ".md") {
		note.ID = strings.TrimSuffix(base, ".md")
	} else {
		note.ID = base
	}

	// Parse frontmatter and content
	title, noteContent := parseFrontmatterAndContent(content)
	note.Title = title
	note.Content = noteContent

	// Fallback to ID if title is empty
	if note.Title == "" {
		note.Title = note.ID
	}

	return note, nil
}

// parseFrontmatterAndContent extracts title from YAML frontmatter and returns content
func parseFrontmatterAndContent(content string) (string, string) {
	// Check if content starts with YAML frontmatter marker
	if !strings.HasPrefix(content, "---") {
		// No frontmatter, extract title from first heading
		return extractTitleFromMarkdown(content), content
	}

	// Split by frontmatter markers
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		// Invalid frontmatter, treat as plain content
		return extractTitleFromMarkdown(content), content
	}

	// parts[0] is empty (before first ---)
	// parts[1] is the frontmatter
	// parts[2] is the content
	frontmatter := parts[1]
	bodyContent := strings.TrimSpace(parts[2])

	// Extract title from frontmatter
	title := extractTitleFromFrontmatter(frontmatter)

	// If no title in frontmatter, try markdown heading
	if title == "" {
		title = extractTitleFromMarkdown(bodyContent)
	}

	return title, bodyContent
}

// extractTitleFromFrontmatter extracts title from YAML frontmatter
func extractTitleFromFrontmatter(frontmatter string) string {
	lines := strings.Split(frontmatter, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			// Extract value after "title:"
			title := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			// Remove quotes if present
			title = strings.Trim(title, "\"'")
			if title != "" {
				return title
			}
		}
	}
	return ""
}

// extractTitleFromMarkdown extracts title from first markdown heading
func extractTitleFromMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// editNote opens a note in the configured editor
func editNote(storage types.StorageBackend, note *types.Note, editorOverride string) error {
	// Create temporary file for editing
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "kbvault-edit-"+note.ID+".md")

	// Write current content to temp file
	if err := os.WriteFile(tempFile, []byte(note.Content), 0644); err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Clean up temp file when done
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			fmt.Printf("Warning: failed to clean up temp file: %v\n", err)
		}
	}()

	// Get file modification time before editing
	stat, err := os.Stat(tempFile)
	if err != nil {
		return fmt.Errorf("failed to stat temp file: %w", err)
	}
	originalModTime := stat.ModTime()

	// Open in editor
	if err := openInEditorWithOverride(tempFile, editorOverride); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Check if file was modified
	stat, err = os.Stat(tempFile)
	if err != nil {
		return fmt.Errorf("failed to stat temp file after editing: %w", err)
	}

	if stat.ModTime().Equal(originalModTime) {
		fmt.Println("No changes made.")
		return nil
	}

	// Read modified content
	modifiedContent, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to read modified content: %w", err)
	}

	// Write back to storage
	if err := storage.Write(context.TODO(), note.FilePath, modifiedContent); err != nil {
		return fmt.Errorf("failed to save changes: %w", err)
	}

	fmt.Printf("Note '%s' updated successfully.\n", note.Title)
	return nil
}

// createAndEditNote creates a new note and opens it for editing
func createAndEditNote(storage types.StorageBackend, title, editorOverride string) error {
	// Generate note ID from title
	noteID := generateNoteID(title)
	filePath := noteID + ".md"

	// Create initial content
	content := fmt.Sprintf("# %s\n\n", title)

	// Create temporary file for editing
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "kbvault-new-"+noteID+".md")

	// Write initial content to temp file
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Clean up temp file when done
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			fmt.Printf("Warning: failed to clean up temp file: %v\n", err)
		}
	}()

	// Open in editor
	if err := openInEditorWithOverride(tempFile, editorOverride); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Read final content
	finalContent, err := os.ReadFile(tempFile)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}

	// Write to storage
	if err := storage.Write(context.TODO(), filePath, finalContent); err != nil {
		return fmt.Errorf("failed to save new note: %w", err)
	}

	fmt.Printf("New note '%s' created successfully at %s\n", title, filePath)
	return nil
}

// openInEditorWithOverride opens a file in the specified editor
func openInEditorWithOverride(filePath, editorOverride string) error {
	editor := editorOverride
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "nano" // Default fallback
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// generateNoteID creates a URL-safe ID from a title
func generateNoteID(title string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	id := strings.ToLower(title)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Clean up multiple consecutive hyphens
	cleaned := strings.ReplaceAll(result.String(), "--", "-")
	cleaned = strings.Trim(cleaned, "-")

	// Ensure it's not empty
	if cleaned == "" {
		cleaned = "note"
	}

	return cleaned
}

// listAllNotesGeneric lists all notes using the generic storage interface
func listAllNotesGeneric(storage types.StorageBackend) ([]*types.NoteMetadata, error) {
	files, err := storage.List(context.TODO(), "")
	if err != nil {
		return nil, err
	}

	var notes []*types.NoteMetadata
	for _, file := range files {
		if !strings.HasSuffix(file, ".md") {
			continue
		}

		note, err := readNote(storage, file)
		if err != nil {
			continue // Skip files that can't be read
		}

		metadata := note.ToMetadata()
		notes = append(notes, &metadata)
	}

	return notes, nil
}
