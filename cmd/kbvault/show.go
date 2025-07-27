package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage/local"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/ulid"
)

func newShowCmd() *cobra.Command {
	var (
		showMetadata bool
		showContent  bool
		format       string
	)

	cmd := &cobra.Command{
		Use:   "show <note-id>",
		Short: "Display note content",
		Long: `Display the content of a note by its ULID identifier.
By default, shows both metadata and content.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noteID := args[0]

			// Validate ULID
			if !ulid.IsValid(noteID) {
				return fmt.Errorf("invalid note ID: %s", noteID)
			}

			// Load configuration
			config, err := loadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Initialize storage
			storage, err := local.New(config.Storage.Local)
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}

			// Find and load note
			note, err := findAndLoadNote(storage, noteID)
			if err != nil {
				return fmt.Errorf("failed to load note: %w", err)
			}

			// Display note based on format and options
			switch format {
			case "json":
				return displayNoteJSON(note)
			case "markdown":
				return displayNoteMarkdown(note, showMetadata, showContent)
			default:
				return displayNoteDefault(note, showMetadata, showContent)
			}
		},
	}

	cmd.Flags().BoolVarP(&showMetadata, "metadata", "m", true, "Show note metadata")
	cmd.Flags().BoolVarP(&showContent, "content", "c", true, "Show note content")
	cmd.Flags().StringVarP(&format, "format", "f", "default", "Output format (default, markdown, json)")

	return cmd
}

func findAndLoadNote(storage *local.Storage, noteID string) (*types.Note, error) {
	// For now, this is simplified - return a placeholder
	fmt.Printf("Show note not yet implemented - would show note ID: %s\n", noteID)
	return &types.Note{
		ID:    noteID,
		Title: "Placeholder Note",
	}, nil
}

func displayNoteDefault(note *types.Note, showMetadata, showContent bool) error {
	if showMetadata {
		fmt.Printf("📝 %s\n", note.Title)
		fmt.Printf("🆔 ID: %s\n", note.ID)
		fmt.Printf("📁 Path: %s\n", note.FilePath)
		fmt.Printf("🏷️  Tags: %s\n", strings.Join(note.Frontmatter.Tags, ", "))
		fmt.Printf("📅 Created: %s\n", note.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("📅 Updated: %s\n", note.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("💾 Storage: %s\n", note.StorageBackend)
		
		if showContent {
			fmt.Println("\n" + strings.Repeat("─", 50))
		}
	}

	if showContent {
		fmt.Println(note.Content)
	}

	return nil
}

func displayNoteMarkdown(note *types.Note, showMetadata, showContent bool) error {
	if showMetadata {
		fmt.Println("---")
		fmt.Printf("id: %s\n", note.ID)
		fmt.Printf("title: %s\n", note.Title)
		fmt.Printf("tags: [%s]\n", strings.Join(note.Frontmatter.Tags, ", "))
		fmt.Printf("created: %s\n", note.CreatedAt.Format("2006-01-02T15:04:05Z"))
		fmt.Printf("updated: %s\n", note.UpdatedAt.Format("2006-01-02T15:04:05Z"))
		fmt.Printf("storage: %s\n", note.StorageBackend)
		fmt.Println("---")
		
		if showContent {
			fmt.Println()
		}
	}

	if showContent {
		fmt.Println(note.Content)
	}

	return nil
}

func displayNoteJSON(note *types.Note) error {
	// This is a simplified JSON output - in production you'd use json.Marshal
	fmt.Printf(`{
  "id": "%s",
  "title": "%s",
  "content": %q,
  "file_path": "%s",
  "storage_backend": "%s",
  "frontmatter": {
    "id": "%s",
    "title": "%s",
    "tags": [%s],
    "created": "%s",
    "updated": "%s",
    "storage": "%s"
  }
}
`, 
		note.ID,
		note.Title,
		note.Content,
		note.FilePath,
		note.StorageBackend,
		note.Frontmatter.ID,
		note.Frontmatter.Title,
		formatTagsJSON(note.Frontmatter.Tags),
		note.CreatedAt.Format("2006-01-02T15:04:05Z"),
		note.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		note.Frontmatter.Storage,
	)
	
	return nil
}

func formatTagsJSON(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	
	quoted := make([]string, len(tags))
	for i, tag := range tags {
		quoted[i] = fmt.Sprintf(`"%s"`, tag)
	}
	
	return strings.Join(quoted, ", ")
}