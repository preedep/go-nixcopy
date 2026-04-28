//go:build integration

package storage_test

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func newLocalStore(t *testing.T) repository.Storage {
	t.Helper()
	s, err := storage.NewLocalStorage(&config.LocalConfig{BasePath: t.TempDir()})
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}
	if err := s.Connect(context.Background()); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { s.Disconnect(context.Background()) })
	return s
}

func TestLocalStorage_WriteAndRead(t *testing.T) {
	ctx := context.Background()
	s := newLocalStore(t)

	content := []byte("hello integration test")
	if err := s.Write(ctx, "sub/file.txt", bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("Write: %v", err)
	}

	rc, size, err := s.Read(ctx, "sub/file.txt")
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

func TestLocalStorage_List(t *testing.T) {
	ctx := context.Background()
	s := newLocalStore(t)

	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		data := []byte(name)
		if err := s.Write(ctx, name, bytes.NewReader(data), int64(len(data))); err != nil {
			t.Fatalf("Write %s: %v", name, err)
		}
	}

	files, err := s.List(ctx, "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("List returned %d files, want 3", len(files))
	}
}

func TestLocalStorage_Stat(t *testing.T) {
	ctx := context.Background()
	s := newLocalStore(t)

	content := []byte("stat me")
	s.Write(ctx, "stat.txt", bytes.NewReader(content), int64(len(content)))

	info, err := s.Stat(ctx, "stat.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Stat size = %d, want %d", info.Size, len(content))
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	ctx := context.Background()
	s := newLocalStore(t)

	content := []byte("delete me")
	s.Write(ctx, "del.txt", bytes.NewReader(content), int64(len(content)))

	if err := s.Delete(ctx, "del.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Stat(ctx, "del.txt"); err == nil {
		t.Error("Stat should fail after Delete")
	}
}

func TestLocalStorage_CreateDirectory(t *testing.T) {
	ctx := context.Background()
	s := newLocalStore(t)

	nested := filepath.Join("deep", "nested", "dir")
	if err := s.CreateDirectory(ctx, nested); err != nil {
		t.Fatalf("CreateDirectory: %v", err)
	}

	info, err := s.Stat(ctx, nested)
	if err != nil {
		t.Fatalf("Stat after CreateDirectory: %v", err)
	}
	if !info.IsDirectory {
		t.Error("expected IsDirectory = true")
	}
}

func TestLocalStorage_ReadFrom(t *testing.T) {
	ctx := context.Background()
	s := newLocalStore(t)

	resumer, ok := s.(repository.Resumer)
	if !ok {
		t.Skip("LocalStorage does not implement Resumer")
	}

	content := []byte("Hello, World!")
	if err := s.Write(ctx, "readfrom.txt", bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("Write: %v", err)
	}

	offset := int64(7)
	rc, size, err := resumer.ReadFrom(ctx, "readfrom.txt", offset)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	defer rc.Close()

	want := content[offset:]
	if size != int64(len(want)) {
		t.Errorf("size = %d, want %d", size, len(want))
	}
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, want) {
		t.Errorf("ReadFrom content = %q, want %q", got, want)
	}
}

func TestLocalStorage_AppendWrite(t *testing.T) {
	ctx := context.Background()
	s := newLocalStore(t)

	resumer, ok := s.(repository.Resumer)
	if !ok {
		t.Skip("LocalStorage does not implement Resumer")
	}

	fullContent := []byte("Hello, World!")
	partial := fullContent[:7]
	tail := fullContent[7:]

	if err := s.Write(ctx, "append.txt", bytes.NewReader(partial), int64(len(partial))); err != nil {
		t.Fatalf("Write partial: %v", err)
	}

	if err := resumer.AppendWrite(ctx, "append.txt", bytes.NewReader(tail), int64(len(tail)), int64(len(partial))); err != nil {
		t.Fatalf("AppendWrite: %v", err)
	}

	rc, size, err := s.Read(ctx, "append.txt")
	if err != nil {
		t.Fatalf("Read after AppendWrite: %v", err)
	}
	defer rc.Close()

	if size != int64(len(fullContent)) {
		t.Errorf("size = %d, want %d", size, len(fullContent))
	}
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, fullContent) {
		t.Errorf("AppendWrite result = %q, want %q", got, fullContent)
	}
}

func TestLocalStorage_Resume_ReadFromThenAppend(t *testing.T) {
	ctx := context.Background()
	src := newLocalStore(t)
	dst := newLocalStore(t)

	srcResumer, ok := src.(repository.Resumer)
	if !ok {
		t.Skip("LocalStorage does not implement Resumer")
	}
	dstResumer, _ := dst.(repository.Resumer)

	fullContent := []byte("The quick brown fox jumps over the lazy dog.")
	partial := fullContent[:10]

	if err := src.Write(ctx, "source.txt", bytes.NewReader(fullContent), int64(len(fullContent))); err != nil {
		t.Fatalf("Write source: %v", err)
	}
	if err := dst.Write(ctx, "dest.txt", bytes.NewReader(partial), int64(len(partial))); err != nil {
		t.Fatalf("Write partial dest: %v", err)
	}

	offset := int64(len(partial))
	rc, remaining, err := srcResumer.ReadFrom(ctx, "source.txt", offset)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	defer rc.Close()

	if err := dstResumer.AppendWrite(ctx, "dest.txt", rc, remaining, offset); err != nil {
		t.Fatalf("AppendWrite: %v", err)
	}

	rcFull, size, err := dst.Read(ctx, "dest.txt")
	if err != nil {
		t.Fatalf("Read final: %v", err)
	}
	defer rcFull.Close()

	if size != int64(len(fullContent)) {
		t.Errorf("final size = %d, want %d", size, len(fullContent))
	}
	got, _ := io.ReadAll(rcFull)
	if !bytes.Equal(got, fullContent) {
		t.Errorf("final content = %q, want %q", got, fullContent)
	}
}
