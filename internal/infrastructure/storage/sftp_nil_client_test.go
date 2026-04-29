package storage_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func TestSFTPStorage_List_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	_, err := s.List(context.Background(), "/path")
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_Read_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	_, _, err := s.Read(context.Background(), "/path/file.txt")
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_Stat_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	_, err := s.Stat(context.Background(), "/path/file.txt")
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_Write_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	err := s.Write(context.Background(), "/path/file.txt", bytes.NewReader(nil), 0)
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_Delete_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	err := s.Delete(context.Background(), "/path/file.txt")
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_CreateDirectory_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	err := s.CreateDirectory(context.Background(), "/path/dir")
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_ReadFrom_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	resumer, ok := s.(repository.Resumer)
	if !ok {
		t.Fatal("SFTPStorage does not implement repository.Resumer")
	}
	_, _, err := resumer.ReadFrom(context.Background(), "/path/file.txt", 0)
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_AppendWrite_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	resumer, ok := s.(repository.Resumer)
	if !ok {
		t.Fatal("SFTPStorage does not implement repository.Resumer")
	}
	err := resumer.AppendWrite(context.Background(), "/path/file.txt", bytes.NewReader(nil), 0, 0)
	assertNotConnectedError(t, err)
}

func TestSFTPStorage_Disconnect_NilClient(t *testing.T) {
	s := storage.NewSFTPStorage(&config.SFTPConfig{Host: "h", Port: 22})
	if err := s.Disconnect(context.Background()); err != nil {
		t.Errorf("Disconnect on nil client: %v", err)
	}
}
