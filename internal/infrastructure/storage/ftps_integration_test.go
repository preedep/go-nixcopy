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

// newFTPSStore connects using environment variables.
//
// Required env vars (set automatically by make integration-test-local):
//
//	FTPS_HOST         e.g. localhost
//	FTPS_PORT         e.g. 21
//	FTPS_USERNAME     e.g. testuser
//	FTPS_PASSWORD     e.g. testpass
//	FTPS_TLS_MODE     explicit | implicit  (default: explicit)
//	FTPS_SKIP_VERIFY  true | false         (default: false; set true for self-signed certs)
func newFTPSStore(t *testing.T) repository.Storage {
	t.Helper()

	host := os.Getenv("FTPS_HOST")
	if host == "" {
		t.Skip("FTPS_HOST not set — skipping FTPS integration tests")
	}

	tlsMode := os.Getenv("FTPS_TLS_MODE")
	if tlsMode == "" {
		tlsMode = "explicit"
	}

	cfg := &config.FTPSConfig{
		Host:       host,
		Port:       getEnvIntOr("FTPS_PORT", 21),
		Username:   os.Getenv("FTPS_USERNAME"),
		Password:   os.Getenv("FTPS_PASSWORD"),
		Timeout:    30 * time.Second,
		TLSMode:    tlsMode,
		SkipVerify: os.Getenv("FTPS_SKIP_VERIFY") == "true",
	}

	s := storage.NewFTPSStorage(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("FTPS Connect: %v", err)
	}
	t.Cleanup(func() { s.Disconnect(context.Background()) })
	return s
}

func TestFTPSStorage_WriteAndRead(t *testing.T) {
	ctx := context.Background()
	s := newFTPSStore(t)

	path := fmt.Sprintf("/integration-%d.txt", time.Now().UnixNano())
	content := []byte("ftps integration test content")

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

func TestFTPSStorage_Stat(t *testing.T) {
	ctx := context.Background()
	s := newFTPSStore(t)

	path := fmt.Sprintf("/stat-%d.txt", time.Now().UnixNano())
	content := []byte("stat target")
	s.Write(ctx, path, bytes.NewReader(content), int64(len(content)))
	t.Cleanup(func() { s.Delete(context.Background(), path) })

	info, err := s.Stat(ctx, path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Stat size = %d, want %d", info.Size, len(content))
	}
}

func TestFTPSStorage_List(t *testing.T) {
	ctx := context.Background()
	s := newFTPSStore(t)

	dir := fmt.Sprintf("/list-%d", time.Now().UnixNano())
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

	var regular int
	for _, f := range files {
		if !f.IsDirectory {
			regular++
		}
	}
	if regular != 3 {
		t.Errorf("List returned %d regular files, want 3", regular)
	}
}

func TestFTPSStorage_Delete(t *testing.T) {
	ctx := context.Background()
	s := newFTPSStore(t)

	path := fmt.Sprintf("/del-%d.txt", time.Now().UnixNano())
	content := []byte("delete me")
	s.Write(ctx, path, bytes.NewReader(content), int64(len(content)))

	if err := s.Delete(ctx, path); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Stat(ctx, path); err == nil {
		t.Error("Stat should fail after Delete")
	}
}

func TestFTPSStorage_CreateDirectory(t *testing.T) {
	ctx := context.Background()
	s := newFTPSStore(t)

	dir := fmt.Sprintf("/mkdir-%d/nested/deep", time.Now().UnixNano())
	if err := s.CreateDirectory(ctx, dir); err != nil {
		t.Fatalf("CreateDirectory: %v", err)
	}

	// Verify by writing a file inside the created directory.
	path := dir + "/probe.txt"
	data := []byte("probe")
	if err := s.Write(ctx, path, bytes.NewReader(data), int64(len(data))); err != nil {
		t.Fatalf("Write inside created dir: %v", err)
	}
	t.Cleanup(func() { s.Delete(context.Background(), path) })
}
