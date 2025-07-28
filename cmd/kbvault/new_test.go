package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestLoadTemplate(t *testing.T) {
	// Create a temporary directory structure that matches what loadTemplate expects
	tempDir := t.TempDir()
	templatesDir := filepath.Join(tempDir, "templates")
	err := os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Create a test template
	testTemplate := `# {{.Title}}

Created: {{.Created.Format "2006-01-02"}}
{{if .Tags}}Tags: {{join .Tags ", "}}{{end}}

Content goes here...`

	templatePath := filepath.Join(templatesDir, "test.md")
	err = os.WriteFile(templatePath, []byte(testTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Change to the temp directory so loadTemplate can find the templates
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	tests := []struct {
		name         string
		templateName string
		wantErr      bool
	}{
		{
			name:         "valid_template",
			templateName: "test",
			wantErr:      false,
		},
		{
			name:         "nonexistent_template",
			templateName: "nonexistent",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple config for testing
			config := &types.Config{}
			
			result, err := loadTemplate(config, tt.templateName)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("loadTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Basic check that we got some content
				if result == "" {
					t.Errorf("loadTemplate() returned empty result")
				}
			}
		})
	}
}

func TestOpenInEditor(t *testing.T) {
	// Create a temporary file
	tempFile := filepath.Join(t.TempDir(), "test.md")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		editor   string
		wantErr  bool
	}{
		{
			name:     "with_echo_editor",
			filePath: tempFile,
			editor:   "echo", // Use echo as a safe "editor" for testing
			wantErr:  false,
		},
		{
			name:     "with_nonexistent_file",
			filePath: "/nonexistent/path/file.md",
			editor:   "echo",
			wantErr:  true, // May or may not error depending on the editor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the EDITOR environment variable
			oldEditor := os.Getenv("EDITOR")
			if err := os.Setenv("EDITOR", tt.editor); err != nil {
				t.Fatalf("Failed to set EDITOR: %v", err)
			}
			defer func() {
				_ = os.Setenv("EDITOR", oldEditor)
			}()

			err := openInEditor(tt.filePath)
			
			// For echo command, it should succeed regardless of file existence
			// since echo just prints its arguments
			if tt.editor == "echo" && err != nil {
				t.Errorf("openInEditor() with echo should not error, got: %v", err)
			}
		})
	}
}

func TestOpenInEditorDefaultEditor(t *testing.T) {
	// Test default editor behavior
	tempFile := filepath.Join(t.TempDir(), "test.md")
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Clear EDITOR environment variable to test default
	oldEditor := os.Getenv("EDITOR")
	if err := os.Unsetenv("EDITOR"); err != nil {
		t.Fatalf("Failed to unset EDITOR: %v", err)
	}
	defer func() {
		_ = os.Setenv("EDITOR", oldEditor)
	}()

	// This will try to use nano as default, which may not be available
	// We'll just check that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("openInEditor() panicked: %v", r)
		}
	}()

	// Call the function - it may error if nano is not available, which is fine
	_ = openInEditor(tempFile)
}

func TestFindVaultRoot(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	
	// Create a vault structure
	vaultDir := filepath.Join(tempDir, "test-vault")
	kbvaultDir := filepath.Join(vaultDir, ".kbvault")
	subDir := filepath.Join(vaultDir, "subdir")
	
	err := os.MkdirAll(kbvaultDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .kbvault directory: %v", err)
	}
	
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create config file
	configPath := filepath.Join(kbvaultDir, "config.toml")
	err = os.WriteFile(configPath, []byte("[vault]\nname = \"test\""), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	tests := []struct {
		name    string
		workDir string
		want    string
		wantErr bool
	}{
		{
			name:    "from_vault_root",
			workDir: vaultDir,
			want:    vaultDir,
			wantErr: false,
		},
		{
			name:    "from_subdirectory",
			workDir: subDir,
			want:    vaultDir,
			wantErr: false,
		},
		{
			name:    "from_outside_vault",
			workDir: tempDir,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Change to the test directory
			oldWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer func() { _ = os.Chdir(oldWd) }()

			err = os.Chdir(tt.workDir)
			if err != nil {
				t.Fatalf("Failed to change directory: %v", err)
			}

			got, err := findVaultRoot()
			if (err != nil) != tt.wantErr {
				t.Errorf("findVaultRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Resolve both paths to handle symlinks on macOS
				gotResolved, err1 := filepath.EvalSymlinks(got)
				wantResolved, err2 := filepath.EvalSymlinks(tt.want)
				if err1 == nil && err2 == nil {
					if gotResolved != wantResolved {
						t.Errorf("findVaultRoot() = %v (resolved: %v), want %v (resolved: %v)", got, gotResolved, tt.want, wantResolved)
					}
				} else if got != tt.want {
					t.Errorf("findVaultRoot() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary vault structure
	tempDir := t.TempDir()
	vaultDir := filepath.Join(tempDir, "test-vault")
	kbvaultDir := filepath.Join(vaultDir, ".kbvault")
	
	err := os.MkdirAll(kbvaultDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .kbvault directory: %v", err)
	}

	// Create config file
	configContent := `[vault]
name = "test-vault"
notes_dir = "notes"
max_file_size = 1048576

[storage]
type = "local"

[storage.local]
path = "./data"
lock_timeout = 30

[server.http]
host = "localhost"
port = 8080
read_timeout = 30

[logging]
level = "info"
output = "stdout"
`
	configPath := filepath.Join(kbvaultDir, "config.toml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Change to vault directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	err = os.Chdir(vaultDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test loading config - skip this test as we now use profile-aware configuration
	t.Skip("loadConfig test skipped - replaced with profile-aware configuration")
}

func TestLoadConfigNotInVault(t *testing.T) {
	// Change to a directory that's not a vault
	tempDir := t.TempDir()
	
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Test loading config should fail - skip this test as we now use profile-aware configuration
	t.Skip("loadConfig test skipped - replaced with profile-aware configuration")
}