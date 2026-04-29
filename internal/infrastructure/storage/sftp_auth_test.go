package storage

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

// writePlainKeyFile generates an RSA private key and writes it unencrypted to path.
func writePlainKeyFile(t *testing.T, path string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		t.Fatalf("ssh.MarshalPrivateKey: %v", err)
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// writeEncryptedKeyFile generates an RSA private key and writes it encrypted with passphrase to path.
func writeEncryptedKeyFile(t *testing.T, path, passphrase string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	block, err := ssh.MarshalPrivateKeyWithPassphrase(priv, "", []byte(passphrase))
	if err != nil {
		t.Fatalf("ssh.MarshalPrivateKeyWithPassphrase: %v", err)
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestBuildAuthMethods_NoAuth(t *testing.T) {
	methods, err := buildAuthMethods(&config.SFTPConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 0 {
		t.Errorf("expected 0 methods, got %d", len(methods))
	}
}

func TestBuildAuthMethods_PasswordOnly(t *testing.T) {
	methods, err := buildAuthMethods(&config.SFTPConfig{Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(methods))
	}
}

func TestBuildAuthMethods_PrivateKeyOnly(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_rsa")
	writePlainKeyFile(t, keyPath)

	methods, err := buildAuthMethods(&config.SFTPConfig{PrivateKeyPath: keyPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(methods))
	}
}

func TestBuildAuthMethods_PrivateKeyWithPassphrase(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_rsa_enc")
	writeEncryptedKeyFile(t, keyPath, "mypassphrase")

	methods, err := buildAuthMethods(&config.SFTPConfig{
		PrivateKeyPath: keyPath,
		PrivateKeyPass: "mypassphrase",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(methods))
	}
}

func TestBuildAuthMethods_PasswordAndPrivateKey(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_rsa")
	writePlainKeyFile(t, keyPath)

	methods, err := buildAuthMethods(&config.SFTPConfig{
		Password:       "secret",
		PrivateKeyPath: keyPath,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(methods))
	}
}

func TestBuildAuthMethods_BadKeyPath(t *testing.T) {
	_, err := buildAuthMethods(&config.SFTPConfig{PrivateKeyPath: "/nonexistent/id_rsa"})
	if err == nil {
		t.Fatal("expected error for nonexistent key path, got nil")
	}
}

func TestBuildAuthMethods_InvalidKeyBytes(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "bad_key")
	if err := os.WriteFile(keyPath, []byte("not a valid pem key"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := buildAuthMethods(&config.SFTPConfig{PrivateKeyPath: keyPath})
	if err == nil {
		t.Fatal("expected error for invalid key bytes, got nil")
	}
}

func TestBuildAuthMethods_WrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_rsa_enc")
	writeEncryptedKeyFile(t, keyPath, "correctpassphrase")

	_, err := buildAuthMethods(&config.SFTPConfig{
		PrivateKeyPath: keyPath,
		PrivateKeyPass: "wrongpassphrase",
	})
	if err == nil {
		t.Fatal("expected error for wrong passphrase, got nil")
	}
}
