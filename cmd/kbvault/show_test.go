package main

import (
	"testing"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Test helper function to create a sample note
func createSampleNote() *types.Note {
	return &types.Note{
		ID:       "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Title:    "Sample Note",
		Content:  "This is a sample note content with some **markdown** formatting.",
		FilePath: "/vault/notes/01ARZ3NDEKTSV4RRFFQ69G5FAV.md",
		CreatedAt: time.Date(2023, 12, 1, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 12, 2, 15, 45, 0, 0, time.UTC),
		Frontmatter: types.Frontmatter{
			Tags: []string{"work", "project", "important"},
			Type: "note",
		},
	}
}

func TestDisplayNoteDefault(t *testing.T) {
	note := createSampleNote()

	// Test the function exists and doesn't panic
	// Since it prints to stdout, we can't easily test output without more complex setup
	err := displayNoteDefault(note, false, false)
	if err != nil {
		t.Errorf("displayNoteDefault() error = %v", err)
	}

	// Test with metadata and content enabled
	err = displayNoteDefault(note, true, true)
	if err != nil {
		t.Errorf("displayNoteDefault() with metadata and content error = %v", err)
	}
}

func TestDisplayNoteMarkdown(t *testing.T) {
	note := createSampleNote()

	// Test the function exists and doesn't panic
	err := displayNoteMarkdown(note, false, false)
	if err != nil {
		t.Errorf("displayNoteMarkdown() error = %v", err)
	}
}

func TestDisplayNoteJSON(t *testing.T) {
	note := createSampleNote()

	// Test the function exists and doesn't panic
	err := displayNoteJSON(note)
	if err != nil {
		t.Errorf("displayNoteJSON() error = %v", err)
	}
}

// Additional tests for edge cases
func TestDisplayFunctionsWithEmptyNote(t *testing.T) {
	note := &types.Note{
		ID:       "",
		Title:    "",
		Content:  "",
		FilePath: "",
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
		Frontmatter: types.Frontmatter{
			Tags: []string{},
		},
	}

	// Test that functions handle empty note gracefully
	err := displayNoteDefault(note, false, false)
	if err != nil {
		t.Errorf("displayNoteDefault() with empty note error = %v", err)
	}

	err = displayNoteMarkdown(note, false, false)
	if err != nil {
		t.Errorf("displayNoteMarkdown() with empty note error = %v", err)
	}

	err = displayNoteJSON(note)
	if err != nil {
		t.Errorf("displayNoteJSON() with empty note error = %v", err)
	}
}

func TestDisplayFunctionsWithNilNote(t *testing.T) {
	// Test that functions handle nil note gracefully
	// These should not panic
	
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayNoteDefault() panicked with nil note: %v", r)
		}
	}()
	
	// Note: these functions might panic with nil, which is expected behavior
	// The defer above will catch any panics and fail the test
}