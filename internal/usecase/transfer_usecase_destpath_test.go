package usecase

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// resolveDestPathForTest mirrors the CLI resolveDestPath logic so the usecase
// integration tests can verify the end-to-end behaviour without importing the
// cli package (which would create a circular dependency).
func resolveDestPathForTest(srcPath, destPath string) string {
	if strings.HasSuffix(destPath, "/") {
		return destPath + filepath.Base(srcPath)
	}
	return destPath
}

// TestTransfer_TrailingSlashDest_FileLandsAtResolvedPath verifies that when the
// CLI resolves a trailing-slash dest ("/home/azureuser/data_file/") to a full path
// ("/home/azureuser/data_file/archive.zip"), the Transfer call writes to that
// exact location in destination storage.
func TestTransfer_TrailingSlashDest_FileLandsAtResolvedPath(t *testing.T) {
	content := []byte("archive content")
	srcPath := "/cmdb/archive.zip"
	destDir := "/home/azureuser/data_file/"

	src := mocks.NewMockStorage()
	src.AddFile(srcPath, content, &entity.FileInfo{
		Path:         srcPath,
		Name:         "archive.zip",
		Size:         int64(len(content)),
		ModifiedTime: time.Now(),
		IsDirectory:  false,
	})
	dst := mocks.NewMockStorage()

	cfg := &entity.TransferConfig{BufferSize: 1024}
	uc := NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())

	resolvedDest := resolveDestPathForTest(srcPath, destDir)
	if resolvedDest != "/home/azureuser/data_file/archive.zip" {
		t.Fatalf("resolveDestPathForTest returned %q, want /home/azureuser/data_file/archive.zip", resolvedDest)
	}

	result, err := uc.Transfer(context.Background(), srcPath, resolvedDest, nil)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("status = %v, want Completed", result.Status)
	}

	got, ok := dst.FileContent[resolvedDest]
	if !ok {
		t.Fatalf("file not found at resolved dest %q in destination storage", resolvedDest)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
	if _, exists := dst.FileContent[destDir]; exists {
		t.Errorf("file must not exist at bare directory path %q", destDir)
	}
}

// TestTransfer_ExplicitDestFilename_UnchangedPath verifies that when an explicit
// destination filename is provided (no trailing slash) the path is used as-is.
func TestTransfer_ExplicitDestFilename_UnchangedPath(t *testing.T) {
	content := []byte("hello explicit")
	srcPath := "/cmdb/archive.zip"
	destPath := "/backup/renamed_archive.zip"

	src := mocks.NewMockStorage()
	src.AddFile(srcPath, content, &entity.FileInfo{
		Path:         srcPath,
		Name:         "archive.zip",
		Size:         int64(len(content)),
		ModifiedTime: time.Now(),
		IsDirectory:  false,
	})
	dst := mocks.NewMockStorage()

	cfg := &entity.TransferConfig{BufferSize: 1024}
	uc := NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())

	resolvedDest := resolveDestPathForTest(srcPath, destPath)
	if resolvedDest != destPath {
		t.Fatalf("resolveDestPathForTest(%q, %q) = %q, want unchanged %q",
			srcPath, destPath, resolvedDest, destPath)
	}

	result, err := uc.Transfer(context.Background(), srcPath, resolvedDest, nil)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("status = %v, want Completed", result.Status)
	}

	if _, ok := dst.FileContent[destPath]; !ok {
		t.Errorf("file not found at explicit dest %q", destPath)
	}
}

// TestTransfer_TrailingSlash_MultipleFiles_EachFilenameAppended verifies that when
// multiple files are transferred into a trailing-slash dest directory, each file
// lands at its own resolved path (simulating what the CLI does via TransferBatch).
func TestTransfer_TrailingSlash_MultipleFiles_EachFilenameAppended(t *testing.T) {
	destDir := "/remote/output/"
	files := []struct {
		srcPath string
		content []byte
	}{
		{"/src/a.txt", []byte("aaa")},
		{"/src/b.txt", []byte("bbb")},
		{"/src/c.csv", []byte("ccc")},
	}

	src := mocks.NewMockStorage()
	for _, f := range files {
		src.AddFile(f.srcPath, f.content, &entity.FileInfo{
			Path:         f.srcPath,
			Name:         filepath.Base(f.srcPath),
			Size:         int64(len(f.content)),
			ModifiedTime: time.Now(),
			IsDirectory:  false,
		})
	}
	dst := mocks.NewMockStorage()

	cfg := &entity.TransferConfig{BufferSize: 1024, ConcurrentFiles: 3}
	uc := NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())

	srcPaths := make([]string, len(files))
	for i, f := range files {
		srcPaths[i] = f.srcPath
	}

	results, err := uc.TransferBatch(context.Background(), srcPaths, destDir, nil)
	if err != nil {
		t.Fatalf("TransferBatch failed: %v", err)
	}

	for i, result := range results {
		if result.Status != entity.TransferStatusCompleted {
			t.Errorf("file %d status = %v, want Completed", i, result.Status)
		}
	}

	for _, f := range files {
		expected := destDir + filepath.Base(f.srcPath)
		got, ok := dst.FileContent[expected]
		if !ok {
			t.Errorf("file not found at %q", expected)
			continue
		}
		if string(got) != string(f.content) {
			t.Errorf("%q content = %q, want %q", expected, got, f.content)
		}
	}
}
