package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Transfer.BufferSize != 32*1024*1024 {
		t.Errorf("BufferSize = %v, want %v", cfg.Transfer.BufferSize, 32*1024*1024)
	}

	if cfg.Transfer.ConcurrentFiles != 4 {
		t.Errorf("ConcurrentFiles = %v, want 4", cfg.Transfer.ConcurrentFiles)
	}

	if cfg.Transfer.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %v, want 3", cfg.Transfer.RetryAttempts)
	}

	if cfg.Transfer.RetryDelay != 5*time.Second {
		t.Errorf("RetryDelay = %v, want 5s", cfg.Transfer.RetryDelay)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %v, want info", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %v, want json", cfg.Logging.Format)
	}
}

func TestStorageType(t *testing.T) {
	tests := []struct {
		name string
		st   StorageType
		want string
	}{
		{"SFTP", StorageTypeSFTP, "sftp"},
		{"FTPS", StorageTypeFTPS, "ftps"},
		{"Blob", StorageTypeBlobStorage, "blob"},
		{"S3", StorageTypeS3, "s3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.st) != tt.want {
				t.Errorf("StorageType = %v, want %v", tt.st, tt.want)
			}
		})
	}
}

func TestS3AuthType(t *testing.T) {
	tests := []struct {
		name string
		at   S3AuthType
		want string
	}{
		{"AccessKey", S3AuthAccessKey, "access_key"},
		{"IAMRole", S3AuthIAMRole, "iam_role"},
		{"InstanceProfile", S3AuthInstanceProfile, "instance_profile"},
		{"AssumeRole", S3AuthAssumeRole, "assume_role"},
		{"WebIdentity", S3AuthWebIdentity, "web_identity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.at) != tt.want {
				t.Errorf("S3AuthType = %v, want %v", tt.at, tt.want)
			}
		})
	}
}

func TestBlobAuthType(t *testing.T) {
	tests := []struct {
		name string
		at   BlobAuthType
		want string
	}{
		{"SharedKey", BlobAuthSharedKey, "shared_key"},
		{"SASToken", BlobAuthSASToken, "sas_token"},
		{"ManagedIdentity", BlobAuthManagedIdentity, "managed_identity"},
		{"ServicePrincipal", BlobAuthServicePrincipal, "service_principal"},
		{"ConnectionString", BlobAuthConnectionString, "connection_string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.at) != tt.want {
				t.Errorf("BlobAuthType = %v, want %v", tt.at, tt.want)
			}
		})
	}
}

func TestSFTPConfig(t *testing.T) {
	cfg := SFTPConfig{
		Host:          "sftp.example.com",
		Port:          22,
		Username:      "user",
		Password:      "pass",
		Timeout:       30 * time.Second,
		MaxPacketSize: 32768,
	}

	if cfg.Host != "sftp.example.com" {
		t.Errorf("Host = %v, want sftp.example.com", cfg.Host)
	}

	if cfg.Port != 22 {
		t.Errorf("Port = %v, want 22", cfg.Port)
	}

	if cfg.MaxPacketSize != 32768 {
		t.Errorf("MaxPacketSize = %v, want 32768", cfg.MaxPacketSize)
	}
}

func TestS3Config(t *testing.T) {
	cfg := S3Config{
		Region:          "ap-southeast-1",
		Bucket:          "my-bucket",
		AuthType:        S3AuthAccessKey,
		AccessKeyID:     "AKIA...",
		SecretAccessKey: "secret",
	}

	if cfg.Region != "ap-southeast-1" {
		t.Errorf("Region = %v, want ap-southeast-1", cfg.Region)
	}

	if cfg.Bucket != "my-bucket" {
		t.Errorf("Bucket = %v, want my-bucket", cfg.Bucket)
	}

	if cfg.AuthType != S3AuthAccessKey {
		t.Errorf("AuthType = %v, want %v", cfg.AuthType, S3AuthAccessKey)
	}
}

func TestBlobConfig(t *testing.T) {
	cfg := BlobConfig{
		AccountName:   "mystorageaccount",
		ContainerName: "mycontainer",
		AuthType:      BlobAuthSharedKey,
		AccountKey:    "key123",
	}

	if cfg.AccountName != "mystorageaccount" {
		t.Errorf("AccountName = %v, want mystorageaccount", cfg.AccountName)
	}

	if cfg.ContainerName != "mycontainer" {
		t.Errorf("ContainerName = %v, want mycontainer", cfg.ContainerName)
	}

	if cfg.AuthType != BlobAuthSharedKey {
		t.Errorf("AuthType = %v, want %v", cfg.AuthType, BlobAuthSharedKey)
	}
}
