package cli

import (
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

func TestApplyCliFlags_Source(t *testing.T) {
	sourceType = "sftp"
	sourceHost = "sftp.example.com"
	sourcePort = 22
	sourceUsername = "testuser"
	sourcePassword = "testpass"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

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
	destType = "s3"
	destRegion = "us-east-1"
	destBucket = "my-bucket"
	destAuthType = "access_key"
	destAccessKey = "AKIA..."
	destSecretKey = "secret"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

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
	bufferSize = 67108864 // 64MB
	concurrentFiles = 8
	retryAttempts = 5

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

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

func TestApplyCliFlags_SkipExisting(t *testing.T) {
	resetTransferFlags()
	skipExisting = true

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if !cfg.Transfer.SkipExisting {
		t.Error("Transfer.SkipExisting = false, want true when --skip-existing is set")
	}
}

func TestApplyCliFlags_EnableResume(t *testing.T) {
	resetTransferFlags()
	enableResume = true

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if !cfg.Transfer.EnableResume {
		t.Error("Transfer.EnableResume = false, want true when --resume is set")
	}
}

func TestApplyCliFlags_SkipExisting_False_DoesNotOverrideConfigFile(t *testing.T) {
	resetTransferFlags()
	skipExisting = false // flag not set

	cfg := config.DefaultConfig()
	cfg.Transfer.SkipExisting = true // loaded from config file
	applyCliFlags(cfg)

	if !cfg.Transfer.SkipExisting {
		t.Error("Transfer.SkipExisting was cleared; false flag should not override a config-file true value")
	}
}

// resetTransferFlags zeros out all package-level flag vars so tests don't bleed state into each other.
func resetTransferFlags() {
	sourceType = ""
	sourceHost = ""
	sourcePort = 0
	sourceUsername = ""
	sourcePassword = ""
	sourcePrivateKey = ""
	sourceRegion = ""
	sourceBucket = ""
	sourceAccessKey = ""
	sourceSecretKey = ""
	sourceAuthType = ""
	sourceAccountName = ""
	sourceAccountKey = ""
	sourceContainer = ""

	destType = ""
	destHost = ""
	destPort = 0
	destUsername = ""
	destPassword = ""
	destPrivateKey = ""
	destRegion = ""
	destBucket = ""
	destAccessKey = ""
	destSecretKey = ""
	destAuthType = ""
	destAccountName = ""
	destAccountKey = ""
	destContainer = ""

	bufferSize = 0
	concurrentFiles = 0
	retryAttempts = 0
	enableResume = false
	skipExisting = false
	compress = ""
}

func TestApplyCliFlags_SourceFTPS(t *testing.T) {
	resetTransferFlags()
	sourceType = "ftps"
	sourceHost = "ftps.example.com"
	sourcePort = 990
	sourceUsername = "ftpuser"
	sourcePassword = "ftppass"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.Type != config.StorageTypeFTPS {
		t.Errorf("Source.Type = %v, want %v", cfg.Source.Type, config.StorageTypeFTPS)
	}
	if cfg.Source.FTPS == nil {
		t.Fatal("Source.FTPS should not be nil")
	}
	if cfg.Source.FTPS.Host != "ftps.example.com" {
		t.Errorf("FTPS.Host = %q, want ftps.example.com", cfg.Source.FTPS.Host)
	}
	if cfg.Source.FTPS.Port != 990 {
		t.Errorf("FTPS.Port = %d, want 990", cfg.Source.FTPS.Port)
	}
	if cfg.Source.FTPS.Username != "ftpuser" {
		t.Errorf("FTPS.Username = %q, want ftpuser", cfg.Source.FTPS.Username)
	}
	if cfg.Source.FTPS.Password != "ftppass" {
		t.Errorf("FTPS.Password = %q, want ftppass", cfg.Source.FTPS.Password)
	}
}

func TestApplyCliFlags_SourceBlob(t *testing.T) {
	resetTransferFlags()
	sourceType = "blob"
	sourceAccountName = "srcaccount"
	sourceContainer = "srccontainer"
	sourceAuthType = "shared_key"
	sourceAccountKey = "srcaccountkey=="

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.Type != config.StorageTypeBlobStorage {
		t.Errorf("Source.Type = %v, want %v", cfg.Source.Type, config.StorageTypeBlobStorage)
	}
	if cfg.Source.BlobStorage == nil {
		t.Fatal("Source.BlobStorage should not be nil")
	}
	if cfg.Source.BlobStorage.AccountName != "srcaccount" {
		t.Errorf("BlobStorage.AccountName = %q, want srcaccount", cfg.Source.BlobStorage.AccountName)
	}
	if cfg.Source.BlobStorage.ContainerName != "srccontainer" {
		t.Errorf("BlobStorage.ContainerName = %q, want srccontainer", cfg.Source.BlobStorage.ContainerName)
	}
	if cfg.Source.BlobStorage.AccountKey != "srcaccountkey==" {
		t.Errorf("BlobStorage.AccountKey = %q, want srcaccountkey==", cfg.Source.BlobStorage.AccountKey)
	}
}

func TestApplyCliFlags_DestSFTP(t *testing.T) {
	resetTransferFlags()
	destType = "sftp"
	destHost = "dest.sftp.example.com"
	destPort = 22
	destUsername = "destuser"
	destPassword = "destpass"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.Type != config.StorageTypeSFTP {
		t.Errorf("Destination.Type = %v, want %v", cfg.Destination.Type, config.StorageTypeSFTP)
	}
	if cfg.Destination.SFTP == nil {
		t.Fatal("Destination.SFTP should not be nil")
	}
	if cfg.Destination.SFTP.Host != "dest.sftp.example.com" {
		t.Errorf("SFTP.Host = %q, want dest.sftp.example.com", cfg.Destination.SFTP.Host)
	}
	if cfg.Destination.SFTP.Port != 22 {
		t.Errorf("SFTP.Port = %d, want 22", cfg.Destination.SFTP.Port)
	}
	if cfg.Destination.SFTP.Username != "destuser" {
		t.Errorf("SFTP.Username = %q, want destuser", cfg.Destination.SFTP.Username)
	}
	if cfg.Destination.SFTP.Password != "destpass" {
		t.Errorf("SFTP.Password = %q, want destpass", cfg.Destination.SFTP.Password)
	}
}

func TestApplyCliFlags_DestBlob(t *testing.T) {
	resetTransferFlags()
	destType = "blob"
	destAccountName = "destaccount"
	destContainer = "destcontainer"
	destAuthType = "managed_identity"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.Type != config.StorageTypeBlobStorage {
		t.Errorf("Destination.Type = %v, want %v", cfg.Destination.Type, config.StorageTypeBlobStorage)
	}
	if cfg.Destination.BlobStorage == nil {
		t.Fatal("Destination.BlobStorage should not be nil")
	}
	if cfg.Destination.BlobStorage.AccountName != "destaccount" {
		t.Errorf("BlobStorage.AccountName = %q, want destaccount", cfg.Destination.BlobStorage.AccountName)
	}
	if cfg.Destination.BlobStorage.ContainerName != "destcontainer" {
		t.Errorf("BlobStorage.ContainerName = %q, want destcontainer", cfg.Destination.BlobStorage.ContainerName)
	}
	if cfg.Destination.BlobStorage.AuthType != config.BlobAuthManagedIdentity {
		t.Errorf("BlobStorage.AuthType = %q, want %q", cfg.Destination.BlobStorage.AuthType, config.BlobAuthManagedIdentity)
	}
}

func TestApplyCliFlags_Compress(t *testing.T) {
	resetTransferFlags()
	compress = "gzip"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Transfer.Compression != "gzip" {
		t.Errorf("Transfer.Compression = %q, want gzip", cfg.Transfer.Compression)
	}
}

func TestCliFlag_WinsOver_EnvVar(t *testing.T) {
	// Verify precedence: CLI flags must win over env vars for the same field.
	// The call order in runTransfer is LoadFromEnv then applyCliFlags.
	t.Setenv("NIXCOPY_SOURCE_TYPE", "sftp")
	t.Setenv("NIXCOPY_SOURCE_HOST", "env-host")

	resetTransferFlags()
	sourceType = "sftp"
	sourceHost = "cli-host"
	sourcePort = 22
	sourceUsername = "user"

	cfg := config.DefaultConfig()
	config.LoadFromEnv(cfg)
	applyCliFlags(cfg)

	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP should not be nil")
	}
	if cfg.Source.SFTP.Host != "cli-host" {
		t.Errorf("Source.SFTP.Host = %q, want cli-host (CLI flag must win over env var)", cfg.Source.SFTP.Host)
	}
}

func TestApplyCliFlags_Defaults(t *testing.T) {
	sourceType = ""
	destType = ""
	bufferSize = 0
	concurrentFiles = 0
	retryAttempts = 0

	cfg := &config.Config{}
	applyCliFlags(cfg)

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
