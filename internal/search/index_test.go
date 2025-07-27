package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIndex_Add(t *testing.T) {
	idx := NewIndex()
	
	doc := &IndexedDocument{
		ID:      "test-1",
		Title:   "Test Document",
		Content: "This is a test document with some content",
		Tags:    []string{"test", "example"},
		Type:    "note",
	}
	
	idx.Add(doc)
	
	// Verify document was added
	assert.Equal(t, 1, idx.Size())
	
	// Verify document can be retrieved
	retrieved, exists := idx.GetDocument("test-1")
	assert.True(t, exists)
	assert.Equal(t, doc.Title, retrieved.Title)
}

func TestIndex_Remove(t *testing.T) {
	idx := NewIndex()
	
	// Add documents
	doc1 := &IndexedDocument{
		ID:      "doc-1",
		Title:   "Document One",
		Content: "First document content",
		Tags:    []string{"tag1"},
		Type:    "note",
	}
	
	doc2 := &IndexedDocument{
		ID:      "doc-2",
		Title:   "Document Two",
		Content: "Second document content",
		Tags:    []string{"tag2"},
		Type:    "note",
	}
	
	idx.Add(doc1)
	idx.Add(doc2)
	
	assert.Equal(t, 2, idx.Size())
	
	// Remove first document
	idx.Remove("doc-1")
	
	assert.Equal(t, 1, idx.Size())
	
	// Verify doc1 is gone
	_, exists := idx.GetDocument("doc-1")
	assert.False(t, exists)
	
	// Verify doc2 still exists
	_, exists = idx.GetDocument("doc-2")
	assert.True(t, exists)
}

func TestIndex_Search(t *testing.T) {
	idx := NewIndex()
	
	// Add test documents
	docs := []*IndexedDocument{
		{
			ID:      "1",
			Title:   "Go Programming",
			Content: "Learn Go programming language",
		},
		{
			ID:      "2",
			Title:   "Python Guide",
			Content: "Python programming tutorial",
		},
		{
			ID:      "3",
			Title:   "JavaScript Basics",
			Content: "Introduction to JavaScript programming",
		},
	}
	
	for _, doc := range docs {
		idx.Add(doc)
	}
	
	// Search for "programming" in content
	results := idx.Search("programming", "content")
	assert.Len(t, results, 3) // All documents have "programming" in content
	
	// Search for "go" in title
	results = idx.Search("go", "title")
	assert.Len(t, results, 1)
	assert.Equal(t, "1", results[0].ID)
	
	// Search for non-existent term
	results = idx.Search("rust", "title")
	assert.Len(t, results, 0)
}

func TestIndex_SearchByTag(t *testing.T) {
	idx := NewIndex()
	
	docs := []*IndexedDocument{
		{
			ID:   "1",
			Tags: []string{"golang", "tutorial"},
		},
		{
			ID:   "2",
			Tags: []string{"python", "tutorial"},
		},
		{
			ID:   "3",
			Tags: []string{"golang", "advanced"},
		},
	}
	
	for _, doc := range docs {
		idx.Add(doc)
	}
	
	// Search by tag
	results := idx.SearchByTag("golang")
	assert.Len(t, results, 2)
	
	results = idx.SearchByTag("tutorial")
	assert.Len(t, results, 2)
	
	results = idx.SearchByTag("advanced")
	assert.Len(t, results, 1)
	assert.Equal(t, "3", results[0].ID)
}

func TestIndex_SearchByType(t *testing.T) {
	idx := NewIndex()
	
	docs := []*IndexedDocument{
		{ID: "1", Type: "note"},
		{ID: "2", Type: "daily"},
		{ID: "3", Type: "note"},
		{ID: "4", Type: "template"},
	}
	
	for _, doc := range docs {
		idx.Add(doc)
	}
	
	// Search by type
	results := idx.SearchByType("note")
	assert.Len(t, results, 2)
	
	results = idx.SearchByType("daily")
	assert.Len(t, results, 1)
	assert.Equal(t, "2", results[0].ID)
}

func TestIndex_Clear(t *testing.T) {
	idx := NewIndex()
	
	// Add some documents
	for i := 0; i < 5; i++ {
		idx.Add(&IndexedDocument{
			ID:    string(rune('0' + i)),
			Title: "Document",
		})
	}
	
	assert.Equal(t, 5, idx.Size())
	
	// Clear index
	idx.Clear()
	
	assert.Equal(t, 0, idx.Size())
	assert.Len(t, idx.GetAllDocuments(), 0)
}

func TestIndex_Tokenize(t *testing.T) {
	idx := NewIndex()
	
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Test123 with-hyphens and_underscores",
			expected: []string{"test123", "with", "hyphens", "and", "underscores"},
		},
		{
			input:    "CamelCase and UPPERCASE",
			expected: []string{"camelcase", "and", "uppercase"},
		},
		{
			input:    "email@example.com",
			expected: []string{"email", "example", "com"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := idx.tokenize(tt.input)
			assert.Equal(t, tt.expected, tokens)
		})
	}
}

func TestIndex_ToMetadata(t *testing.T) {
	doc := &IndexedDocument{
		ID:        "test-id",
		Title:     "Test Title",
		Tags:      []string{"tag1", "tag2"},
		Type:      "note",
		FilePath:  "/path/to/note.md",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Size:      1024,
	}
	
	metadata := doc.ToMetadata()
	
	assert.Equal(t, doc.ID, metadata.ID)
	assert.Equal(t, doc.Title, metadata.Title)
	assert.Equal(t, doc.Tags, metadata.Tags)
	assert.Equal(t, doc.Type, metadata.Type)
	assert.Equal(t, doc.FilePath, metadata.FilePath)
	assert.Equal(t, doc.CreatedAt, metadata.CreatedAt)
	assert.Equal(t, doc.UpdatedAt, metadata.UpdatedAt)
	assert.Equal(t, doc.Size, metadata.Size)
}

func TestIndex_UpdateDocument(t *testing.T) {
	idx := NewIndex()
	
	// Add initial document
	doc := &IndexedDocument{
		ID:      "update-test",
		Title:   "Original Title",
		Content: "Original content",
		Tags:    []string{"original"},
	}
	
	idx.Add(doc)
	
	// Verify original was indexed
	results := idx.Search("original", "title")
	assert.Len(t, results, 1)
	
	// Update document
	updatedDoc := &IndexedDocument{
		ID:      "update-test",
		Title:   "Updated Title",
		Content: "Updated content",
		Tags:    []string{"updated"},
	}
	
	idx.Add(updatedDoc)
	
	// Verify old content is gone
	results = idx.Search("original", "title")
	assert.Len(t, results, 0)
	
	// Verify new content is indexed
	results = idx.Search("updated", "title")
	assert.Len(t, results, 1)
	
	// Verify tags were updated
	results = idx.SearchByTag("original")
	assert.Len(t, results, 0)
	
	results = idx.SearchByTag("updated")
	assert.Len(t, results, 1)
}

func TestIndex_ConcurrentAccess(t *testing.T) {
	idx := NewIndex()
	
	// Test concurrent adds
	done := make(chan bool)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			doc := &IndexedDocument{
				ID:      string(rune('0' + id)),
				Title:   "Concurrent Document",
				Content: "Content for concurrent testing",
			}
			idx.Add(doc)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all documents were added
	assert.Equal(t, 10, idx.Size())
	
	// Test concurrent searches
	searchDone := make(chan int)
	
	for i := 0; i < 5; i++ {
		go func() {
			results := idx.Search("concurrent", "title")
			searchDone <- len(results)
		}()
	}
	
	// Verify searches complete without panic
	for i := 0; i < 5; i++ {
		count := <-searchDone
		assert.Equal(t, 10, count)
	}
}