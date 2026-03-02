package config

import (
	"time"
)

type Config struct {
	Source      SourceConfig      `yaml:"source" json:"source"`
	Destination DestinationConfig `yaml:"destination" json:"destination"`
	Transfer    TransferConfig    `yaml:"transfer" json:"transfer"`
	Logging     LoggingConfig     `yaml:"logging" json:"logging"`
}

type StorageType string

const (
	StorageTypeSFTP        StorageType = "sftp"
	StorageTypeFTPS        StorageType = "ftps"
	StorageTypeBlobStorage StorageType = "blob"
	StorageTypeS3          StorageType = "s3"
)

type SourceConfig struct {
	Type        StorageType `yaml:"type" json:"type"`
	SFTP        *SFTPConfig `yaml:"sftp,omitempty" json:"sftp,omitempty"`
	FTPS        *FTPSConfig `yaml:"ftps,omitempty" json:"ftps,omitempty"`
	BlobStorage *BlobConfig `yaml:"blob,omitempty" json:"blob,omitempty"`
	S3          *S3Config   `yaml:"s3,omitempty" json:"s3,omitempty"`
}

type DestinationConfig struct {
	Type        StorageType `yaml:"type" json:"type"`
	SFTP        *SFTPConfig `yaml:"sftp,omitempty" json:"sftp,omitempty"`
	FTPS        *FTPSConfig `yaml:"ftps,omitempty" json:"ftps,omitempty"`
	BlobStorage *BlobConfig `yaml:"blob,omitempty" json:"blob,omitempty"`
	S3          *S3Config   `yaml:"s3,omitempty" json:"s3,omitempty"`
}

type SFTPConfig struct {
	Host           string        `yaml:"host" json:"host"`
	Port           int           `yaml:"port" json:"port"`
	Username       string        `yaml:"username" json:"username"`
	Password       string        `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKeyPath string        `yaml:"private_key_path,omitempty" json:"private_key_path,omitempty"`
	PrivateKeyPass string        `yaml:"private_key_passphrase,omitempty" json:"private_key_passphrase,omitempty"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout"`
	MaxPacketSize  int           `yaml:"max_packet_size" json:"max_packet_size"`
}

type FTPSConfig struct {
	Host       string        `yaml:"host" json:"host"`
	Port       int           `yaml:"port" json:"port"`
	Username   string        `yaml:"username" json:"username"`
	Password   string        `yaml:"password" json:"password"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
	TLSMode    string        `yaml:"tls_mode" json:"tls_mode"`
	SkipVerify bool          `yaml:"skip_verify" json:"skip_verify"`
}

type BlobAuthType string

const (
	BlobAuthSharedKey        BlobAuthType = "shared_key"
	BlobAuthSASToken         BlobAuthType = "sas_token"
	BlobAuthManagedIdentity  BlobAuthType = "managed_identity"
	BlobAuthServicePrincipal BlobAuthType = "service_principal"
	BlobAuthConnectionString BlobAuthType = "connection_string"
)

type BlobConfig struct {
	AccountName   string       `yaml:"account_name" json:"account_name"`
	ContainerName string       `yaml:"container_name" json:"container_name"`
	Endpoint      string       `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	AuthType      BlobAuthType `yaml:"auth_type" json:"auth_type"`

	AccountKey       string `yaml:"account_key,omitempty" json:"account_key,omitempty"`
	SASToken         string `yaml:"sas_token,omitempty" json:"sas_token,omitempty"`
	ConnectionString string `yaml:"connection_string,omitempty" json:"connection_string,omitempty"`

	TenantID     string `yaml:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	ClientID     string `yaml:"client_id,omitempty" json:"client_id,omitempty"`
	ClientSecret string `yaml:"client_secret,omitempty" json:"client_secret,omitempty"`

	UseManagedIdentity bool `yaml:"use_managed_identity,omitempty" json:"use_managed_identity,omitempty"`
}

type S3AuthType string

const (
	S3AuthAccessKey       S3AuthType = "access_key"
	S3AuthIAMRole         S3AuthType = "iam_role"
	S3AuthInstanceProfile S3AuthType = "instance_profile"
	S3AuthAssumeRole      S3AuthType = "assume_role"
	S3AuthWebIdentity     S3AuthType = "web_identity"
)

type S3Config struct {
	Region       string     `yaml:"region" json:"region"`
	Bucket       string     `yaml:"bucket" json:"bucket"`
	Endpoint     string     `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	UsePathStyle bool       `yaml:"use_path_style" json:"use_path_style"`
	AuthType     S3AuthType `yaml:"auth_type" json:"auth_type"`

	AccessKeyID     string `yaml:"access_key_id,omitempty" json:"access_key_id,omitempty"`
	SecretAccessKey string `yaml:"secret_access_key,omitempty" json:"secret_access_key,omitempty"`
	SessionToken    string `yaml:"session_token,omitempty" json:"session_token,omitempty"`

	RoleARN         string `yaml:"role_arn,omitempty" json:"role_arn,omitempty"`
	RoleSessionName string `yaml:"role_session_name,omitempty" json:"role_session_name,omitempty"`
	ExternalID      string `yaml:"external_id,omitempty" json:"external_id,omitempty"`

	WebIdentityTokenFile string `yaml:"web_identity_token_file,omitempty" json:"web_identity_token_file,omitempty"`

	Profile string `yaml:"profile,omitempty" json:"profile,omitempty"`
}

type TransferConfig struct {
	BufferSize      int           `yaml:"buffer_size" json:"buffer_size"`
	ConcurrentFiles int           `yaml:"concurrent_files" json:"concurrent_files"`
	RetryAttempts   int           `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay      time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	VerifyChecksum  bool          `yaml:"verify_checksum" json:"verify_checksum"`
}

type LoggingConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"`
	OutputPath string `yaml:"output_path" json:"output_path"`
}

func DefaultConfig() *Config {
	return &Config{
		Transfer: TransferConfig{
			BufferSize:      32 * 1024 * 1024,
			ConcurrentFiles: 4,
			RetryAttempts:   3,
			RetryDelay:      5 * time.Second,
			Timeout:         30 * time.Minute,
			VerifyChecksum:  false,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
		},
	}
}
