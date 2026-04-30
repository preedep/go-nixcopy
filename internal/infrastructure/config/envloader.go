package config

import (
	"os"
	"strconv"
	"time"
)

// LoadFromEnv overlays NIXCOPY_* env vars onto cfg: sets source/dest types and
// transfer settings, then applies backend-specific vars for whatever type is set.
// Precedence: CLI flags > env vars > config file > defaults.
// When the type is supplied via a CLI flag rather than NIXCOPY_SOURCE/DEST_TYPE,
// call ApplyBackendEnv after setting the type so backend vars land on the right backend.
func LoadFromEnv(cfg *Config) {
	if v := os.Getenv("NIXCOPY_SOURCE_TYPE"); v != "" {
		cfg.Source.Type = StorageType(v)
	}
	if v := os.Getenv("NIXCOPY_DEST_TYPE"); v != "" {
		cfg.Destination.Type = StorageType(v)
	}
	applyTransferEnv(&cfg.Transfer)
	ApplyBackendEnv(cfg)
}

// ApplyBackendEnv applies backend-specific NIXCOPY_* env vars using the type
// already set on cfg. Call this after the final source/dest type is resolved
// (e.g. after CLI flags have set the type) so vars go to the correct backend.
func ApplyBackendEnv(cfg *Config) {
	applySourceEnv(&cfg.Source)
	applyDestEnv(&cfg.Destination)
}

func applySourceEnv(src *SourceConfig) {
	switch src.Type {
	case StorageTypeLocal:
		if src.Local == nil {
			src.Local = &LocalConfig{}
		}
		setEnvString(&src.Local.BasePath, "NIXCOPY_SOURCE_BASE_PATH")

	case StorageTypeSFTP:
		if src.SFTP == nil {
			src.SFTP = &SFTPConfig{}
		}
		setEnvString(&src.SFTP.Host, "NIXCOPY_SOURCE_HOST")
		setEnvInt(&src.SFTP.Port, "NIXCOPY_SOURCE_PORT")
		setEnvString(&src.SFTP.Username, "NIXCOPY_SOURCE_USERNAME")
		setEnvString(&src.SFTP.Password, "NIXCOPY_SOURCE_PASSWORD")
		setEnvString(&src.SFTP.PrivateKeyPath, "NIXCOPY_SOURCE_PRIVATE_KEY")
		setEnvString(&src.SFTP.PrivateKeyPass, "NIXCOPY_SOURCE_PRIVATE_KEY_PASS")
		setEnvDuration(&src.SFTP.Timeout, "NIXCOPY_SOURCE_TIMEOUT")
		setEnvInt(&src.SFTP.MaxPacketSize, "NIXCOPY_SOURCE_MAX_PACKET_SIZE")

	case StorageTypeFTPS:
		if src.FTPS == nil {
			src.FTPS = &FTPSConfig{}
		}
		setEnvString(&src.FTPS.Host, "NIXCOPY_SOURCE_HOST")
		setEnvInt(&src.FTPS.Port, "NIXCOPY_SOURCE_PORT")
		setEnvString(&src.FTPS.Username, "NIXCOPY_SOURCE_USERNAME")
		setEnvString(&src.FTPS.Password, "NIXCOPY_SOURCE_PASSWORD")
		setEnvString(&src.FTPS.TLSMode, "NIXCOPY_SOURCE_TLS_MODE")
		setEnvBool(&src.FTPS.SkipVerify, "NIXCOPY_SOURCE_SKIP_VERIFY")
		setEnvDuration(&src.FTPS.Timeout, "NIXCOPY_SOURCE_TIMEOUT")

	case StorageTypeS3:
		if src.S3 == nil {
			src.S3 = &S3Config{}
		}
		setEnvString(&src.S3.Region, "NIXCOPY_SOURCE_REGION")
		setEnvString(&src.S3.Bucket, "NIXCOPY_SOURCE_BUCKET")
		setEnvString(&src.S3.Endpoint, "NIXCOPY_SOURCE_ENDPOINT")
		setEnvBool(&src.S3.UsePathStyle, "NIXCOPY_SOURCE_USE_PATH_STYLE")
		if v := os.Getenv("NIXCOPY_SOURCE_AUTH_TYPE"); v != "" {
			src.S3.AuthType = S3AuthType(v)
		}
		setEnvString(&src.S3.AccessKeyID, "NIXCOPY_SOURCE_ACCESS_KEY")
		setEnvString(&src.S3.SecretAccessKey, "NIXCOPY_SOURCE_SECRET_KEY")
		setEnvString(&src.S3.SessionToken, "NIXCOPY_SOURCE_SESSION_TOKEN")
		setEnvString(&src.S3.RoleARN, "NIXCOPY_SOURCE_ROLE_ARN")
		setEnvString(&src.S3.RoleSessionName, "NIXCOPY_SOURCE_ROLE_SESSION_NAME")
		setEnvString(&src.S3.ExternalID, "NIXCOPY_SOURCE_EXTERNAL_ID")
		setEnvString(&src.S3.WebIdentityTokenFile, "NIXCOPY_SOURCE_WEB_IDENTITY_TOKEN_FILE")
		setEnvString(&src.S3.Profile, "NIXCOPY_SOURCE_PROFILE")

	case StorageTypeBlobStorage:
		if src.BlobStorage == nil {
			src.BlobStorage = &BlobConfig{}
		}
		setEnvString(&src.BlobStorage.AccountName, "NIXCOPY_SOURCE_ACCOUNT_NAME")
		setEnvString(&src.BlobStorage.ContainerName, "NIXCOPY_SOURCE_CONTAINER")
		setEnvString(&src.BlobStorage.Endpoint, "NIXCOPY_SOURCE_ENDPOINT")
		if v := os.Getenv("NIXCOPY_SOURCE_AUTH_TYPE"); v != "" {
			src.BlobStorage.AuthType = BlobAuthType(v)
		}
		setEnvString(&src.BlobStorage.AccountKey, "NIXCOPY_SOURCE_ACCOUNT_KEY")
		setEnvString(&src.BlobStorage.SASToken, "NIXCOPY_SOURCE_SAS_TOKEN")
		setEnvString(&src.BlobStorage.ConnectionString, "NIXCOPY_SOURCE_CONNECTION_STRING")
		setEnvString(&src.BlobStorage.TenantID, "NIXCOPY_SOURCE_TENANT_ID")
		setEnvString(&src.BlobStorage.ClientID, "NIXCOPY_SOURCE_CLIENT_ID")
		setEnvString(&src.BlobStorage.ClientSecret, "NIXCOPY_SOURCE_CLIENT_SECRET")
		setEnvBool(&src.BlobStorage.UseManagedIdentity, "NIXCOPY_SOURCE_USE_MANAGED_IDENTITY")

	case StorageTypeGCS:
		if src.GCS == nil {
			src.GCS = &GCSConfig{}
		}
		setEnvString(&src.GCS.ProjectID, "NIXCOPY_SOURCE_GCS_PROJECT")
		setEnvString(&src.GCS.Bucket, "NIXCOPY_SOURCE_BUCKET")
		setEnvString(&src.GCS.Endpoint, "NIXCOPY_SOURCE_ENDPOINT")
		if v := os.Getenv("NIXCOPY_SOURCE_AUTH_TYPE"); v != "" {
			src.GCS.AuthType = GCSAuthType(v)
		}
		setEnvString(&src.GCS.CredentialsFile, "NIXCOPY_SOURCE_CREDENTIALS_FILE")
		setEnvString(&src.GCS.CredentialsJSON, "NIXCOPY_SOURCE_CREDENTIALS_JSON")
		setEnvString(&src.GCS.ImpersonateServiceAccount, "NIXCOPY_SOURCE_IMPERSONATE_SA")
		setEnvString(&src.GCS.AccessToken, "NIXCOPY_SOURCE_ACCESS_TOKEN")
	}
}

func applyDestEnv(dst *DestinationConfig) {
	switch dst.Type {
	case StorageTypeLocal:
		if dst.Local == nil {
			dst.Local = &LocalConfig{}
		}
		setEnvString(&dst.Local.BasePath, "NIXCOPY_DEST_BASE_PATH")

	case StorageTypeSFTP:
		if dst.SFTP == nil {
			dst.SFTP = &SFTPConfig{}
		}
		setEnvString(&dst.SFTP.Host, "NIXCOPY_DEST_HOST")
		setEnvInt(&dst.SFTP.Port, "NIXCOPY_DEST_PORT")
		setEnvString(&dst.SFTP.Username, "NIXCOPY_DEST_USERNAME")
		setEnvString(&dst.SFTP.Password, "NIXCOPY_DEST_PASSWORD")
		setEnvString(&dst.SFTP.PrivateKeyPath, "NIXCOPY_DEST_PRIVATE_KEY")
		setEnvString(&dst.SFTP.PrivateKeyPass, "NIXCOPY_DEST_PRIVATE_KEY_PASS")
		setEnvDuration(&dst.SFTP.Timeout, "NIXCOPY_DEST_TIMEOUT")
		setEnvInt(&dst.SFTP.MaxPacketSize, "NIXCOPY_DEST_MAX_PACKET_SIZE")

	case StorageTypeFTPS:
		if dst.FTPS == nil {
			dst.FTPS = &FTPSConfig{}
		}
		setEnvString(&dst.FTPS.Host, "NIXCOPY_DEST_HOST")
		setEnvInt(&dst.FTPS.Port, "NIXCOPY_DEST_PORT")
		setEnvString(&dst.FTPS.Username, "NIXCOPY_DEST_USERNAME")
		setEnvString(&dst.FTPS.Password, "NIXCOPY_DEST_PASSWORD")
		setEnvString(&dst.FTPS.TLSMode, "NIXCOPY_DEST_TLS_MODE")
		setEnvBool(&dst.FTPS.SkipVerify, "NIXCOPY_DEST_SKIP_VERIFY")
		setEnvDuration(&dst.FTPS.Timeout, "NIXCOPY_DEST_TIMEOUT")

	case StorageTypeS3:
		if dst.S3 == nil {
			dst.S3 = &S3Config{}
		}
		setEnvString(&dst.S3.Region, "NIXCOPY_DEST_REGION")
		setEnvString(&dst.S3.Bucket, "NIXCOPY_DEST_BUCKET")
		setEnvString(&dst.S3.Endpoint, "NIXCOPY_DEST_ENDPOINT")
		setEnvBool(&dst.S3.UsePathStyle, "NIXCOPY_DEST_USE_PATH_STYLE")
		if v := os.Getenv("NIXCOPY_DEST_AUTH_TYPE"); v != "" {
			dst.S3.AuthType = S3AuthType(v)
		}
		setEnvString(&dst.S3.AccessKeyID, "NIXCOPY_DEST_ACCESS_KEY")
		setEnvString(&dst.S3.SecretAccessKey, "NIXCOPY_DEST_SECRET_KEY")
		setEnvString(&dst.S3.SessionToken, "NIXCOPY_DEST_SESSION_TOKEN")
		setEnvString(&dst.S3.RoleARN, "NIXCOPY_DEST_ROLE_ARN")
		setEnvString(&dst.S3.RoleSessionName, "NIXCOPY_DEST_ROLE_SESSION_NAME")
		setEnvString(&dst.S3.ExternalID, "NIXCOPY_DEST_EXTERNAL_ID")
		setEnvString(&dst.S3.WebIdentityTokenFile, "NIXCOPY_DEST_WEB_IDENTITY_TOKEN_FILE")
		setEnvString(&dst.S3.Profile, "NIXCOPY_DEST_PROFILE")

	case StorageTypeBlobStorage:
		if dst.BlobStorage == nil {
			dst.BlobStorage = &BlobConfig{}
		}
		setEnvString(&dst.BlobStorage.AccountName, "NIXCOPY_DEST_ACCOUNT_NAME")
		setEnvString(&dst.BlobStorage.ContainerName, "NIXCOPY_DEST_CONTAINER")
		setEnvString(&dst.BlobStorage.Endpoint, "NIXCOPY_DEST_ENDPOINT")
		if v := os.Getenv("NIXCOPY_DEST_AUTH_TYPE"); v != "" {
			dst.BlobStorage.AuthType = BlobAuthType(v)
		}
		setEnvString(&dst.BlobStorage.AccountKey, "NIXCOPY_DEST_ACCOUNT_KEY")
		setEnvString(&dst.BlobStorage.SASToken, "NIXCOPY_DEST_SAS_TOKEN")
		setEnvString(&dst.BlobStorage.ConnectionString, "NIXCOPY_DEST_CONNECTION_STRING")
		setEnvString(&dst.BlobStorage.TenantID, "NIXCOPY_DEST_TENANT_ID")
		setEnvString(&dst.BlobStorage.ClientID, "NIXCOPY_DEST_CLIENT_ID")
		setEnvString(&dst.BlobStorage.ClientSecret, "NIXCOPY_DEST_CLIENT_SECRET")
		setEnvBool(&dst.BlobStorage.UseManagedIdentity, "NIXCOPY_DEST_USE_MANAGED_IDENTITY")

	case StorageTypeGCS:
		if dst.GCS == nil {
			dst.GCS = &GCSConfig{}
		}
		setEnvString(&dst.GCS.ProjectID, "NIXCOPY_DEST_GCS_PROJECT")
		setEnvString(&dst.GCS.Bucket, "NIXCOPY_DEST_BUCKET")
		setEnvString(&dst.GCS.Endpoint, "NIXCOPY_DEST_ENDPOINT")
		if v := os.Getenv("NIXCOPY_DEST_AUTH_TYPE"); v != "" {
			dst.GCS.AuthType = GCSAuthType(v)
		}
		setEnvString(&dst.GCS.CredentialsFile, "NIXCOPY_DEST_CREDENTIALS_FILE")
		setEnvString(&dst.GCS.CredentialsJSON, "NIXCOPY_DEST_CREDENTIALS_JSON")
		setEnvString(&dst.GCS.ImpersonateServiceAccount, "NIXCOPY_DEST_IMPERSONATE_SA")
		setEnvString(&dst.GCS.AccessToken, "NIXCOPY_DEST_ACCESS_TOKEN")
	}
}

func applyTransferEnv(t *TransferConfig) {
	setEnvInt(&t.BufferSize, "NIXCOPY_BUFFER_SIZE")
	setEnvInt(&t.ConcurrentFiles, "NIXCOPY_CONCURRENT_FILES")
	setEnvInt(&t.RetryAttempts, "NIXCOPY_RETRY_ATTEMPTS")
	setEnvDuration(&t.RetryDelay, "NIXCOPY_RETRY_DELAY")
	setEnvDuration(&t.Timeout, "NIXCOPY_TIMEOUT")
	setEnvBool(&t.VerifyChecksum, "NIXCOPY_VERIFY_CHECKSUM")
	setEnvBool(&t.EnableResume, "NIXCOPY_ENABLE_RESUME")
	setEnvBool(&t.SkipExisting, "NIXCOPY_SKIP_EXISTING")
	setEnvInt64(&t.BandwidthLimit, "NIXCOPY_BANDWIDTH_LIMIT")
	setEnvString(&t.Compression, "NIXCOPY_COMPRESSION")
}

func setEnvString(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func setEnvInt(dst *int, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*dst = n
		}
	}
}

func setEnvBool(dst *bool, key string) {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			*dst = b
		}
	}
}

func setEnvInt64(dst *int64, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			*dst = n
		}
	}
}

func setEnvDuration(dst *time.Duration, key string) {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			*dst = d
		}
	}
}
