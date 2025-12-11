package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/ulid"
)

func newNewCmd() *cobra.Command {
	var (
		title    string
		template string
		tags     []string
		open     bool
	)

	cmd := &cobra.Command{
		Use:   "new [title]",
		Short: "Create a new note",
		Long: `Create a new note with a unique ULID identifier.
The note will be created in the vault's notes directory.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use profile-aware configuration
			config := getConfig()
			if config == nil {
				return fmt.Errorf("configuration not initialized")
			}

			// Determine title
			if len(args) > 0 {
				title = args[0]
			}
			if title == "" {
				title = "Untitled Note"
			}

			// Create storage backend
			storageBackend, err := storage.CreateStorage(config.Storage)
			if err != nil {
				return fmt.Errorf("failed to create storage backend: %w", err)
			}
			defer func() {
				if err := storageBackend.Close(); err != nil {
					// Log error but don't fail the command
					fmt.Printf("Warning: failed to close storage: %v\n", err)
				}
			}()

			// Create new note
			note, err := createNewNote(config, title, template, tags)
			if err != nil {
				return fmt.Errorf("failed to create note: %w", err)
			}

			// Save the note to storage
			ctx := context.Background()
			if err := saveNote(ctx, storageBackend, note); err != nil {
				return fmt.Errorf("failed to save note: %w", err)
			}

			fmt.Printf("✅ Created note: %s\n", note.Title)
			fmt.Printf("📝 ID: %s\n", note.ID)
			fmt.Printf("📁 Path: %s\n", note.FilePath)
			fmt.Printf("💾 Storage: %s\n", config.Storage.Type)

			// Open in editor if requested
			if open {
				// Open editor and get modified content
				editedContent, err := openInEditorAndRead(note.FilePath)
				if err != nil {
					return fmt.Errorf("failed to edit note: %w", err)
				}

				// If user made changes, update the note and resave with frontmatter
				if editedContent != "" {
					note.Content = editedContent
					note.Frontmatter.Updated = time.Now().Format("2006-01-02T15:04:05Z")
					note.UpdatedAt = time.Now()

					// Resave with updated frontmatter
					if err := saveNote(ctx, storageBackend, note); err != nil {
						return fmt.Errorf("failed to save edited note: %w", err)
					}
					fmt.Printf("✅ Note updated with your edits\n")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "Title for the new note")
	cmd.Flags().StringVar(&template, "template", "default", "Template to use for the note")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Tags for the note (comma-separated)")
	cmd.Flags().BoolVarP(&open, "open", "o", false, "Open the note in default editor after creation")

	return cmd
}

func createNewNote(config *types.Config, title, template string, tags []string) (*types.Note, error) {
	// Generate ULID
	id := ulid.New()

	// Create filename
	filename := ulid.ToFilename(id)
	filePath := filepath.Join(config.Vault.NotesDir, filename)

	// Create note structure
	note := &types.Note{
		ID:             id,
		Title:          title,
		Content:        "",
		FilePath:       filePath,
		StorageBackend: config.Storage.Type,
		Frontmatter: types.Frontmatter{
			ID:      id,
			Title:   title,
			Tags:    tags,
			Type:    "note",
			Storage: string(config.Storage.Type),
			Created: time.Now().Format("2006-01-02T15:04:05Z"),
			Updated: time.Now().Format("2006-01-02T15:04:05Z"),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Apply template if specified
	if template != "default" {
		templateContent, err := loadTemplate(config, template)
		if err != nil {
			return nil, fmt.Errorf("failed to load template %s: %w", template, err)
		}
		note.Content = templateContent
	} else {
		// Use default content
		note.Content = fmt.Sprintf("# %s\n\nContent goes here...\n", title)
	}

	return note, nil
}

func loadTemplate(config *types.Config, templateName string) (string, error) {
	templatePath := filepath.Join("templates", templateName+".md")

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template %s not found at %s", templateName, templatePath)
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	return string(content), nil
}

func openInEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // fallback editor
	}

	fmt.Printf("Opening in editor: %s %s\n", editor, filePath)

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// openInEditorAndRead opens a file in the editor and returns the edited content
func openInEditorAndRead(filePath string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // fallback editor
	}

	fmt.Printf("Opening in editor: %s %s\n", editor, filePath)

	// Open the file in the editor
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read the edited file content (excluding frontmatter)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited file: %w", err)
	}

	// Extract content after frontmatter
	content := string(data)
	// If file starts with ---, skip the frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			// Return content after the closing ---
			return strings.TrimSpace(parts[2]), nil
		}
	}

	// Return all content if no frontmatter
	return strings.TrimSpace(content), nil
}

func saveNote(ctx context.Context, storage types.StorageBackend, note *types.Note) error {
	// Format note content with frontmatter
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")
	buf.WriteString(fmt.Sprintf("id: %s\n", note.ID))
	buf.WriteString(fmt.Sprintf("title: %s\n", note.Title))
	if len(note.Frontmatter.Tags) > 0 {
		buf.WriteString("tags:\n")
		for _, tag := range note.Frontmatter.Tags {
			buf.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}
	buf.WriteString(fmt.Sprintf("type: %s\n", note.Frontmatter.Type))
	buf.WriteString(fmt.Sprintf("storage: %s\n", note.Frontmatter.Storage))
	buf.WriteString(fmt.Sprintf("created: %s\n", note.Frontmatter.Created))
	buf.WriteString(fmt.Sprintf("updated: %s\n", note.Frontmatter.Updated))
	buf.WriteString("---\n\n")

	// Write content
	buf.WriteString(note.Content)

	// Save to storage
	return storage.Write(ctx, note.FilePath, buf.Bytes())
}

func findVaultRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		configPath := filepath.Join(dir, ".kbvault", "config.toml")
		if _, err := os.Stat(configPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no .kbvault directory found")
}
