package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

// ---- helpers ----

func newLocal(t *testing.T, base string) repository.Storage {
	t.Helper()
	s, err := storage.NewLocalStorage(&config.LocalConfig{BasePath: base})
	if err != nil {
		t.Fatalf("NewLocalStorage: %v", err)
	}
	return s
}

func ctx() context.Context { return context.Background() }

// ---- NewLocalStorage ----

func TestNewLocalStorage_EmptyPath(t *testing.T) {
	_, err := storage.NewLocalStorage(&config.LocalConfig{BasePath: ""})
	if err == nil {
		t.Fatal("expected error for empty base path")
	}
}

func TestNewLocalStorage_ValidPath(t *testing.T) {
	_, err := storage.NewLocalStorage(&config.LocalConfig{BasePath: t.TempDir()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---- Connect ----

func TestLocalStorage_Connect_ExistingDir(t *testing.T) {
	s := newLocal(t, t.TempDir())
	if err := s.Connect(ctx()); err != nil {
		t.Fatalf("Connect on existing dir: %v", err)
	}
}

func TestLocalStorage_Connect_CreatesNonExistingDir(t *testing.T) {
	newPath := filepath.Join(t.TempDir(), "auto-created")
	s := newLocal(t, newPath)
	if err := s.Connect(ctx()); err != nil {
		t.Fatalf("Connect should create missing dir: %v", err)
	}
	info, err := os.Stat(newPath)
	if err != nil || !info.IsDir() {
		t.Fatal("expected directory to be created")
	}
}

func TestLocalStorage_Connect_PathIsFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(f, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, f)
	if err := s.Connect(ctx()); err == nil {
		t.Fatal("expected error when base path is a file")
	}
}

// ---- Disconnect ----

func TestLocalStorage_Disconnect(t *testing.T) {
	s := newLocal(t, t.TempDir())
	if err := s.Disconnect(ctx()); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
}

// ---- List ----

func TestLocalStorage_List_EmptyDir(t *testing.T) {
	s := newLocal(t, t.TempDir())
	files, err := s.List(ctx(), "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestLocalStorage_List_WithFiles(t *testing.T) {
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "b.txt"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	files, err := s.List(ctx(), "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestLocalStorage_List_IncludesSubdirs(t *testing.T) {
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	files, err := s.List(ctx(), "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var dirCount, fileCount int
	for _, f := range files {
		if f.IsDirectory {
			dirCount++
		} else {
			fileCount++
		}
	}
	if dirCount != 1 || fileCount != 1 {
		t.Errorf("expected 1 dir + 1 file, got %d dirs + %d files", dirCount, fileCount)
	}
}

func TestLocalStorage_List_NonexistentPath(t *testing.T) {
	s := newLocal(t, t.TempDir())
	_, err := s.List(ctx(), "does-not-exist")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

// ---- Read ----

func TestLocalStorage_Read_Success(t *testing.T) {
	base := t.TempDir()
	content := []byte("hello storage")
	if err := os.WriteFile(filepath.Join(base, "data.txt"), content, 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	rc, size, err := s.Read(ctx(), "data.txt")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	defer rc.Close()
	if size != int64(len(content)) {
		t.Errorf("size = %d, want %d", size, len(content))
	}
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestLocalStorage_Read_Directory(t *testing.T) {
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "dir"), 0755); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	_, _, err := s.Read(ctx(), "dir")
	if err == nil {
		t.Fatal("expected error when reading a directory")
	}
}

func TestLocalStorage_Read_NonexistentFile(t *testing.T) {
	s := newLocal(t, t.TempDir())
	_, _, err := s.Read(ctx(), "ghost.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// ---- Write ----

func TestLocalStorage_Write_CreatesFile(t *testing.T) {
	base := t.TempDir()
	s := newLocal(t, base)
	content := []byte("written content")
	if err := s.Write(ctx(), "out.txt", bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(base, "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("file content = %q, want %q", got, content)
	}
}

func TestLocalStorage_Write_CreatesParentDirs(t *testing.T) {
	base := t.TempDir()
	s := newLocal(t, base)
	content := []byte("nested")
	if err := s.Write(ctx(), "a/b/c/file.txt", bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("Write with nested path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "a/b/c/file.txt")); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestLocalStorage_Write_SizeMismatch(t *testing.T) {
	base := t.TempDir()
	s := newLocal(t, base)
	content := []byte("short")
	err := s.Write(ctx(), "mismatch.txt", bytes.NewReader(content), 100) // claim 100 bytes
	if err == nil {
		t.Fatal("expected size mismatch error")
	}
}

func TestLocalStorage_Write_ZeroSize_SkipsCheck(t *testing.T) {
	base := t.TempDir()
	s := newLocal(t, base)
	content := []byte("any size")
	if err := s.Write(ctx(), "nocheck.txt", bytes.NewReader(content), 0); err != nil {
		t.Fatalf("Write with size=0 should not check size: %v", err)
	}
}

// ---- Delete ----

func TestLocalStorage_Delete_Success(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "todelete.txt")
	if err := os.WriteFile(path, []byte("bye"), 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	if err := s.Delete(ctx(), "todelete.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected file to be deleted")
	}
}

func TestLocalStorage_Delete_Directory(t *testing.T) {
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "dir"), 0755); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	if err := s.Delete(ctx(), "dir"); err == nil {
		t.Fatal("expected error when deleting a directory")
	}
}

func TestLocalStorage_Delete_NonexistentFile(t *testing.T) {
	s := newLocal(t, t.TempDir())
	if err := s.Delete(ctx(), "ghost.txt"); err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// ---- Stat ----

func TestLocalStorage_Stat_File(t *testing.T) {
	base := t.TempDir()
	content := []byte("stat me")
	if err := os.WriteFile(filepath.Join(base, "info.txt"), content, 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	info, err := s.Stat(ctx(), "info.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", info.Size, len(content))
	}
	if info.IsDirectory {
		t.Error("IsDirectory = true, want false for file")
	}
	if info.Name != "info.txt" {
		t.Errorf("Name = %q, want info.txt", info.Name)
	}
}

func TestLocalStorage_Stat_Directory(t *testing.T) {
	base := t.TempDir()
	if err := os.Mkdir(filepath.Join(base, "mydir"), 0755); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	info, err := s.Stat(ctx(), "mydir")
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if !info.IsDirectory {
		t.Error("IsDirectory = false, want true for directory")
	}
}

func TestLocalStorage_Stat_NonexistentFile(t *testing.T) {
	s := newLocal(t, t.TempDir())
	_, err := s.Stat(ctx(), "ghost.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

// ---- CreateDirectory ----

func TestLocalStorage_CreateDirectory_Nested(t *testing.T) {
	base := t.TempDir()
	s := newLocal(t, base)
	if err := s.CreateDirectory(ctx(), "deep/nested/dir"); err != nil {
		t.Fatalf("CreateDirectory: %v", err)
	}
	info, err := os.Stat(filepath.Join(base, "deep/nested/dir"))
	if err != nil || !info.IsDir() {
		t.Fatal("expected nested directory to be created")
	}
}

func TestLocalStorage_CreateDirectory_Idempotent(t *testing.T) {
	base := t.TempDir()
	s := newLocal(t, base)
	if err := s.CreateDirectory(ctx(), "mydir"); err != nil {
		t.Fatalf("first CreateDirectory: %v", err)
	}
	if err := s.CreateDirectory(ctx(), "mydir"); err != nil {
		t.Fatalf("second CreateDirectory (idempotent): %v", err)
	}
}

// ---- ReadFrom (Resumer interface) ----

func TestLocalStorage_ReadFrom_ZeroOffset(t *testing.T) {
	base := t.TempDir()
	content := []byte("full content")
	if err := os.WriteFile(filepath.Join(base, "r.txt"), content, 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	resumer, ok := s.(interface {
		ReadFrom(context.Context, string, int64) (io.ReadCloser, int64, error)
	})
	if !ok {
		t.Skip("LocalStorage does not implement Resumer")
	}
	rc, remaining, err := resumer.ReadFrom(ctx(), "r.txt", 0)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	defer rc.Close()
	if remaining != int64(len(content)) {
		t.Errorf("remaining = %d, want %d", remaining, len(content))
	}
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestLocalStorage_ReadFrom_WithOffset(t *testing.T) {
	base := t.TempDir()
	content := []byte("0123456789")
	if err := os.WriteFile(filepath.Join(base, "seek.txt"), content, 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	resumer, ok := s.(interface {
		ReadFrom(context.Context, string, int64) (io.ReadCloser, int64, error)
	})
	if !ok {
		t.Skip("LocalStorage does not implement Resumer")
	}
	rc, remaining, err := resumer.ReadFrom(ctx(), "seek.txt", 4)
	if err != nil {
		t.Fatalf("ReadFrom offset=4: %v", err)
	}
	defer rc.Close()
	if remaining != 6 {
		t.Errorf("remaining = %d, want 6", remaining)
	}
	got, _ := io.ReadAll(rc)
	if string(got) != "456789" {
		t.Errorf("content after offset = %q, want \"456789\"", got)
	}
}

// ---- AppendWrite (Resumer interface) ----

func TestLocalStorage_AppendWrite_Success(t *testing.T) {
	base := t.TempDir()
	// Create a file with initial content
	path := filepath.Join(base, "resume.txt")
	if err := os.WriteFile(path, []byte("AAAA______"), 0644); err != nil {
		t.Fatal(err)
	}
	s := newLocal(t, base)
	resumer, ok := s.(interface {
		AppendWrite(context.Context, string, io.Reader, int64, int64) error
	})
	if !ok {
		t.Skip("LocalStorage does not implement Resumer")
	}
	// Overwrite the last 6 bytes (offset=4) with "BBBBBB"
	if err := resumer.AppendWrite(ctx(), "resume.txt", bytes.NewReader([]byte("BBBBBB")), 6, 4); err != nil {
		t.Fatalf("AppendWrite: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "AAAABBBBBB" {
		t.Errorf("after AppendWrite = %q, want \"AAAABBBBBB\"", string(got))
	}
}

func TestLocalStorage_AppendWrite_NonexistentFile(t *testing.T) {
	s := newLocal(t, t.TempDir())
	resumer, ok := s.(interface {
		AppendWrite(context.Context, string, io.Reader, int64, int64) error
	})
	if !ok {
		t.Skip("LocalStorage does not implement Resumer")
	}
	err := resumer.AppendWrite(ctx(), "ghost.txt", bytes.NewReader([]byte("data")), 4, 0)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
