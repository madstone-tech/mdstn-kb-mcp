package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIIntegration(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "test-kbvault", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build kbvault: %v", err)
	}
	defer func() { _ = os.Remove("test-kbvault") }()

	binaryPath, err := filepath.Abs("test-kbvault")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Run("version command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--version")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run version command: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "kbvault version") {
			t.Errorf("Version output doesn't contain expected text: %s", outputStr)
		}
	})

	t.Run("help command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run help command: %v", err)
		}

		outputStr := string(output)
		expectedCommands := []string{"init", "new", "show", "list", "config"}
		for _, command := range expectedCommands {
			if !strings.Contains(outputStr, command) {
				t.Errorf("Help output doesn't contain command '%s': %s", command, outputStr)
			}
		}
	})

	t.Run("init command help", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "init", "--help")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run init help command: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Initialize a new kbVault") {
			t.Errorf("Init help output doesn't contain expected text: %s", outputStr)
		}
	})

	t.Run("vault initialization workflow", func(t *testing.T) {
		// Create temporary directory for test vault
		testDir := t.TempDir()
		
		// Test vault initialization
		cmd := exec.Command(binaryPath, "init", "--path", testDir, "--name", "test-vault")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to initialize vault: %v, output: %s", err, string(output))
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Initialized kbVault") {
			t.Errorf("Init output doesn't contain success message: %s", outputStr)
		}

		// Check that directories were created
		expectedDirs := []string{".kbvault", "notes", "templates"}
		for _, dir := range expectedDirs {
			dirPath := filepath.Join(testDir, dir)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				t.Errorf("Expected directory %s was not created", dirPath)
			}
		}

		// Check that config file was created
		configPath := filepath.Join(testDir, ".kbvault", "config.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}
	})

	t.Run("config commands", func(t *testing.T) {
		// Create temporary directory for test vault
		testDir := t.TempDir()
		
		// Initialize vault first
		cmd := exec.Command(binaryPath, "init", "--path", testDir, "--name", "test-vault")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize vault: %v", err)
		}

		// Change to vault directory for config commands
		oldDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer func() { _ = os.Chdir(oldDir) }()
		
		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("Failed to change to test directory: %v", err)
		}

		// Test config show
		cmd = exec.Command(binaryPath, "config", "show")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run config show: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "test-vault") {
			t.Errorf("Config show doesn't contain vault name: %s", outputStr)
		}

		// Test config path
		cmd = exec.Command(binaryPath, "config", "path")
		output, err = cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run config path: %v", err)
		}

		pathStr := strings.TrimSpace(string(output))
		expectedPath := filepath.Join(testDir, ".kbvault", "config.toml")
		
		// Resolve both paths to handle symlinks (e.g., /private on macOS)
		resolvedOutput, err := filepath.EvalSymlinks(pathStr)
		if err != nil {
			resolvedOutput = pathStr
		}
		resolvedExpected, err := filepath.EvalSymlinks(expectedPath)
		if err != nil {
			resolvedExpected = expectedPath
		}
		
		if resolvedOutput != resolvedExpected {
			t.Errorf("Config path output doesn't match expected: got %s, expected %s", resolvedOutput, resolvedExpected)
		}
	})

	t.Run("new and show commands", func(t *testing.T) {
		// Create temporary directory for test vault
		testDir := t.TempDir()
		
		// Initialize vault first
		cmd := exec.Command(binaryPath, "init", "--path", testDir, "--name", "test-vault")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to initialize vault: %v", err)
		}

		// Change to vault directory
		oldDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer func() { _ = os.Chdir(oldDir) }()
		
		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("Failed to change to test directory: %v", err)
		}

		// Test note creation (placeholder)
		cmd = exec.Command(binaryPath, "new", "--title", "Test Note", "--tags", "test,integration")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Created note: Test Note") {
			t.Errorf("New note output doesn't contain success message: %s", outputStr)
		}

		// Test list command (placeholder)
		cmd = exec.Command(binaryPath, "list")
		output, err = cmd.Output()
		if err != nil {
			t.Fatalf("Failed to list notes: %v", err)
		}

		outputStr = string(output)
		if !strings.Contains(outputStr, "not yet implemented") {
			t.Errorf("List output doesn't contain expected placeholder: %s", outputStr)
		}
	})
}