package cli

import (
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

func TestApplyCliFlags_Source(t *testing.T) {
	// Reset flags
	sourceType = "sftp"
	sourceHost = "sftp.example.com"
	sourcePort = 22
	sourceUsername = "testuser"
	sourcePassword = "testpass"

	cfg := config.DefaultConfig()
	err := applyCliFlags(cfg)

	if err != nil {
		t.Fatalf("applyCliFlags() error = %v", err)
	}

	if cfg.Source.Type != config.StorageTypeSFTP {
		t.Errorf("Source.Type = %v, want %v", cfg.Source.Type, config.StorageTypeSFTP)
	}

	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP should not be nil")
	}

	if cfg.Source.SFTP.Host != "sftp.example.com" {
		t.Errorf("Source.SFTP.Host = %v, want sftp.example.com", cfg.Source.SFTP.Host)
	}

	if cfg.Source.SFTP.Port != 22 {
		t.Errorf("Source.SFTP.Port = %v, want 22", cfg.Source.SFTP.Port)
	}

	if cfg.Source.SFTP.Username != "testuser" {
		t.Errorf("Source.SFTP.Username = %v, want testuser", cfg.Source.SFTP.Username)
	}
}

func TestApplyCliFlags_Destination(t *testing.T) {
	// Reset flags
	destType = "s3"
	destRegion = "us-east-1"
	destBucket = "my-bucket"
	destAuthType = "access_key"
	destAccessKey = "AKIA..."
	destSecretKey = "secret"

	cfg := config.DefaultConfig()
	err := applyCliFlags(cfg)

	if err != nil {
		t.Fatalf("applyCliFlags() error = %v", err)
	}

	if cfg.Destination.Type != config.StorageTypeS3 {
		t.Errorf("Destination.Type = %v, want %v", cfg.Destination.Type, config.StorageTypeS3)
	}

	if cfg.Destination.S3 == nil {
		t.Fatal("Destination.S3 should not be nil")
	}

	if cfg.Destination.S3.Region != "us-east-1" {
		t.Errorf("Destination.S3.Region = %v, want us-east-1", cfg.Destination.S3.Region)
	}

	if cfg.Destination.S3.Bucket != "my-bucket" {
		t.Errorf("Destination.S3.Bucket = %v, want my-bucket", cfg.Destination.S3.Bucket)
	}

	if cfg.Destination.S3.AuthType != config.S3AuthAccessKey {
		t.Errorf("Destination.S3.AuthType = %v, want %v", cfg.Destination.S3.AuthType, config.S3AuthAccessKey)
	}
}

func TestApplyCliFlags_Transfer(t *testing.T) {
	// Reset flags
	bufferSize = 67108864 // 64MB
	concurrentFiles = 8
	retryAttempts = 5

	cfg := config.DefaultConfig()
	err := applyCliFlags(cfg)

	if err != nil {
		t.Fatalf("applyCliFlags() error = %v", err)
	}

	if cfg.Transfer.BufferSize != 67108864 {
		t.Errorf("Transfer.BufferSize = %v, want 67108864", cfg.Transfer.BufferSize)
	}

	if cfg.Transfer.ConcurrentFiles != 8 {
		t.Errorf("Transfer.ConcurrentFiles = %v, want 8", cfg.Transfer.ConcurrentFiles)
	}

	if cfg.Transfer.RetryAttempts != 5 {
		t.Errorf("Transfer.RetryAttempts = %v, want 5", cfg.Transfer.RetryAttempts)
	}
}

func TestApplyCliFlags_Defaults(t *testing.T) {
	// Reset all flags to empty
	sourceType = ""
	destType = ""
	bufferSize = 0
	concurrentFiles = 0
	retryAttempts = 0

	cfg := &config.Config{}
	err := applyCliFlags(cfg)

	if err != nil {
		t.Fatalf("applyCliFlags() error = %v", err)
	}

	// Check defaults are applied
	if cfg.Transfer.BufferSize != 32*1024*1024 {
		t.Errorf("Transfer.BufferSize = %v, want %v", cfg.Transfer.BufferSize, 32*1024*1024)
	}

	if cfg.Transfer.ConcurrentFiles != 4 {
		t.Errorf("Transfer.ConcurrentFiles = %v, want 4", cfg.Transfer.ConcurrentFiles)
	}

	if cfg.Transfer.RetryAttempts != 3 {
		t.Errorf("Transfer.RetryAttempts = %v, want 3", cfg.Transfer.RetryAttempts)
	}

	if cfg.Transfer.RetryDelay != 5*time.Second {
		t.Errorf("Transfer.RetryDelay = %v, want 5s", cfg.Transfer.RetryDelay)
	}
}

func TestValidateConfig_MissingSourceType(t *testing.T) {
	cfg := &config.Config{}

	err := validateConfig(cfg)

	if err == nil {
		t.Error("validateConfig() should return error for missing source type")
	}

	if err.Error() != "source type is required" {
		t.Errorf("Error message = %v, want 'source type is required'", err.Error())
	}
}

func TestValidateConfig_MissingDestinationType(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeSFTP,
			SFTP: &config.SFTPConfig{
				Host:     "sftp.example.com",
				Username: "user",
			},
		},
	}

	err := validateConfig(cfg)

	if err == nil {
		t.Error("validateConfig() should return error for missing destination type")
	}

	if err.Error() != "destination type is required" {
		t.Errorf("Error message = %v, want 'destination type is required'", err.Error())
	}
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		Source: config.SourceConfig{
			Type: config.StorageTypeSFTP,
			SFTP: &config.SFTPConfig{
				Host:     "sftp.example.com",
				Username: "user",
			},
		},
		Destination: config.DestinationConfig{
			Type: config.StorageTypeS3,
			S3: &config.S3Config{
				Region: "us-east-1",
				Bucket: "my-bucket",
			},
		},
	}

	err := validateConfig(cfg)

	if err != nil {
		t.Errorf("validateConfig() should not return error for valid config, got: %v", err)
	}
}
