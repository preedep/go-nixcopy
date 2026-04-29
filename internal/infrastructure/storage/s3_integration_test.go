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

// newS3Store connects to MinIO (or real S3) using environment variables.
//
// Required env vars (set automatically in CI via the minio service container):
//
//	S3_ENDPOINT        e.g. http://localhost:9000
//	S3_ACCESS_KEY      e.g. minioadmin
//	S3_SECRET_KEY      e.g. minioadmin
//	S3_BUCKET          e.g. test-bucket
//	S3_REGION          e.g. us-east-1  (MinIO accepts any value)
func newS3Store(t *testing.T) repository.Storage {
	t.Helper()

	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("S3_ENDPOINT not set — skipping S3 integration tests")
	}

	cfg := &config.S3Config{
		Region:          getEnvOr("S3_REGION", "us-east-1"),
		Bucket:          getEnvOr("S3_BUCKET", "test-bucket"),
		AuthType:        config.S3AuthAccessKey,
		AccessKeyID:     os.Getenv("S3_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("S3_SECRET_KEY"),
		Endpoint:        endpoint,
		UsePathStyle:    true, // required for MinIO
	}

	s := storage.NewS3Storage(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("S3 Connect: %v", err)
	}
	t.Cleanup(func() { s.Disconnect(context.Background()) })
	return s
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestS3Storage_WriteAndRead(t *testing.T) {
	ctx := context.Background()
	s := newS3Store(t)

	key := fmt.Sprintf("integration-test/%d/file.txt", time.Now().UnixNano())
	content := []byte("s3 integration test content")

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

func TestS3Storage_Stat(t *testing.T) {
	ctx := context.Background()
	s := newS3Store(t)

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

func TestS3Storage_List(t *testing.T) {
	ctx := context.Background()
	s := newS3Store(t)

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

func TestS3Storage_Delete(t *testing.T) {
	ctx := context.Background()
	s := newS3Store(t)

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
