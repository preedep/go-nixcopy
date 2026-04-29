package storage_test

import (
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

// ---- NewStorageFromSourceConfig ----

func TestNewStorageFromSourceConfig_Local(t *testing.T) {
	cfg := &config.SourceConfig{
		Type:  config.StorageTypeLocal,
		Local: &config.LocalConfig{BasePath: t.TempDir()},
	}
	s, err := storage.NewStorageFromSourceConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromSourceConfig_Local_NilConfig(t *testing.T) {
	cfg := &config.SourceConfig{Type: config.StorageTypeLocal}
	_, err := storage.NewStorageFromSourceConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil Local config")
	}
}

func TestNewStorageFromSourceConfig_SFTP(t *testing.T) {
	cfg := &config.SourceConfig{
		Type: config.StorageTypeSFTP,
		SFTP: &config.SFTPConfig{Host: "sftp.example.com", Port: 22, Username: "user"},
	}
	s, err := storage.NewStorageFromSourceConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromSourceConfig_SFTP_NilConfig(t *testing.T) {
	cfg := &config.SourceConfig{Type: config.StorageTypeSFTP}
	_, err := storage.NewStorageFromSourceConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil SFTP config")
	}
}

func TestNewStorageFromSourceConfig_FTPS(t *testing.T) {
	cfg := &config.SourceConfig{
		Type: config.StorageTypeFTPS,
		FTPS: &config.FTPSConfig{Host: "ftps.example.com", Port: 21},
	}
	s, err := storage.NewStorageFromSourceConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromSourceConfig_FTPS_NilConfig(t *testing.T) {
	cfg := &config.SourceConfig{Type: config.StorageTypeFTPS}
	_, err := storage.NewStorageFromSourceConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil FTPS config")
	}
}

func TestNewStorageFromSourceConfig_Blob(t *testing.T) {
	cfg := &config.SourceConfig{
		Type: config.StorageTypeBlobStorage,
		BlobStorage: &config.BlobConfig{
			AccountName:   "account",
			ContainerName: "container",
		},
	}
	s, err := storage.NewStorageFromSourceConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromSourceConfig_Blob_NilConfig(t *testing.T) {
	cfg := &config.SourceConfig{Type: config.StorageTypeBlobStorage}
	_, err := storage.NewStorageFromSourceConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil BlobStorage config")
	}
}

func TestNewStorageFromSourceConfig_S3(t *testing.T) {
	cfg := &config.SourceConfig{
		Type: config.StorageTypeS3,
		S3:   &config.S3Config{Region: "us-east-1", Bucket: "my-bucket"},
	}
	s, err := storage.NewStorageFromSourceConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromSourceConfig_S3_NilConfig(t *testing.T) {
	cfg := &config.SourceConfig{Type: config.StorageTypeS3}
	_, err := storage.NewStorageFromSourceConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil S3 config")
	}
}

func TestNewStorageFromSourceConfig_Unknown(t *testing.T) {
	cfg := &config.SourceConfig{Type: "gcs"}
	_, err := storage.NewStorageFromSourceConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unknown storage type")
	}
}

// ---- NewStorageFromDestConfig ----

func TestNewStorageFromDestConfig_Local(t *testing.T) {
	cfg := &config.DestinationConfig{
		Type:  config.StorageTypeLocal,
		Local: &config.LocalConfig{BasePath: t.TempDir()},
	}
	s, err := storage.NewStorageFromDestConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromDestConfig_Local_NilConfig(t *testing.T) {
	cfg := &config.DestinationConfig{Type: config.StorageTypeLocal}
	_, err := storage.NewStorageFromDestConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil Local config")
	}
}

func TestNewStorageFromDestConfig_SFTP(t *testing.T) {
	cfg := &config.DestinationConfig{
		Type: config.StorageTypeSFTP,
		SFTP: &config.SFTPConfig{Host: "sftp.example.com", Port: 22, Username: "user"},
	}
	s, err := storage.NewStorageFromDestConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromDestConfig_SFTP_NilConfig(t *testing.T) {
	cfg := &config.DestinationConfig{Type: config.StorageTypeSFTP}
	_, err := storage.NewStorageFromDestConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil SFTP config")
	}
}

func TestNewStorageFromDestConfig_FTPS(t *testing.T) {
	cfg := &config.DestinationConfig{
		Type: config.StorageTypeFTPS,
		FTPS: &config.FTPSConfig{Host: "ftps.example.com", Port: 21},
	}
	s, err := storage.NewStorageFromDestConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromDestConfig_FTPS_NilConfig(t *testing.T) {
	cfg := &config.DestinationConfig{Type: config.StorageTypeFTPS}
	_, err := storage.NewStorageFromDestConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil FTPS config")
	}
}

func TestNewStorageFromDestConfig_Blob(t *testing.T) {
	cfg := &config.DestinationConfig{
		Type: config.StorageTypeBlobStorage,
		BlobStorage: &config.BlobConfig{
			AccountName:   "account",
			ContainerName: "container",
		},
	}
	s, err := storage.NewStorageFromDestConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromDestConfig_Blob_NilConfig(t *testing.T) {
	cfg := &config.DestinationConfig{Type: config.StorageTypeBlobStorage}
	_, err := storage.NewStorageFromDestConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil BlobStorage config")
	}
}

func TestNewStorageFromDestConfig_S3(t *testing.T) {
	cfg := &config.DestinationConfig{
		Type: config.StorageTypeS3,
		S3:   &config.S3Config{Region: "ap-southeast-1", Bucket: "dest-bucket"},
	}
	s, err := storage.NewStorageFromDestConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil storage")
	}
}

func TestNewStorageFromDestConfig_S3_NilConfig(t *testing.T) {
	cfg := &config.DestinationConfig{Type: config.StorageTypeS3}
	_, err := storage.NewStorageFromDestConfig(cfg)
	if err == nil {
		t.Fatal("expected error for nil S3 config")
	}
}

func TestNewStorageFromDestConfig_Unknown(t *testing.T) {
	cfg := &config.DestinationConfig{Type: "gcs"}
	_, err := storage.NewStorageFromDestConfig(cfg)
	if err == nil {
		t.Fatal("expected error for unknown storage type")
	}
}
