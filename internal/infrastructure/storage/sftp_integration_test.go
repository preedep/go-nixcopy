//go:build integration

package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

// newSFTPStore connects using environment variables.
//
// Required env vars (set automatically in CI via the atmoz/sftp service container):
//
//	SFTP_HOST      e.g. localhost
//	SFTP_PORT      e.g. 2222
//	SFTP_USERNAME  e.g. testuser
//	SFTP_PASSWORD  e.g. testpass
func newSFTPStore(t *testing.T) repository.Storage {
	t.Helper()

	host := os.Getenv("SFTP_HOST")
	if host == "" {
		t.Skip("SFTP_HOST not set — skipping SFTP integration tests")
	}

	cfg := &config.SFTPConfig{
		Host:          host,
		Port:          getEnvIntOr("SFTP_PORT", 2222),
		Username:      os.Getenv("SFTP_USERNAME"),
		Password:      os.Getenv("SFTP_PASSWORD"),
		Timeout:       30 * time.Second,
		MaxPacketSize: 32768,
	}

	s := storage.NewSFTPStorage(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("SFTP Connect: %v", err)
	}
	t.Cleanup(func() { s.Disconnect(context.Background()) })
	return s
}

func getEnvIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	fmt.Sscanf(v, "%d", &n)
	if n == 0 {
		return fallback
	}
	return n
}

func TestSFTPStorage_WriteAndRead(t *testing.T) {
	ctx := context.Background()
	s := newSFTPStore(t)

	path := fmt.Sprintf("/upload/integration-%d.txt", time.Now().UnixNano())
	content := []byte("sftp integration test content")

	if err := s.Write(ctx, path, bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("Write: %v", err)
	}
	t.Cleanup(func() { s.Delete(context.Background(), path) })

	rc, size, err := s.Read(ctx, path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	defer rc.Close()

	if size != int64(len(content)) {
		t.Errorf("size = %d, want %d", size, len(content))
	}
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestSFTPStorage_List(t *testing.T) {
	ctx := context.Background()
	s := newSFTPStore(t)

	dir := fmt.Sprintf("/upload/list-%d", time.Now().UnixNano())
	s.CreateDirectory(ctx, dir)

	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("%s/file%d.txt", dir, i)
		data := []byte(name)
		if err := s.Write(ctx, name, bytes.NewReader(data), int64(len(data))); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	files, err := s.List(ctx, dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("List returned %d files, want 3", len(files))
	}
}

func TestSFTPStorage_Delete(t *testing.T) {
	ctx := context.Background()
	s := newSFTPStore(t)

	path := fmt.Sprintf("/upload/del-%d.txt", time.Now().UnixNano())
	content := []byte("delete me")
	s.Write(ctx, path, bytes.NewReader(content), int64(len(content)))

	if err := s.Delete(ctx, path); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Stat(ctx, path); err == nil {
		t.Error("Stat should fail after Delete")
	}
}
