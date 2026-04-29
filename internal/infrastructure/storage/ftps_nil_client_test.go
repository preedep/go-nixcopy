package storage_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func assertNotConnectedError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error when client is not connected")
	}
	if !strings.Contains(err.Error(), "not connected") {
		t.Errorf("error = %v, want message containing 'not connected'", err)
	}
}

func TestFTPSStorage_List_NilClient(t *testing.T) {
	s := storage.NewFTPSStorage(&config.FTPSConfig{Host: "h", Port: 21})
	_, err := s.List(context.Background(), "/path")
	assertNotConnectedError(t, err)
}

func TestFTPSStorage_Read_NilClient(t *testing.T) {
	s := storage.NewFTPSStorage(&config.FTPSConfig{Host: "h", Port: 21})
	_, _, err := s.Read(context.Background(), "/path/file.txt")
	assertNotConnectedError(t, err)
}

func TestFTPSStorage_Stat_NilClient(t *testing.T) {
	s := storage.NewFTPSStorage(&config.FTPSConfig{Host: "h", Port: 21})
	_, err := s.Stat(context.Background(), "/path/file.txt")
	assertNotConnectedError(t, err)
}

func TestFTPSStorage_Delete_NilClient(t *testing.T) {
	s := storage.NewFTPSStorage(&config.FTPSConfig{Host: "h", Port: 21})
	err := s.Delete(context.Background(), "/path/file.txt")
	assertNotConnectedError(t, err)
}

func TestFTPSStorage_Disconnect_NilClient(t *testing.T) {
	s := storage.NewFTPSStorage(&config.FTPSConfig{Host: "h", Port: 21})
	// Disconnect with no active connection must return nil, not panic
	if err := s.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect on nil client: %v", err)
	}
}

func TestFTPSStorage_Write_NilClient_RootPath(t *testing.T) {
	s := storage.NewFTPSStorage(&config.FTPSConfig{Host: "h", Port: 21})
	err := s.Write(context.Background(), "file.txt", bytes.NewReader(nil), 0)
	assertNotConnectedError(t, err)
}
