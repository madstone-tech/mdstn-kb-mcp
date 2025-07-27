package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEngine_Render(t *testing.T) {
	// Create temporary directory for templates
	tempDir := t.TempDir()
	
	// Create engine
	engine := NewEngine(tempDir)
	
	// Create test template
	testTemplate := `# {{.Title}}

Created: {{.Created.Format "2006-01-02"}}
{{if .Tags}}Tags: {{join .Tags ", "}}{{end}}

Content goes here...`
	
	templatePath := filepath.Join(tempDir, "test.md")
	err := os.WriteFile(templatePath, []byte(testTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}
	
	// Test data
	data := TemplateData{
		ID:      "test-id",
		Title:   "Test Note",
		Tags:    []string{"test", "example"},
		Created: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	
	// Render template
	result, err := engine.Render("test", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}
	
	// Check result
	expected := `# Test Note

Created: 2023-01-01
Tags: test, example

Content goes here...`
	
	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Unexpected result:\nGot:\n%s\n\nExpected:\n%s", result, expected)
	}
}

func TestEngine_CreateDefaultTemplates(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create engine
	engine := NewEngine(tempDir)
	
	// Create default templates
	err := engine.CreateDefaultTemplates()
	if err != nil {
		t.Fatalf("Failed to create default templates: %v", err)
	}
	
	// Check that templates were created
	expectedTemplates := []string{"default", "daily", "meeting", "book"}
	for _, name := range expectedTemplates {
		templatePath := filepath.Join(tempDir, name+".md")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			t.Errorf("Template %s was not created", name)
		}
	}
}

func TestEngine_ListTemplates(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create engine
	engine := NewEngine(tempDir)
	
	// Create some test templates
	templates := []string{"template1.md", "template2.md", "not-template.txt"}
	for _, tmpl := range templates {
		templatePath := filepath.Join(tempDir, tmpl)
		err := os.WriteFile(templatePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test template %s: %v", tmpl, err)
		}
	}
	
	// List templates
	result, err := engine.ListTemplates()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}
	
	// Check result - should only include .md files
	expected := []string{"template1", "template2"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d templates, got %d", len(expected), len(result))
	}
	
	for _, expectedName := range expected {
		found := false
		for _, resultName := range result {
			if resultName == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected template %s not found in result", expectedName)
		}
	}
}

func TestEngine_ValidateTemplate(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	// Create engine
	engine := NewEngine(tempDir)
	
	// Create valid template
	validTemplate := `# {{.Title}}
Created: {{.Created.Format "2006-01-02"}}`
	
	validPath := filepath.Join(tempDir, "valid.md")
	err := os.WriteFile(validPath, []byte(validTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid template: %v", err)
	}
	
	// Create invalid template
	invalidTemplate := `# {{.Title}}
Created: {{.Created.Format "2006-01-02"`  // Missing closing brace
	
	invalidPath := filepath.Join(tempDir, "invalid.md")
	err = os.WriteFile(invalidPath, []byte(invalidTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid template: %v", err)
	}
	
	// Test valid template
	err = engine.ValidateTemplate("valid")
	if err != nil {
		t.Errorf("Valid template failed validation: %v", err)
	}
	
	// Test invalid template
	err = engine.ValidateTemplate("invalid")
	if err == nil {
		t.Error("Invalid template passed validation")
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := templateFuncs()
	
	// Test functions exist
	expectedFuncs := []string{"lower", "upper", "title", "join", "now", "date", "add", "seq"}
	for _, funcName := range expectedFuncs {
		if _, exists := funcs[funcName]; !exists {
			t.Errorf("Expected function %s not found", funcName)
		}
	}
	
	// Test add function
	addFunc := funcs["add"].(func(int, int) int)
	if result := addFunc(2, 3); result != 5 {
		t.Errorf("add(2, 3) = %d, expected 5", result)
	}
	
	// Test seq function
	seqFunc := funcs["seq"].(func(int, int) []int)
	result := seqFunc(1, 3)
	expected := []int{1, 2, 3}
	if len(result) != len(expected) {
		t.Errorf("seq(1, 3) length = %d, expected %d", len(result), len(expected))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("seq(1, 3)[%d] = %d, expected %d", i, result[i], v)
		}
	}
}