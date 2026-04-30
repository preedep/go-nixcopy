package entity

import (
	"io"
	"time"
)

type TransferStatus string

const (
	TransferStatusPending    TransferStatus = "pending"
	TransferStatusInProgress TransferStatus = "in_progress"
	TransferStatusCompleted  TransferStatus = "completed"
	TransferStatusFailed     TransferStatus = "failed"
	TransferStatusSkipped    TransferStatus = "skipped"
)

type FileInfo struct {
	Path         string
	Name         string
	Size         int64
	ModifiedTime time.Time
	IsDirectory  bool
}

type TransferProgress struct {
	FileName         string
	TotalBytes       int64
	TransferredBytes int64
	Speed            float64
	StartTime        time.Time
	EstimatedTime    time.Duration
	Status           TransferStatus
	Error            error
}

type TransferConfig struct {
	BufferSize      int
	ConcurrentFiles int
	RetryAttempts   int
	RetryDelay      time.Duration
	Timeout         time.Duration
	VerifyChecksum  bool
	EnableResume    bool
	SkipExisting    bool   // skip transfer if destination already has a file with matching size
	BandwidthLimit  int64  // max bytes per second per file transfer; 0 = unlimited
	Compression     string // "gzip", "zstd", or "" (no compression)
}

type TransferResult struct {
	SourcePath       string
	DestinationPath  string
	BytesTransferred int64
	Duration         time.Duration
	Status           TransferStatus
	Error            error
	Checksum         string // SHA256 hex of transferred content; set when VerifyChecksum is true
	ResumedFrom      int64  // byte offset the transfer resumed from; 0 if started from the beginning
}

type StreamReader interface {
	io.ReadCloser
	Size() int64
}
