package usecase

import (
	"context"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

func TestTransferUseCase_SkipExisting_SameSize(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	log := applog.NewNopLogger()

	content := []byte("hello idempotent world")
	source.AddFile("/src/file.txt", content, &entity.FileInfo{
		Path: "/src/file.txt", Name: "file.txt",
		Size: int64(len(content)), ModifiedTime: time.Now(),
	})
	// Pre-populate destination with same size — simulates a completed prior run
	dest.AddFile("/dst/file.txt", content, &entity.FileInfo{
		Path: "/dst/file.txt", Name: "file.txt",
		Size: int64(len(content)), ModifiedTime: time.Now(),
	})

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 1, RetryAttempts: 0,
		SkipExisting: true,
	}
	uc := NewTransferUseCase(source, dest, cfg, log)

	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entity.TransferStatusSkipped {
		t.Errorf("Status = %q, want %q", result.Status, entity.TransferStatusSkipped)
	}
	if dest.WriteCalled {
		t.Error("Write was called — file should have been skipped without touching destination")
	}
}

func TestTransferUseCase_SkipExisting_DifferentSize_Retransfers(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	log := applog.NewNopLogger()

	srcContent := []byte("updated content that is longer")
	source.AddFile("/src/file.txt", srcContent, &entity.FileInfo{
		Path: "/src/file.txt", Name: "file.txt",
		Size: int64(len(srcContent)), ModifiedTime: time.Now(),
	})
	// Destination has a file but with a different size (e.g. partial/stale)
	oldContent := []byte("old")
	dest.AddFile("/dst/file.txt", oldContent, &entity.FileInfo{
		Path: "/dst/file.txt", Name: "file.txt",
		Size: int64(len(oldContent)), ModifiedTime: time.Now(),
	})

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 1, RetryAttempts: 0,
		SkipExisting: true,
	}
	uc := NewTransferUseCase(source, dest, cfg, log)

	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %q, want %q (different size should trigger re-transfer)", result.Status, entity.TransferStatusCompleted)
	}
	if !dest.WriteCalled {
		t.Error("Write was not called — file with different size should be re-transferred")
	}
}

func TestTransferUseCase_SkipExisting_NotExists_Transfers(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	log := applog.NewNopLogger()

	content := []byte("brand new file")
	source.AddFile("/src/file.txt", content, &entity.FileInfo{
		Path: "/src/file.txt", Name: "file.txt",
		Size: int64(len(content)), ModifiedTime: time.Now(),
	})
	// Destination is empty

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 1, RetryAttempts: 0,
		SkipExisting: true,
	}
	uc := NewTransferUseCase(source, dest, cfg, log)

	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %q, want %q (non-existing dest should be transferred)", result.Status, entity.TransferStatusCompleted)
	}
}

func TestTransferUseCase_SkipExisting_Disabled_AlwaysTransfers(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	log := applog.NewNopLogger()

	content := []byte("some content")
	source.AddFile("/src/file.txt", content, &entity.FileInfo{
		Path: "/src/file.txt", Name: "file.txt",
		Size: int64(len(content)), ModifiedTime: time.Now(),
	})
	// Destination already has same file
	dest.AddFile("/dst/file.txt", content, &entity.FileInfo{
		Path: "/dst/file.txt", Name: "file.txt",
		Size: int64(len(content)), ModifiedTime: time.Now(),
	})

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 1, RetryAttempts: 0,
		SkipExisting: false, // disabled
	}
	uc := NewTransferUseCase(source, dest, cfg, log)

	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status == entity.TransferStatusSkipped {
		t.Error("Status = skipped, but SkipExisting is false — should always transfer")
	}
	if !dest.WriteCalled {
		t.Error("Write not called — should overwrite when SkipExisting is false")
	}
}

func TestTransferUseCase_SkipExisting_BatchIdempotent(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	log := applog.NewNopLogger()

	files := []string{"/src/a.txt", "/src/b.txt", "/src/c.txt"}
	for _, f := range files {
		c := []byte("content of " + f)
		info := &entity.FileInfo{Path: f, Name: f[5:], Size: int64(len(c)), ModifiedTime: time.Now()}
		source.AddFile(f, c, info)
		// a.txt and c.txt already at destination with matching size
		if f != "/src/b.txt" {
			destPath := "/dst/" + f[5:]
			dest.AddFile(destPath, c, &entity.FileInfo{Path: destPath, Name: f[5:], Size: int64(len(c)), ModifiedTime: time.Now()})
		}
	}

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 3, RetryAttempts: 0,
		SkipExisting: true,
	}
	uc := NewTransferUseCase(source, dest, cfg, log)

	results, err := uc.TransferBatch(context.Background(), files, "/dst/", nil)
	if err != nil {
		t.Fatalf("unexpected batch error: %v", err)
	}

	var skipped, completed int
	for _, r := range results {
		switch r.Status {
		case entity.TransferStatusSkipped:
			skipped++
		case entity.TransferStatusCompleted:
			completed++
		}
	}

	if skipped != 2 {
		t.Errorf("skipped = %d, want 2 (a.txt and c.txt already exist)", skipped)
	}
	if completed != 1 {
		t.Errorf("completed = %d, want 1 (only b.txt should transfer)", completed)
	}
}
