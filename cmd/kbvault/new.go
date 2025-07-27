package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/config"
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
			// Load configuration
			config, err := loadConfig()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Determine title
			if len(args) > 0 {
				title = args[0]
			}
			if title == "" {
				title = "Untitled Note"
			}

			// Create new note
			note, err := createNewNote(config, title, template, tags)
			if err != nil {
				return fmt.Errorf("failed to create note: %w", err)
			}

			fmt.Printf("✅ Created note: %s\n", note.Title)
			fmt.Printf("📝 ID: %s\n", note.ID)
			fmt.Printf("📁 Path: %s\n", note.FilePath)
			fmt.Println("Note creation not yet fully implemented - this is a placeholder")

			// Open in editor if requested
			if open {
				return openInEditor(note.FilePath)
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
	filePath := filepath.Join("notes", filename)

	// Create note structure
	note := &types.Note{
		ID:             id,
		Title:          title,
		Content:        "",
		FilePath:       filePath,
		StorageBackend: types.StorageTypeLocal,
		Frontmatter: types.Frontmatter{
			ID:      id,
			Title:   title,
			Tags:    tags,
			Type:    "note",
			Storage: "local",
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

	cmd := fmt.Sprintf("%s %s", editor, filePath)
	fmt.Printf("Opening in editor: %s\n", cmd)
	
	// This is a simplified version - in production you'd use exec.Command
	fmt.Printf("💡 Run: %s\n", cmd)
	
	return nil
}

func loadConfig() (*types.Config, error) {
	// Try to find vault root
	vaultRoot, err := findVaultRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a kbvault directory: %w", err)
	}

	// Load configuration
	manager := config.NewManager()
	configPath := filepath.Join(vaultRoot, ".kbvault", "config.toml")
	
	return manager.LoadFromFile(configPath)
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