package config

import (
	"testing"
	"time"
)

func TestLoadFromEnv_SourceSFTP(t *testing.T) {
	t.Setenv("NIXCOPY_SOURCE_TYPE", "sftp")
	t.Setenv("NIXCOPY_SOURCE_HOST", "sftp.example.com")
	t.Setenv("NIXCOPY_SOURCE_PORT", "2222")
	t.Setenv("NIXCOPY_SOURCE_USERNAME", "user")
	t.Setenv("NIXCOPY_SOURCE_PASSWORD", "secret")
	t.Setenv("NIXCOPY_SOURCE_TIMEOUT", "45s")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Source.Type != StorageTypeSFTP {
		t.Errorf("Source.Type = %q, want %q", cfg.Source.Type, StorageTypeSFTP)
	}
	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP is nil")
	}
	if cfg.Source.SFTP.Host != "sftp.example.com" {
		t.Errorf("SFTP.Host = %q, want %q", cfg.Source.SFTP.Host, "sftp.example.com")
	}
	if cfg.Source.SFTP.Port != 2222 {
		t.Errorf("SFTP.Port = %d, want 2222", cfg.Source.SFTP.Port)
	}
	if cfg.Source.SFTP.Username != "user" {
		t.Errorf("SFTP.Username = %q, want %q", cfg.Source.SFTP.Username, "user")
	}
	if cfg.Source.SFTP.Password != "secret" {
		t.Errorf("SFTP.Password = %q, want %q", cfg.Source.SFTP.Password, "secret")
	}
	if cfg.Source.SFTP.Timeout != 45*time.Second {
		t.Errorf("SFTP.Timeout = %v, want 45s", cfg.Source.SFTP.Timeout)
	}
}

func TestLoadFromEnv_DestS3(t *testing.T) {
	t.Setenv("NIXCOPY_DEST_TYPE", "s3")
	t.Setenv("NIXCOPY_DEST_REGION", "ap-southeast-1")
	t.Setenv("NIXCOPY_DEST_BUCKET", "my-bucket")
	t.Setenv("NIXCOPY_DEST_AUTH_TYPE", "access_key")
	t.Setenv("NIXCOPY_DEST_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE")
	t.Setenv("NIXCOPY_DEST_SECRET_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Destination.Type != StorageTypeS3 {
		t.Errorf("Destination.Type = %q, want %q", cfg.Destination.Type, StorageTypeS3)
	}
	if cfg.Destination.S3 == nil {
		t.Fatal("Destination.S3 is nil")
	}
	if cfg.Destination.S3.Region != "ap-southeast-1" {
		t.Errorf("S3.Region = %q, want %q", cfg.Destination.S3.Region, "ap-southeast-1")
	}
	if cfg.Destination.S3.Bucket != "my-bucket" {
		t.Errorf("S3.Bucket = %q, want %q", cfg.Destination.S3.Bucket, "my-bucket")
	}
	if cfg.Destination.S3.AuthType != S3AuthAccessKey {
		t.Errorf("S3.AuthType = %q, want %q", cfg.Destination.S3.AuthType, S3AuthAccessKey)
	}
	if cfg.Destination.S3.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("S3.AccessKeyID = %q, want AKIAIOSFODNN7EXAMPLE", cfg.Destination.S3.AccessKeyID)
	}
}

func TestLoadFromEnv_DestBlob_ManagedIdentity(t *testing.T) {
	t.Setenv("NIXCOPY_DEST_TYPE", "blob")
	t.Setenv("NIXCOPY_DEST_ACCOUNT_NAME", "mystorageaccount")
	t.Setenv("NIXCOPY_DEST_CONTAINER", "mycontainer")
	t.Setenv("NIXCOPY_DEST_AUTH_TYPE", "managed_identity")
	t.Setenv("NIXCOPY_DEST_USE_MANAGED_IDENTITY", "true")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Destination.BlobStorage == nil {
		t.Fatal("Destination.BlobStorage is nil")
	}
	if cfg.Destination.BlobStorage.AccountName != "mystorageaccount" {
		t.Errorf("BlobStorage.AccountName = %q, want mystorageaccount", cfg.Destination.BlobStorage.AccountName)
	}
	if cfg.Destination.BlobStorage.AuthType != BlobAuthManagedIdentity {
		t.Errorf("BlobStorage.AuthType = %q, want %q", cfg.Destination.BlobStorage.AuthType, BlobAuthManagedIdentity)
	}
	if !cfg.Destination.BlobStorage.UseManagedIdentity {
		t.Error("BlobStorage.UseManagedIdentity = false, want true")
	}
}

func TestLoadFromEnv_TransferSettings(t *testing.T) {
	t.Setenv("NIXCOPY_BUFFER_SIZE", "67108864")
	t.Setenv("NIXCOPY_CONCURRENT_FILES", "8")
	t.Setenv("NIXCOPY_RETRY_ATTEMPTS", "5")
	t.Setenv("NIXCOPY_RETRY_DELAY", "10s")
	t.Setenv("NIXCOPY_VERIFY_CHECKSUM", "true")
	t.Setenv("NIXCOPY_ENABLE_RESUME", "true")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Transfer.BufferSize != 67108864 {
		t.Errorf("Transfer.BufferSize = %d, want 67108864", cfg.Transfer.BufferSize)
	}
	if cfg.Transfer.ConcurrentFiles != 8 {
		t.Errorf("Transfer.ConcurrentFiles = %d, want 8", cfg.Transfer.ConcurrentFiles)
	}
	if cfg.Transfer.RetryAttempts != 5 {
		t.Errorf("Transfer.RetryAttempts = %d, want 5", cfg.Transfer.RetryAttempts)
	}
	if cfg.Transfer.RetryDelay != 10*time.Second {
		t.Errorf("Transfer.RetryDelay = %v, want 10s", cfg.Transfer.RetryDelay)
	}
	if !cfg.Transfer.VerifyChecksum {
		t.Error("Transfer.VerifyChecksum = false, want true")
	}
	if !cfg.Transfer.EnableResume {
		t.Error("Transfer.EnableResume = false, want true")
	}
}

func TestLoadFromEnv_SkipExisting(t *testing.T) {
	t.Setenv("NIXCOPY_SKIP_EXISTING", "true")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if !cfg.Transfer.SkipExisting {
		t.Error("Transfer.SkipExisting = false, want true (NIXCOPY_SKIP_EXISTING=true)")
	}
}

func TestLoadFromEnv_EmptyVarsNoChange(t *testing.T) {
	cfg := DefaultConfig()
	before := cfg.Transfer.BufferSize

	LoadFromEnv(cfg)

	if cfg.Transfer.BufferSize != before {
		t.Errorf("BufferSize changed from %d to %d with no env vars", before, cfg.Transfer.BufferSize)
	}
}

func TestLoadFromEnv_EnvOverridesConfigFile(t *testing.T) {
	// Simulate config-file-loaded value being overridden by env var
	cfg := DefaultConfig()
	cfg.Source.Type = StorageTypeSFTP
	cfg.Source.SFTP = &SFTPConfig{Host: "original.host", Port: 22, Username: "original"}

	t.Setenv("NIXCOPY_SOURCE_HOST", "env.host")
	t.Setenv("NIXCOPY_SOURCE_PORT", "2222")

	LoadFromEnv(cfg)

	if cfg.Source.SFTP.Host != "env.host" {
		t.Errorf("SFTP.Host = %q, want env.host (env should override config file)", cfg.Source.SFTP.Host)
	}
	if cfg.Source.SFTP.Port != 2222 {
		t.Errorf("SFTP.Port = %d, want 2222 (env should override config file)", cfg.Source.SFTP.Port)
	}
	// Unset field should be untouched
	if cfg.Source.SFTP.Username != "original" {
		t.Errorf("SFTP.Username = %q, want original (env not set, should keep config value)", cfg.Source.SFTP.Username)
	}
}
