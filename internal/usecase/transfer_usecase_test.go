package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// basicStorage wraps MockStorage but does not expose repository.Resumer,
// letting us test the fall-back path when EnableResume is true but the
// backend does not support resume.
type basicStorage struct{ m *mocks.MockStorage }

func (b *basicStorage) Connect(ctx context.Context) error    { return b.m.Connect(ctx) }
func (b *basicStorage) Disconnect(ctx context.Context) error { return b.m.Disconnect(ctx) }
func (b *basicStorage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	return b.m.List(ctx, path)
}
func (b *basicStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	return b.m.Read(ctx, path)
}
func (b *basicStorage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	return b.m.Stat(ctx, path)
}
func (b *basicStorage) Write(ctx context.Context, path string, r io.Reader, size int64) error {
	return b.m.Write(ctx, path, r, size)
}
func (b *basicStorage) CreateDirectory(ctx context.Context, path string) error {
	return b.m.CreateDirectory(ctx, path)
}
func (b *basicStorage) Delete(ctx context.Context, path string) error { return b.m.Delete(ctx, path) }

// Verify basicStorage satisfies Storage but not Resumer.
var _ repository.Storage = (*basicStorage)(nil)

func TestTransferUseCase_Transfer_Success(t *testing.T) {
	// Setup
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	// Add test file to source
	testContent := []byte("test file content")
	source.AddFile("/source/test.txt", testContent, &entity.FileInfo{
		Path:         "/source/test.txt",
		Name:         "test.txt",
		Size:         int64(len(testContent)),
		ModifiedTime: time.Now(),
		IsDirectory:  false,
	})

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		Timeout:         time.Minute,
		VerifyChecksum:  false,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)

	// Execute
	ctx := context.Background()
	result, err := useCase.Transfer(ctx, "/source/test.txt", "/dest/test.txt", nil)

	// Assert
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want %v", result.Status, entity.TransferStatusCompleted)
	}

	if result.BytesTransferred != int64(len(testContent)) {
		t.Errorf("BytesTransferred = %v, want %v", result.BytesTransferred, len(testContent))
	}

	if !source.ReadCalled {
		t.Error("Source.Read() was not called")
	}

	if !dest.WriteCalled {
		t.Error("Dest.Write() was not called")
	}

	// Verify content was written
	writtenContent, ok := dest.FileContent["/dest/test.txt"]
	if !ok {
		t.Fatal("File was not written to destination")
	}

	if string(writtenContent) != string(testContent) {
		t.Errorf("Written content = %v, want %v", string(writtenContent), string(testContent))
	}
}

func TestTransferUseCase_Transfer_SourceNotFound(t *testing.T) {
	// Setup
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   3,
		RetryDelay:      time.Millisecond,
		Timeout:         time.Minute,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)

	// Execute
	ctx := context.Background()
	result, err := useCase.Transfer(ctx, "/source/nonexistent.txt", "/dest/test.txt", nil)

	// Assert
	if err == nil {
		t.Error("Transfer() should return error for nonexistent file")
	}

	if result.Status != entity.TransferStatusFailed {
		t.Errorf("Status = %v, want %v", result.Status, entity.TransferStatusFailed)
	}
}

func TestTransferUseCase_TransferBatch_Success(t *testing.T) {
	// Setup
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	// Add multiple test files
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, filename := range files {
		content := []byte("content of " + filename)
		path := "/source/" + filename
		source.AddFile(path, content, &entity.FileInfo{
			Path:         path,
			Name:         filename,
			Size:         int64(len(content)),
			ModifiedTime: time.Now(),
			IsDirectory:  false,
		})
	}

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 2,
		RetryAttempts:   3,
		RetryDelay:      time.Millisecond,
		Timeout:         time.Minute,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)

	// Execute
	ctx := context.Background()
	sourcePaths := []string{"/source/file1.txt", "/source/file2.txt", "/source/file3.txt"}
	results, err := useCase.TransferBatch(ctx, sourcePaths, "/dest/", nil)

	// Assert
	if err != nil {
		t.Fatalf("TransferBatch() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Results count = %v, want 3", len(results))
	}

	successCount := 0
	for _, result := range results {
		if result.Status == entity.TransferStatusCompleted {
			successCount++
		}
	}

	if successCount != 3 {
		t.Errorf("Success count = %v, want 3", successCount)
	}

	// Verify all files were written
	for _, filename := range files {
		destPath := "/dest/" + filename
		if _, ok := dest.FileContent[destPath]; !ok {
			t.Errorf("File %s was not written to destination", filename)
		}
	}
}

func TestTransferUseCase_Transfer_ChecksumVerification_Success(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	testContent := []byte("checksum test content")
	source.AddFile("/source/file.txt", testContent, &entity.FileInfo{
		Path:         "/source/file.txt",
		Name:         "file.txt",
		Size:         int64(len(testContent)),
		ModifiedTime: time.Now(),
	})

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		VerifyChecksum:  true,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)
	result, err := useCase.Transfer(context.Background(), "/source/file.txt", "/dest/file.txt", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}

	h := sha256.Sum256(testContent)
	want := hex.EncodeToString(h[:])
	if result.Checksum != want {
		t.Errorf("Checksum = %q, want %q", result.Checksum, want)
	}
}

func TestTransferUseCase_Transfer_ChecksumVerification_Mismatch(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	dest.CorruptWrite = true // destination flips first byte — checksum will differ
	logger := applog.NewNopLogger()

	testContent := []byte("checksum test content")
	source.AddFile("/source/file.txt", testContent, &entity.FileInfo{
		Path:         "/source/file.txt",
		Name:         "file.txt",
		Size:         int64(len(testContent)),
		ModifiedTime: time.Now(),
	})

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0, // no retries so test finishes quickly
		RetryDelay:      time.Millisecond,
		VerifyChecksum:  true,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)
	result, err := useCase.Transfer(context.Background(), "/source/file.txt", "/dest/file.txt", nil)

	if err == nil {
		t.Fatal("Transfer() expected checksum mismatch error, got nil")
	}
	if result.Status != entity.TransferStatusFailed {
		t.Errorf("Status = %v, want failed", result.Status)
	}
	if result.Checksum != "" {
		t.Errorf("Checksum should be empty on mismatch, got %q", result.Checksum)
	}
}

func TestTransferUseCase_Transfer_Resume_WithPartialDest(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	fullContent := []byte("Hello, World! This is a test file.")
	partialContent := fullContent[:6] // dest already has "Hello,"

	source.AddFile("/source/file.txt", fullContent, &entity.FileInfo{
		Path:         "/source/file.txt",
		Name:         "file.txt",
		Size:         int64(len(fullContent)),
		ModifiedTime: time.Now(),
	})

	// Pre-seed destination with partial content
	dest.AddFile("/dest/file.txt", partialContent, &entity.FileInfo{
		Path:         "/dest/file.txt",
		Name:         "file.txt",
		Size:         int64(len(partialContent)),
		ModifiedTime: time.Now(),
	})

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)
	result, err := useCase.Transfer(context.Background(), "/source/file.txt", "/dest/file.txt", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	if result.ResumedFrom != int64(len(partialContent)) {
		t.Errorf("ResumedFrom = %d, want %d", result.ResumedFrom, len(partialContent))
	}
	if result.BytesTransferred != int64(len(fullContent)) {
		t.Errorf("BytesTransferred = %d, want %d", result.BytesTransferred, len(fullContent))
	}

	got := string(dest.FileContent["/dest/file.txt"])
	if got != string(fullContent) {
		t.Errorf("dest content = %q, want %q", got, string(fullContent))
	}
}

func TestTransferUseCase_Transfer_Resume_NoPartialDest(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	fullContent := []byte("Hello, World!")
	source.AddFile("/source/file.txt", fullContent, &entity.FileInfo{
		Path:         "/source/file.txt",
		Name:         "file.txt",
		Size:         int64(len(fullContent)),
		ModifiedTime: time.Now(),
	})

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)
	result, err := useCase.Transfer(context.Background(), "/source/file.txt", "/dest/file.txt", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	if result.ResumedFrom != 0 {
		t.Errorf("ResumedFrom = %d, want 0 (full transfer)", result.ResumedFrom)
	}

	got := string(dest.FileContent["/dest/file.txt"])
	if got != string(fullContent) {
		t.Errorf("dest content = %q, want %q", got, string(fullContent))
	}
}

func TestTransferUseCase_Transfer_DirectorySource(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	source.AddFile("/source/mydir", nil, &entity.FileInfo{
		Path:        "/source/mydir",
		Name:        "mydir",
		IsDirectory: true,
	})

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)
	result, err := useCase.Transfer(context.Background(), "/source/mydir", "/dest/mydir", nil)

	if err == nil {
		t.Fatal("Transfer() expected error for directory source, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result for directory error, got %+v", result)
	}
}

func TestTransferUseCase_Transfer_AllRetriesExhausted(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	source.AddFile("/source/file.txt", []byte("content"), &entity.FileInfo{
		Path: "/source/file.txt",
		Name: "file.txt",
		Size: 7,
	})
	source.ReadError = errors.New("simulated read error")

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   2,
		RetryDelay:      time.Millisecond,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)
	result, err := useCase.Transfer(context.Background(), "/source/file.txt", "/dest/file.txt", nil)

	if err == nil {
		t.Fatal("Transfer() expected error after retries exhausted, got nil")
	}
	if result.Status != entity.TransferStatusFailed {
		t.Errorf("Status = %v, want failed", result.Status)
	}
}

func TestTransferUseCase_Transfer_Resume_NonResumerBackend(t *testing.T) {
	inner := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	fullContent := []byte("Hello, World!")
	inner.AddFile("/source/file.txt", fullContent, &entity.FileInfo{
		Path: "/source/file.txt",
		Name: "file.txt",
		Size: int64(len(fullContent)),
	})

	// basicStorage wraps inner but does NOT implement repository.Resumer
	source := &basicStorage{m: inner}

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)
	result, err := useCase.Transfer(context.Background(), "/source/file.txt", "/dest/file.txt", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	if result.ResumedFrom != 0 {
		t.Errorf("ResumedFrom = %d, want 0 (full transfer fallback)", result.ResumedFrom)
	}
	got := string(dest.FileContent["/dest/file.txt"])
	if got != string(fullContent) {
		t.Errorf("dest content = %q, want %q", got, string(fullContent))
	}
}

// Compile-time check: basicStorage must NOT satisfy repository.Resumer.
// If this ever compiles with the assertion below, something has changed.
var _ = func() { var _ repository.Storage = (*basicStorage)(nil) }

func TestTransferUseCase_TransferBatch_PartialFailure(t *testing.T) {
	// Setup
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	// Add only 2 out of 3 files
	source.AddFile("/source/file1.txt", []byte("content1"), &entity.FileInfo{
		Path: "/source/file1.txt",
		Name: "file1.txt",
		Size: 8,
	})
	source.AddFile("/source/file2.txt", []byte("content2"), &entity.FileInfo{
		Path: "/source/file2.txt",
		Name: "file2.txt",
		Size: 8,
	})

	config := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 2,
		RetryAttempts:   1,
		RetryDelay:      time.Millisecond,
		Timeout:         time.Minute,
	}

	useCase := NewTransferUseCase(source, dest, config, logger)

	// Execute - try to transfer 3 files but only 2 exist
	ctx := context.Background()
	sourcePaths := []string{"/source/file1.txt", "/source/file2.txt", "/source/file3.txt"}
	results, err := useCase.TransferBatch(ctx, sourcePaths, "/dest/", nil)

	// Assert
	if err != nil {
		t.Fatalf("TransferBatch() error = %v", err)
	}

	successCount := 0
	failCount := 0
	for _, result := range results {
		if result.Status == entity.TransferStatusCompleted {
			successCount++
		} else if result.Status == entity.TransferStatusFailed {
			failCount++
		}
	}

	if successCount != 2 {
		t.Errorf("Success count = %v, want 2", successCount)
	}

	if failCount != 1 {
		t.Errorf("Fail count = %v, want 1", failCount)
	}
}
