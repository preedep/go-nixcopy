package cli

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"github.com/preedep/go-nixcopy/internal/infrastructure/config"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// mockFactory satisfies storageFactory using pre-built MockStorage instances.
type mockFactory struct {
	src *mocks.MockStorage
	dst *mocks.MockStorage
}

func (f *mockFactory) NewSource(_ *config.SourceConfig) (repository.Storage, error) {
	return f.src, nil
}

func (f *mockFactory) NewDest(_ *config.DestinationConfig) (repository.Storage, error) {
	return f.dst, nil
}

// withMockFactory installs a mockFactory for the test and restores the real factory on cleanup.
func withMockFactory(t *testing.T, src, dst *mocks.MockStorage) {
	t.Helper()
	orig := activeStorageFactory
	activeStorageFactory = &mockFactory{src: src, dst: dst}
	t.Cleanup(func() { activeStorageFactory = orig })
}

// silenceStdout discards stdout during the test (JSON summary goes there).
func silenceStdout(t *testing.T) {
	t.Helper()
	orig := os.Stdout
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open /dev/null: %v", err)
	}
	os.Stdout = null
	t.Cleanup(func() {
		_ = null.Close()
		os.Stdout = orig
	})
}

// newFileSource returns a MockStorage pre-populated with one file.
func newFileSource(path string, content []byte) *mocks.MockStorage {
	src := mocks.NewMockStorage()
	src.Files[path] = &entity.FileInfo{
		Path:         path,
		Name:         path[strings.LastIndex(path, "/")+1:],
		Size:         int64(len(content)),
		ModifiedTime: time.Now(),
	}
	src.FileContent[path] = content
	return src
}

func TestRunTransfer_SingleFile_Success(t *testing.T) {
	resetTransferFlags()
	silenceStdout(t)

	content := []byte("hello nixcopy")
	src := newFileSource("/src/hello.txt", content)
	dst := mocks.NewMockStorage()

	withMockFactory(t, src, dst)

	sourcePath = "/src/hello.txt"
	destPath = "/dst/hello.txt"
	sourceType = "local"
	destType = "local"

	err := runTransfer(transferCmd, nil)
	if err != nil {
		t.Fatalf("runTransfer() error = %v", err)
	}
	if !dst.WriteCalled {
		t.Error("expected dest.Write to be called")
	}
	if got := dst.FileContent["/dst/hello.txt"]; string(got) != string(content) {
		t.Errorf("dest content = %q, want %q", got, content)
	}
}

func TestRunTransfer_SkipExisting(t *testing.T) {
	resetTransferFlags()
	silenceStdout(t)

	content := []byte("already there")
	src := newFileSource("/src/file.txt", content)

	dst := mocks.NewMockStorage()
	// Pre-populate destination with same size so skip-existing fires.
	dst.Files["/dst/file.txt"] = &entity.FileInfo{Name: "file.txt", Size: int64(len(content))}

	withMockFactory(t, src, dst)

	sourcePath = "/src/file.txt"
	destPath = "/dst/file.txt"
	sourceType = "local"
	destType = "local"
	skipExisting = true

	err := runTransfer(transferCmd, nil)
	if err != nil {
		t.Fatalf("runTransfer() error = %v", err)
	}
	if dst.WriteCalled {
		t.Error("dest.Write must not be called when file is skipped")
	}
}

func TestRunTransfer_NoSourceSpecified(t *testing.T) {
	resetTransferFlags()

	src := mocks.NewMockStorage()
	dst := mocks.NewMockStorage()
	withMockFactory(t, src, dst)

	// sourcePath intentionally left empty
	destPath = "/dst/out.txt"
	sourceType = "local"
	destType = "local"

	err := runTransfer(transferCmd, nil)
	if err == nil {
		t.Fatal("expected error when no source is specified")
	}
	if !strings.Contains(err.Error(), "no source") {
		t.Errorf("error should mention 'no source', got: %v", err)
	}
}

func TestRunTransfer_SourceConnectError(t *testing.T) {
	resetTransferFlags()

	src := mocks.NewMockStorage()
	src.ConnectError = context.DeadlineExceeded
	dst := mocks.NewMockStorage()

	withMockFactory(t, src, dst)

	sourcePath = "/src/file.txt"
	destPath = "/dst/file.txt"
	sourceType = "local"
	destType = "local"

	err := runTransfer(transferCmd, nil)
	if err == nil {
		t.Fatal("expected error when source Connect fails")
	}
	if !strings.Contains(err.Error(), "source") {
		t.Errorf("error should mention 'source', got: %v", err)
	}
}

func TestRunTransfer_SourceStatError_ReturnsError(t *testing.T) {
	resetTransferFlags()
	silenceStdout(t)

	src := mocks.NewMockStorage()
	src.StatError = io.ErrUnexpectedEOF // file not found — no retry loop entered

	dst := mocks.NewMockStorage()

	withMockFactory(t, src, dst)

	sourcePath = "/src/missing.txt"
	destPath = "/dst/out.txt"
	sourceType = "local"
	destType = "local"

	err := runTransfer(transferCmd, nil)
	if err == nil {
		t.Fatal("expected error when source Stat fails")
	}
	if dst.WriteCalled {
		t.Error("dest.Write must not be called when source Stat fails")
	}
}

func TestRunTransfer_BatchTransfer_Success(t *testing.T) {
	resetTransferFlags()
	silenceStdout(t)

	files := map[string][]byte{
		"/src/a.txt": []byte("aaa"),
		"/src/b.txt": []byte("bbb"),
	}

	src := mocks.NewMockStorage()
	for path, data := range files {
		name := path[strings.LastIndex(path, "/")+1:]
		src.Files[path] = &entity.FileInfo{
			Path: path, Name: name, Size: int64(len(data)),
		}
		src.FileContent[path] = data
	}

	dst := mocks.NewMockStorage()
	withMockFactory(t, src, dst)

	// Supply both paths explicitly so no glob expansion is needed.
	sourcePaths = []string{"/src/a.txt", "/src/b.txt"}
	destPath = "/dst/"
	sourceType = "local"
	destType = "local"
	concurrentFiles = 2

	err := runTransfer(transferCmd, nil)
	if err != nil {
		t.Fatalf("runTransfer() error = %v", err)
	}
	if !dst.WriteCalled {
		t.Error("expected dest.Write to be called for batch")
	}
}
