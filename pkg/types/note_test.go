package types

import (
	"testing"
	"time"
)

func TestNote_ToMetadata(t *testing.T) {
	note := &Note{
		ID:             "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Title:          "Test Note",
		Content:        "This is a test note",
		FilePath:       "notes/01ARZ3NDEKTSV4RRFFQ69G5FAV.md",
		StorageBackend: StorageTypeLocal,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Size:           18,
		Frontmatter: Frontmatter{
			ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			Title:   "Test Note",
			Tags:    []string{"test", "example"},
			Type:    "note",
			Created: "2024-01-15T14:23:00Z",
			Updated: "2024-01-15T14:23:00Z",
			Storage: "local",
		},
	}

	metadata := note.ToMetadata()

	if metadata.ID != note.ID {
		t.Errorf("Expected ID %s, got %s", note.ID, metadata.ID)
	}

	if metadata.Title != note.Title {
		t.Errorf("Expected title %s, got %s", note.Title, metadata.Title)
	}

	if len(metadata.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(metadata.Tags))
	}

	if metadata.StorageBackend != note.StorageBackend {
		t.Errorf("Expected storage backend %s, got %s", note.StorageBackend, metadata.StorageBackend)
	}
}

func TestNote_HasTag(t *testing.T) {
	note := &Note{
		Frontmatter: Frontmatter{
			Tags: []string{"test", "example", "demo"},
		},
	}

	testCases := []struct {
		tag      string
		expected bool
	}{
		{"test", true},
		{"example", true},
		{"demo", true},
		{"nonexistent", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := note.HasTag(tc.tag)
		if result != tc.expected {
			t.Errorf("HasTag(%s) = %v, expected %v", tc.tag, result, tc.expected)
		}
	}
}

func TestNote_AddTag(t *testing.T) {
	note := &Note{
		Frontmatter: Frontmatter{
			Tags: []string{"existing"},
		},
	}

	// Add new tag
	note.AddTag("new")
	if !note.HasTag("new") {
		t.Error("AddTag should add new tag")
	}

	// Add existing tag (should not duplicate)
	initialLen := len(note.Frontmatter.Tags)
	note.AddTag("existing")
	if len(note.Frontmatter.Tags) != initialLen {
		t.Error("AddTag should not duplicate existing tag")
	}
}

func TestNote_RemoveTag(t *testing.T) {
	note := &Note{
		Frontmatter: Frontmatter{
			Tags: []string{"tag1", "tag2", "tag3"},
		},
	}

	// Remove existing tag
	note.RemoveTag("tag2")
	if note.HasTag("tag2") {
		t.Error("RemoveTag should remove existing tag")
	}

	// Verify other tags remain
	if !note.HasTag("tag1") || !note.HasTag("tag3") {
		t.Error("RemoveTag should not affect other tags")
	}

	// Remove non-existent tag (should not cause error)
	initialLen := len(note.Frontmatter.Tags)
	note.RemoveTag("nonexistent")
	if len(note.Frontmatter.Tags) != initialLen {
		t.Error("RemoveTag of non-existent tag should not change tags")
	}
}

func TestNote_IsDaily(t *testing.T) {
	testCases := []struct {
		noteType string
		expected bool
	}{
		{"daily", true},
		{"note", false},
		{"template", false},
		{"", false},
	}

	for _, tc := range testCases {
		note := &Note{
			Frontmatter: Frontmatter{Type: tc.noteType},
		}

		result := note.IsDaily()
		if result != tc.expected {
			t.Errorf("IsDaily() with type %s = %v, expected %v", tc.noteType, result, tc.expected)
		}
	}
}

func TestNote_IsTemplate(t *testing.T) {
	testCases := []struct {
		noteType string
		expected bool
	}{
		{"template", true},
		{"note", false},
		{"daily", false},
		{"", false},
	}

	for _, tc := range testCases {
		note := &Note{
			Frontmatter: Frontmatter{Type: tc.noteType},
		}

		result := note.IsTemplate()
		if result != tc.expected {
			t.Errorf("IsTemplate() with type %s = %v, expected %v", tc.noteType, result, tc.expected)
		}
	}
}

func TestNote_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		note        *Note
		expectError bool
	}{
		{
			name: "valid note",
			note: &Note{
				ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				Title:   "Test Note",
				Content: "This is a test",
				Size:    14,
			},
			expectError: false,
		},
		{
			name: "empty ID",
			note: &Note{
				ID:      "",
				Title:   "Test Note",
				Content: "This is a test",
			},
			expectError: true,
		},
		{
			name: "empty title",
			note: &Note{
				ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				Title:   "",
				Content: "This is a test",
			},
			expectError: true,
		},
		{
			name: "content too large",
			note: &Note{
				ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
				Title:   "Test Note",
				Content: string(make([]byte, 11*1024*1024)), // 11MB
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.note.Validate()

			if tc.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestCreateNoteRequest_Validation(t *testing.T) {
	// This would typically use a validation library like go-playground/validator
	// For now, we'll test the struct fields are properly defined

	req := CreateNoteRequest{
		Title:   "Test Note",
		Content: "This is test content",
		Tags:    []string{"test", "example"},
		Type:    "note",
	}

	if req.Title == "" {
		t.Error("Title should not be empty")
	}

	if len(req.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(req.Tags))
	}
}

func TestUpdateNoteRequest_PointerFields(t *testing.T) {
	title := "Updated Title"
	content := "Updated Content"
	noteType := "updated"

	req := UpdateNoteRequest{
		ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Title:   &title,
		Content: &content,
		Type:    &noteType,
		Tags:    []string{"updated"},
	}

	if req.ID == "" {
		t.Error("ID should not be empty")
	}

	if req.Title == nil || *req.Title != title {
		t.Error("Title pointer should be set correctly")
	}

	if req.Content == nil || *req.Content != content {
		t.Error("Content pointer should be set correctly")
	}

	if req.Type == nil || *req.Type != noteType {
		t.Error("Type pointer should be set correctly")
	}
}

func TestLinkType_Constants(t *testing.T) {
	// Test that link type constants are defined correctly
	if LinkTypeWiki != "wiki" {
		t.Errorf("Expected LinkTypeWiki to be 'wiki', got %s", LinkTypeWiki)
	}

	if LinkTypeMarkdown != "markdown" {
		t.Errorf("Expected LinkTypeMarkdown to be 'markdown', got %s", LinkTypeMarkdown)
	}

	if LinkTypeID != "id" {
		t.Errorf("Expected LinkTypeID to be 'id', got %s", LinkTypeID)
	}
}

func TestLink_Structure(t *testing.T) {
	link := Link{
		SourceID:    "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		TargetID:    "01B3NVXKF8V9PQRZ6XYMD7CT42",
		TargetTitle: "Target Note",
		LinkText:    "Link to target",
		Position:    125,
		IsValid:     true,
		Type:        LinkTypeWiki,
	}

	if link.SourceID == "" || link.TargetID == "" {
		t.Error("Link should have source and target IDs")
	}

	if link.Type != LinkTypeWiki {
		t.Errorf("Expected link type %s, got %s", LinkTypeWiki, link.Type)
	}

	if !link.IsValid {
		t.Error("Link should be marked as valid")
	}
}
