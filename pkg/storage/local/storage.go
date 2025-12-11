package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Storage implements the StorageBackend interface for local filesystem storage
type Storage struct {
	config    types.LocalStorageConfig
	locks     map[string]*sync.Mutex
	lockMutex sync.RWMutex
	closed    bool
	closeMux  sync.RWMutex
}

// New creates a new local storage backend
func New(config types.LocalStorageConfig) (*Storage, error) {
	// Set defaults if not provided
	if config.DirPerms == "" {
		config.DirPerms = "0755"
	}
	if config.FilePerms == "" {
		config.FilePerms = "0644"
	}
	if config.LockTimeout == 0 {
		config.LockTimeout = 10 // 10 seconds default
	}

	// Validate permissions
	if _, err := strconv.ParseUint(config.DirPerms, 8, 32); err != nil {
		return nil, fmt.Errorf("invalid directory permissions: %s", config.DirPerms)
	}
	if _, err := strconv.ParseUint(config.FilePerms, 8, 32); err != nil {
		return nil, fmt.Errorf("invalid file permissions: %s", config.FilePerms)
	}

	storage := &Storage{
		config: config,
		locks:  make(map[string]*sync.Mutex),
	}

	// Create root directory if it doesn't exist and CreateDirs is enabled
	if config.CreateDirs {
		if err := storage.ensureDir(config.Path); err != nil {
			return nil, fmt.Errorf("failed to create root directory: %w", err)
		}
	}

	// Verify the root directory exists and is accessible
	if err := storage.Health(context.Background()); err != nil {
		return nil, fmt.Errorf("storage health check failed: %w", err)
	}

	return storage, nil
}

// Type returns the storage backend type
func (s *Storage) Type() types.StorageType {
	return types.StorageTypeLocal
}

// Read retrieves a file's content by path
func (s *Storage) Read(ctx context.Context, path string) ([]byte, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}

	fullPath := s.getFullPath(path)

	if s.config.EnableLocking {
		unlock, err := s.lockFile(ctx, fullPath, unix.LOCK_SH)
		if err != nil {
			return nil, types.NewStorageError(s.Type(), "read", path, err, true)
		}
		defer unlock()
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.NewStorageError(s.Type(), "read", path, err, false)
		}
		return nil, types.NewStorageError(s.Type(), "read", path, err, true)
	}

	return data, nil
}

// Write stores content at the given path
func (s *Storage) Write(ctx context.Context, path string, data []byte) error {
	if err := s.checkClosed(); err != nil {
		return err
	}

	fullPath := s.getFullPath(path)

	// Ensure directory exists
	if s.config.CreateDirs {
		if err := s.ensureDir(filepath.Dir(fullPath)); err != nil {
			return types.NewStorageError(s.Type(), "write", path, err, true)
		}
	}

	// Use atomic write with temp file
	tempPath := fullPath + ".tmp." + strconv.FormatInt(time.Now().UnixNano(), 10)

	if s.config.EnableLocking {
		unlock, err := s.lockFile(ctx, fullPath, unix.LOCK_EX)
		if err != nil {
			return types.NewStorageError(s.Type(), "write", path, err, true)
		}
		defer unlock()
	}

	// Write to temp file
	filePerms, err := s.getFilePerms()
	if err != nil {
		return types.NewStorageError(s.Type(), "write", path, err, false)
	}

	if err := os.WriteFile(tempPath, data, filePerms); err != nil {
		return types.NewStorageError(s.Type(), "write", path, err, true)
	}

	// Atomic rename
	if err := os.Rename(tempPath, fullPath); err != nil {
		_ = os.Remove(tempPath) // Clean up temp file (ignore error as we're already handling one)
		return types.NewStorageError(s.Type(), "write", path, err, true)
	}

	return nil
}

// Delete removes a file at the given path
func (s *Storage) Delete(ctx context.Context, path string) error {
	if err := s.checkClosed(); err != nil {
		return err
	}

	fullPath := s.getFullPath(path)

	if s.config.EnableLocking {
		unlock, err := s.lockFile(ctx, fullPath, unix.LOCK_EX)
		if err != nil {
			return types.NewStorageError(s.Type(), "delete", path, err, true)
		}
		defer unlock()
	}

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return types.NewStorageError(s.Type(), "delete", path, err, false)
		}
		return types.NewStorageError(s.Type(), "delete", path, err, true)
	}

	return nil
}

// Exists checks if a file exists at the given path
func (s *Storage) Exists(ctx context.Context, path string) (bool, error) {
	if err := s.checkClosed(); err != nil {
		return false, err
	}

	fullPath := s.getFullPath(path)
	_, err := os.Stat(fullPath)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, types.NewStorageError(s.Type(), "exists", path, err, true)
}

// List returns all files matching the given prefix
func (s *Storage) List(ctx context.Context, prefix string) ([]string, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}

	// For prefix like "list/", we want to list all files in that directory
	prefixPath := s.getFullPath(prefix)

	// If prefix ends with "/", treat it as a directory
	var searchDir string
	var namePrefix string

	if strings.HasSuffix(prefix, "/") {
		searchDir = prefixPath
		namePrefix = ""
	} else {
		searchDir = filepath.Dir(prefixPath)
		namePrefix = filepath.Base(prefixPath)
	}

	entries, err := os.ReadDir(searchDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, types.NewStorageError(s.Type(), "list", prefix, err, true)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if namePrefix == "" || strings.HasPrefix(entry.Name(), namePrefix) {
			// Return relative path from storage root
			relPath, err := filepath.Rel(s.config.Path, filepath.Join(searchDir, entry.Name()))
			if err != nil {
				continue
			}
			matches = append(matches, relPath)
		}
	}

	return matches, nil
}

// Stat returns metadata about a file
func (s *Storage) Stat(ctx context.Context, path string) (*types.FileInfo, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}

	fullPath := s.getFullPath(path)
	stat, err := os.Stat(fullPath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.NewStorageError(s.Type(), "stat", path, err, false)
		}
		return nil, types.NewStorageError(s.Type(), "stat", path, err, true)
	}

	fileInfo := &types.FileInfo{
		Path:        path,
		Size:        stat.Size(),
		ModTime:     stat.ModTime().Unix(),
		ContentType: "text/markdown", // Default for .md files
	}

	return fileInfo, nil
}

// ReadStream returns a reader for streaming large files
func (s *Storage) ReadStream(ctx context.Context, path string) (io.ReadCloser, error) {
	if err := s.checkClosed(); err != nil {
		return nil, err
	}

	fullPath := s.getFullPath(path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.NewStorageError(s.Type(), "read_stream", path, err, false)
		}
		return nil, types.NewStorageError(s.Type(), "read_stream", path, err, true)
	}

	// Note: File locking for streams would need to be handled differently
	// as we need to maintain the lock for the duration of the stream
	return file, nil
}

// WriteStream writes data from a reader to the given path
func (s *Storage) WriteStream(ctx context.Context, path string, reader io.Reader) error {
	if err := s.checkClosed(); err != nil {
		return err
	}

	fullPath := s.getFullPath(path)

	// Ensure directory exists
	if s.config.CreateDirs {
		if err := s.ensureDir(filepath.Dir(fullPath)); err != nil {
			return types.NewStorageError(s.Type(), "write_stream", path, err, true)
		}
	}

	// Use atomic write with temp file
	tempPath := fullPath + ".tmp." + strconv.FormatInt(time.Now().UnixNano(), 10)

	if s.config.EnableLocking {
		unlock, err := s.lockFile(ctx, fullPath, unix.LOCK_EX)
		if err != nil {
			return types.NewStorageError(s.Type(), "write_stream", path, err, true)
		}
		defer unlock()
	}

	filePerms, err := s.getFilePerms()
	if err != nil {
		return types.NewStorageError(s.Type(), "write_stream", path, err, false)
	}

	// Create temp file
	tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerms)
	if err != nil {
		return types.NewStorageError(s.Type(), "write_stream", path, err, true)
	}

	// Copy data from reader to temp file
	_, err = io.Copy(tempFile, reader)
	_ = tempFile.Close() // Close temp file (ignore close error)

	if err != nil {
		_ = os.Remove(tempPath) // Clean up on error (ignore removal error)
		return types.NewStorageError(s.Type(), "write_stream", path, err, true)
	}

	// Atomic rename
	if err := os.Rename(tempPath, fullPath); err != nil {
		_ = os.Remove(tempPath) // Clean up on error (ignore removal error)
		return types.NewStorageError(s.Type(), "write_stream", path, err, true)
	}

	return nil
}

// Copy copies a file from src to dst within the same backend
func (s *Storage) Copy(ctx context.Context, src, dst string) error {
	if err := s.checkClosed(); err != nil {
		return err
	}

	// Read source file
	data, err := s.Read(ctx, src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination
	if err := s.Write(ctx, dst, data); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// Move moves/renames a file from src to dst
func (s *Storage) Move(ctx context.Context, src, dst string) error {
	if err := s.checkClosed(); err != nil {
		return err
	}

	srcPath := s.getFullPath(src)
	dstPath := s.getFullPath(dst)

	// Ensure destination directory exists
	if s.config.CreateDirs {
		if err := s.ensureDir(filepath.Dir(dstPath)); err != nil {
			return types.NewStorageError(s.Type(), "move", src+" -> "+dst, err, true)
		}
	}

	if s.config.EnableLocking {
		// Lock both files
		unlockSrc, err := s.lockFile(ctx, srcPath, unix.LOCK_EX)
		if err != nil {
			return types.NewStorageError(s.Type(), "move", src+" -> "+dst, err, true)
		}
		defer unlockSrc()

		unlockDst, err := s.lockFile(ctx, dstPath, unix.LOCK_EX)
		if err != nil {
			return types.NewStorageError(s.Type(), "move", src+" -> "+dst, err, true)
		}
		defer unlockDst()
	}

	err := os.Rename(srcPath, dstPath)
	if err != nil {
		return types.NewStorageError(s.Type(), "move", src+" -> "+dst, err, true)
	}

	return nil
}

// Health performs a health check on the storage backend
func (s *Storage) Health(ctx context.Context) error {
	if err := s.checkClosed(); err != nil {
		return err
	}

	// Check if root directory exists and is accessible
	stat, err := os.Stat(s.config.Path)
	if err != nil {
		return fmt.Errorf("storage root not accessible: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("storage root is not a directory: %s", s.config.Path)
	}

	// Test write permissions by creating a temp file
	tempPath := filepath.Join(s.config.Path, ".health_check_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("storage not writable: %w", err)
	}
	_ = file.Close()        // Ignore close error for temp file
	_ = os.Remove(tempPath) // Ignore removal error for temp file

	return nil
}

// Close cleanly shuts down the storage backend
func (s *Storage) Close() error {
	s.closeMux.Lock()
	defer s.closeMux.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Clear lock map
	s.lockMutex.Lock()
	s.locks = make(map[string]*sync.Mutex)
	s.lockMutex.Unlock()

	return nil
}

// Helper methods

func (s *Storage) checkClosed() error {
	s.closeMux.RLock()
	defer s.closeMux.RUnlock()

	if s.closed {
		return fmt.Errorf("storage is closed")
	}
	return nil
}

func (s *Storage) getFullPath(path string) string {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	// Remove leading slash if present
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	// Ensure the path is relative and doesn't escape
	if strings.Contains(cleanPath, "..") {
		// Further sanitize by removing any remaining ".." components
		parts := strings.Split(cleanPath, string(filepath.Separator))
		var sanitized []string
		for _, part := range parts {
			if part != ".." && part != "." && part != "" {
				sanitized = append(sanitized, part)
			}
		}
		cleanPath = strings.Join(sanitized, string(filepath.Separator))
	}

	return filepath.Join(s.config.Path, cleanPath)
}

func (s *Storage) ensureDir(dir string) error {
	dirPerms, err := s.getDirPerms()
	if err != nil {
		return err
	}

	return os.MkdirAll(dir, dirPerms)
}

func (s *Storage) getDirPerms() (os.FileMode, error) {
	perms, err := strconv.ParseUint(s.config.DirPerms, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid directory permissions: %s", s.config.DirPerms)
	}
	return os.FileMode(perms), nil
}

func (s *Storage) getFilePerms() (os.FileMode, error) {
	perms, err := strconv.ParseUint(s.config.FilePerms, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid file permissions: %s", s.config.FilePerms)
	}
	return os.FileMode(perms), nil
}

func (s *Storage) lockFile(ctx context.Context, path string, lockType int) (func(), error) {
	// Get or create mutex for this path
	s.lockMutex.Lock()
	mutex, exists := s.locks[path]
	if !exists {
		mutex = &sync.Mutex{}
		s.locks[path] = mutex
	}
	s.lockMutex.Unlock()

	// Acquire mutex with timeout
	done := make(chan bool, 1)
	go func() {
		mutex.Lock()
		done <- true
	}()

	timeout := time.Duration(s.config.LockTimeout) * time.Second
	select {
	case <-done:
		// Mutex acquired, now try to get file lock
		// For read operations, don't create the file if it doesn't exist
		var flags int
		if lockType == unix.LOCK_SH {
			flags = os.O_RDONLY // Read lock - don't create file
		} else {
			flags = os.O_RDONLY | os.O_CREATE // Write lock - create if needed
		}

		file, err := os.OpenFile(path, flags, 0644)
		if err != nil {
			mutex.Unlock()
			return nil, err
		}

		// Try to get flock with timeout
		flockDone := make(chan error, 1)
		go func() {
			flockDone <- unix.Flock(int(file.Fd()), lockType|unix.LOCK_NB)
		}()

		select {
		case err := <-flockDone:
			if err != nil {
				_ = file.Close() // Ignore close error when handling lock error
				mutex.Unlock()
				return nil, fmt.Errorf("failed to acquire file lock: %w", err)
			}

			// Return unlock function
			return func() {
				_ = unix.Flock(int(file.Fd()), unix.LOCK_UN) // Ignore unlock error
				_ = file.Close()                             // Ignore close error in cleanup
				mutex.Unlock()
			}, nil

		case <-time.After(timeout):
			_ = file.Close() // Ignore close error on timeout
			mutex.Unlock()
			return nil, fmt.Errorf("timeout acquiring file lock after %v", timeout)
		}

	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout acquiring mutex lock after %v", timeout)

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
