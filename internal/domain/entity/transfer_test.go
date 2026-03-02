package entity

import (
	"testing"
	"time"
)

func TestTransferProgress(t *testing.T) {
	progress := TransferProgress{
		FileName:         "test.pdf",
		TotalBytes:       1000000,
		TransferredBytes: 500000,
		Speed:            125000, // 125 KB/s
		StartTime:        time.Now(),
		Status:           TransferStatusInProgress,
	}

	if progress.FileName != "test.pdf" {
		t.Errorf("FileName = %v, want test.pdf", progress.FileName)
	}

	if progress.Status != TransferStatusInProgress {
		t.Errorf("Status = %v, want %v", progress.Status, TransferStatusInProgress)
	}

	percentage := float64(progress.TransferredBytes) / float64(progress.TotalBytes) * 100
	if percentage != 50.0 {
		t.Errorf("Percentage = %v, want 50.0", percentage)
	}
}

func TestTransferResult(t *testing.T) {
	result := TransferResult{
		SourcePath:       "/source/file.txt",
		DestinationPath:  "/dest/file.txt",
		BytesTransferred: 1024,
		Duration:         time.Second,
		Status:           TransferStatusCompleted,
		Error:            nil,
	}

	if result.Status != TransferStatusCompleted {
		t.Errorf("Status = %v, want %v", result.Status, TransferStatusCompleted)
	}

	if result.Error != nil {
		t.Errorf("Error should be nil, got %v", result.Error)
	}

	speed := float64(result.BytesTransferred) / result.Duration.Seconds()
	if speed != 1024.0 {
		t.Errorf("Speed = %v, want 1024.0", speed)
	}
}

func TestTransferConfig_Defaults(t *testing.T) {
	config := TransferConfig{
		BufferSize:      32 * 1024 * 1024,
		ConcurrentFiles: 4,
		RetryAttempts:   3,
		RetryDelay:      5 * time.Second,
		Timeout:         30 * time.Minute,
		VerifyChecksum:  false,
	}

	if config.BufferSize != 32*1024*1024 {
		t.Errorf("BufferSize = %v, want %v", config.BufferSize, 32*1024*1024)
	}

	if config.ConcurrentFiles != 4 {
		t.Errorf("ConcurrentFiles = %v, want 4", config.ConcurrentFiles)
	}

	if config.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %v, want 3", config.RetryAttempts)
	}
}

func TestFileInfo(t *testing.T) {
	now := time.Now()
	fileInfo := FileInfo{
		Path:         "/data/file.txt",
		Name:         "file.txt",
		Size:         1024,
		ModifiedTime: now,
		IsDirectory:  false,
	}

	if fileInfo.Name != "file.txt" {
		t.Errorf("Name = %v, want file.txt", fileInfo.Name)
	}

	if fileInfo.Size != 1024 {
		t.Errorf("Size = %v, want 1024", fileInfo.Size)
	}

	if fileInfo.IsDirectory {
		t.Errorf("IsDirectory should be false")
	}

	if !fileInfo.ModifiedTime.Equal(now) {
		t.Errorf("ModifiedTime = %v, want %v", fileInfo.ModifiedTime, now)
	}
}

func TestBatchTransferResult(t *testing.T) {
	result := BatchTransferResult{
		TotalFiles:      10,
		SuccessfulFiles: 8,
		FailedFiles:     2,
		TotalBytes:      10240,
		TotalDuration:   10 * time.Second,
		AverageSpeed:    1024.0,
	}

	if result.TotalFiles != 10 {
		t.Errorf("TotalFiles = %v, want 10", result.TotalFiles)
	}

	if result.SuccessfulFiles != 8 {
		t.Errorf("SuccessfulFiles = %v, want 8", result.SuccessfulFiles)
	}

	if result.FailedFiles != 2 {
		t.Errorf("FailedFiles = %v, want 2", result.FailedFiles)
	}

	successRate := float64(result.SuccessfulFiles) / float64(result.TotalFiles) * 100
	if successRate != 80.0 {
		t.Errorf("Success rate = %v, want 80.0", successRate)
	}
}
