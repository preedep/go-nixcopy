package storage

import (
	"fmt"

	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

func NewStorageFromSourceConfig(cfg *config.SourceConfig) (repository.Storage, error) {
	switch cfg.Type {
	case config.StorageTypeLocal:
		if cfg.Local == nil {
			return nil, fmt.Errorf("Local configuration is required")
		}
		return NewLocalStorage(cfg.Local)
	case config.StorageTypeSFTP:
		if cfg.SFTP == nil {
			return nil, fmt.Errorf("SFTP configuration is required")
		}
		return NewSFTPStorage(cfg.SFTP), nil
	case config.StorageTypeFTPS:
		if cfg.FTPS == nil {
			return nil, fmt.Errorf("FTPS configuration is required")
		}
		return NewFTPSStorage(cfg.FTPS), nil
	case config.StorageTypeBlobStorage:
		if cfg.BlobStorage == nil {
			return nil, fmt.Errorf("Blob Storage configuration is required")
		}
		return NewBlobStorage(cfg.BlobStorage), nil
	case config.StorageTypeS3:
		if cfg.S3 == nil {
			return nil, fmt.Errorf("S3 configuration is required")
		}
		return NewS3Storage(cfg.S3), nil
	default:
		return nil, fmt.Errorf("unsupported source storage type: %s", cfg.Type)
	}
}

func NewStorageFromDestConfig(cfg *config.DestinationConfig) (repository.Storage, error) {
	switch cfg.Type {
	case config.StorageTypeLocal:
		if cfg.Local == nil {
			return nil, fmt.Errorf("Local configuration is required")
		}
		return NewLocalStorage(cfg.Local)
	case config.StorageTypeSFTP:
		if cfg.SFTP == nil {
			return nil, fmt.Errorf("SFTP configuration is required")
		}
		return NewSFTPStorage(cfg.SFTP), nil
	case config.StorageTypeFTPS:
		if cfg.FTPS == nil {
			return nil, fmt.Errorf("FTPS configuration is required")
		}
		return NewFTPSStorage(cfg.FTPS), nil
	case config.StorageTypeBlobStorage:
		if cfg.BlobStorage == nil {
			return nil, fmt.Errorf("Blob Storage configuration is required")
		}
		return NewBlobStorage(cfg.BlobStorage), nil
	case config.StorageTypeS3:
		if cfg.S3 == nil {
			return nil, fmt.Errorf("S3 configuration is required")
		}
		return NewS3Storage(cfg.S3), nil
	default:
		return nil, fmt.Errorf("unsupported destination storage type: %s", cfg.Type)
	}
}
