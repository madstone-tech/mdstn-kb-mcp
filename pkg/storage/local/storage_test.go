package local

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestNew(t *testing.T) {
	tempDir := t.TempDir()
	
	config := types.LocalStorageConfig{
		Path:          tempDir,
		CreateDirs:    true,
		EnableLocking: true,
		LockTimeout:   5,
		DirPerms:      "0755",
		FilePerms:     "0644",
	}

	storage, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if storage.Type() != types.StorageTypeLocal {
		t.Errorf("Expected storage type 'local', got %s", storage.Type())
	}
}

func TestNew_Defaults(t *testing.T) {
	tempDir := t.TempDir()
	
	config := types.LocalStorageConfig{
		Path:          tempDir,
		CreateDirs:    true,
		EnableLocking: true,
	}

	storage, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if storage.config.DirPerms != "0755" {
		t.Errorf("Expected default dir perms '0755', got %s", storage.config.DirPerms)
	}

	if storage.config.FilePerms != "0644" {
		t.Errorf("Expected default file perms '0644', got %s", storage.config.FilePerms)
	}

	if storage.config.LockTimeout != 10 {
		t.Errorf("Expected default lock timeout 10, got %d", storage.config.LockTimeout)
	}
}

func TestStorage_WriteAndRead(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	testData := []byte("Hello, World!")
	testPath := "test/file.md"

	// Write data
	err := storage.Write(ctx, testPath, testData)
	if err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	// Read data back
	readData, err := storage.Read(ctx, testPath)
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("Data mismatch: expected %s, got %s", testData, readData)
	}
}

func TestStorage_WriteAndReadStream(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	testData := "This is a test file with multiple lines\nLine 2\nLine 3"
	testPath := "stream/test.md"

	// Write using stream
	reader := strings.NewReader(testData)
	err := storage.WriteStream(ctx, testPath, reader)
	if err != nil {
		t.Fatalf("Failed to write stream: %v", err)
	}

	// Read using stream
	readStream, err := storage.ReadStream(ctx, testPath)
	if err != nil {
		t.Fatalf("Failed to read stream: %v", err)
	}
	defer readStream.Close()

	readData, err := io.ReadAll(readStream)
	if err != nil {
		t.Fatalf("Failed to read from stream: %v", err)
	}

	if string(readData) != testData {
		t.Errorf("Stream data mismatch: expected %s, got %s", testData, readData)
	}
}

func TestStorage_Delete(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	testPath := "delete/test.md"

	// Write a file
	err := storage.Write(ctx, testPath, []byte("test data"))
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify it exists
	exists, err := storage.Exists(ctx, testPath)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Fatal("File should exist before deletion")
	}

	// Delete the file
	err = storage.Delete(ctx, testPath)
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify it no longer exists
	exists, err = storage.Exists(ctx, testPath)
	if err != nil {
		t.Fatalf("Failed to check existence after deletion: %v", err)
	}
	if exists {
		t.Fatal("File should not exist after deletion")
	}
}

func TestStorage_Exists(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()

	// Test non-existent file
	exists, err := storage.Exists(ctx, "nonexistent.md")
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Non-existent file should not exist")
	}

	// Create a file and test again
	testPath := "exists/test.md"
	err = storage.Write(ctx, testPath, []byte("test"))
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	exists, err = storage.Exists(ctx, testPath)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Existing file should exist")
	}
}

func TestStorage_List(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()

	// Create test files
	testFiles := []string{
		"list/test1.md",
		"list/test2.md", 
		"list/test3.txt",
		"other/file.md",
	}

	for _, file := range testFiles {
		err := storage.Write(ctx, file, []byte("test content"))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// List files with prefix "list/"
	files, err := storage.List(ctx, "list/")
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	// Should find 3 files in list/ directory
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d: %v", len(files), files)
	}

	// Check for specific files
	expectedFiles := []string{"list/test1.md", "list/test2.md", "list/test3.txt"}
	for _, expected := range expectedFiles {
		found := false
		for _, actual := range files {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %s not found in list: %v", expected, files)
		}
	}
}

func TestStorage_Stat(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	testData := []byte("test data for stat")
	testPath := "stat/test.md"

	// Write test file
	err := storage.Write(ctx, testPath, testData)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Get file info
	info, err := storage.Stat(ctx, testPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Path != testPath {
		t.Errorf("Expected path %s, got %s", testPath, info.Path)
	}

	if info.Size != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), info.Size)
	}

	if info.ContentType != "text/markdown" {
		t.Errorf("Expected content type 'text/markdown', got %s", info.ContentType)
	}

	// ModTime should be recent
	now := time.Now().Unix()
	if info.ModTime < now-10 || info.ModTime > now+10 {
		t.Errorf("ModTime seems incorrect: %d (now: %d)", info.ModTime, now)
	}
}

func TestStorage_Copy(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	testData := []byte("data to copy")
	srcPath := "copy/source.md"
	dstPath := "copy/destination.md"

	// Write source file
	err := storage.Write(ctx, srcPath, testData)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Copy file
	err = storage.Copy(ctx, srcPath, dstPath)
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	// Verify both files exist and have same content
	srcData, err := storage.Read(ctx, srcPath)
	if err != nil {
		t.Fatalf("Failed to read source after copy: %v", err)
	}

	dstData, err := storage.Read(ctx, dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination after copy: %v", err)
	}

	if string(srcData) != string(dstData) {
		t.Errorf("Copy data mismatch: src=%s, dst=%s", srcData, dstData)
	}
}

func TestStorage_Move(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	testData := []byte("data to move")
	srcPath := "move/source.md"
	dstPath := "move/destination.md"

	// Write source file
	err := storage.Write(ctx, srcPath, testData)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Move file
	err = storage.Move(ctx, srcPath, dstPath)
	if err != nil {
		t.Fatalf("Failed to move file: %v", err)
	}

	// Source should not exist
	exists, err := storage.Exists(ctx, srcPath)
	if err != nil {
		t.Fatalf("Failed to check source existence: %v", err)
	}
	if exists {
		t.Error("Source file should not exist after move")
	}

	// Destination should exist with correct data
	dstData, err := storage.Read(ctx, dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination after move: %v", err)
	}

	if string(dstData) != string(testData) {
		t.Errorf("Move data mismatch: expected=%s, got=%s", testData, dstData)
	}
}

func TestStorage_Health(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()

	// Health check should pass
	err := storage.Health(ctx)
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

func TestStorage_Close(t *testing.T) {
	storage := createTestStorage(t)
	
	ctx := context.Background()

	// Should work before close
	err := storage.Write(ctx, "test.md", []byte("test"))
	if err != nil {
		t.Fatalf("Write should work before close: %v", err)
	}

	// Close storage
	err = storage.Close()
	if err != nil {
		t.Fatalf("Failed to close storage: %v", err)
	}

	// Operations should fail after close
	err = storage.Write(ctx, "test2.md", []byte("test"))
	if err == nil {
		t.Error("Write should fail after close")
	}

	// Subsequent closes should not error
	err = storage.Close()
	if err != nil {
		t.Errorf("Second close should not error: %v", err)
	}
}

func TestStorage_ConcurrentAccess(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	const numGoroutines = 10
	const numOperations = 5

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Launch concurrent writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			for j := 0; j < numOperations; j++ {
				path := fmt.Sprintf("concurrent/file_%d_%d.md", id, j)
				data := []byte(fmt.Sprintf("data from goroutine %d, operation %d", id, j))
				
				if err := storage.Write(ctx, path, data); err != nil {
					errors <- fmt.Errorf("write error goroutine %d: %w", id, err)
					return
				}
				
				// Verify read
				readData, err := storage.Read(ctx, path)
				if err != nil {
					errors <- fmt.Errorf("read error goroutine %d: %w", id, err)
					return
				}
				
				if string(readData) != string(data) {
					errors <- fmt.Errorf("data mismatch goroutine %d", id)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestStorage_PathSecurity(t *testing.T) {
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()

	// Test directory traversal attempts
	maliciousPaths := []string{
		"../../../etc/passwd",
		"/etc/passwd",
		"test/../../../etc/passwd",
		"test/../../outside.txt",
	}

	for _, path := range maliciousPaths {
		err := storage.Write(ctx, path, []byte("malicious content"))
		if err != nil {
			// Error is expected and good for security
			continue
		}

		// If write succeeded, check that it's contained within storage root
		fullPath := storage.getFullPath(path)
		if !strings.HasPrefix(fullPath, storage.config.Path) {
			t.Errorf("Path %s escaped storage root: %s", path, fullPath)
		}
	}
}

func TestStorage_ErrorHandling(t *testing.T) {
	// Test with non-existent directory and CreateDirs=false
	tempDir := filepath.Join(t.TempDir(), "nonexistent")
	
	config := types.LocalStorageConfig{
		Path:       tempDir,
		CreateDirs: false, // Don't create directories
	}

	_, err := New(config)
	if err == nil {
		t.Error("Should fail when directory doesn't exist and CreateDirs=false")
	}

	// Test read from non-existent file
	storage := createTestStorage(t)
	defer storage.Close()
	
	ctx := context.Background()
	_, err = storage.Read(ctx, "nonexistent.md")
	
	if err == nil {
		t.Error("Should error when reading non-existent file")
		return
	}

	// Should be a StorageError
	if storageErr, ok := err.(*types.StorageError); ok {
		if storageErr.Backend != types.StorageTypeLocal {
			t.Errorf("Error should indicate local backend, got %s", storageErr.Backend)
		}
		if storageErr.Operation != "read" {
			t.Errorf("Error should indicate read operation, got %s", storageErr.Operation)
		}
	} else {
		t.Errorf("Should return StorageError, got %T: %v", err, err)
	}
}

func TestStorage_InvalidPermissions(t *testing.T) {
	tempDir := t.TempDir()
	
	config := types.LocalStorageConfig{
		Path:      tempDir,
		DirPerms:  "invalid",
		FilePerms: "0644",
	}

	_, err := New(config)
	if err == nil {
		t.Error("Should fail with invalid directory permissions")
	}

	config.DirPerms = "0755"
	config.FilePerms = "invalid"

	_, err = New(config)
	if err == nil {
		t.Error("Should fail with invalid file permissions")
	}
}

// Helper function to create a test storage instance
func createTestStorage(t *testing.T) *Storage {
	tempDir := t.TempDir()
	
	config := types.LocalStorageConfig{
		Path:          tempDir,
		CreateDirs:    true,
		EnableLocking: true,
		LockTimeout:   5,
		DirPerms:      "0755",
		FilePerms:     "0644",
	}

	storage, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create test storage: %v", err)
	}

	return storage
}

// Benchmark tests
func BenchmarkStorage_Write(b *testing.B) {
	storage := createBenchStorage(b)
	defer storage.Close()
	
	ctx := context.Background()
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("bench/write_%d.md", i)
		if err := storage.Write(ctx, path, data); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
	}
}

func BenchmarkStorage_Read(b *testing.B) {
	storage := createBenchStorage(b)
	defer storage.Close()
	
	ctx := context.Background()
	data := []byte("benchmark test data")
	path := "bench/read_test.md"

	// Setup
	if err := storage.Write(ctx, path, data); err != nil {
		b.Fatalf("Setup write failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := storage.Read(ctx, path); err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

func createBenchStorage(b *testing.B) *Storage {
	tempDir := b.TempDir()
	
	config := types.LocalStorageConfig{
		Path:          tempDir,
		CreateDirs:    true,
		EnableLocking: false, // Disable locking for benchmarks
		DirPerms:      "0755",
		FilePerms:     "0644",
	}

	storage, err := New(config)
	if err != nil {
		b.Fatalf("Failed to create benchmark storage: %v", err)
	}

	return storage
}