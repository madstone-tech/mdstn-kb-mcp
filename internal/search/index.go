package search

import (
	"strings"
	"sync"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Index provides an in-memory inverted index for fast text search
type Index struct {
	mu sync.RWMutex
	
	// Inverted index: term -> field -> document IDs
	terms map[string]map[string]map[string]bool
	
	// Document store: ID -> document
	documents map[string]*IndexedDocument
	
	// Metadata indices
	tagIndex  map[string]map[string]bool // tag -> document IDs
	typeIndex map[string]map[string]bool // type -> document IDs
}

// IndexedDocument represents a document in the search index
type IndexedDocument struct {
	ID        string
	Title     string
	Content   string
	Tags      []string
	Type      string
	FilePath  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Size      int64
}

// NewIndex creates a new search index
func NewIndex() *Index {
	return &Index{
		terms:     make(map[string]map[string]map[string]bool),
		documents: make(map[string]*IndexedDocument),
		tagIndex:  make(map[string]map[string]bool),
		typeIndex: make(map[string]map[string]bool),
	}
}

// Add indexes a document
func (idx *Index) Add(doc *IndexedDocument) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	
	// Remove old version if exists
	idx.removeUnsafe(doc.ID)
	
	// Store document
	idx.documents[doc.ID] = doc
	
	// Index title
	idx.indexField(doc.ID, "title", doc.Title)
	
	// Index content
	idx.indexField(doc.ID, "content", doc.Content)
	
	// Index tags
	for _, tag := range doc.Tags {
		idx.indexField(doc.ID, "tags", tag)
		
		// Add to tag index
		if idx.tagIndex[tag] == nil {
			idx.tagIndex[tag] = make(map[string]bool)
		}
		idx.tagIndex[tag][doc.ID] = true
	}
	
	// Index type
	if doc.Type != "" {
		if idx.typeIndex[doc.Type] == nil {
			idx.typeIndex[doc.Type] = make(map[string]bool)
		}
		idx.typeIndex[doc.Type][doc.ID] = true
	}
}

// Remove deletes a document from the index
func (idx *Index) Remove(docID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	
	idx.removeUnsafe(docID)
}

// removeUnsafe removes a document without locking (must be called with lock held)
func (idx *Index) removeUnsafe(docID string) {
	doc, exists := idx.documents[docID]
	if !exists {
		return
	}
	
	// Remove from term index
	for term, fields := range idx.terms {
		for field, docs := range fields {
			delete(docs, docID)
			
			// Clean up empty maps
			if len(docs) == 0 {
				delete(fields, field)
			}
		}
		
		if len(fields) == 0 {
			delete(idx.terms, term)
		}
	}
	
	// Remove from tag index
	for _, tag := range doc.Tags {
		delete(idx.tagIndex[tag], docID)
		if len(idx.tagIndex[tag]) == 0 {
			delete(idx.tagIndex, tag)
		}
	}
	
	// Remove from type index
	if doc.Type != "" {
		delete(idx.typeIndex[doc.Type], docID)
		if len(idx.typeIndex[doc.Type]) == 0 {
			delete(idx.typeIndex, doc.Type)
		}
	}
	
	// Remove document
	delete(idx.documents, docID)
}

// Search finds documents containing the given term in the specified field
func (idx *Index) Search(term string, field string) []*IndexedDocument {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	term = strings.ToLower(term)
	
	var results []*IndexedDocument
	
	if fields, ok := idx.terms[term]; ok {
		if docIDs, ok := fields[field]; ok {
			for docID := range docIDs {
				if doc, exists := idx.documents[docID]; exists {
					results = append(results, doc)
				}
			}
		}
	}
	
	return results
}

// SearchByTag finds documents with the given tag
func (idx *Index) SearchByTag(tag string) []*IndexedDocument {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	var results []*IndexedDocument
	
	if docIDs, ok := idx.tagIndex[tag]; ok {
		for docID := range docIDs {
			if doc, exists := idx.documents[docID]; exists {
				results = append(results, doc)
			}
		}
	}
	
	return results
}

// SearchByType finds documents of the given type
func (idx *Index) SearchByType(docType string) []*IndexedDocument {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	var results []*IndexedDocument
	
	if docIDs, ok := idx.typeIndex[docType]; ok {
		for docID := range docIDs {
			if doc, exists := idx.documents[docID]; exists {
				results = append(results, doc)
			}
		}
	}
	
	return results
}

// GetDocument retrieves a document by ID
func (idx *Index) GetDocument(docID string) (*IndexedDocument, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	doc, exists := idx.documents[docID]
	return doc, exists
}

// GetAllDocuments returns all indexed documents
func (idx *Index) GetAllDocuments() []*IndexedDocument {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	results := make([]*IndexedDocument, 0, len(idx.documents))
	for _, doc := range idx.documents {
		results = append(results, doc)
	}
	
	return results
}

// Size returns the number of indexed documents
func (idx *Index) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	return len(idx.documents)
}

// Clear removes all documents from the index
func (idx *Index) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	
	idx.terms = make(map[string]map[string]map[string]bool)
	idx.documents = make(map[string]*IndexedDocument)
	idx.tagIndex = make(map[string]map[string]bool)
	idx.typeIndex = make(map[string]map[string]bool)
}

// indexField indexes the content of a field for a document
func (idx *Index) indexField(docID, field, content string) {
	// Tokenize content
	tokens := idx.tokenize(content)
	
	// Index each token
	for _, token := range tokens {
		if idx.terms[token] == nil {
			idx.terms[token] = make(map[string]map[string]bool)
		}
		
		if idx.terms[token][field] == nil {
			idx.terms[token][field] = make(map[string]bool)
		}
		
		idx.terms[token][field][docID] = true
	}
}

// tokenize splits text into indexable tokens
func (idx *Index) tokenize(text string) []string {
	text = strings.ToLower(text)
	
	// Simple tokenization - split on non-alphanumeric characters
	var tokens []string
	var current strings.Builder
	
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
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

// ToMetadata converts an IndexedDocument to NoteMetadata
func (d *IndexedDocument) ToMetadata() *types.NoteMetadata {
	return &types.NoteMetadata{
		ID:        d.ID,
		Title:     d.Title,
		Tags:      d.Tags,
		Type:      d.Type,
		FilePath:  d.FilePath,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		Size:      d.Size,
	}
}