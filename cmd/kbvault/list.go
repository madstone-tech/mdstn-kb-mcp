package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func newListCmd() *cobra.Command {
	var (
		format    string
		sortBy    string
		reverse   bool
		limit     int
		tags      []string
		showPaths bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all notes with metadata",
		Long: `List all notes in the vault with their metadata.
Supports filtering by tags and various sorting options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get profile-aware configuration
			config := getConfig()
			if config == nil {
				return fmt.Errorf("configuration not initialized")
			}

			// Initialize storage backend
			storage, err := storage.CreateStorage(config.Storage)
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			defer func() {
				if err := storage.Close(); err != nil {
					// Log error but don't fail the command
					fmt.Printf("Warning: failed to close storage: %v\n", err)
				}
			}()

			// List all notes
			notes, err := listAllNotes(storage)
			if err != nil {
				return fmt.Errorf("failed to list notes: %w", err)
			}

			// Filter by tags if specified
			if len(tags) > 0 {
				notes = filterNotesByTags(notes, tags)
			}

			// Sort notes
			sortNotes(notes, sortBy, reverse)

			// Apply limit
			if limit > 0 && len(notes) > limit {
				notes = notes[:limit]
			}

			// Display results
			switch format {
			case "json":
				return displayNotesJSON(notes)
			case "compact":
				return displayNotesCompact(notes, showPaths)
			default:
				return displayNotesDefault(notes, showPaths)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "default", "Output format (default, compact, json)")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "updated", "Sort by field (title, created, updated)")
	cmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "Reverse sort order")
	cmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of results (0 = no limit)")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by tags (comma-separated)")
	cmd.Flags().BoolVarP(&showPaths, "paths", "p", false, "Show file paths")

	return cmd
}

func listAllNotes(storage types.StorageBackend) ([]*types.Note, error) {
	ctx := context.Background()

	// Try to list files from common directories
	dirs := []string{"", "notes/", "daily/"}
	var allFiles []string

	for _, dir := range dirs {
		files, err := storage.List(ctx, dir)
		if err != nil {
			// Continue if directory doesn't exist
			continue
		}
		allFiles = append(allFiles, files...)
	}

	// Remove duplicates (in case files appear in multiple directories)
	fileSet := make(map[string]bool)
	var uniqueFiles []string
	for _, f := range allFiles {
		if !fileSet[f] {
			fileSet[f] = true
			uniqueFiles = append(uniqueFiles, f)
		}
	}

	var notes []*types.Note

	for _, file := range uniqueFiles {
		// Filter for markdown files only
		if !strings.HasSuffix(file, ".md") {
			continue
		}

		// Read and parse the note
		note, err := readAndParseNote(storage, file)
		if err != nil {
			// Skip files that can't be parsed
			// but don't fail the entire list command
			continue
		}

		notes = append(notes, note)
	}

	return notes, nil
}

// readAndParseNote reads a note file and extracts its metadata
func readAndParseNote(storage types.StorageBackend, filePath string) (*types.Note, error) {
	ctx := context.Background()

	// Read file content
	data, err := storage.Read(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	content := string(data)
	note := &types.Note{
		FilePath: filePath,
	}

	// Extract ID from filename (remove .md extension)
	base := filepath.Base(filePath)
	if strings.HasSuffix(base, ".md") {
		note.ID = strings.TrimSuffix(base, ".md")
	} else {
		note.ID = base
	}

	// Parse frontmatter and content to extract title and metadata
	parsedContent := parseNoteMetadata(content, note)

	// Fallback to ID if title is still empty
	if note.Title == "" {
		note.Title = note.ID
	}

	// Get file metadata for size and timestamps
	fileInfo, err := storage.Stat(ctx, filePath)
	if err == nil {
		note.Size = fileInfo.Size
		// Use ModTime for both created and updated if available
		if fileInfo.ModTime > 0 {
			modTime := time.Unix(0, fileInfo.ModTime*int64(time.Millisecond))
			note.UpdatedAt = modTime
			// If CreatedAt is zero, use UpdatedAt as default
			if note.CreatedAt.IsZero() {
				note.CreatedAt = modTime
			}
		}
	}

	// Set storage backend type
	note.StorageBackend = storage.Type()

	// Set content for the note
	note.Content = parsedContent

	return note, nil
}

// parseNoteMetadata extracts metadata from frontmatter and content
func parseNoteMetadata(content string, note *types.Note) string {
	// Check if content starts with YAML frontmatter marker
	if !strings.HasPrefix(content, "---") {
		// No frontmatter, extract title from markdown and use defaults
		note.Title = extractTitleFromMarkdown(content)
		note.Frontmatter = types.Frontmatter{
			ID:   note.ID,
			Type: "note",
			Tags: []string{},
		}
		return content
	}

	// Split by frontmatter markers
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		// Invalid frontmatter, treat as plain content
		note.Title = extractTitleFromMarkdown(content)
		note.Frontmatter = types.Frontmatter{
			ID:   note.ID,
			Type: "note",
			Tags: []string{},
		}
		return content
	}

	// parts[0] is empty (before first ---)
	// parts[1] is the frontmatter
	// parts[2] is the content
	frontmatter := parts[1]
	bodyContent := strings.TrimSpace(parts[2])

	// Parse frontmatter fields
	parseFrontmatterFields(frontmatter, note)

	// Extract title from frontmatter
	if note.Title == "" {
		note.Title = extractTitleFromFrontmatter(frontmatter)
	}

	// If no title in frontmatter, try markdown heading
	if note.Title == "" {
		note.Title = extractTitleFromMarkdown(bodyContent)
	}

	return bodyContent
}

// parseFrontmatterFields extracts structured fields from YAML frontmatter
func parseFrontmatterFields(frontmatter string, note *types.Note) {
	lines := strings.Split(frontmatter, "\n")
	fm := types.Frontmatter{
		ID:   note.ID,
		Type: "note",
		Tags: []string{},
	}

	// Track if we're in a tags array
	inTagsArray := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Check if this line starts a tags array
		if strings.HasPrefix(trimmedLine, "tags:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "tags:"))
			// If value is empty, tags are on following lines
			if value == "" {
				inTagsArray = true
				continue
			}
			// Handle inline tags: tags: [tag1, tag2]
			value = strings.TrimPrefix(value, "[")
			value = strings.TrimSuffix(value, "]")
			if value != "" {
				tagParts := strings.Split(value, ",")
				for _, tag := range tagParts {
					tag = strings.TrimSpace(tag)
					tag = strings.Trim(tag, "\"'")
					if tag != "" {
						fm.Tags = append(fm.Tags, tag)
					}
				}
			}
			continue
		}

		// Parse YAML list items (tags array)
		if inTagsArray {
			if strings.HasPrefix(trimmedLine, "- ") {
				tag := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "-"))
				tag = strings.Trim(tag, "\"'")
				if tag != "" {
					fm.Tags = append(fm.Tags, tag)
				}
			} else if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				// End of tags array (next non-indented line)
				inTagsArray = false
				// Re-process this line as a normal field
				i := i
				_ = i // avoid unused variable
				// Continue to normal processing below
			} else {
				// Might be a continuation, check if it's still a list item
				if !strings.HasPrefix(trimmedLine, "- ") {
					inTagsArray = false
				}
			}

			// Don't process this line further if we're still in tags array
			if inTagsArray && strings.HasPrefix(trimmedLine, "- ") {
				continue
			}
		}

		// Parse key: value pairs
		if !strings.Contains(trimmedLine, ":") {
			continue
		}

		parts := strings.SplitN(trimmedLine, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "id":
			value = strings.Trim(value, "\"'")
			if value != "" {
				fm.ID = value
			}

		case "title":
			value = strings.Trim(value, "\"'")
			note.Title = value
			fm.Title = value

		case "type":
			value = strings.Trim(value, "\"'")
			if value != "" {
				fm.Type = value
			}

		case "created":
			value = strings.Trim(value, "\"'")
			fm.Created = value
			// Try to parse as ISO timestamp
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				note.CreatedAt = t
			}

		case "updated":
			value = strings.Trim(value, "\"'")
			fm.Updated = value
			// Try to parse as ISO timestamp
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				note.UpdatedAt = t
			}

		case "storage":
			value = strings.Trim(value, "\"'")
			if value != "" {
				fm.Storage = value
			}

		case "template":
			value = strings.Trim(value, "\"'")
			if value != "" {
				fm.Template = value
			}
		}
	}

	note.Frontmatter = fm
}

func filterNotesByTags(notes []*types.Note, filterTags []string) []*types.Note {
	var filtered []*types.Note

	for _, note := range notes {
		if hasAnyTag(note.Frontmatter.Tags, filterTags) {
			filtered = append(filtered, note)
		}
	}

	return filtered
}

func hasAnyTag(noteTags, filterTags []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range noteTags {
		tagSet[strings.ToLower(tag)] = true
	}

	for _, filterTag := range filterTags {
		if tagSet[strings.ToLower(filterTag)] {
			return true
		}
	}

	return false
}

func sortNotes(notes []*types.Note, sortBy string, reverse bool) {
	sort.Slice(notes, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "title":
			less = strings.ToLower(notes[i].Title) < strings.ToLower(notes[j].Title)
		case "created":
			less = notes[i].CreatedAt.Before(notes[j].CreatedAt)
		case "updated":
			less = notes[i].UpdatedAt.Before(notes[j].UpdatedAt)
		default:
			// Default to updated time
			less = notes[i].UpdatedAt.Before(notes[j].UpdatedAt)
		}

		if reverse {
			return !less
		}
		return less
	})
}

func displayNotesDefault(notes []*types.Note, showPaths bool) error {
	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	fmt.Printf("Found %d note(s):\n\n", len(notes))

	for i, note := range notes {
		fmt.Printf("%d. 📝 %s\n", i+1, note.Title)
		fmt.Printf("   🆔 %s\n", note.ID)

		if len(note.Frontmatter.Tags) > 0 {
			fmt.Printf("   🏷️  %s\n", strings.Join(note.Frontmatter.Tags, ", "))
		}

		fmt.Printf("   📅 Updated: %s\n", formatRelativeTime(note.UpdatedAt))

		if showPaths {
			fmt.Printf("   📁 %s\n", note.FilePath)
		}

		if i < len(notes)-1 {
			fmt.Println()
		}
	}

	return nil
}

func displayNotesCompact(notes []*types.Note, showPaths bool) error {
	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	for _, note := range notes {
		line := fmt.Sprintf("%s | %s", note.ID[:8], note.Title)

		if len(note.Frontmatter.Tags) > 0 {
			line += fmt.Sprintf(" | %s", strings.Join(note.Frontmatter.Tags, ","))
		}

		if showPaths {
			line += fmt.Sprintf(" | %s", note.FilePath)
		}

		fmt.Println(line)
	}

	return nil
}

func displayNotesJSON(notes []*types.Note) error {
	fmt.Println("[")

	for i, note := range notes {
		fmt.Printf(`  {
    "id": "%s",
    "title": "%s",
    "file_path": "%s",
    "tags": [%s],
    "created": "%s",
    "updated": "%s"
  }`,
			note.ID,
			note.Title,
			note.FilePath,
			formatTagsJSON(note.Frontmatter.Tags),
			note.CreatedAt.Format("2006-01-02T15:04:05Z"),
			note.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		)

		if i < len(notes)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}

	fmt.Println("]")
	return nil
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d minute(s) ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hour(s) ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d day(s) ago", days)
	default:
		return t.Format("2006-01-02")
	}
}

func formatTagsJSON(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	quotedTags := make([]string, len(tags))
	for i, tag := range tags {
		quotedTags[i] = fmt.Sprintf(`"%s"`, tag)
	}
	return strings.Join(quotedTags, ", ")
}
