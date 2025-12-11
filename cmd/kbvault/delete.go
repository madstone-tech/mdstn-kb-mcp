package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var (
		force       bool
		dryRun      bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "delete [note-id-or-title]",
		Short: "Delete a note from the vault",
		Long: `Delete a note from the vault by ID or title.

This command will permanently delete the note file. Use with caution!

Examples:
  # Delete by note ID
  kbvault delete note-123

  # Delete by title (interactive selection if multiple matches)
  kbvault delete "old meeting notes"

  # Force delete without confirmation
  kbvault delete note-123 --force

  # Dry run to see what would be deleted
  kbvault delete "temp*" --dry-run

  # Interactive mode for safer deletion
  kbvault delete note-123 --interactive`,
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

			// Find notes to delete
			notes, err := findNotesToDelete(storageBackend, query)
			if err != nil {
				return err
			}

			if len(notes) == 0 {
				fmt.Printf("No notes found matching '%s'\n", query)
				return nil
			}

			// Show what will be deleted
			if err := showDeletionPlan(notes, dryRun); err != nil {
				return err
			}

			if dryRun {
				return nil
			}

			// Confirm deletion
			if !force && !confirmDeletion(notes, interactive) {
				fmt.Println("Deletion cancelled.")
				return nil
			}

			// Perform deletion
			return deleteNotes(storageBackend, notes)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Delete without confirmation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without actually deleting")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive confirmation for each note")

	return cmd
}

// findNotesToDelete finds notes matching the query pattern
func findNotesToDelete(storage types.StorageBackend, query string) ([]*types.Note, error) {
	// Check if query has wildcards
	if strings.Contains(query, "*") {
		return findNotesByPattern(storage, query)
	}

	// Find single note by ID or title
	note, err := findNoteByQuery(storage, query)
	if err != nil {
		// Try partial matches
		matches, err := findNotesByTitle(storage, query)
		if err != nil {
			return nil, err
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("no notes found matching '%s'", query)
		}

		if len(matches) == 1 {
			return matches, nil
		}

		// Multiple matches - let user choose
		selected, err := selectFromMultipleNotes(matches, query)
		if err != nil {
			return nil, err
		}

		return []*types.Note{selected}, nil
	}

	return []*types.Note{note}, nil
}

// findNotesByPattern finds notes matching a wildcard pattern
func findNotesByPattern(storage types.StorageBackend, pattern string) ([]*types.Note, error) {
	allNotes, err := listAllNotesGeneric(storage)
	if err != nil {
		return nil, err
	}

	var matches []*types.Note

	// Convert simple wildcard pattern to Go-compatible
	// For now, just support suffix matching with *
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		for _, metadata := range allNotes {
			if strings.HasSuffix(metadata.Title, suffix) || strings.HasSuffix(metadata.ID, suffix) {
				note, err := loadNoteByID(storage, metadata.ID)
				if err == nil {
					matches = append(matches, note)
				}
			}
		}
	} else if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		for _, metadata := range allNotes {
			if strings.HasPrefix(metadata.Title, prefix) || strings.HasPrefix(metadata.ID, prefix) {
				note, err := loadNoteByID(storage, metadata.ID)
				if err == nil {
					matches = append(matches, note)
				}
			}
		}
	} else {
		// Middle wildcard or complex patterns - simple contains match
		searchTerm := strings.ReplaceAll(pattern, "*", "")
		for _, metadata := range allNotes {
			if strings.Contains(metadata.Title, searchTerm) || strings.Contains(metadata.ID, searchTerm) {
				note, err := loadNoteByID(storage, metadata.ID)
				if err == nil {
					matches = append(matches, note)
				}
			}
		}
	}

	return matches, nil
}

// showDeletionPlan displays what will be deleted
func showDeletionPlan(notes []*types.Note, dryRun bool) error {
	if dryRun {
		fmt.Printf("DRY RUN: The following %d note(s) would be deleted:\n\n", len(notes))
	} else {
		fmt.Printf("The following %d note(s) will be deleted:\n\n", len(notes))
	}

	for i, note := range notes {
		fmt.Printf("%d. %s\n", i+1, note.Title)
		fmt.Printf("   ID: %s\n", note.ID)
		fmt.Printf("   Path: %s\n", note.FilePath)

		if len(note.Frontmatter.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(note.Frontmatter.Tags, ", "))
		}

		// Show snippet of content
		contentLines := strings.Split(note.Content, "\n")
		var snippet string
		for _, line := range contentLines {
			if line != "" && !strings.HasPrefix(line, "#") {
				snippet = line
				break
			}
		}
		if snippet != "" {
			if len(snippet) > 60 {
				snippet = snippet[:60] + "..."
			}
			fmt.Printf("   Content: %s\n", snippet)
		}

		fmt.Println()
	}

	return nil
}

// confirmDeletion asks for user confirmation
func confirmDeletion(notes []*types.Note, interactive bool) bool {
	if interactive {
		return confirmInteractive(notes)
	}

	if len(notes) == 1 {
		fmt.Printf("Are you sure you want to delete '%s'? (y/N): ", notes[0].Title)
	} else {
		fmt.Printf("Are you sure you want to delete these %d notes? (y/N): ", len(notes))
	}

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// confirmInteractive asks for confirmation for each note individually
func confirmInteractive(notes []*types.Note) bool {
	confirmedNotes := 0

	for i, note := range notes {
		fmt.Printf("\n--- Note %d of %d ---\n", i+1, len(notes))
		fmt.Printf("Title: %s\n", note.Title)
		fmt.Printf("ID: %s\n", note.ID)
		fmt.Printf("Path: %s\n", note.FilePath)

		fmt.Print("Delete this note? (y/N/q): ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			continue
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "q" || response == "quit" {
			fmt.Println("Deletion cancelled.")
			return false
		}

		if response == "y" || response == "yes" {
			confirmedNotes++
		}
	}

	if confirmedNotes == 0 {
		fmt.Println("No notes selected for deletion.")
		return false
	}

	fmt.Printf("\n%d notes will be deleted. Proceed? (y/N): ", confirmedNotes)

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// deleteNotes performs the actual deletion
func deleteNotes(storage types.StorageBackend, notes []*types.Note) error {
	var errors []string
	deletedCount := 0

	for _, note := range notes {
		fmt.Printf("Deleting '%s'...", note.Title)

		if err := storage.Delete(context.TODO(), note.FilePath); err != nil {
			fmt.Printf(" FAILED: %v\n", err)
			errors = append(errors, fmt.Sprintf("%s: %v", note.Title, err))
		} else {
			fmt.Printf(" OK\n")
			deletedCount++
		}
	}

	fmt.Printf("\nDeleted %d of %d notes.\n", deletedCount, len(notes))

	if len(errors) > 0 {
		fmt.Printf("\nErrors occurred:\n")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("some deletions failed")
	}

	return nil
}
