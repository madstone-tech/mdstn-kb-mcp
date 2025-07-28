package main

import (
	"fmt"
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
			defer storage.Close()

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
	// For now, this is simplified - in a full implementation we'd properly parse
	// markdown files and extract frontmatter
	fmt.Println("Note listing not yet implemented - this is a placeholder")
	return []*types.Note{}, nil
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
