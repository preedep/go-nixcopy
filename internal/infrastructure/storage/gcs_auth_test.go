package storage_test

import (
	"context"
	"strings"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func TestGCSStorage_Connect_ServiceAccount_MissingCredentials(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:   "test-bucket",
		AuthType: config.GCSAuthServiceAccount,
		// neither CredentialsFile nor CredentialsJSON set
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error when service_account has no credentials")
	}
	if !strings.Contains(err.Error(), "credentials_file") {
		t.Errorf("error should mention credentials_file, got: %v", err)
	}
}

func TestGCSStorage_Connect_Impersonate_MissingServiceAccount(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:   "test-bucket",
		AuthType: config.GCSAuthImpersonate,
		// ImpersonateServiceAccount not set
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error when impersonate_service_account is empty")
	}
	if !strings.Contains(err.Error(), "impersonate_service_account") {
		t.Errorf("error should mention impersonate_service_account, got: %v", err)
	}
}

func TestGCSStorage_Connect_AccessToken_MissingToken(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:   "test-bucket",
		AuthType: config.GCSAuthAccessToken,
		// AccessToken not set
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error when access_token is empty")
	}
	if !strings.Contains(err.Error(), "access_token") {
		t.Errorf("error should mention access_token, got: %v", err)
	}
}

// TestGCSStorage_Connect_AccessToken_BuildsClient verifies that Connect() succeeds
// with a static access token. The token is only validated by GCS on the first API
// call, not at client construction time, so no real GCP credentials are needed.
func TestGCSStorage_Connect_AccessToken_BuildsClient(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:      "test-bucket",
		AuthType:    config.GCSAuthAccessToken,
		AccessToken: "ya29.dummy-access-token",
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect() with access_token returned unexpected error: %v", err)
	}
}

func TestGCSStorage_Connect_UnknownAuthType(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:   "test-bucket",
		AuthType: "invalid_auth",
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for unknown auth type")
	}
	if !strings.Contains(err.Error(), "unsupported GCS auth type") {
		t.Errorf("error should mention unsupported auth type, got: %v", err)
	}
}

// TestGCSStorage_Connect_WorkloadIdentityAlias verifies that "workload_identity"
// is accepted as an alias for application_default and does not return a validation error.
// The ADC lookup may fail in CI without GCP credentials, so we only check
// that the error is not a "unsupported auth type" error.
func TestGCSStorage_Connect_WorkloadIdentityAlias(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:   "test-bucket",
		AuthType: "workload_identity",
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err != nil && strings.Contains(err.Error(), "unsupported GCS auth type") {
		t.Errorf("workload_identity should be accepted as alias for application_default, got: %v", err)
	}
}

// Gap 1: service_account + credentials_file path.
// The file doesn't exist, so storage.NewClient fails — but the credentials_file
// branch inside Connect is covered.
func TestGCSStorage_Connect_ServiceAccount_CredentialsFile_BadPath(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:          "test-bucket",
		AuthType:        config.GCSAuthServiceAccount,
		CredentialsFile: "/tmp/this-file-does-not-exist-nixcopy.json",
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error when credentials file does not exist")
	}
}

// Gap 2: service_account + credentials_json path.
// Invalid JSON causes storage.NewClient to fail — but the credentials_json
// branch inside Connect is covered.
func TestGCSStorage_Connect_ServiceAccount_CredentialsJSON_Invalid(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:          "test-bucket",
		AuthType:        config.GCSAuthServiceAccount,
		CredentialsJSON: `{"type":"service_account","invalid_json":}`,
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid credentials JSON")
	}
}

// Gap 3: impersonate with SA set.
// Without ADC in CI, impersonate.CredentialsTokenSource returns an error —
// but the code path through CredentialsTokenSource and its error branch is covered.
func TestGCSStorage_Connect_Impersonate_WithServiceAccount_FailsWithoutADC(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:                    "test-bucket",
		AuthType:                  config.GCSAuthImpersonate,
		ImpersonateServiceAccount: "sa@my-project.iam.gserviceaccount.com",
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	// Error is expected (no ADC in CI), but must not be "missing impersonate_service_account".
	if err == nil {
		t.Log("Connect succeeded (GCP credentials available in this environment)")
		return
	}
	if strings.Contains(err.Error(), "impersonate_service_account") {
		t.Errorf("should have passed SA validation, got: %v", err)
	}
}

// Gap 4: endpoint option branch.
// access_token auth + non-empty Endpoint exercises the option.WithEndpoint line.
func TestGCSStorage_Connect_WithEndpoint(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:      "test-bucket",
		AuthType:    config.GCSAuthAccessToken,
		AccessToken: "ya29.dummy-token",
		Endpoint:    "https://storage.googleapis.com",
	}
	err := storage.NewGCSStorage(cfg).Connect(context.Background())
	if err != nil {
		t.Fatalf("Connect() with endpoint returned unexpected error: %v", err)
	}
}

// Gap 5: Disconnect when client is connected.
// Uses access_token (no real GCP needed) to build a real client, then disconnects.
func TestGCSStorage_Disconnect_Connected(t *testing.T) {
	cfg := &config.GCSConfig{
		Bucket:      "test-bucket",
		AuthType:    config.GCSAuthAccessToken,
		AccessToken: "ya29.dummy-token",
	}
	s := storage.NewGCSStorage(cfg)
	if err := s.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	if err := s.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect() on connected client returned error: %v", err)
	}
}
