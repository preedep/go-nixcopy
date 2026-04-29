package usecase

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// ---- progressReader.Close ----

func TestProgressReader_Close_WithCloser(t *testing.T) {
	inner := io.NopCloser(bytes.NewReader([]byte("data")))
	pr := &progressReader{reader: inner}
	if err := pr.Close(); err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}
}

func TestProgressReader_Close_WithoutCloser(t *testing.T) {
	// bytes.Reader does not implement io.Closer
	pr := &progressReader{reader: bytes.NewReader([]byte("data"))}
	if err := pr.Close(); err != nil {
		t.Errorf("Close() unexpected error for non-Closer reader: %v", err)
	}
}

// ---- Resume + Compression (warns, falls back to full transfer) ----

func TestTransfer_Resume_And_Compression_Warns(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	content := []byte("Hello compressed world!")
	source.AddFile("/src/file.txt", content, &entity.FileInfo{
		Path: "/src/file.txt",
		Name: "file.txt",
		Size: int64(len(content)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
		Compression:     "gzip", // incompatible with resume — should warn and complete
	}

	uc := NewTransferUseCase(source, dest, cfg, logger)
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err != nil {
		t.Fatalf("Transfer() unexpected error: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
}

// ---- Checksum + resumed transfer (checksum skipped, warns) ----

func TestTransfer_Checksum_Skipped_When_Resumed(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	full := []byte("Hello, World! This is a full file content.")
	partial := full[:10]

	source.AddFile("/src/file.txt", full, &entity.FileInfo{
		Path: "/src/file.txt",
		Name: "file.txt",
		Size: int64(len(full)),
	})
	// Seed dest with partial content so resume offset > 0
	dest.AddFile("/dst/file.txt", partial, &entity.FileInfo{
		Path: "/dst/file.txt",
		Name: "file.txt",
		Size: int64(len(partial)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
		VerifyChecksum:  true, // should be skipped for resumed transfers with a warning
	}

	uc := NewTransferUseCase(source, dest, cfg, logger)
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err != nil {
		t.Fatalf("Transfer() unexpected error: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	// Checksum field is empty because verification was skipped
	if result.Checksum != "" {
		t.Errorf("Checksum = %q, want empty (skipped for resumed transfer)", result.Checksum)
	}
}

// ---- Checksum + Compression (skips verification, warns) ----

func TestTransfer_Checksum_Skipped_When_Compressed(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	content := []byte("some data that will be gzip-compressed")
	source.AddFile("/src/file.txt", content, &entity.FileInfo{
		Path: "/src/file.txt",
		Name: "file.txt",
		Size: int64(len(content)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		VerifyChecksum:  true,
		Compression:     "gzip", // checksum must be skipped because byte stream differs
	}

	uc := NewTransferUseCase(source, dest, cfg, logger)
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err != nil {
		t.Fatalf("Transfer() unexpected error: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	if result.Checksum != "" {
		t.Errorf("Checksum = %q, want empty (skipped for compressed transfer)", result.Checksum)
	}
}

// ---- Resume + ReadFrom error (exhausts retries) ----

func TestTransfer_Resume_ReadFrom_Error(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	full := []byte("some file content for resume error test")
	partial := full[:5]

	source.AddFile("/src/file.txt", full, &entity.FileInfo{
		Path: "/src/file.txt",
		Name: "file.txt",
		Size: int64(len(full)),
	})
	// Partial dest triggers resume path
	dest.AddFile("/dst/file.txt", partial, &entity.FileInfo{
		Path: "/dst/file.txt",
		Name: "file.txt",
		Size: int64(len(partial)),
	})
	// ReadError makes ReadFrom fail on every attempt
	source.ReadError = io.ErrUnexpectedEOF

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   1,
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
	}

	uc := NewTransferUseCase(source, dest, cfg, logger)
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)
	if err == nil {
		t.Fatal("Transfer() expected error when ReadFrom always fails")
	}
	if result.Status != entity.TransferStatusFailed {
		t.Errorf("Status = %v, want failed", result.Status)
	}
}
