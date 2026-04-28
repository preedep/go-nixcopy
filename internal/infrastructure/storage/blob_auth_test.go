package storage_test

import (
	"context"
	"strings"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func TestBlobStorage_Connect_SharedKey_MissingAccountKey(t *testing.T) {
	cfg := &config.BlobConfig{
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
		AuthType:      config.BlobAuthSharedKey,
		AccountKey:    "",
	}
	err := storage.NewBlobStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing account_key")
	}
	if !strings.Contains(err.Error(), "account_key") {
		t.Errorf("error should mention account_key, got: %v", err)
	}
}

func TestBlobStorage_Connect_SASToken_MissingToken(t *testing.T) {
	cfg := &config.BlobConfig{
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
		AuthType:      config.BlobAuthSASToken,
		SASToken:      "",
	}
	err := storage.NewBlobStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing sas_token")
	}
	if !strings.Contains(err.Error(), "sas_token") {
		t.Errorf("error should mention sas_token, got: %v", err)
	}
}

func TestBlobStorage_Connect_ConnectionString_Missing(t *testing.T) {
	cfg := &config.BlobConfig{
		ContainerName:    "testcontainer",
		AuthType:         config.BlobAuthConnectionString,
		ConnectionString: "",
	}
	err := storage.NewBlobStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for missing connection_string")
	}
	if !strings.Contains(err.Error(), "connection_string") {
		t.Errorf("error should mention connection_string, got: %v", err)
	}
}

func TestBlobStorage_Connect_ServicePrincipal_MissingFields(t *testing.T) {
	cases := []struct {
		name         string
		tenantID     string
		clientID     string
		clientSecret string
	}{
		{"missing all", "", "", ""},
		{"missing tenant", "", "client-id", "secret"},
		{"missing client_id", "tenant", "", "secret"},
		{"missing secret", "tenant", "client-id", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.BlobConfig{
				AccountName:   "testaccount",
				ContainerName: "testcontainer",
				AuthType:      config.BlobAuthServicePrincipal,
				TenantID:      tc.tenantID,
				ClientID:      tc.clientID,
				ClientSecret:  tc.clientSecret,
			}
			err := storage.NewBlobStorage(cfg).Connect(context.Background())
			if err == nil {
				t.Fatal("expected error for missing service principal fields")
			}
		})
	}
}

func TestBlobStorage_Connect_ManagedIdentity_SystemAssigned(t *testing.T) {
	cfg := &config.BlobConfig{
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
		AuthType:      config.BlobAuthManagedIdentity,
		ClientID:      "", // system-assigned: must NOT pass a ClientID
	}
	// NewManagedIdentityCredential and NewClient only create structs; no network call.
	err := storage.NewBlobStorage(cfg).Connect(context.Background())
	if err != nil {
		t.Fatalf("system-assigned managed identity Connect() returned unexpected error: %v", err)
	}
}

func TestBlobStorage_Connect_ManagedIdentity_UserAssigned(t *testing.T) {
	cfg := &config.BlobConfig{
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
		AuthType:      config.BlobAuthManagedIdentity,
		ClientID:      "00000000-0000-0000-0000-000000000001",
	}
	err := storage.NewBlobStorage(cfg).Connect(context.Background())
	if err != nil {
		t.Fatalf("user-assigned managed identity Connect() returned unexpected error: %v", err)
	}
}

func TestBlobStorage_Connect_UnsupportedAuthType(t *testing.T) {
	cfg := &config.BlobConfig{
		AccountName:   "testaccount",
		ContainerName: "testcontainer",
		AuthType:      "unknown_auth",
	}
	err := storage.NewBlobStorage(cfg).Connect(context.Background())
	if err == nil {
		t.Fatal("expected error for unsupported auth type")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error should mention unsupported, got: %v", err)
	}
}
