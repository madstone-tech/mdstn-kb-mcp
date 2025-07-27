package search

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorage implements a simple in-memory storage for testing
type mockStorage struct {
	files map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		files: make(map[string][]byte),
	}
}

func (m *mockStorage) Type() types.StorageType {
	return types.StorageTypeLocal
}

func (m *mockStorage) Read(ctx context.Context, path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, types.NewStorageError(m.Type(), "read", path, nil, false)
	}
	return data, nil
}

func (m *mockStorage) Write(ctx context.Context, path string, data []byte) error {
	m.files[path] = data
	return nil
}

func (m *mockStorage) Delete(ctx context.Context, path string) error {
	delete(m.files, path)
	return nil
}

func (m *mockStorage) Exists(ctx context.Context, path string) (bool, error) {
	_, ok := m.files[path]
	return ok, nil
}

func (m *mockStorage) List(ctx context.Context, prefix string) ([]string, error) {
	var results []string
	for path := range m.files {
		results = append(results, path)
	}
	return results, nil
}

func (m *mockStorage) Stat(ctx context.Context, path string) (*types.FileInfo, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, types.NewStorageError(m.Type(), "stat", path, nil, false)
	}
	return &types.FileInfo{
		Path:    path,
		Size:    int64(len(data)),
		ModTime: time.Now().Unix(),
	}, nil
}

func (m *mockStorage) ReadStream(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockStorage) WriteStream(ctx context.Context, path string, reader io.Reader) error {
	return nil
}

func (m *mockStorage) Copy(ctx context.Context, src, dst string) error {
	data, ok := m.files[src]
	if !ok {
		return types.NewStorageError(m.Type(), "copy", src, nil, false)
	}
	m.files[dst] = data
	return nil
}

func (m *mockStorage) Move(ctx context.Context, src, dst string) error {
	data, ok := m.files[src]
	if !ok {
		return types.NewStorageError(m.Type(), "move", src, nil, false)
	}
	m.files[dst] = data
	delete(m.files, src)
	return nil
}

func (m *mockStorage) Health(ctx context.Context) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestEngine_Search(t *testing.T) {
	storage := newMockStorage()
	engine := New(storage, DefaultOptions())
	
	// Add test documents
	docs := []*IndexedDocument{
		{
			ID:        "1",
			Title:     "Getting Started with Go",
			Content:   "Go is a statically typed, compiled programming language",
			Tags:      []string{"golang", "programming", "tutorial"},
			Type:      "note",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        "2",
			Title:     "Advanced Go Patterns",
			Content:   "Design patterns and best practices for Go development",
			Tags:      []string{"golang", "patterns", "advanced"},
			Type:      "note",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},
		{
			ID:        "3",
			Title:     "Python Tutorial",
			Content:   "Introduction to Python programming for beginners",
			Tags:      []string{"python", "programming", "tutorial"},
			Type:      "note",
			CreatedAt: time.Now().Add(-6 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},
	}
	
	// Build index
	for _, doc := range docs {
		engine.index.Add(doc)
	}
	
	tests := []struct {
		name     string
		query    SearchQuery
		expected int
		validate func(t *testing.T, results []SearchResult)
	}{
		{
			name: "search by keyword",
			query: SearchQuery{
				Query: "Go",
			},
			expected: 2,
			validate: func(t *testing.T, results []SearchResult) {
				assert.Len(t, results, 2)
				// Should match documents 1 and 2
				ids := []string{results[0].Note.ID, results[1].Note.ID}
				assert.Contains(t, ids, "1")
				assert.Contains(t, ids, "2")
			},
		},
		{
			name: "search by tag",
			query: SearchQuery{
				Tags: []string{"golang"},
			},
			expected: 2,
			validate: func(t *testing.T, results []SearchResult) {
				assert.Len(t, results, 2)
			},
		},
		{
			name: "search by multiple tags",
			query: SearchQuery{
				Tags: []string{"golang", "advanced"},
			},
			expected: 1,
			validate: func(t *testing.T, results []SearchResult) {
				assert.Len(t, results, 1)
				assert.Equal(t, "2", results[0].Note.ID)
			},
		},
		{
			name: "search with field constraint",
			query: SearchQuery{
				Query:  "programming",
				Fields: []string{"content"},
			},
			expected: 2,
			validate: func(t *testing.T, results []SearchResult) {
				assert.Len(t, results, 2)
			},
		},
		{
			name: "search with sorting by created date",
			query: SearchQuery{
				Query:  "programming",
				SortBy: "created",
			},
			expected: 2,
			validate: func(t *testing.T, results []SearchResult) {
				assert.Len(t, results, 2)
				// Older document should come first
				assert.Equal(t, "1", results[0].Note.ID)
			},
		},
		{
			name: "search with limit",
			query: SearchQuery{
				Query: "programming",
				Limit: 1,
			},
			expected: 1,
			validate: func(t *testing.T, results []SearchResult) {
				assert.Len(t, results, 1)
			},
		},
	}
	
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.Search(ctx, tt.query)
			require.NoError(t, err)
			
			if tt.validate != nil {
				tt.validate(t, results)
			} else {
				assert.Len(t, results, tt.expected)
			}
		})
	}
}

func TestEngine_BuildIndex(t *testing.T) {
	storage := newMockStorage()
	engine := New(storage, DefaultOptions())
	
	// Add test files
	testFiles := map[string]string{
		"note1.md": "# Test Note 1\n\nThis is a test note about Go programming.",
		"note2.md": "# Test Note 2\n\nAnother note about Python development.",
		"note3.md": "# Test Note 3\n\nA third note about JavaScript.",
	}
	
	for path, content := range testFiles {
		err := storage.Write(context.Background(), path, []byte(content))
		require.NoError(t, err)
	}
	
	// Build index
	ctx := context.Background()
	err := engine.BuildIndex(ctx)
	require.NoError(t, err)
	
	// Verify index was built
	assert.Equal(t, 3, engine.index.Size())
	
	// Test search after building index
	results, err := engine.Search(ctx, SearchQuery{Query: "programming"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestEngine_IndexNote(t *testing.T) {
	storage := newMockStorage()
	engine := New(storage, DefaultOptions())
	
	note := &types.Note{
		ID:      "test-123",
		Title:   "Test Note",
		Content: "This is a test note content",
		Frontmatter: types.Frontmatter{
			Tags: []string{"test", "example"},
			Type: "note",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	ctx := context.Background()
	err := engine.IndexNote(ctx, note)
	require.NoError(t, err)
	
	// Verify note was indexed
	results, err := engine.Search(ctx, SearchQuery{Query: "test"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "test-123", results[0].Note.ID)
}

func TestEngine_RemoveFromIndex(t *testing.T) {
	storage := newMockStorage()
	engine := New(storage, DefaultOptions())
	
	// Add a document
	doc := &IndexedDocument{
		ID:      "remove-test",
		Title:   "Document to Remove",
		Content: "This document will be removed",
	}
	engine.index.Add(doc)
	
	// Verify it's indexed
	ctx := context.Background()
	results, err := engine.Search(ctx, SearchQuery{Query: "remove"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	
	// Remove from index
	err = engine.RemoveFromIndex(ctx, "remove-test")
	require.NoError(t, err)
	
	// Verify it's gone
	results, err = engine.Search(ctx, SearchQuery{Query: "remove"})
	require.NoError(t, err)
	assert.Len(t, results, 0)
}