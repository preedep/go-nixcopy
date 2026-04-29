package storage_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func newDisconnectedS3() *config.S3Config {
	return &config.S3Config{
		Region: "us-east-1",
		Bucket: "test-bucket",
	}
}

func TestS3Storage_List_NilClient(t *testing.T) {
	s := storage.NewS3Storage(newDisconnectedS3())
	_, err := s.List(context.Background(), "prefix/")
	if err == nil {
		t.Fatal("expected error when S3 client is not connected")
	}
}

func TestS3Storage_Read_NilClient(t *testing.T) {
	s := storage.NewS3Storage(newDisconnectedS3())
	_, _, err := s.Read(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when S3 client is not connected")
	}
}

func TestS3Storage_Stat_NilClient(t *testing.T) {
	s := storage.NewS3Storage(newDisconnectedS3())
	_, err := s.Stat(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when S3 client is not connected")
	}
}

func TestS3Storage_Write_NilClient(t *testing.T) {
	s := storage.NewS3Storage(newDisconnectedS3())
	err := s.Write(context.Background(), "file.txt", bytes.NewReader(nil), 0)
	if err == nil {
		t.Fatal("expected error when S3 client is not connected")
	}
}

func TestS3Storage_Delete_NilClient(t *testing.T) {
	s := storage.NewS3Storage(newDisconnectedS3())
	err := s.Delete(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when S3 client is not connected")
	}
}

func TestS3Storage_Disconnect_ReturnsNil(t *testing.T) {
	s := storage.NewS3Storage(newDisconnectedS3())
	if err := s.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect = %v, want nil", err)
	}
}

func TestS3Storage_CreateDirectory_ReturnsNil(t *testing.T) {
	s := storage.NewS3Storage(newDisconnectedS3())
	if err := s.CreateDirectory(context.Background(), "prefix/"); err != nil {
		t.Errorf("CreateDirectory = %v, want nil (S3 is a no-op)", err)
	}
}
