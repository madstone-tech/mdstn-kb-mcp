package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/madstone-tech/mdstn-kb-mcp/internal/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCommand(t *testing.T) {
	// Create temporary vault
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	
	// Initialize vault
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{vaultPath})
	err := initCmd.Execute()
	require.NoError(t, err)
	
	// Create test notes
	testNotes := []struct {
		filename string
		content  string
	}{
		{
			filename: "golang-basics.md",
			content: `# Golang Basics

Tags: #golang #programming #tutorial

This note covers the basics of Go programming language.
Topics include variables, functions, and goroutines.`,
		},
		{
			filename: "python-tutorial.md",
			content: `# Python Tutorial

Tags: #python #programming #tutorial

Introduction to Python programming for beginners.
Learn about data types, loops, and functions.`,
		},
		{
			filename: "daily-2024-01-15.md",
			content: `# Daily Note - 2024-01-15

Tags: #daily

Today I learned about goroutines and channels in Go.
Also reviewed Python decorators.`,
		},
		{
			filename: "meeting-notes.md",
			content: `# Team Meeting Notes

Tags: #meeting #work

Discussed project timeline and deliverables.
Action items: review code, update documentation.`,
		},
	}
	
	// Create notes directory as expected by vault structure
	notesDir := filepath.Join(vaultPath, "notes")
	err = os.MkdirAll(notesDir, 0755)
	require.NoError(t, err)
	
	// Write test notes
	for _, note := range testNotes {
		notePath := filepath.Join(notesDir, note.filename)
		err := os.WriteFile(notePath, []byte(note.content), 0644)
		require.NoError(t, err)
	}
	
	// Change to vault directory for tests to work
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	
	err = os.Chdir(vaultPath)
	require.NoError(t, err)
	
	tests := []struct {
		name        string
		args        []string
		wantOutput  []string
		dontWant    []string
		wantErr     bool
		checkJSON   bool
	}{
		{
			name: "simple text search",
			args: []string{"search", "golang"},
			wantOutput: []string{
				"Golang Basics",
				"daily-2024-01-15", // Mentions goroutines
			},
			dontWant: []string{
				"Python Tutorial",
				"Team Meeting",
			},
		},
		{
			name: "search with tag filter",
			args: []string{"search", "programming", "--tag", "golang"},
			wantOutput: []string{
				"Golang Basics",
			},
			dontWant: []string{
				"Python Tutorial", // Has programming tag but not golang
			},
		},
		{
			name: "search by tag only",
			args: []string{"search", "--tag", "daily"},
			wantOutput: []string{
				"Daily Note",
			},
		},
		{
			name: "search with multiple tags",
			args: []string{"search", "--tag", "programming", "--tag", "tutorial"},
			wantOutput: []string{
				"Golang Basics",
				"Python Tutorial",
			},
			dontWant: []string{
				"Daily Note",
			},
		},
		{
			name: "search in specific field",
			args: []string{"search", "variables", "--field", "content"},
			wantOutput: []string{
				"Golang Basics", // Contains "variables" in content
			},
		},
		{
			name: "search with limit",
			args: []string{"search", "programming", "--limit", "1"},
			wantOutput: []string{
				"Found 1 result",
			},
		},
		{
			name: "empty search",
			args: []string{"search"},
			wantOutput: []string{
				"No results found",
			},
		},
		{
			name: "search with JSON output",
			args: []string{"search", "golang", "--json"},
			checkJSON: true,
		},
		{
			name: "search with detailed output",
			args: []string{"search", "goroutines", "--detailed"},
			wantOutput: []string{
				"=== Result",
				"Title:",
				"Snippet:",
				"goroutines",
			},
		},
		{
			name: "build index",
			args: []string{"search", "--build-index"},
			wantOutput: []string{
				"Building search index",
				"Search index built successfully",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			
			output := buf.String()
			
			if tt.checkJSON {
				// Verify valid JSON output
				var result struct {
					Count   int                   `json:"count"`
					Results []search.SearchResult `json:"results"`
				}
				err := json.Unmarshal(buf.Bytes(), &result)
				assert.NoError(t, err)
				assert.Greater(t, result.Count, 0)
			} else {
				// Check expected output
				for _, want := range tt.wantOutput {
					assert.Contains(t, output, want)
				}
				
				// Check unwanted output
				for _, dontWant := range tt.dontWant {
					assert.NotContains(t, output, dontWant)
				}
			}
		})
	}
}

func TestSearchCommand_DateRange(t *testing.T) {
	// Create temporary vault
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	
	// Initialize vault
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{vaultPath})
	err := initCmd.Execute()
	require.NoError(t, err)
	
	// Change to vault directory for tests to work
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	
	err = os.Chdir(vaultPath)
	require.NoError(t, err)
	
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid date range",
			args: []string{"search", "--after", "2024-01-01", "--before", "2024-12-31"},
		},
		{
			name: "invalid after date",
			args: []string{"search", "--after", "invalid-date"},
			wantErr: true,
			errMsg:  "invalid after date",
		},
		{
			name: "invalid before date",
			args: []string{"search", "--before", "2024-13-45"},
			wantErr: true,
			errMsg:  "invalid before date",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)
			
			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchCommand_Sorting(t *testing.T) {
	// Create temporary vault
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	
	// Initialize vault
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{vaultPath})
	err := initCmd.Execute()
	require.NoError(t, err)
	
	// Create notes directory
	notesDir := filepath.Join(vaultPath, "notes")
	err = os.MkdirAll(notesDir, 0755)
	require.NoError(t, err)
	
	// Create notes with different titles
	notes := []struct {
		filename string
		title    string
	}{
		{"alpha.md", "# Alpha Note\nContent about programming"},
		{"beta.md", "# Beta Note\nContent about programming"},
		{"gamma.md", "# Gamma Note\nContent about programming"},
	}
	
	for _, note := range notes {
		notePath := filepath.Join(notesDir, note.filename)
		err := os.WriteFile(notePath, []byte(note.title), 0644)
		require.NoError(t, err)
	}
	
	// Change to vault directory for tests to work
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	
	err = os.Chdir(vaultPath)
	require.NoError(t, err)
	
	// Test sorting by title
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"search", "programming", "--sort", "title"})
	
	err = cmd.Execute()
	require.NoError(t, err)
	
	output := buf.String()
	
	// Check that results appear in alphabetical order
	alphaPos := strings.Index(output, "Alpha")
	betaPos := strings.Index(output, "Beta")
	gammaPos := strings.Index(output, "Gamma")
	
	assert.True(t, alphaPos < betaPos, "Alpha should appear before Beta")
	assert.True(t, betaPos < gammaPos, "Beta should appear before Gamma")
}

func TestSearchCommand_Pagination(t *testing.T) {
	// Create temporary vault
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	
	// Initialize vault
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{vaultPath})
	err := initCmd.Execute()
	require.NoError(t, err)
	
	// Create notes directory
	notesDir := filepath.Join(vaultPath, "notes")
	err = os.MkdirAll(notesDir, 0755)
	require.NoError(t, err)
	
	// Create multiple notes
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("note-%02d.md", i)
		content := fmt.Sprintf("# Note %d\nThis is test note number %d with search keyword", i, i)
		notePath := filepath.Join(notesDir, filename)
		err := os.WriteFile(notePath, []byte(content), 0644)
		require.NoError(t, err)
	}
	
	// Change to vault directory for tests to work
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	
	err = os.Chdir(vaultPath)
	require.NoError(t, err)
	
	// Test with limit
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"search", "keyword", "--limit", "3", "--json"})
	
	err = cmd.Execute()
	require.NoError(t, err)
	
	var result struct {
		Count   int `json:"count"`
		Results []struct{} `json:"results"`
	}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 3, result.Count)
	
	// Test with offset
	cmd = newRootCmd()
	buf.Reset()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"search", "keyword", "--limit", "3", "--offset", "5", "--json"})
	
	err = cmd.Execute()
	require.NoError(t, err)
	
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 3, result.Count)
}