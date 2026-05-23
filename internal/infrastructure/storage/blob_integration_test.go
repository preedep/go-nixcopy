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

// newBlobStore connects to Azurite (or real Azure) using environment variables.
//
// Required env vars (set automatically by make integration-test-local / test-integration.sh):
//
//	BLOB_CONNECTION_STRING  e.g. DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;...
//	BLOB_CONTAINER          e.g. test-container
func newBlobStore(t *testing.T) repository.Storage {
	t.Helper()

	connStr := os.Getenv("BLOB_CONNECTION_STRING")
	if connStr == "" {
		t.Skip("BLOB_CONNECTION_STRING not set — skipping Azure Blob integration tests")
	}

	cfg := &config.BlobConfig{
		ContainerName:    getEnvOr("BLOB_CONTAINER", "test-container"),
		AuthType:         config.BlobAuthConnectionString,
		ConnectionString: connStr,
	}

	s := storage.NewBlobStorage(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("Blob Connect: %v", err)
	}
	t.Cleanup(func() { s.Disconnect(context.Background()) })
	return s
}

func TestBlobStorage_WriteAndRead(t *testing.T) {
	ctx := context.Background()
	s := newBlobStore(t)

	key := fmt.Sprintf("integration-test/%d/file.txt", time.Now().UnixNano())
	content := []byte("azure blob integration test content")

	if err := s.Write(ctx, key, bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("Write: %v", err)
	}
	t.Cleanup(func() { s.Delete(context.Background(), key) })

	rc, size, err := s.Read(ctx, key)
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

func TestBlobStorage_Stat(t *testing.T) {
	ctx := context.Background()
	s := newBlobStore(t)

	key := fmt.Sprintf("integration-test/%d/stat.txt", time.Now().UnixNano())
	content := []byte("stat target")
	s.Write(ctx, key, bytes.NewReader(content), int64(len(content)))
	t.Cleanup(func() { s.Delete(context.Background(), key) })

	info, err := s.Stat(ctx, key)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Stat size = %d, want %d", info.Size, len(content))
	}
}

func TestBlobStorage_List(t *testing.T) {
	ctx := context.Background()
	s := newBlobStore(t)

	prefix := fmt.Sprintf("integration-test/%d/", time.Now().UnixNano())
	keys := []string{prefix + "one.txt", prefix + "two.txt", prefix + "three.txt"}

	for _, key := range keys {
		data := []byte(key)
		if err := s.Write(ctx, key, bytes.NewReader(data), int64(len(data))); err != nil {
			t.Fatalf("Write %s: %v", key, err)
		}
		t.Cleanup(func() { s.Delete(context.Background(), key) })
	}

	files, err := s.List(ctx, prefix)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("List returned %d files, want 3", len(files))
	}
}

func TestBlobStorage_Delete(t *testing.T) {
	ctx := context.Background()
	s := newBlobStore(t)

	key := fmt.Sprintf("integration-test/%d/delete.txt", time.Now().UnixNano())
	content := []byte("delete me")
	s.Write(ctx, key, bytes.NewReader(content), int64(len(content)))

	if err := s.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Stat(ctx, key); err == nil {
		t.Error("Stat should fail after Delete")
	}
}
