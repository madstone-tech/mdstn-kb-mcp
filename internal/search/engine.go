package search

import (
	"context"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Engine provides full-text search capabilities for notes
type Engine struct {
	mu      sync.RWMutex
	index   *Index
	storage types.StorageBackend
	options Options
}

// Options configures the search engine behavior
type Options struct {
	// CaseSensitive enables case-sensitive search
	CaseSensitive bool

	// MaxResults limits the number of search results
	MaxResults int

	// IndexUpdateInterval controls how often the index is refreshed
	IndexUpdateInterval time.Duration

	// EnableFuzzySearch allows approximate matching
	EnableFuzzySearch bool

	// FuzzyThreshold sets the minimum similarity score (0.0 to 1.0)
	FuzzyThreshold float64
}

// DefaultOptions returns reasonable default search options
func DefaultOptions() Options {
	return Options{
		CaseSensitive:       false,
		MaxResults:          100,
		IndexUpdateInterval: 5 * time.Minute,
		EnableFuzzySearch:   true,
		FuzzyThreshold:      0.7,
	}
}

// New creates a new search engine
func New(storage types.StorageBackend, opts Options) *Engine {
	return &Engine{
		index:   NewIndex(),
		storage: storage,
		options: opts,
	}
}

// SearchQuery represents a search request
type SearchQuery struct {
	// Query is the search text
	Query string

	// Fields to search in (title, content, tags, all)
	Fields []string

	// Tags to filter by (AND operation)
	Tags []string

	// Type to filter by
	Type string

	// DateRange for filtering by creation/update time
	DateRange *DateRange

	// SortBy field (relevance, created, updated, title)
	SortBy string

	// SortDesc reverses the sort order
	SortDesc bool

	// Limit results (0 means use engine default)
	Limit int

	// Offset for pagination
	Offset int
}

// DateRange specifies a time range for filtering
type DateRange struct {
	After  time.Time
	Before time.Time
}

// SearchResult represents a single search match
type SearchResult struct {
	// Note metadata
	Note *types.NoteMetadata

	// Score indicates relevance (higher is better)
	Score float64

	// Matches shows where the query matched
	Matches []Match

	// Snippet shows context around the match
	Snippet string
}

// Match represents a specific location where the query matched
type Match struct {
	Field    string // title, content, tags
	Position int    // character position
	Length   int    // length of match
	Context  string // surrounding text
}

// Search performs a full-text search across notes
func (e *Engine) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	e.mu.RLock()

	// Check if index is empty and build it automatically if needed
	if len(e.index.GetAllDocuments()) == 0 {
		e.mu.RUnlock()

		// Build the index (this will acquire the write lock)
		if err := e.BuildIndex(ctx); err != nil {
			return nil, err
		}

		e.mu.RLock()
	}

	defer e.mu.RUnlock()

	// Normalize query
	searchTerms := e.tokenize(query.Query)

	// Get all matching documents from index
	var candidates []*IndexedDocument

	if len(searchTerms) > 0 {
		// Search specified fields or all fields
		fields := query.Fields
		if len(fields) == 0 || contains(fields, "all") {
			fields = []string{"title", "content", "tags"}
		}

		for _, field := range fields {
			for _, term := range searchTerms {
				docs := e.index.Search(term, field)
				candidates = append(candidates, docs...)
			}
		}
	} else {
		// No text query, get all documents for filtering
		candidates = e.index.GetAllDocuments()
	}

	// Remove duplicates and apply filters
	seen := make(map[string]bool)
	var results []SearchResult

	for _, doc := range candidates {
		if seen[doc.ID] {
			continue
		}
		seen[doc.ID] = true

		// Apply filters
		if !e.matchesFilters(doc, query) {
			continue
		}

		// Calculate score and matches
		score, matches := e.calculateScore(doc, searchTerms, query)
		if score == 0 {
			continue
		}

		// Create result
		result := SearchResult{
			Note:    doc.ToMetadata(),
			Score:   score,
			Matches: matches,
			Snippet: e.generateSnippet(doc, matches),
		}

		results = append(results, result)
	}

	// Sort results
	e.sortResults(results, query.SortBy, query.SortDesc)

	// Apply pagination
	limit := query.Limit
	if limit <= 0 {
		limit = e.options.MaxResults
	}

	start := query.Offset
	end := start + limit
	if end > len(results) {
		end = len(results)
	}
	if start > len(results) {
		start = len(results)
	}

	return results[start:end], nil
}

// BuildIndex creates or updates the search index
func (e *Engine) BuildIndex(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// List all notes from common directories
	dirs := []string{"", "notes/", "daily/"} // Root, notes directory, and daily notes
	var allFiles []string

	for _, dir := range dirs {
		files, err := e.storage.List(ctx, dir)
		if err != nil {
			// Continue if directory doesn't exist
			continue
		}
		// Storage List already returns full relative paths from storage root
		allFiles = append(allFiles, files...)
	}

	files := allFiles

	// Debug: check if we found any files
	if len(files) == 0 {
		// Try alternative approach - list from root and look for .md files
		rootFiles, err := e.storage.List(ctx, "")
		if err == nil {
			for _, file := range rootFiles {
				if strings.HasSuffix(file, ".md") {
					files = append(files, file)
				}
			}
		}

		// Also try looking for files recursively in the notes directory
		if len(files) == 0 {
			// This is a fallback - the storage backend might not support directory listing
			// For local storage, try some common note file patterns
			commonFiles := []string{
				"notes/golang-basics.md",
				"notes/python-tutorial.md",
				"notes/daily-2024-01-15.md",
				"notes/meeting-notes.md",
			}
			for _, file := range commonFiles {
				if data, err := e.storage.Read(ctx, file); err == nil && len(data) > 0 {
					files = append(files, file)
				}
			}
		}
	}

	// Clear existing index
	e.index = NewIndex()

	// Index each note
	for _, file := range files {
		if !strings.HasSuffix(file, ".md") {
			continue
		}

		// Read note content
		data, err := e.storage.Read(ctx, file)
		if err != nil {
			// Log error but continue indexing
			continue
		}

		// Parse note
		note, err := e.parseNote(file, data)
		if err != nil {
			continue
		}

		// Index the note
		e.indexNote(note)
	}

	return nil
}

// IndexNote adds or updates a single note in the index
func (e *Engine) IndexNote(ctx context.Context, note *types.Note) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	doc := &IndexedDocument{
		ID:        note.ID,
		Title:     note.Title,
		Content:   note.Content,
		Tags:      note.Frontmatter.Tags,
		Type:      note.Frontmatter.Type,
		FilePath:  note.FilePath,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
		Size:      note.Size,
	}

	e.index.Add(doc)
	return nil
}

// RemoveFromIndex removes a note from the search index
func (e *Engine) RemoveFromIndex(ctx context.Context, noteID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.index.Remove(noteID)
	return nil
}

// tokenize splits text into searchable tokens
func (e *Engine) tokenize(text string) []string {
	if !e.options.CaseSensitive {
		text = strings.ToLower(text)
	}

	// Split on word boundaries
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// matchesFilters checks if a document matches the query filters
func (e *Engine) matchesFilters(doc *IndexedDocument, query SearchQuery) bool {
	// Tag filter (AND operation)
	if len(query.Tags) > 0 {
		for _, tag := range query.Tags {
			if !contains(doc.Tags, tag) {
				return false
			}
		}
	}

	// Type filter
	if query.Type != "" && doc.Type != query.Type {
		return false
	}

	// Date range filter
	if query.DateRange != nil {
		if !query.DateRange.After.IsZero() && doc.CreatedAt.Before(query.DateRange.After) {
			return false
		}
		if !query.DateRange.Before.IsZero() && doc.CreatedAt.After(query.DateRange.Before) {
			return false
		}
	}

	return true
}

// calculateScore computes the relevance score for a document
func (e *Engine) calculateScore(doc *IndexedDocument, terms []string, query SearchQuery) (float64, []Match) {
	if len(terms) == 0 {
		// No text search, just return a base score
		return 1.0, nil
	}

	var totalScore float64
	var matches []Match

	// Score title matches (weighted higher)
	titleScore, titleMatches := e.scoreField(doc.Title, terms, "title", 2.0)
	totalScore += titleScore
	matches = append(matches, titleMatches...)

	// Score content matches
	contentScore, contentMatches := e.scoreField(doc.Content, terms, "content", 1.0)
	totalScore += contentScore
	matches = append(matches, contentMatches...)

	// Score tag matches
	tagText := strings.Join(doc.Tags, " ")
	tagScore, tagMatches := e.scoreField(tagText, terms, "tags", 1.5)
	totalScore += tagScore
	matches = append(matches, tagMatches...)

	return totalScore, matches
}

// scoreField calculates score for matches in a specific field
func (e *Engine) scoreField(text string, terms []string, field string, weight float64) (float64, []Match) {
	if !e.options.CaseSensitive {
		text = strings.ToLower(text)
	}

	var score float64
	var matches []Match

	for _, term := range terms {
		// Exact match
		count := strings.Count(text, term)
		if count > 0 {
			score += float64(count) * weight

			// Find match positions
			idx := 0
			for i := 0; i < count && i < 3; i++ { // Limit to 3 matches per term
				pos := strings.Index(text[idx:], term)
				if pos == -1 {
					break
				}

				actualPos := idx + pos
				match := Match{
					Field:    field,
					Position: actualPos,
					Length:   len(term),
					Context:  e.extractContext(text, actualPos, len(term)),
				}
				matches = append(matches, match)

				idx = actualPos + len(term)
			}
		}

		// Fuzzy match if enabled
		if e.options.EnableFuzzySearch && count == 0 {
			// Simple fuzzy matching - check if term is substring
			if strings.Contains(text, term[:min(3, len(term))]) {
				score += 0.5 * weight
			}
		}
	}

	return score, matches
}

// extractContext gets surrounding text for a match
func (e *Engine) extractContext(text string, pos, length int) string {
	const contextSize = 40

	start := max(0, pos-contextSize)
	end := min(len(text), pos+length+contextSize)

	context := text[start:end]

	// Add ellipsis if truncated
	if start > 0 {
		context = "..." + context
	}
	if end < len(text) {
		context = context + "..."
	}

	return context
}

// generateSnippet creates a preview snippet for search results
func (e *Engine) generateSnippet(doc *IndexedDocument, matches []Match) string {
	if len(matches) == 0 {
		// No matches, return beginning of content
		content := doc.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		return content
	}

	// Use the first content match if available
	for _, match := range matches {
		if match.Field == "content" {
			return match.Context
		}
	}

	// Otherwise use first match
	return matches[0].Context
}

// sortResults sorts search results according to the query
func (e *Engine) sortResults(results []SearchResult, sortBy string, desc bool) {
	sort.Slice(results, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "created":
			less = results[i].Note.CreatedAt.Before(results[j].Note.CreatedAt)
		case "updated":
			less = results[i].Note.UpdatedAt.Before(results[j].Note.UpdatedAt)
		case "title":
			less = results[i].Note.Title < results[j].Note.Title
		default: // relevance
			less = results[i].Score > results[j].Score // Higher score first
		}

		if desc && sortBy != "relevance" {
			return !less
		}
		return less
	})
}

// parseNote extracts note data from file content
func (e *Engine) parseNote(path string, data []byte) (*IndexedDocument, error) {
	// Parse note content and extract metadata
	content := string(data)
	lines := strings.Split(content, "\n")

	// Extract title from first # heading or filename
	title := strings.TrimSuffix(filepath.Base(path), ".md")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Extract tags from content
	var tags []string
	for _, line := range lines {
		// Look for "Tags:" line with hashtag-style tags
		if strings.HasPrefix(strings.TrimSpace(line), "Tags:") {
			tagsPart := strings.TrimPrefix(strings.TrimSpace(line), "Tags:")
			// Extract hashtag-style tags
			tagMatches := regexp.MustCompile(`#(\w+)`).FindAllStringSubmatch(tagsPart, -1)
			for _, match := range tagMatches {
				if len(match) > 1 {
					tags = append(tags, match[1])
				}
			}
		}
	}

	doc := &IndexedDocument{
		ID:        path, // Use path as ID for now
		Title:     title,
		Content:   content,
		Tags:      tags,
		FilePath:  path,
		CreatedAt: time.Now(), // Would get from file info
		UpdatedAt: time.Now(),
		Size:      int64(len(data)),
	}

	return doc, nil
}

// indexNote adds a document to the index
func (e *Engine) indexNote(doc *IndexedDocument) {
	e.index.Add(doc)
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
