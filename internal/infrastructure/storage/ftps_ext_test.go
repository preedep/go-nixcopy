package storage_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func newDisconnectedFTPS() *config.FTPSConfig {
	return &config.FTPSConfig{
		Host:     "localhost",
		Port:     21,
		Username: "user",
		Password: "pass",
	}
}

func TestFTPSStorage_Write_NilClient(t *testing.T) {
	s := storage.NewFTPSStorage(newDisconnectedFTPS())
	err := s.Write(context.Background(), "/some/path/file.txt", bytes.NewReader(nil), 0)
	if err == nil {
		t.Fatal("expected error when client is not connected")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("error = %v, want 'not connected'", err)
	}
}

func TestFTPSStorage_CreateDirectory_NilClient(t *testing.T) {
	s := storage.NewFTPSStorage(newDisconnectedFTPS())
	err := s.CreateDirectory(context.Background(), "/some/nested/dir")
	if err == nil {
		t.Fatal("expected error when client is not connected")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("error = %v, want 'not connected'", err)
	}
}
