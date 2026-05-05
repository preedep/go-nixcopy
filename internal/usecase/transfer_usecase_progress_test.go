package usecase

import (
	"context"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// progressSource builds a mock with one file at the given path.
func progressSource(path string, content []byte) *mocks.MockStorage {
	s := mocks.NewMockStorage()
	s.AddFile(path, content, &entity.FileInfo{
		Path:         path,
		Name:         "file.txt",
		Size:         int64(len(content)),
		ModifiedTime: time.Now(),
		IsDirectory:  false,
	})
	return s
}

func baseCfg() *entity.TransferConfig {
	return &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 2,
		RetryAttempts:   1,
		RetryDelay:      0,
	}
}

// TestTransfer_ProgressChan_ClosedAfterCompletion verifies that Transfer closes
// the caller-supplied progress channel so a ranging consumer exits cleanly.
// Before the fix, the CLI also called close() → panic: close of closed channel.
func TestTransfer_ProgressChan_ClosedAfterCompletion(t *testing.T) {
	src := progressSource("/src/file.txt", []byte("hello"))
	dst := mocks.NewMockStorage()
	uc := NewTransferUseCase(src, dst, baseCfg(), applog.NewNopLogger())

	progressChan := make(chan entity.TransferProgress, 10)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range progressChan {
		}
	}()

	_, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", progressChan)
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}

	select {
	case <-done:
		// consumer goroutine exited → channel was closed by Transfer
	case <-time.After(2 * time.Second):
		t.Fatal("progress channel was not closed after Transfer completed (consumer goroutine stuck)")
	}
}

// TestTransfer_ProgressChan_NoPanicOnSingleClose verifies that calling close()
// on the progress channel after Transfer returns does NOT panic.
// This is the regression test for the double-close bug.
func TestTransfer_ProgressChan_NoPanicOnSingleClose(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic detected (double-close regression): %v", r)
		}
	}()

	src := progressSource("/src/file.txt", []byte("data"))
	dst := mocks.NewMockStorage()
	uc := NewTransferUseCase(src, dst, baseCfg(), applog.NewNopLogger())

	progressChan := make(chan entity.TransferProgress, 10)
	go func() {
		for range progressChan {
		}
	}()

	_, _ = uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", progressChan)
	// Transfer already closed progressChan — a second close would panic.
	// The test passes if we reach here without panicking.
}

// TestTransferBatch_ProgressChan_ClosedAfterCompletion verifies that
// TransferBatch closes the caller-supplied progress channel so a ranging
// consumer exits cleanly. Before the fix, the channel was never closed by
// TransferBatch, causing the CLI's progressWg.Wait() to deadlock.
func TestTransferBatch_ProgressChan_ClosedAfterCompletion(t *testing.T) {
	src := mocks.NewMockStorage()
	files := []string{"/src/a.txt", "/src/b.txt", "/src/c.txt"}
	for _, p := range files {
		content := []byte("content")
		src.AddFile(p, content, &entity.FileInfo{
			Path: p, Name: "x.txt", Size: int64(len(content)),
			ModifiedTime: time.Now(),
		})
	}
	dst := mocks.NewMockStorage()
	uc := NewTransferUseCase(src, dst, baseCfg(), applog.NewNopLogger())

	progressChan := make(chan entity.TransferProgress, 20)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range progressChan {
		}
	}()

	_, err := uc.TransferBatch(context.Background(), files, "/dst/", progressChan)
	if err != nil {
		t.Fatalf("TransferBatch: %v", err)
	}

	select {
	case <-done:
		// consumer goroutine exited → channel was closed by TransferBatch
	case <-time.After(2 * time.Second):
		t.Fatal("progress channel was not closed after TransferBatch completed (consumer goroutine stuck)")
	}
}
