package cli

import (
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

// ---- validateConfig: source backends ----

func TestValidateConfig_SFTP_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{Type: config.StorageTypeSFTP},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "SFTP source configuration is required" {
		t.Errorf("err = %v, want SFTP source configuration is required", err)
	}
}

func TestValidateConfig_SFTP_MissingHost(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeSFTP,
			SFTP: &config.SFTPConfig{Username: "user"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "source SFTP host is required" {
		t.Errorf("err = %v, want source SFTP host is required", err)
	}
}

func TestValidateConfig_SFTP_MissingUsername(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeSFTP,
			SFTP: &config.SFTPConfig{Host: "sftp.example.com"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "source SFTP username is required" {
		t.Errorf("err = %v, want source SFTP username is required", err)
	}
}

func TestValidateConfig_FTPS_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{Type: config.StorageTypeFTPS},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "FTPS source configuration is required" {
		t.Errorf("err = %v, want FTPS source configuration is required", err)
	}
}

func TestValidateConfig_FTPS_MissingHost(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeFTPS,
			FTPS: &config.FTPSConfig{},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "source FTPS host is required" {
		t.Errorf("err = %v, want source FTPS host is required", err)
	}
}

func TestValidateConfig_S3_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{Type: config.StorageTypeS3},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "S3 source configuration is required" {
		t.Errorf("err = %v, want S3 source configuration is required", err)
	}
}

func TestValidateConfig_S3_MissingRegion(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeS3,
			S3:   &config.S3Config{Bucket: "my-bucket"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "source S3 region is required" {
		t.Errorf("err = %v, want source S3 region is required", err)
	}
}

func TestValidateConfig_S3_MissingBucket(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeS3,
			S3:   &config.S3Config{Region: "us-east-1"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "source S3 bucket is required" {
		t.Errorf("err = %v, want source S3 bucket is required", err)
	}
}

func TestValidateConfig_Blob_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{Type: config.StorageTypeBlobStorage},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "Blob Storage source configuration is required" {
		t.Errorf("err = %v, want Blob Storage source configuration is required", err)
	}
}

func TestValidateConfig_Blob_MissingAccountName(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type:        config.StorageTypeBlobStorage,
			BlobStorage: &config.BlobConfig{ContainerName: "container"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "source Blob Storage account name is required" {
		t.Errorf("err = %v, want source Blob Storage account name is required", err)
	}
}

func TestValidateConfig_Blob_MissingContainerName(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type:        config.StorageTypeBlobStorage,
			BlobStorage: &config.BlobConfig{AccountName: "account"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "source Blob Storage container name is required" {
		t.Errorf("err = %v, want source Blob Storage container name is required", err)
	}
}

// ---- validateConfig: destination backends ----

func validSource() config.SourceConfig {
	return config.SourceConfig{
		Type: config.StorageTypeSFTP,
		SFTP: &config.SFTPConfig{Host: "sftp.example.com", Username: "user"},
	}
}

func TestValidateConfig_Dest_SFTP_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source:      validSource(),
		Destination: config.DestinationConfig{Type: config.StorageTypeSFTP},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "SFTP destination configuration is required" {
		t.Errorf("err = %v, want SFTP destination configuration is required", err)
	}
}

func TestValidateConfig_Dest_SFTP_MissingHost(t *testing.T) {
	cfg := &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type: config.StorageTypeSFTP,
			SFTP: &config.SFTPConfig{Username: "user"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "destination SFTP host is required" {
		t.Errorf("err = %v, want destination SFTP host is required", err)
	}
}

func TestValidateConfig_Dest_SFTP_MissingUsername(t *testing.T) {
	cfg := &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type: config.StorageTypeSFTP,
			SFTP: &config.SFTPConfig{Host: "dest.example.com"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "destination SFTP username is required" {
		t.Errorf("err = %v, want destination SFTP username is required", err)
	}
}

func TestValidateConfig_Dest_FTPS_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source:      validSource(),
		Destination: config.DestinationConfig{Type: config.StorageTypeFTPS},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "FTPS destination configuration is required" {
		t.Errorf("err = %v, want FTPS destination configuration is required", err)
	}
}

func TestValidateConfig_Dest_FTPS_MissingHost(t *testing.T) {
	cfg := &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type: config.StorageTypeFTPS,
			FTPS: &config.FTPSConfig{},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "destination FTPS host is required" {
		t.Errorf("err = %v, want destination FTPS host is required", err)
	}
}

func TestValidateConfig_Dest_S3_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source:      validSource(),
		Destination: config.DestinationConfig{Type: config.StorageTypeS3},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "S3 destination configuration is required" {
		t.Errorf("err = %v, want S3 destination configuration is required", err)
	}
}

func TestValidateConfig_Dest_S3_MissingRegion(t *testing.T) {
	cfg := &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type: config.StorageTypeS3,
			S3:   &config.S3Config{Bucket: "dest-bucket"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "destination S3 region is required" {
		t.Errorf("err = %v, want destination S3 region is required", err)
	}
}

func TestValidateConfig_Dest_S3_MissingBucket(t *testing.T) {
	cfg := &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type: config.StorageTypeS3,
			S3:   &config.S3Config{Region: "ap-southeast-1"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "destination S3 bucket is required" {
		t.Errorf("err = %v, want destination S3 bucket is required", err)
	}
}

func TestValidateConfig_Dest_Blob_NilConfig(t *testing.T) {
	cfg := &config.Config{
		Source:      validSource(),
		Destination: config.DestinationConfig{Type: config.StorageTypeBlobStorage},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "Blob Storage destination configuration is required" {
		t.Errorf("err = %v, want Blob Storage destination configuration is required", err)
	}
}

func TestValidateConfig_Dest_Blob_MissingAccountName(t *testing.T) {
	cfg := &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type:        config.StorageTypeBlobStorage,
			BlobStorage: &config.BlobConfig{ContainerName: "container"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "destination Blob Storage account name is required" {
		t.Errorf("err = %v, want destination Blob Storage account name is required", err)
	}
}

func TestValidateConfig_Dest_Blob_MissingContainerName(t *testing.T) {
	cfg := &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type:        config.StorageTypeBlobStorage,
			BlobStorage: &config.BlobConfig{AccountName: "account"},
		},
	}
	err := validateConfig(cfg)
	if err == nil || err.Error() != "destination Blob Storage container name is required" {
		t.Errorf("err = %v, want destination Blob Storage container name is required", err)
	}
}

// ---- validateConfig: compression ----

func fullValidConfig() *config.Config {
	return &config.Config{
		Source: validSource(),
		Destination: config.DestinationConfig{
			Type: config.StorageTypeS3,
			S3:   &config.S3Config{Region: "us-east-1", Bucket: "bucket"},
		},
	}
}

func TestValidateConfig_InvalidCompression(t *testing.T) {
	cfg := fullValidConfig()
	cfg.Transfer.Compression = "brotli"
	err := validateConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid compression")
	}
}

func TestValidateConfig_Compression_Gzip(t *testing.T) {
	cfg := fullValidConfig()
	cfg.Transfer.Compression = "gzip"
	if err := validateConfig(cfg); err != nil {
		t.Errorf("unexpected error for gzip: %v", err)
	}
}

func TestValidateConfig_Compression_Zstd(t *testing.T) {
	cfg := fullValidConfig()
	cfg.Transfer.Compression = "zstd"
	if err := validateConfig(cfg); err != nil {
		t.Errorf("unexpected error for zstd: %v", err)
	}
}

func TestValidateConfig_Compression_Empty(t *testing.T) {
	cfg := fullValidConfig()
	cfg.Transfer.Compression = ""
	if err := validateConfig(cfg); err != nil {
		t.Errorf("unexpected error for empty compression: %v", err)
	}
}

// ---- formatSize ----

func TestFormatSize(t *testing.T) {
	cases := []struct {
		size int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{int64(1.5 * 1024 * 1024), "1.50 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{int64(2.5 * 1024 * 1024 * 1024), "2.50 GB"},
	}
	for _, tc := range cases {
		got := formatSize(tc.size)
		if got != tc.want {
			t.Errorf("formatSize(%d) = %q, want %q", tc.size, got, tc.want)
		}
	}
}
