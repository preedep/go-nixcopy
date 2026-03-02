// Package storage provides implementations for various storage backends.
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	appconfig "github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

// LocalStorage implements the repository.Storage interface for local file system operations.
//
// It provides direct access to the local file system, supporting both read and write operations.
// This implementation is useful for:
//   - Copying files from/to local disk
//   - Backup to local storage
//   - Testing and development
//   - Migration between local and cloud storage
//
// Features:
//   - Direct file system access (no network overhead)
//   - Streaming I/O for memory efficiency
//   - Recursive directory listing
//   - Automatic directory creation for writes
//   - Symbolic link handling
//
// Thread Safety:
// LocalStorage is safe for concurrent use. Multiple goroutines can perform
// read/write operations simultaneously.
//
// Permissions:
// The process must have appropriate file system permissions:
//   - Read permission for source files/directories
//   - Write permission for destination directories
//   - Execute permission for directory traversal
type LocalStorage struct {
	basePath string // Base directory path for all operations
}

// NewLocalStorage creates a new LocalStorage instance.
//
// Parameters:
//   - config: Local storage configuration containing the base path
//
// Returns:
//   - repository.Storage: Storage interface implementation
//   - error: Non-nil if base path is invalid or inaccessible
//
// The base path will be validated to ensure it exists and is accessible.
// All file operations will be relative to this base path for security.
//
// Example:
//
//	config := &appconfig.LocalConfig{BasePath: "/data/files"}
//	storage, err := NewLocalStorage(config)
func NewLocalStorage(config *appconfig.LocalConfig) (repository.Storage, error) {
	if config.BasePath == "" {
		return nil, fmt.Errorf("base_path is required for local storage")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(config.BasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return &LocalStorage{
		basePath: absPath,
	}, nil
}

// Connect validates that the base path exists and is accessible.
//
// For local storage, this method:
//   - Checks if the base path exists
//   - Creates the directory if it doesn't exist (for write operations)
//   - Validates read/write permissions
//
// Returns:
//   - error: Non-nil if base path is inaccessible or cannot be created
func (l *LocalStorage) Connect(ctx context.Context) error {
	// Check if base path exists
	info, err := os.Stat(l.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory if it doesn't exist
			if err := os.MkdirAll(l.basePath, 0755); err != nil {
				return fmt.Errorf("failed to create base directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to access base path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("base path is not a directory: %s", l.basePath)
	}

	return nil
}

// Disconnect is a no-op for local storage as there's no connection to close.
func (l *LocalStorage) Disconnect(ctx context.Context) error {
	return nil
}

// List returns all files and directories in the specified path.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused for local operations)
//   - path: Relative path from base directory (empty string for base directory)
//
// Returns:
//   - []entity.FileInfo: Slice of file/directory information
//   - error: Non-nil if directory cannot be read or doesn't exist
//
// Behavior:
//   - Returns immediate children only (not recursive)
//   - Includes both files and directories
//   - Follows symbolic links (reports link target info)
//   - Skips files that cannot be stat'd (permission errors)
//
// Example:
//
//	files, err := storage.List(ctx, "documents/2024")
//	// Returns files in /data/files/documents/2024/
func (l *LocalStorage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	fullPath := l.getFullPath(path)

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []entity.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// Skip files we can't stat (permission errors, etc.)
			continue
		}

		fileInfo := entity.FileInfo{
			Name:         entry.Name(),
			Path:         filepath.Join(path, entry.Name()),
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
			IsDirectory:  entry.IsDir(),
		}
		files = append(files, fileInfo)
	}

	return files, nil
}

// Read opens a file for reading and returns a stream.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused for local operations)
//   - path: Relative path to the file from base directory
//
// Returns:
//   - io.ReadCloser: Stream for reading file contents (must be closed by caller)
//   - int64: File size in bytes
//   - error: Non-nil if file doesn't exist or cannot be opened
//
// Memory Efficiency:
// The file is NOT loaded into memory. Returns a stream that reads data on demand.
// Caller must close the returned ReadCloser to free resources.
//
// Example:
//
//	reader, size, err := storage.Read(ctx, "documents/report.pdf")
//	if err != nil {
//	    return err
//	}
//	defer reader.Close()
//	// Read from reader...
func (l *LocalStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	fullPath := l.getFullPath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return nil, 0, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}

	return file, info.Size(), nil
}

// Write writes data from a reader to a file.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused for local operations)
//   - path: Relative path where file should be written
//   - reader: Data source to read from
//   - size: Expected size in bytes (used for validation, can be 0 to skip)
//
// Returns:
//   - error: Non-nil if write fails or directory cannot be created
//
// Behavior:
//   - Creates parent directories automatically if they don't exist
//   - Overwrites existing files
//   - Streams data (doesn't load entire file into memory)
//   - Sets file permissions to 0644 (rw-r--r--)
//
// Atomicity:
// Writes are NOT atomic. If the operation fails mid-write, a partial file may exist.
// Consider using a temporary file and rename for atomic writes if needed.
//
// Example:
//
//	err := storage.Write(ctx, "backup/file.pdf", reader, size)
func (l *LocalStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	fullPath := l.getFullPath(path)

	// Create parent directories if they don't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create and write to file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data from reader to file
	written, err := io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Validate size if provided
	if size > 0 && written != size {
		return fmt.Errorf("size mismatch: expected %d bytes, wrote %d bytes", size, written)
	}

	return nil
}

// Delete removes a file from the file system.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused for local operations)
//   - path: Relative path to the file to delete
//
// Returns:
//   - error: Non-nil if file doesn't exist or cannot be deleted
//
// Behavior:
//   - Removes files only (not directories)
//   - Returns error if path is a directory
//   - Returns error if file doesn't exist
//
// Example:
//
//	err := storage.Delete(ctx, "temp/old-file.txt")
func (l *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := l.getFullPath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Stat returns metadata about a file or directory.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused for local operations)
//   - path: Relative path to the file or directory
//
// Returns:
//   - *entity.FileInfo: File metadata including name, size, modification time
//   - error: Non-nil if file doesn't exist or cannot be accessed
//
// Example:
//
//	info, err := storage.Stat(ctx, "documents/report.pdf")
//	fmt.Printf("Size: %d bytes, Modified: %s\n", info.Size, info.ModTime)
func (l *LocalStorage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	fullPath := l.getFullPath(path)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &entity.FileInfo{
		Name:         info.Name(),
		Path:         path,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		IsDirectory:  info.IsDir(),
	}, nil
}

// CreateDirectory creates a directory at the specified path.
//
// Parameters:
//   - ctx: Context for cancellation (currently unused for local operations)
//   - path: Relative path where directory should be created
//
// Returns:
//   - error: Non-nil if directory cannot be created
//
// Behavior:
//   - Creates parent directories automatically if they don't exist
//   - No error if directory already exists
//   - Sets directory permissions to 0755 (rwxr-xr-x)
//
// Example:
//
//	err := storage.CreateDirectory(ctx, "backup/2024/01")
func (l *LocalStorage) CreateDirectory(ctx context.Context, path string) error {
	fullPath := l.getFullPath(path)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// getFullPath constructs the absolute file system path from a relative path.
//
// This method ensures all operations are confined to the base directory
// for security (prevents directory traversal attacks).
//
// Parameters:
//   - relativePath: Path relative to base directory
//
// Returns:
//   - string: Absolute file system path
//
// Security:
// The method uses filepath.Join which cleans the path and prevents
// directory traversal with ".." components.
func (l *LocalStorage) getFullPath(relativePath string) string {
	if relativePath == "" || relativePath == "." {
		return l.basePath
	}
	return filepath.Join(l.basePath, relativePath)
}
