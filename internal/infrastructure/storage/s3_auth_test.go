package storage_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func TestS3Storage_Connect_AccessKey_MissingFields(t *testing.T) {
	cases := []struct {
		name      string
		accessKey string
		secretKey string
	}{
		{"missing both", "", ""},
		{"missing access_key_id", "", "secret"},
		{"missing secret_access_key", "key", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.S3Config{
				Region:          "us-east-1",
				Bucket:          "test-bucket",
				AuthType:        config.S3AuthAccessKey,
				AccessKeyID:     tc.accessKey,
				SecretAccessKey: tc.secretKey,
			}
			err := storage.NewS3Storage(cfg).Connect(context.Background())
			if err == nil {
				t.Fatal("expected error for missing access key fields")
			}
		})
	}
}

func TestS3Storage_Connect_AssumeRole_MissingRoleARN(t *testing.T) {
	cfg := &config.S3Config{
		Region:   "us-east-1",
		Bucket:   "test-bucket",
		AuthType: config.S3AuthAssumeRole,
		RoleARN:  "",
	}
	err := storage.NewS3Storage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing role_arn")
	}
	if !strings.Contains(err.Error(), "role_arn") {
		t.Errorf("error should mention role_arn, got: %v", err)
	}
}

func TestS3Storage_Connect_WebIdentity_MissingFields(t *testing.T) {
	cases := []struct {
		name      string
		roleARN   string
		tokenFile string
	}{
		{"missing both", "", ""},
		{"missing token_file", "arn:aws:iam::123456789012:role/Test", ""},
		{"missing role_arn", "", "/var/run/secrets/token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.S3Config{
				Region:               "us-east-1",
				Bucket:               "test-bucket",
				AuthType:             config.S3AuthWebIdentity,
				RoleARN:              tc.roleARN,
				WebIdentityTokenFile: tc.tokenFile,
			}
			err := storage.NewS3Storage(cfg).Connect(context.Background())
			if err == nil {
				t.Fatal("expected error for missing web identity fields")
			}
		})
	}
}

// TestS3Storage_Connect_AssumeRole_BuildsProvider verifies that Connect() wires
// up the STS AssumeRole provider without returning an error. The actual STS call
// is deferred until the first credential retrieval, so no real AWS access is needed.
func TestS3Storage_Connect_AssumeRole_BuildsProvider(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy-secret")
	t.Setenv("AWS_SESSION_TOKEN", "")

	cfg := &config.S3Config{
		Region:          "us-east-1",
		Bucket:          "test-bucket",
		AuthType:        config.S3AuthAssumeRole,
		RoleARN:         "arn:aws:iam::123456789012:role/TestRole",
		RoleSessionName: "test-session",
		ExternalID:      "ext-123",
	}
	err := storage.NewS3Storage(cfg).Connect(context.Background())
	if err != nil {
		t.Fatalf("AssumeRole Connect() returned unexpected error: %v", err)
	}
}

// TestS3Storage_Connect_WebIdentity_BuildsProvider verifies that Connect() wires
// up the web identity provider for IRSA without returning an error. The actual STS
// call is deferred until the first credential retrieval.
func TestS3Storage_Connect_WebIdentity_BuildsProvider(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "dummy-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "dummy-secret")
	t.Setenv("AWS_SESSION_TOKEN", "")

	f, err := os.CreateTemp(t.TempDir(), "web-identity-token-*")
	if err != nil {
		t.Fatalf("create temp token file: %v", err)
	}
	_, _ = f.WriteString("dummy-oidc-token")
	f.Close()

	cfg := &config.S3Config{
		Region:               "us-east-1",
		Bucket:               "test-bucket",
		AuthType:             config.S3AuthWebIdentity,
		RoleARN:              "arn:aws:iam::123456789012:role/TestRole",
		WebIdentityTokenFile: f.Name(),
	}
	err = storage.NewS3Storage(cfg).Connect(context.Background())
	if err != nil {
		t.Fatalf("WebIdentity Connect() returned unexpected error: %v", err)
	}
}
