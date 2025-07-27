package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/internal/search"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage/local"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var (
		tags       []string
		noteType   string
		sortBy     string
		sortDesc   bool
		limit      int
		offset     int
		fields     []string
		outputJSON bool
		detailed   bool
		buildIndex bool
		after      string
		before     string
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search notes using full-text search",
		Long: `Search through your notes using full-text search with various filters.

Examples:
  # Simple text search
  kbvault search "golang concurrency"
  
  # Search with tag filter
  kbvault search "patterns" --tag golang --tag advanced
  
  # Search by type
  kbvault search --type daily
  
  # Search with date range
  kbvault search "meeting" --after 2024-01-01 --before 2024-12-31
  
  # Search in specific fields
  kbvault search "TODO" --field content
  
  # JSON output with pagination
  kbvault search "api" --json --limit 10 --offset 20
  
  # Build/rebuild search index
  kbvault search --build-index`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Find vault root for proper storage initialization
			vaultRoot, err := findVaultRoot()
			if err != nil {
				return fmt.Errorf("not in a kbvault directory: %w", err)
			}

			// Initialize storage with vault root as base path
			storageConfig := cfg.Storage.Local
			storageConfig.Path = vaultRoot // Use vault root instead of relative path
			storage, err := local.New(storageConfig)
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			defer storage.Close()

			// Create search engine
			searchOpts := search.DefaultOptions()
			searchOpts.MaxResults = limit
			engine := search.New(storage, searchOpts)

			ctx := context.Background()

			// Handle index building
			if buildIndex {
				fmt.Fprintln(cmd.OutOrStdout(), "Building search index...")
				if err := engine.BuildIndex(ctx); err != nil {
					return fmt.Errorf("failed to build index: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Search index built successfully")
				return nil
			}

			// Build search query
			query := search.SearchQuery{
				Query:    strings.Join(args, " "),
				Tags:     tags,
				Type:     noteType,
				Fields:   fields,
				SortBy:   sortBy,
				SortDesc: sortDesc,
				Limit:    limit,
				Offset:   offset,
			}

			// Parse date range if provided
			if after != "" || before != "" {
				dateRange := &search.DateRange{}
				
				if after != "" {
					t, err := time.Parse("2006-01-02", after)
					if err != nil {
						return fmt.Errorf("invalid after date: %w", err)
					}
					dateRange.After = t
				}
				
				if before != "" {
					t, err := time.Parse("2006-01-02", before)
					if err != nil {
						return fmt.Errorf("invalid before date: %w", err)
					}
					dateRange.Before = t
				}
				
				query.DateRange = dateRange
			}

			// First, build the index (in production, this would be cached)
			if err := engine.BuildIndex(ctx); err != nil {
				return fmt.Errorf("failed to build index: %w", err)
			}

			// Perform search
			results, err := engine.Search(ctx, query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			// Output results
			if outputJSON {
				return outputSearchJSON(cmd.OutOrStdout(), results)
			}

			if detailed {
				return outputSearchDetailed(cmd.OutOrStdout(), results)
			}

			return outputSearchList(cmd.OutOrStdout(), results)
		},
	}

	// Add flags
	cmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "Filter by tags (AND operation)")
	cmd.Flags().StringVar(&noteType, "type", "", "Filter by note type")
	cmd.Flags().StringVar(&sortBy, "sort", "relevance", "Sort results by: relevance, created, updated, title")
	cmd.Flags().BoolVar(&sortDesc, "desc", false, "Sort in descending order")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of results to skip")
	cmd.Flags().StringSliceVarP(&fields, "field", "f", nil, "Fields to search in: title, content, tags, all")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed results with snippets")
	cmd.Flags().BoolVar(&buildIndex, "build-index", false, "Build or rebuild the search index")
	cmd.Flags().StringVar(&after, "after", "", "Only show notes created after this date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&before, "before", "", "Only show notes created before this date (YYYY-MM-DD)")

	return cmd
}

func outputSearchList(w io.Writer, results []search.SearchResult) error {
	if len(results) == 0 {
		fmt.Fprintln(w, "No results found")
		return nil
	}

	fmt.Fprintf(w, "Found %d results:\n\n", len(results))

	for i, result := range results {
		fmt.Fprintf(w, "%d. %s\n", i+1, result.Note.Title)
		fmt.Fprintf(w, "   ID: %s\n", result.Note.ID)
		
		if len(result.Note.Tags) > 0 {
			fmt.Fprintf(w, "   Tags: %s\n", strings.Join(result.Note.Tags, ", "))
		}
		
		fmt.Fprintf(w, "   Score: %.2f\n", result.Score)
		fmt.Fprintf(w, "   Updated: %s\n", result.Note.UpdatedAt.Format("2006-01-02 15:04"))
		fmt.Fprintln(w)
	}

	return nil
}

func outputSearchDetailed(w io.Writer, results []search.SearchResult) error {
	if len(results) == 0 {
		fmt.Fprintln(w, "No results found")
		return nil
	}

	fmt.Fprintf(w, "Found %d results:\n\n", len(results))

	for i, result := range results {
		fmt.Fprintf(w, "=== Result %d ===\n", i+1)
		fmt.Fprintf(w, "Title: %s\n", result.Note.Title)
		fmt.Fprintf(w, "ID: %s\n", result.Note.ID)
		fmt.Fprintf(w, "Path: %s\n", result.Note.FilePath)
		fmt.Fprintf(w, "Type: %s\n", result.Note.Type)
		
		if len(result.Note.Tags) > 0 {
			fmt.Fprintf(w, "Tags: %s\n", strings.Join(result.Note.Tags, ", "))
		}
		
		fmt.Fprintf(w, "Score: %.2f\n", result.Score)
		fmt.Fprintf(w, "Created: %s\n", result.Note.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(w, "Updated: %s\n", result.Note.UpdatedAt.Format("2006-01-02 15:04:05"))
		
		if result.Snippet != "" {
			fmt.Fprintf(w, "\nSnippet:\n%s\n", result.Snippet)
		}
		
		if len(result.Matches) > 0 {
			fmt.Fprintf(w, "\nMatches:\n")
			for _, match := range result.Matches {
				fmt.Fprintf(w, "  - %s: %s\n", match.Field, match.Context)
			}
		}
		
		fmt.Fprintln(w)
	}

	return nil
}

func outputSearchJSON(w io.Writer, results []search.SearchResult) error {
	output := struct {
		Count   int                   `json:"count"`
		Results []search.SearchResult `json:"results"`
	}{
		Count:   len(results),
		Results: results,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// Helper function to read note for searching
func readNoteForSearch(storage types.StorageBackend, path string) (*types.Note, error) {
	ctx := context.Background()
	
	// Read file content
	data, err := storage.Read(ctx, path)
	if err != nil {
		return nil, err
	}

	// Parse frontmatter and content
	content := string(data)
	note := &types.Note{
		FilePath: path,
		Content:  content,
	}

	// Extract ID from filename
	base := filepath.Base(path)
	if strings.HasSuffix(base, ".md") {
		note.ID = strings.TrimSuffix(base, ".md")
	}

	// Get file info
	info, err := storage.Stat(ctx, path)
	if err == nil {
		note.Size = info.Size
		note.CreatedAt = time.Unix(info.ModTime, 0)
		note.UpdatedAt = time.Unix(info.ModTime, 0)
	}

	// Simple title extraction (first # heading or filename)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			note.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}
	if note.Title == "" {
		note.Title = strings.TrimSuffix(filepath.Base(path), ".md")
	}

	return note, nil
}