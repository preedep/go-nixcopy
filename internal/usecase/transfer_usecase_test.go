package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
	"go.uber.org/zap"
)

func TestTransferUseCase_Transfer_Success(t *testing.T) {
	// Setup
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := zap.NewNop()

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
	logger := zap.NewNop()

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
	logger := zap.NewNop()

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

func TestTransferUseCase_TransferBatch_PartialFailure(t *testing.T) {
	// Setup
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := zap.NewNop()

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
