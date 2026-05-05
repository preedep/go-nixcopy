package config

import (
	"os"
	"testing"
	"time"
)

// writeTempYAML writes content to a temp .yaml file and returns its path.
func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cfg-*.yaml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadFile_SFTPSource_LocalDest(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: sftp
  sftp:
    host: sftp.example.com
    port: 2222
    username: user
    password: secret
    private_key_path: /home/user/.ssh/id_rsa
    private_key_passphrase: passphrase
    timeout: 45s
    max_packet_size: 32768
destination:
  type: local
  local:
    base_path: /tmp/dest
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.Type != StorageTypeSFTP {
		t.Errorf("Source.Type = %q, want sftp", cfg.Source.Type)
	}
	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP is nil")
	}
	if cfg.Source.SFTP.Host != "sftp.example.com" {
		t.Errorf("SFTP.Host = %q, want sftp.example.com", cfg.Source.SFTP.Host)
	}
	if cfg.Source.SFTP.Port != 2222 {
		t.Errorf("SFTP.Port = %d, want 2222", cfg.Source.SFTP.Port)
	}
	if cfg.Source.SFTP.Username != "user" {
		t.Errorf("SFTP.Username = %q, want user", cfg.Source.SFTP.Username)
	}
	if cfg.Source.SFTP.Password != "secret" {
		t.Errorf("SFTP.Password = %q, want secret", cfg.Source.SFTP.Password)
	}
	if cfg.Source.SFTP.PrivateKeyPath != "/home/user/.ssh/id_rsa" {
		t.Errorf("SFTP.PrivateKeyPath = %q, want /home/user/.ssh/id_rsa", cfg.Source.SFTP.PrivateKeyPath)
	}
	if cfg.Source.SFTP.PrivateKeyPass != "passphrase" {
		t.Errorf("SFTP.PrivateKeyPass = %q, want passphrase", cfg.Source.SFTP.PrivateKeyPass)
	}
	if cfg.Source.SFTP.Timeout != 45*time.Second {
		t.Errorf("SFTP.Timeout = %v, want 45s", cfg.Source.SFTP.Timeout)
	}
	if cfg.Source.SFTP.MaxPacketSize != 32768 {
		t.Errorf("SFTP.MaxPacketSize = %d, want 32768", cfg.Source.SFTP.MaxPacketSize)
	}

	if cfg.Destination.Type != StorageTypeLocal {
		t.Errorf("Destination.Type = %q, want local", cfg.Destination.Type)
	}
	if cfg.Destination.Local == nil {
		t.Fatal("Destination.Local is nil")
	}
	if cfg.Destination.Local.BasePath != "/tmp/dest" {
		t.Errorf("Local.BasePath = %q, want /tmp/dest", cfg.Destination.Local.BasePath)
	}
}

func TestLoadFile_FTPSSource_S3Dest(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: ftps
  ftps:
    host: ftps.example.com
    port: 21
    username: ftpuser
    password: ftppass
    timeout: 30s
    tls_mode: explicit
    skip_verify: true
destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-bucket
    auth_type: access_key
    access_key_id: AKIAIOSFODNN7EXAMPLE
    secret_access_key: wJalrXUtnFEMI
    endpoint: http://minio:9000
    use_path_style: true
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.Type != StorageTypeFTPS {
		t.Errorf("Source.Type = %q, want ftps", cfg.Source.Type)
	}
	if cfg.Source.FTPS == nil {
		t.Fatal("Source.FTPS is nil")
	}
	if cfg.Source.FTPS.Host != "ftps.example.com" {
		t.Errorf("FTPS.Host = %q, want ftps.example.com", cfg.Source.FTPS.Host)
	}
	if cfg.Source.FTPS.Port != 21 {
		t.Errorf("FTPS.Port = %d, want 21", cfg.Source.FTPS.Port)
	}
	if cfg.Source.FTPS.TLSMode != "explicit" {
		t.Errorf("FTPS.TLSMode = %q, want explicit", cfg.Source.FTPS.TLSMode)
	}
	if !cfg.Source.FTPS.SkipVerify {
		t.Error("FTPS.SkipVerify = false, want true")
	}
	if cfg.Source.FTPS.Timeout != 30*time.Second {
		t.Errorf("FTPS.Timeout = %v, want 30s", cfg.Source.FTPS.Timeout)
	}

	if cfg.Destination.Type != StorageTypeS3 {
		t.Errorf("Destination.Type = %q, want s3", cfg.Destination.Type)
	}
	if cfg.Destination.S3 == nil {
		t.Fatal("Destination.S3 is nil")
	}
	if cfg.Destination.S3.Region != "ap-southeast-1" {
		t.Errorf("S3.Region = %q, want ap-southeast-1", cfg.Destination.S3.Region)
	}
	if cfg.Destination.S3.Bucket != "my-bucket" {
		t.Errorf("S3.Bucket = %q, want my-bucket", cfg.Destination.S3.Bucket)
	}
	if cfg.Destination.S3.AuthType != S3AuthAccessKey {
		t.Errorf("S3.AuthType = %q, want access_key", cfg.Destination.S3.AuthType)
	}
	if cfg.Destination.S3.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("S3.AccessKeyID = %q, want AKIAIOSFODNN7EXAMPLE", cfg.Destination.S3.AccessKeyID)
	}
	if cfg.Destination.S3.SecretAccessKey != "wJalrXUtnFEMI" {
		t.Errorf("S3.SecretAccessKey = %q", cfg.Destination.S3.SecretAccessKey)
	}
	if cfg.Destination.S3.Endpoint != "http://minio:9000" {
		t.Errorf("S3.Endpoint = %q, want http://minio:9000", cfg.Destination.S3.Endpoint)
	}
	if !cfg.Destination.S3.UsePathStyle {
		t.Error("S3.UsePathStyle = false, want true")
	}
}

func TestLoadFile_BlobSource_SharedKey(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: blob
  blob:
    account_name: mystorageaccount
    container_name: mycontainer
    auth_type: shared_key
    account_key: base64key==
    endpoint: https://mystorageaccount.blob.core.windows.net/
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.Type != StorageTypeBlobStorage {
		t.Errorf("Source.Type = %q, want blob", cfg.Source.Type)
	}
	if cfg.Source.BlobStorage == nil {
		t.Fatal("Source.BlobStorage is nil")
	}
	if cfg.Source.BlobStorage.AccountName != "mystorageaccount" {
		t.Errorf("BlobStorage.AccountName = %q", cfg.Source.BlobStorage.AccountName)
	}
	if cfg.Source.BlobStorage.ContainerName != "mycontainer" {
		t.Errorf("BlobStorage.ContainerName = %q", cfg.Source.BlobStorage.ContainerName)
	}
	if cfg.Source.BlobStorage.AuthType != BlobAuthSharedKey {
		t.Errorf("BlobStorage.AuthType = %q, want shared_key", cfg.Source.BlobStorage.AuthType)
	}
	if cfg.Source.BlobStorage.AccountKey != "base64key==" {
		t.Errorf("BlobStorage.AccountKey = %q", cfg.Source.BlobStorage.AccountKey)
	}
	if cfg.Source.BlobStorage.Endpoint != "https://mystorageaccount.blob.core.windows.net/" {
		t.Errorf("BlobStorage.Endpoint = %q", cfg.Source.BlobStorage.Endpoint)
	}
}

func TestLoadFile_BlobSource_ServicePrincipal(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: blob
  blob:
    account_name: myaccount
    container_name: mycontainer
    auth_type: service_principal
    tenant_id: tenant-123
    client_id: client-456
    client_secret: client-secret-789
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.BlobStorage == nil {
		t.Fatal("Source.BlobStorage is nil")
	}
	if cfg.Source.BlobStorage.AuthType != BlobAuthServicePrincipal {
		t.Errorf("BlobStorage.AuthType = %q, want service_principal", cfg.Source.BlobStorage.AuthType)
	}
	if cfg.Source.BlobStorage.TenantID != "tenant-123" {
		t.Errorf("BlobStorage.TenantID = %q, want tenant-123", cfg.Source.BlobStorage.TenantID)
	}
	if cfg.Source.BlobStorage.ClientID != "client-456" {
		t.Errorf("BlobStorage.ClientID = %q, want client-456", cfg.Source.BlobStorage.ClientID)
	}
	if cfg.Source.BlobStorage.ClientSecret != "client-secret-789" {
		t.Errorf("BlobStorage.ClientSecret = %q", cfg.Source.BlobStorage.ClientSecret)
	}
}

func TestLoadFile_BlobDest_ManagedIdentity(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: local
  local:
    base_path: /tmp/src
destination:
  type: blob
  blob:
    account_name: destaccount
    container_name: destcontainer
    auth_type: managed_identity
    use_managed_identity: true
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Destination.Type != StorageTypeBlobStorage {
		t.Errorf("Destination.Type = %q, want blob", cfg.Destination.Type)
	}
	if cfg.Destination.BlobStorage == nil {
		t.Fatal("Destination.BlobStorage is nil")
	}
	if cfg.Destination.BlobStorage.AuthType != BlobAuthManagedIdentity {
		t.Errorf("BlobStorage.AuthType = %q, want managed_identity", cfg.Destination.BlobStorage.AuthType)
	}
	if !cfg.Destination.BlobStorage.UseManagedIdentity {
		t.Error("BlobStorage.UseManagedIdentity = false, want true")
	}
}

func TestLoadFile_GCSSource_ServiceAccount(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: gcs
  gcs:
    project_id: my-gcp-project
    bucket: my-bucket
    auth_type: service_account
    credentials_file: /path/to/sa.json
    endpoint: http://gcs-emulator:4443
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.Type != StorageTypeGCS {
		t.Errorf("Source.Type = %q, want gcs", cfg.Source.Type)
	}
	if cfg.Source.GCS == nil {
		t.Fatal("Source.GCS is nil")
	}
	if cfg.Source.GCS.ProjectID != "my-gcp-project" {
		t.Errorf("GCS.ProjectID = %q", cfg.Source.GCS.ProjectID)
	}
	if cfg.Source.GCS.Bucket != "my-bucket" {
		t.Errorf("GCS.Bucket = %q", cfg.Source.GCS.Bucket)
	}
	if cfg.Source.GCS.AuthType != GCSAuthServiceAccount {
		t.Errorf("GCS.AuthType = %q, want service_account", cfg.Source.GCS.AuthType)
	}
	if cfg.Source.GCS.CredentialsFile != "/path/to/sa.json" {
		t.Errorf("GCS.CredentialsFile = %q", cfg.Source.GCS.CredentialsFile)
	}
	if cfg.Source.GCS.Endpoint != "http://gcs-emulator:4443" {
		t.Errorf("GCS.Endpoint = %q", cfg.Source.GCS.Endpoint)
	}
}

func TestLoadFile_GCSSource_Impersonate(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: gcs
  gcs:
    project_id: my-project
    bucket: src-bucket
    auth_type: impersonate
    impersonate_service_account: worker@my-project.iam.gserviceaccount.com
    delegates:
      - mid@my-project.iam.gserviceaccount.com
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.GCS == nil {
		t.Fatal("Source.GCS is nil")
	}
	if cfg.Source.GCS.AuthType != GCSAuthImpersonate {
		t.Errorf("GCS.AuthType = %q, want impersonate", cfg.Source.GCS.AuthType)
	}
	if cfg.Source.GCS.ImpersonateServiceAccount != "worker@my-project.iam.gserviceaccount.com" {
		t.Errorf("GCS.ImpersonateServiceAccount = %q", cfg.Source.GCS.ImpersonateServiceAccount)
	}
	if len(cfg.Source.GCS.Delegates) != 1 || cfg.Source.GCS.Delegates[0] != "mid@my-project.iam.gserviceaccount.com" {
		t.Errorf("GCS.Delegates = %v, want [mid@my-project.iam.gserviceaccount.com]", cfg.Source.GCS.Delegates)
	}
}

func TestLoadFile_GCSDest_ApplicationDefault(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: local
  local:
    base_path: /tmp/src
destination:
  type: gcs
  gcs:
    project_id: dst-project
    bucket: dst-bucket
    auth_type: application_default
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Destination.Type != StorageTypeGCS {
		t.Errorf("Destination.Type = %q, want gcs", cfg.Destination.Type)
	}
	if cfg.Destination.GCS == nil {
		t.Fatal("Destination.GCS is nil")
	}
	if cfg.Destination.GCS.AuthType != GCSAuthApplicationDefault {
		t.Errorf("GCS.AuthType = %q, want application_default", cfg.Destination.GCS.AuthType)
	}
	if cfg.Destination.GCS.Bucket != "dst-bucket" {
		t.Errorf("GCS.Bucket = %q, want dst-bucket", cfg.Destination.GCS.Bucket)
	}
}

func TestLoadFile_S3Source_WebIdentity(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: s3
  s3:
    region: us-east-1
    bucket: src-bucket
    auth_type: web_identity
    web_identity_token_file: /var/run/secrets/token
    role_arn: arn:aws:iam::123456789012:role/MyRole
    role_session_name: my-session
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.S3 == nil {
		t.Fatal("Source.S3 is nil")
	}
	if cfg.Source.S3.AuthType != S3AuthWebIdentity {
		t.Errorf("S3.AuthType = %q, want web_identity", cfg.Source.S3.AuthType)
	}
	if cfg.Source.S3.WebIdentityTokenFile != "/var/run/secrets/token" {
		t.Errorf("S3.WebIdentityTokenFile = %q", cfg.Source.S3.WebIdentityTokenFile)
	}
	if cfg.Source.S3.RoleARN != "arn:aws:iam::123456789012:role/MyRole" {
		t.Errorf("S3.RoleARN = %q", cfg.Source.S3.RoleARN)
	}
	if cfg.Source.S3.RoleSessionName != "my-session" {
		t.Errorf("S3.RoleSessionName = %q", cfg.Source.S3.RoleSessionName)
	}
}

func TestLoadFile_TransferSettings(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: local
  local:
    base_path: /tmp/src
destination:
  type: local
  local:
    base_path: /tmp/dst
transfer:
  buffer_size: 67108864
  concurrent_files: 8
  retry_attempts: 5
  retry_delay: 10s
  timeout: 2h
  verify_checksum: true
  enable_resume: true
  skip_existing: true
  bandwidth_limit: 10485760
  compression: gzip
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

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
	if cfg.Transfer.Timeout != 2*time.Hour {
		t.Errorf("Transfer.Timeout = %v, want 2h", cfg.Transfer.Timeout)
	}
	if !cfg.Transfer.VerifyChecksum {
		t.Error("Transfer.VerifyChecksum = false, want true")
	}
	if !cfg.Transfer.EnableResume {
		t.Error("Transfer.EnableResume = false, want true")
	}
	if !cfg.Transfer.SkipExisting {
		t.Error("Transfer.SkipExisting = false, want true")
	}
	if cfg.Transfer.BandwidthLimit != 10485760 {
		t.Errorf("Transfer.BandwidthLimit = %d, want 10485760", cfg.Transfer.BandwidthLimit)
	}
	if cfg.Transfer.Compression != "gzip" {
		t.Errorf("Transfer.Compression = %q, want gzip", cfg.Transfer.Compression)
	}
}

func TestLoadFile_LoggingConfig(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: local
  local:
    base_path: /tmp/src
destination:
  type: local
  local:
    base_path: /tmp/dst
logging:
  level: debug
  format: console
  output_path: /var/log/nixcopy.log
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want debug", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "console" {
		t.Errorf("Logging.Format = %q, want console", cfg.Logging.Format)
	}
	if cfg.Logging.OutputPath != "/var/log/nixcopy.log" {
		t.Errorf("Logging.OutputPath = %q, want /var/log/nixcopy.log", cfg.Logging.OutputPath)
	}
}

func TestLoadFile_EnvVarExpansion(t *testing.T) {
	t.Setenv("TEST_SFTP_HOST", "env-sftp.example.com")
	t.Setenv("TEST_SFTP_PASS", "env-secret")
	t.Setenv("TEST_S3_KEY", "ENVAKIAKEY")
	t.Setenv("TEST_S3_SECRET", "envsecret")

	path := writeTempYAML(t, `
source:
  type: sftp
  sftp:
    host: ${TEST_SFTP_HOST}
    port: 22
    username: user
    password: ${TEST_SFTP_PASS}
destination:
  type: s3
  s3:
    region: us-east-1
    bucket: my-bucket
    auth_type: access_key
    access_key_id: ${TEST_S3_KEY}
    secret_access_key: ${TEST_S3_SECRET}
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.SFTP == nil {
		t.Fatal("Source.SFTP is nil")
	}
	if cfg.Source.SFTP.Host != "env-sftp.example.com" {
		t.Errorf("SFTP.Host = %q, want env-sftp.example.com (env var not expanded)", cfg.Source.SFTP.Host)
	}
	if cfg.Source.SFTP.Password != "env-secret" {
		t.Errorf("SFTP.Password = %q, want env-secret (env var not expanded)", cfg.Source.SFTP.Password)
	}
	if cfg.Destination.S3 == nil {
		t.Fatal("Destination.S3 is nil")
	}
	if cfg.Destination.S3.AccessKeyID != "ENVAKIAKEY" {
		t.Errorf("S3.AccessKeyID = %q, want ENVAKIAKEY (env var not expanded)", cfg.Destination.S3.AccessKeyID)
	}
	if cfg.Destination.S3.SecretAccessKey != "envsecret" {
		t.Errorf("S3.SecretAccessKey = %q, want envsecret (env var not expanded)", cfg.Destination.S3.SecretAccessKey)
	}
}

func TestLoadFile_DefaultsPreservedWhenTransferOmitted(t *testing.T) {
	// When the YAML omits the transfer section entirely, DefaultConfig values must survive.
	path := writeTempYAML(t, `
source:
  type: local
  local:
    base_path: /tmp/src
destination:
  type: local
  local:
    base_path: /tmp/dst
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Transfer.BufferSize != 32*1024*1024 {
		t.Errorf("Transfer.BufferSize = %d, want %d (default)", cfg.Transfer.BufferSize, 32*1024*1024)
	}
	if cfg.Transfer.ConcurrentFiles != 4 {
		t.Errorf("Transfer.ConcurrentFiles = %d, want 4 (default)", cfg.Transfer.ConcurrentFiles)
	}
	if cfg.Transfer.RetryAttempts != 3 {
		t.Errorf("Transfer.RetryAttempts = %d, want 3 (default)", cfg.Transfer.RetryAttempts)
	}
	if cfg.Transfer.RetryDelay != 5*time.Second {
		t.Errorf("Transfer.RetryDelay = %v, want 5s (default)", cfg.Transfer.RetryDelay)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want info (default)", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want json (default)", cfg.Logging.Format)
	}
}

func TestLoadFile_PartialTransferOverridesDefaults(t *testing.T) {
	// A transfer section that sets only some fields must override those fields
	// while leaving unset fields at their DefaultConfig values.
	path := writeTempYAML(t, `
source:
  type: local
  local:
    base_path: /tmp/src
destination:
  type: local
  local:
    base_path: /tmp/dst
transfer:
  concurrent_files: 16
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Transfer.ConcurrentFiles != 16 {
		t.Errorf("Transfer.ConcurrentFiles = %d, want 16", cfg.Transfer.ConcurrentFiles)
	}
	// Unspecified fields must retain DefaultConfig values.
	if cfg.Transfer.BufferSize != 32*1024*1024 {
		t.Errorf("Transfer.BufferSize = %d, want %d (default unchanged)", cfg.Transfer.BufferSize, 32*1024*1024)
	}
	if cfg.Transfer.RetryAttempts != 3 {
		t.Errorf("Transfer.RetryAttempts = %d, want 3 (default unchanged)", cfg.Transfer.RetryAttempts)
	}
}

func TestLoadFile_FileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadFile returned nil error for non-existent file")
	}
}

func TestLoadFile_MalformedYAML(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: sftp
  sftp:
    host: [invalid yaml
`)
	_, err := LoadFile(path)
	if err == nil {
		t.Error("LoadFile returned nil error for malformed YAML")
	}
}

func TestLoadFile_BlobSource_SASToken(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: blob
  blob:
    account_name: myaccount
    container_name: mycontainer
    auth_type: sas_token
    sas_token: sv=2020-08-04&ss=b&srt=sco&sp=rwdlacupiytfx&se=2024-12-31T00:00:00Z
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.BlobStorage == nil {
		t.Fatal("Source.BlobStorage is nil")
	}
	if cfg.Source.BlobStorage.AuthType != BlobAuthSASToken {
		t.Errorf("BlobStorage.AuthType = %q, want sas_token", cfg.Source.BlobStorage.AuthType)
	}
	if cfg.Source.BlobStorage.SASToken == "" {
		t.Error("BlobStorage.SASToken is empty")
	}
}

func TestLoadFile_BlobSource_ConnectionString(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: blob
  blob:
    account_name: devstoreaccount1
    container_name: testcontainer
    auth_type: connection_string
    connection_string: "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.BlobStorage == nil {
		t.Fatal("Source.BlobStorage is nil")
	}
	if cfg.Source.BlobStorage.AuthType != BlobAuthConnectionString {
		t.Errorf("BlobStorage.AuthType = %q, want connection_string", cfg.Source.BlobStorage.AuthType)
	}
	if cfg.Source.BlobStorage.ConnectionString == "" {
		t.Error("BlobStorage.ConnectionString is empty")
	}
}

func TestLoadFile_S3Source_AssumeRole(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: s3
  s3:
    region: us-west-2
    bucket: src-bucket
    auth_type: assume_role
    role_arn: arn:aws:iam::111122223333:role/CrossAccountRole
    role_session_name: nixcopy-session
    external_id: ext-12345
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.S3 == nil {
		t.Fatal("Source.S3 is nil")
	}
	if cfg.Source.S3.AuthType != S3AuthAssumeRole {
		t.Errorf("S3.AuthType = %q, want assume_role", cfg.Source.S3.AuthType)
	}
	if cfg.Source.S3.RoleARN != "arn:aws:iam::111122223333:role/CrossAccountRole" {
		t.Errorf("S3.RoleARN = %q", cfg.Source.S3.RoleARN)
	}
	if cfg.Source.S3.ExternalID != "ext-12345" {
		t.Errorf("S3.ExternalID = %q, want ext-12345", cfg.Source.S3.ExternalID)
	}
}

func TestLoadFile_GCSSource_AccessToken(t *testing.T) {
	path := writeTempYAML(t, `
source:
  type: gcs
  gcs:
    project_id: my-project
    bucket: my-bucket
    auth_type: access_token
    access_token: ya29.short-lived-token
destination:
  type: local
  local:
    base_path: /tmp/out
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if cfg.Source.GCS == nil {
		t.Fatal("Source.GCS is nil")
	}
	if cfg.Source.GCS.AuthType != GCSAuthAccessToken {
		t.Errorf("GCS.AuthType = %q, want access_token", cfg.Source.GCS.AuthType)
	}
	if cfg.Source.GCS.AccessToken != "ya29.short-lived-token" {
		t.Errorf("GCS.AccessToken = %q", cfg.Source.GCS.AccessToken)
	}
}

// TestLoadFile_ExampleConfig verifies that the shipped config.example.yaml parses
// without error and produces the expected top-level types.
func TestLoadFile_ExampleConfig(t *testing.T) {
	cfg, err := LoadFile("../../../config.example.yaml")
	if err != nil {
		t.Fatalf("LoadFile(config.example.yaml): %v", err)
	}
	if cfg.Source.Type != StorageTypeSFTP {
		t.Errorf("config.example.yaml Source.Type = %q, want sftp", cfg.Source.Type)
	}
	if cfg.Destination.Type != StorageTypeS3 {
		t.Errorf("config.example.yaml Destination.Type = %q, want s3", cfg.Destination.Type)
	}
}
