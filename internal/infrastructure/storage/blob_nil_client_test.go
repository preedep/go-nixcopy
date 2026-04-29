package storage_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func TestBlobStorage_List_NilClient(t *testing.T) {
	s := storage.NewBlobStorage(&config.BlobConfig{AccountName: "a", ContainerName: "c"})
	_, err := s.List(context.Background(), "prefix/")
	if err == nil {
		t.Fatal("expected error when blob client is not connected")
	}
}

func TestBlobStorage_Read_NilClient(t *testing.T) {
	s := storage.NewBlobStorage(&config.BlobConfig{AccountName: "a", ContainerName: "c"})
	_, _, err := s.Read(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when blob client is not connected")
	}
}

func TestBlobStorage_Stat_NilClient(t *testing.T) {
	s := storage.NewBlobStorage(&config.BlobConfig{AccountName: "a", ContainerName: "c"})
	_, err := s.Stat(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when blob client is not connected")
	}
}

func TestBlobStorage_Write_NilClient(t *testing.T) {
	s := storage.NewBlobStorage(&config.BlobConfig{AccountName: "a", ContainerName: "c"})
	err := s.Write(context.Background(), "file.txt", bytes.NewReader(nil), 0)
	if err == nil {
		t.Fatal("expected error when blob client is not connected")
	}
}

func TestBlobStorage_Delete_NilClient(t *testing.T) {
	s := storage.NewBlobStorage(&config.BlobConfig{AccountName: "a", ContainerName: "c"})
	err := s.Delete(context.Background(), "file.txt")
	if err == nil {
		t.Fatal("expected error when blob client is not connected")
	}
}

func TestBlobStorage_Disconnect_ReturnsNil(t *testing.T) {
	s := storage.NewBlobStorage(&config.BlobConfig{AccountName: "a", ContainerName: "c"})
	if err := s.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect = %v, want nil", err)
	}
}

func TestBlobStorage_CreateDirectory_ReturnsNil(t *testing.T) {
	s := storage.NewBlobStorage(&config.BlobConfig{AccountName: "a", ContainerName: "c"})
	if err := s.CreateDirectory(context.Background(), "some/prefix/"); err != nil {
		t.Errorf("CreateDirectory = %v, want nil (blob storage is a no-op)", err)
	}
}
