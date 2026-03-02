package entity

import (
	"io"
	"time"
)

type TransferType string

const (
	TransferTypeSFTPToFTPS        TransferType = "sftp_to_ftps"
	TransferTypeSFTPToBlobStorage TransferType = "sftp_to_blob"
	TransferTypeSFTPToS3          TransferType = "sftp_to_s3"
	TransferTypeFTPSToSFTP        TransferType = "ftps_to_sftp"
	TransferTypeFTPSToBlobStorage TransferType = "ftps_to_blob"
	TransferTypeFTPSToS3          TransferType = "ftps_to_s3"
	TransferTypeBlobStorageToSFTP TransferType = "blob_to_sftp"
	TransferTypeBlobStorageToFTPS TransferType = "blob_to_ftps"
	TransferTypeBlobStorageToS3   TransferType = "blob_to_s3"
	TransferTypeS3ToSFTP          TransferType = "s3_to_sftp"
	TransferTypeS3ToFTPS          TransferType = "s3_to_ftps"
	TransferTypeS3ToBlobStorage   TransferType = "s3_to_blob"
)

type TransferStatus string

const (
	TransferStatusPending    TransferStatus = "pending"
	TransferStatusInProgress TransferStatus = "in_progress"
	TransferStatusCompleted  TransferStatus = "completed"
	TransferStatusFailed     TransferStatus = "failed"
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
}

type TransferResult struct {
	SourcePath       string
	DestinationPath  string
	BytesTransferred int64
	Duration         time.Duration
	Status           TransferStatus
	Error            error
}

type StreamReader interface {
	io.ReadCloser
	Size() int64
}
