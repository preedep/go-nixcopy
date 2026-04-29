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

func TestLoadFromEnv_Compression(t *testing.T) {
	t.Setenv("NIXCOPY_COMPRESSION", "gzip")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Transfer.Compression != "gzip" {
		t.Errorf("Transfer.Compression = %q, want gzip", cfg.Transfer.Compression)
	}
}

func TestLoadFromEnv_BandwidthLimit(t *testing.T) {
	t.Setenv("NIXCOPY_BANDWIDTH_LIMIT", "10485760") // 10 MB/s in raw bytes

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Transfer.BandwidthLimit != 10485760 {
		t.Errorf("Transfer.BandwidthLimit = %d, want 10485760", cfg.Transfer.BandwidthLimit)
	}
}

func TestLoadFromEnv_SourceFTPS(t *testing.T) {
	t.Setenv("NIXCOPY_SOURCE_TYPE", "ftps")
	t.Setenv("NIXCOPY_SOURCE_HOST", "ftps.example.com")
	t.Setenv("NIXCOPY_SOURCE_PORT", "990")
	t.Setenv("NIXCOPY_SOURCE_USERNAME", "ftpuser")
	t.Setenv("NIXCOPY_SOURCE_PASSWORD", "ftppass")
	t.Setenv("NIXCOPY_SOURCE_TLS_MODE", "explicit")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Source.Type != StorageTypeFTPS {
		t.Errorf("Source.Type = %q, want %q", cfg.Source.Type, StorageTypeFTPS)
	}
	if cfg.Source.FTPS == nil {
		t.Fatal("Source.FTPS is nil")
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
	if cfg.Source.FTPS.TLSMode != "explicit" {
		t.Errorf("FTPS.TLSMode = %q, want explicit", cfg.Source.FTPS.TLSMode)
	}
}

func TestLoadFromEnv_SourceBlob(t *testing.T) {
	t.Setenv("NIXCOPY_SOURCE_TYPE", "blob")
	t.Setenv("NIXCOPY_SOURCE_ACCOUNT_NAME", "srcaccount")
	t.Setenv("NIXCOPY_SOURCE_CONTAINER", "srccontainer")
	t.Setenv("NIXCOPY_SOURCE_AUTH_TYPE", "shared_key")
	t.Setenv("NIXCOPY_SOURCE_ACCOUNT_KEY", "srcaccountkey==")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Source.Type != StorageTypeBlobStorage {
		t.Errorf("Source.Type = %q, want %q", cfg.Source.Type, StorageTypeBlobStorage)
	}
	if cfg.Source.BlobStorage == nil {
		t.Fatal("Source.BlobStorage is nil")
	}
	if cfg.Source.BlobStorage.AccountName != "srcaccount" {
		t.Errorf("BlobStorage.AccountName = %q, want srcaccount", cfg.Source.BlobStorage.AccountName)
	}
	if cfg.Source.BlobStorage.ContainerName != "srccontainer" {
		t.Errorf("BlobStorage.ContainerName = %q, want srccontainer", cfg.Source.BlobStorage.ContainerName)
	}
	if cfg.Source.BlobStorage.AuthType != BlobAuthSharedKey {
		t.Errorf("BlobStorage.AuthType = %q, want %q", cfg.Source.BlobStorage.AuthType, BlobAuthSharedKey)
	}
	if cfg.Source.BlobStorage.AccountKey != "srcaccountkey==" {
		t.Errorf("BlobStorage.AccountKey = %q, want srcaccountkey==", cfg.Source.BlobStorage.AccountKey)
	}
}

func TestLoadFromEnv_SourceS3(t *testing.T) {
	t.Setenv("NIXCOPY_SOURCE_TYPE", "s3")
	t.Setenv("NIXCOPY_SOURCE_REGION", "us-west-2")
	t.Setenv("NIXCOPY_SOURCE_BUCKET", "src-bucket")
	t.Setenv("NIXCOPY_SOURCE_AUTH_TYPE", "access_key")
	t.Setenv("NIXCOPY_SOURCE_ACCESS_KEY", "AKIASRCKEY")
	t.Setenv("NIXCOPY_SOURCE_SECRET_KEY", "srcsecretkey")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Source.Type != StorageTypeS3 {
		t.Errorf("Source.Type = %q, want %q", cfg.Source.Type, StorageTypeS3)
	}
	if cfg.Source.S3 == nil {
		t.Fatal("Source.S3 is nil")
	}
	if cfg.Source.S3.Region != "us-west-2" {
		t.Errorf("S3.Region = %q, want us-west-2", cfg.Source.S3.Region)
	}
	if cfg.Source.S3.Bucket != "src-bucket" {
		t.Errorf("S3.Bucket = %q, want src-bucket", cfg.Source.S3.Bucket)
	}
	if cfg.Source.S3.AuthType != S3AuthAccessKey {
		t.Errorf("S3.AuthType = %q, want %q", cfg.Source.S3.AuthType, S3AuthAccessKey)
	}
	if cfg.Source.S3.AccessKeyID != "AKIASRCKEY" {
		t.Errorf("S3.AccessKeyID = %q, want AKIASRCKEY", cfg.Source.S3.AccessKeyID)
	}
	if cfg.Source.S3.SecretAccessKey != "srcsecretkey" {
		t.Errorf("S3.SecretAccessKey = %q, want srcsecretkey", cfg.Source.S3.SecretAccessKey)
	}
}

func TestLoadFromEnv_DestSFTP(t *testing.T) {
	t.Setenv("NIXCOPY_DEST_TYPE", "sftp")
	t.Setenv("NIXCOPY_DEST_HOST", "dest.sftp.example.com")
	t.Setenv("NIXCOPY_DEST_PORT", "22")
	t.Setenv("NIXCOPY_DEST_USERNAME", "destuser")
	t.Setenv("NIXCOPY_DEST_PASSWORD", "destpass")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Destination.Type != StorageTypeSFTP {
		t.Errorf("Destination.Type = %q, want %q", cfg.Destination.Type, StorageTypeSFTP)
	}
	if cfg.Destination.SFTP == nil {
		t.Fatal("Destination.SFTP is nil")
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

func TestLoadFromEnv_DestFTPS(t *testing.T) {
	t.Setenv("NIXCOPY_DEST_TYPE", "ftps")
	t.Setenv("NIXCOPY_DEST_HOST", "dest.ftps.example.com")
	t.Setenv("NIXCOPY_DEST_PORT", "21")
	t.Setenv("NIXCOPY_DEST_USERNAME", "destftpuser")
	t.Setenv("NIXCOPY_DEST_PASSWORD", "destftppass")
	t.Setenv("NIXCOPY_DEST_TLS_MODE", "implicit")

	cfg := DefaultConfig()
	LoadFromEnv(cfg)

	if cfg.Destination.Type != StorageTypeFTPS {
		t.Errorf("Destination.Type = %q, want %q", cfg.Destination.Type, StorageTypeFTPS)
	}
	if cfg.Destination.FTPS == nil {
		t.Fatal("Destination.FTPS is nil")
	}
	if cfg.Destination.FTPS.Host != "dest.ftps.example.com" {
		t.Errorf("FTPS.Host = %q, want dest.ftps.example.com", cfg.Destination.FTPS.Host)
	}
	if cfg.Destination.FTPS.Port != 21 {
		t.Errorf("FTPS.Port = %d, want 21", cfg.Destination.FTPS.Port)
	}
	if cfg.Destination.FTPS.Username != "destftpuser" {
		t.Errorf("FTPS.Username = %q, want destftpuser", cfg.Destination.FTPS.Username)
	}
	if cfg.Destination.FTPS.Password != "destftppass" {
		t.Errorf("FTPS.Password = %q, want destftppass", cfg.Destination.FTPS.Password)
	}
	if cfg.Destination.FTPS.TLSMode != "implicit" {
		t.Errorf("FTPS.TLSMode = %q, want implicit", cfg.Destination.FTPS.TLSMode)
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
