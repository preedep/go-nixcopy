package cli

import (
	"os"
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
	sourcePath = ""
	sourcePaths = nil
	destPath = ""

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
	sourceTLSMode = ""

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
	destTLSMode = ""

	sourceGCSProject = ""
	sourceCredentialsFile = ""
	sourceImpersonateSA = ""
	sourceAccessToken = ""

	destGCSProject = ""
	destCredentialsFile = ""
	destImpersonateSA = ""
	destAccessToken = ""

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
	if sourceType != "" {
		cfg.Source.Type = config.StorageType(sourceType)
	}
	config.ApplyBackendEnv(cfg)
	applyCliFlags(cfg)

	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP should not be nil")
	}
	if cfg.Source.SFTP.Host != "cli-host" {
		t.Errorf("Source.SFTP.Host = %q, want cli-host (CLI flag must win over env var)", cfg.Source.SFTP.Host)
	}
}

// runTransferEnvFlow simulates the env+CLI resolution sequence from runTransfer,
// without connecting to any real storage backend.
func runTransferEnvFlow(cfg *config.Config) {
	config.LoadFromEnv(cfg)
	if sourceType != "" {
		cfg.Source.Type = config.StorageType(sourceType)
	}
	if destType != "" {
		cfg.Destination.Type = config.StorageType(destType)
	}
	config.ApplyBackendEnv(cfg)
	applyCliFlags(cfg)
}

func TestEnvSourceField_AppliedWhenTypeFromCliFlag(t *testing.T) {
	// Regression test for the bug where NIXCOPY_SOURCE_HOST was silently ignored
	// when the type came from --source-type rather than NIXCOPY_SOURCE_TYPE.
	t.Setenv("NIXCOPY_SOURCE_HOST", "env-host")
	// NIXCOPY_SOURCE_TYPE intentionally NOT set — type comes from CLI flag only

	resetTransferFlags()
	sourceType = "sftp"
	sourceUsername = "cliuser"

	cfg := config.DefaultConfig()
	runTransferEnvFlow(cfg)

	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP is nil")
	}
	if cfg.Source.SFTP.Host != "env-host" {
		t.Errorf("SFTP.Host = %q, want env-host (NIXCOPY_SOURCE_HOST must apply when type from --source-type)", cfg.Source.SFTP.Host)
	}
	if cfg.Source.SFTP.Username != "cliuser" {
		t.Errorf("SFTP.Username = %q, want cliuser", cfg.Source.SFTP.Username)
	}
}

func TestEnvDestField_AppliedWhenTypeFromCliFlag(t *testing.T) {
	t.Setenv("NIXCOPY_DEST_REGION", "us-west-2")
	t.Setenv("NIXCOPY_DEST_BUCKET", "env-bucket")
	// NIXCOPY_DEST_TYPE intentionally NOT set

	resetTransferFlags()
	destType = "s3"
	destAccessKey = "CLIKEY"

	cfg := config.DefaultConfig()
	runTransferEnvFlow(cfg)

	if cfg.Destination.S3 == nil {
		t.Fatal("Destination.S3 is nil")
	}
	if cfg.Destination.S3.Region != "us-west-2" {
		t.Errorf("S3.Region = %q, want us-west-2 (NIXCOPY_DEST_REGION must apply when type from --dest-type)", cfg.Destination.S3.Region)
	}
	if cfg.Destination.S3.Bucket != "env-bucket" {
		t.Errorf("S3.Bucket = %q, want env-bucket", cfg.Destination.S3.Bucket)
	}
	if cfg.Destination.S3.AccessKeyID != "CLIKEY" {
		t.Errorf("S3.AccessKeyID = %q, want CLIKEY (CLI field must win over unset env)", cfg.Destination.S3.AccessKeyID)
	}
}

func TestCliField_WinsOver_EnvField_WhenTypeFromCliFlag(t *testing.T) {
	// Both CLI flag and env var supply the same field; CLI must win.
	t.Setenv("NIXCOPY_SOURCE_HOST", "env-host")
	// NIXCOPY_SOURCE_TYPE NOT set

	resetTransferFlags()
	sourceType = "sftp"
	sourceHost = "cli-host" // must beat env-host
	sourceUsername = "user"

	cfg := config.DefaultConfig()
	runTransferEnvFlow(cfg)

	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP is nil")
	}
	if cfg.Source.SFTP.Host != "cli-host" {
		t.Errorf("SFTP.Host = %q, want cli-host (CLI flag must win over env var)", cfg.Source.SFTP.Host)
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

func TestApplyCliFlags_SourceLocal(t *testing.T) {
	resetTransferFlags()
	sourceType = "local"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.Type != config.StorageTypeLocal {
		t.Errorf("Source.Type = %v, want %v", cfg.Source.Type, config.StorageTypeLocal)
	}
}

func TestApplyCliFlags_DestLocal(t *testing.T) {
	resetTransferFlags()
	destType = "local"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.Type != config.StorageTypeLocal {
		t.Errorf("Destination.Type = %v, want %v", cfg.Destination.Type, config.StorageTypeLocal)
	}
}

func TestApplyCliFlags_SourceSFTP_PrivateKey(t *testing.T) {
	resetTransferFlags()
	sourceType = "sftp"
	sourceHost = "sftp.example.com"
	sourceUsername = "user"
	sourcePrivateKey = "/home/user/.ssh/id_rsa"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP is nil")
	}
	if cfg.Source.SFTP.PrivateKeyPath != "/home/user/.ssh/id_rsa" {
		t.Errorf("SFTP.PrivateKeyPath = %q, want /home/user/.ssh/id_rsa", cfg.Source.SFTP.PrivateKeyPath)
	}
}

func TestApplyCliFlags_DestSFTP_PrivateKey(t *testing.T) {
	resetTransferFlags()
	destType = "sftp"
	destHost = "dest.sftp.example.com"
	destUsername = "destuser"
	destPrivateKey = "/home/user/.ssh/id_rsa"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.SFTP == nil {
		t.Fatal("Destination.SFTP is nil")
	}
	if cfg.Destination.SFTP.PrivateKeyPath != "/home/user/.ssh/id_rsa" {
		t.Errorf("SFTP.PrivateKeyPath = %q, want /home/user/.ssh/id_rsa", cfg.Destination.SFTP.PrivateKeyPath)
	}
}

func TestApplyCliFlags_DestFTPS(t *testing.T) {
	resetTransferFlags()
	destType = "ftps"
	destHost = "ftps-dest.example.com"
	destPort = 21
	destUsername = "ftpuser"
	destPassword = "ftppass"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.Type != config.StorageTypeFTPS {
		t.Errorf("Destination.Type = %v, want %v", cfg.Destination.Type, config.StorageTypeFTPS)
	}
	if cfg.Destination.FTPS == nil {
		t.Fatal("Destination.FTPS is nil")
	}
	if cfg.Destination.FTPS.Host != "ftps-dest.example.com" {
		t.Errorf("FTPS.Host = %q, want ftps-dest.example.com", cfg.Destination.FTPS.Host)
	}
	if cfg.Destination.FTPS.Port != 21 {
		t.Errorf("FTPS.Port = %d, want 21", cfg.Destination.FTPS.Port)
	}
}

func TestApplyCliFlags_SourceS3(t *testing.T) {
	resetTransferFlags()
	sourceType = "s3"
	sourceRegion = "ap-southeast-1"
	sourceBucket = "src-bucket"
	sourceAccessKey = "AKIAIOSFODNN7EXAMPLE"
	sourceSecretKey = "wJalrXUtnFEMI"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.Type != config.StorageTypeS3 {
		t.Errorf("Source.Type = %v, want %v", cfg.Source.Type, config.StorageTypeS3)
	}
	if cfg.Source.S3 == nil {
		t.Fatal("Source.S3 is nil")
	}
	if cfg.Source.S3.Region != "ap-southeast-1" {
		t.Errorf("S3.Region = %q, want ap-southeast-1", cfg.Source.S3.Region)
	}
	if cfg.Source.S3.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("S3.AccessKeyID = %q", cfg.Source.S3.AccessKeyID)
	}
	if cfg.Source.S3.SecretAccessKey != "wJalrXUtnFEMI" {
		t.Errorf("S3.SecretAccessKey = %q", cfg.Source.S3.SecretAccessKey)
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

// ---- ${ENV_VAR} expansion in path flags ----

func TestApplyCliFlags_SourcePrivateKey_ExpandsEnvVar(t *testing.T) {
	resetTransferFlags()
	t.Setenv("TEST_KEY_DIR", "/home/user/.ssh")

	sourceType = "sftp"
	sourceHost = "sftp.example.com"
	sourceUsername = "user"
	sourcePrivateKey = "${TEST_KEY_DIR}/id_rsa"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	want := "/home/user/.ssh/id_rsa"
	if cfg.Source.SFTP.PrivateKeyPath != want {
		t.Errorf("Source.SFTP.PrivateKeyPath = %q, want %q", cfg.Source.SFTP.PrivateKeyPath, want)
	}
}

func TestApplyCliFlags_DestPrivateKey_ExpandsEnvVar(t *testing.T) {
	resetTransferFlags()
	t.Setenv("TEST_KEY_DIR", "/home/user/.ssh")

	destType = "sftp"
	destHost = "sftp.example.com"
	destUsername = "user"
	destPrivateKey = "${TEST_KEY_DIR}/id_rsa"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	want := "/home/user/.ssh/id_rsa"
	if cfg.Destination.SFTP.PrivateKeyPath != want {
		t.Errorf("Destination.SFTP.PrivateKeyPath = %q, want %q", cfg.Destination.SFTP.PrivateKeyPath, want)
	}
}

func TestApplyCliFlags_SourceCredentialsFile_ExpandsEnvVar(t *testing.T) {
	resetTransferFlags()
	t.Setenv("TEST_CREDS_DIR", "/etc/gcp")

	sourceType = "gcs"
	sourceBucket = "my-bucket"
	sourceCredentialsFile = "${TEST_CREDS_DIR}/sa-key.json"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	want := "/etc/gcp/sa-key.json"
	if cfg.Source.GCS.CredentialsFile != want {
		t.Errorf("Source.GCS.CredentialsFile = %q, want %q", cfg.Source.GCS.CredentialsFile, want)
	}
}

func TestApplyCliFlags_DestCredentialsFile_ExpandsEnvVar(t *testing.T) {
	resetTransferFlags()
	t.Setenv("TEST_CREDS_DIR", "/etc/gcp")

	destType = "gcs"
	destBucket = "my-bucket"
	destCredentialsFile = "${TEST_CREDS_DIR}/sa-key.json"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	want := "/etc/gcp/sa-key.json"
	if cfg.Destination.GCS.CredentialsFile != want {
		t.Errorf("Destination.GCS.CredentialsFile = %q, want %q", cfg.Destination.GCS.CredentialsFile, want)
	}
}

func TestApplyCliFlags_PathFlag_PlainStringUnchanged(t *testing.T) {
	resetTransferFlags()

	sourceType = "sftp"
	sourceHost = "sftp.example.com"
	sourceUsername = "user"
	sourcePrivateKey = "/home/user/.ssh/id_rsa"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	want := "/home/user/.ssh/id_rsa"
	if cfg.Source.SFTP.PrivateKeyPath != want {
		t.Errorf("Source.SFTP.PrivateKeyPath = %q, want %q (plain path must not be mangled)", cfg.Source.SFTP.PrivateKeyPath, want)
	}
}

// ---- expandTransferPaths ----

func TestExpandTransferPaths_SourcePath(t *testing.T) {
	resetTransferFlags()
	t.Setenv("TEST_DATA_DIR", "/mnt/data")

	sourcePath = "${TEST_DATA_DIR}/input"
	destPath = "/output"
	expandTransferPaths()

	want := "/mnt/data/input"
	if sourcePath != want {
		t.Errorf("sourcePath = %q, want %q", sourcePath, want)
	}
}

func TestExpandTransferPaths_SourcePaths_Slice(t *testing.T) {
	resetTransferFlags()
	t.Setenv("TEST_DATA_DIR", "/mnt/data")

	sourcePaths = []string{"${TEST_DATA_DIR}/a.txt", "${TEST_DATA_DIR}/b.txt", "/static/c.txt"}
	destPath = "/out"
	expandTransferPaths()

	wants := []string{"/mnt/data/a.txt", "/mnt/data/b.txt", "/static/c.txt"}
	for i, want := range wants {
		if sourcePaths[i] != want {
			t.Errorf("sourcePaths[%d] = %q, want %q", i, sourcePaths[i], want)
		}
	}
}

func TestExpandTransferPaths_DestPath(t *testing.T) {
	resetTransferFlags()
	t.Setenv("TEST_OUT_DIR", "/mnt/output")

	sourcePath = "/input"
	destPath = "${TEST_OUT_DIR}/result"
	expandTransferPaths()

	want := "/mnt/output/result"
	if destPath != want {
		t.Errorf("destPath = %q, want %q", destPath, want)
	}
}

// ---- GCS impersonate / access-token flags ----

func TestApplyCliFlags_SourceGCS_ProjectID(t *testing.T) {
	resetTransferFlags()
	sourceType = "gcs"
	sourceBucket = "src-bucket"
	sourceGCSProject = "my-gcp-project"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.GCS == nil {
		t.Fatal("Source.GCS is nil")
	}
	if cfg.Source.GCS.ProjectID != "my-gcp-project" {
		t.Errorf("ProjectID = %q, want my-gcp-project", cfg.Source.GCS.ProjectID)
	}
}

func TestApplyCliFlags_DestGCS_ProjectID(t *testing.T) {
	resetTransferFlags()
	destType = "gcs"
	destBucket = "dest-bucket"
	destGCSProject = "dest-gcp-project"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.GCS == nil {
		t.Fatal("Destination.GCS is nil")
	}
	if cfg.Destination.GCS.ProjectID != "dest-gcp-project" {
		t.Errorf("ProjectID = %q, want dest-gcp-project", cfg.Destination.GCS.ProjectID)
	}
}

func TestApplyCliFlags_SourceGCS_ImpersonateSA(t *testing.T) {
	resetTransferFlags()
	sourceType = "gcs"
	sourceBucket = "src-bucket"
	sourceImpersonateSA = "sa@project.iam.gserviceaccount.com"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.GCS == nil {
		t.Fatal("Source.GCS is nil")
	}
	if cfg.Source.GCS.ImpersonateServiceAccount != "sa@project.iam.gserviceaccount.com" {
		t.Errorf("ImpersonateServiceAccount = %q, want sa@project.iam.gserviceaccount.com",
			cfg.Source.GCS.ImpersonateServiceAccount)
	}
}

func TestApplyCliFlags_SourceGCS_AccessToken(t *testing.T) {
	resetTransferFlags()
	sourceType = "gcs"
	sourceBucket = "src-bucket"
	sourceAccessToken = "ya29.token-abc"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Source.GCS == nil {
		t.Fatal("Source.GCS is nil")
	}
	if cfg.Source.GCS.AccessToken != "ya29.token-abc" {
		t.Errorf("AccessToken = %q, want ya29.token-abc", cfg.Source.GCS.AccessToken)
	}
}

func TestApplyCliFlags_DestGCS_ImpersonateSA(t *testing.T) {
	resetTransferFlags()
	destType = "gcs"
	destBucket = "dest-bucket"
	destImpersonateSA = "dest-sa@project.iam.gserviceaccount.com"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.GCS == nil {
		t.Fatal("Destination.GCS is nil")
	}
	if cfg.Destination.GCS.ImpersonateServiceAccount != "dest-sa@project.iam.gserviceaccount.com" {
		t.Errorf("ImpersonateServiceAccount = %q, want dest-sa@project.iam.gserviceaccount.com",
			cfg.Destination.GCS.ImpersonateServiceAccount)
	}
}

func TestApplyCliFlags_DestGCS_AccessToken(t *testing.T) {
	resetTransferFlags()
	destType = "gcs"
	destBucket = "dest-bucket"
	destAccessToken = "ya29.dest-token-xyz"

	cfg := config.DefaultConfig()
	applyCliFlags(cfg)

	if cfg.Destination.GCS == nil {
		t.Fatal("Destination.GCS is nil")
	}
	if cfg.Destination.GCS.AccessToken != "ya29.dest-token-xyz" {
		t.Errorf("AccessToken = %q, want ya29.dest-token-xyz", cfg.Destination.GCS.AccessToken)
	}
}

func TestExpandTransferPaths_PWD(t *testing.T) {
	resetTransferFlags()
	pwd, _ := os.Getwd()

	sourcePath = "${PWD}/files"
	destPath = "${PWD}/output"
	expandTransferPaths()

	if sourcePath != pwd+"/files" {
		t.Errorf("sourcePath = %q, want %q", sourcePath, pwd+"/files")
	}
	if destPath != pwd+"/output" {
		t.Errorf("destPath = %q, want %q", destPath, pwd+"/output")
	}
}
