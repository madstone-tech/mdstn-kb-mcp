package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/internal/search"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/storage"
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
					// Log error but don't fail the command (ignore write errors)
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to close storage: %v\n", closeErr)
				}
			}()

			// Create search engine
			searchOpts := search.DefaultOptions()
			searchOpts.MaxResults = limit
			engine := search.New(storageBackend, searchOpts)

			ctx := context.Background()

			// Handle index building
			if buildIndex {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Building search index..."); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
				if err := engine.BuildIndex(ctx); err != nil {
					return fmt.Errorf("failed to build index: %w", err)
				}
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Search index built successfully"); err != nil {
					return fmt.Errorf("failed to write output: %w", err)
				}
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
		if _, err := fmt.Fprintln(w, "No results found"); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "Found %d results:\n\n", len(results)); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	for i, result := range results {
		if _, err := fmt.Fprintf(w, "%d. %s\n", i+1, result.Note.Title); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		if _, err := fmt.Fprintf(w, "   ID: %s\n", result.Note.ID); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		
		if len(result.Note.Tags) > 0 {
			if _, err := fmt.Fprintf(w, "   Tags: %s\n", strings.Join(result.Note.Tags, ", ")); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
		
		if _, err := fmt.Fprintf(w, "   Score: %.2f\n", result.Score); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		if _, err := fmt.Fprintf(w, "   Updated: %s\n", result.Note.UpdatedAt.Format("2006-01-02 15:04")); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

func outputSearchDetailed(w io.Writer, results []search.SearchResult) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(w, "No results found")
		return err
	}

	if _, err := fmt.Fprintf(w, "Found %d results:\n\n", len(results)); err != nil {
		return err
	}

	for i, result := range results {
		if _, err := fmt.Fprintf(w, "=== Result %d ===\n", i+1); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Title: %s\n", result.Note.Title); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "ID: %s\n", result.Note.ID); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Path: %s\n", result.Note.FilePath); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Type: %s\n", result.Note.Type); err != nil {
			return err
		}
		
		if len(result.Note.Tags) > 0 {
			if _, err := fmt.Fprintf(w, "Tags: %s\n", strings.Join(result.Note.Tags, ", ")); err != nil {
				return err
			}
		}
		
		if _, err := fmt.Fprintf(w, "Score: %.2f\n", result.Score); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Created: %s\n", result.Note.CreatedAt.Format("2006-01-02 15:04:05")); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Updated: %s\n", result.Note.UpdatedAt.Format("2006-01-02 15:04:05")); err != nil {
			return err
		}
		
		if result.Snippet != "" {
			if _, err := fmt.Fprintf(w, "\nSnippet:\n%s\n", result.Snippet); err != nil {
				return err
			}
		}
		
		if len(result.Matches) > 0 {
			if _, err := fmt.Fprintf(w, "\nMatches:\n"); err != nil {
				return err
			}
			for _, match := range result.Matches {
				if _, err := fmt.Fprintf(w, "  - %s: %s\n", match.Field, match.Context); err != nil {
					return err
				}
			}
		}
		
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
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

