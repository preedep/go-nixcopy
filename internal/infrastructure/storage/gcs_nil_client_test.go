package storage_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func newDisconnectedGCS() *config.GCSConfig {
	return &config.GCSConfig{
		Bucket: "test-bucket",
	}
}

func TestGCSStorage_List_NilClient(t *testing.T) {
	s := storage.NewGCSStorage(newDisconnectedGCS())
	_, err := s.List(context.Background(), "prefix/")
	if err == nil {
		t.Fatal("expected error when GCS client is not connected")
	}
}

func TestGCSStorage_Read_NilClient(t *testing.T) {
	s := storage.NewGCSStorage(newDisconnectedGCS())
	_, _, err := s.Read(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when GCS client is not connected")
	}
}

func TestGCSStorage_Stat_NilClient(t *testing.T) {
	s := storage.NewGCSStorage(newDisconnectedGCS())
	_, err := s.Stat(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when GCS client is not connected")
	}
}

func TestGCSStorage_Write_NilClient(t *testing.T) {
	s := storage.NewGCSStorage(newDisconnectedGCS())
	err := s.Write(context.Background(), "file.txt", bytes.NewReader(nil), 0)
	if err == nil {
		t.Fatal("expected error when GCS client is not connected")
	}
}

func TestGCSStorage_Delete_NilClient(t *testing.T) {
	s := storage.NewGCSStorage(newDisconnectedGCS())
	err := s.Delete(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when GCS client is not connected")
	}
}

func TestGCSStorage_Disconnect_ReturnsNil(t *testing.T) {
	s := storage.NewGCSStorage(newDisconnectedGCS())
	if err := s.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect = %v, want nil", err)
	}
}

func TestGCSStorage_CreateDirectory_ReturnsNil(t *testing.T) {
	s := storage.NewGCSStorage(newDisconnectedGCS())
	if err := s.CreateDirectory(context.Background(), "prefix/"); err != nil {
		t.Errorf("CreateDirectory = %v, want nil (GCS is a no-op)", err)
	}
}
