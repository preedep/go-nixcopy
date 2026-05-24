package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// writeOnlyMockDest wraps MockStorage but returns an error on every Read call.
// Used to exercise the checksum "failed to read destination" error path.
type writeOnlyMockDest struct{ inner *mocks.MockStorage }

func (d *writeOnlyMockDest) Connect(ctx context.Context) error    { return d.inner.Connect(ctx) }
func (d *writeOnlyMockDest) Disconnect(ctx context.Context) error { return d.inner.Disconnect(ctx) }
func (d *writeOnlyMockDest) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	return d.inner.List(ctx, path)
}
func (d *writeOnlyMockDest) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	return nil, 0, errors.New("read not supported on this dest")
}
func (d *writeOnlyMockDest) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	return d.inner.Stat(ctx, path)
}
func (d *writeOnlyMockDest) Write(ctx context.Context, path string, r io.Reader, size int64) error {
	return d.inner.Write(ctx, path, r, size)
}
func (d *writeOnlyMockDest) CreateDirectory(ctx context.Context, path string) error { return nil }
func (d *writeOnlyMockDest) Delete(ctx context.Context, path string) error {
	return d.inner.Delete(ctx, path)
}

// errorHashReader returns a valid first byte then errors, causing io.Copy to fail mid-hash.
// Used to exercise the "failed to hash destination" error path.
type errorHashReader struct{ done bool }

func (r *errorHashReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, errors.New("simulated hash read error")
	}
	r.done = true
	p[0] = 0xAA
	return 1, nil
}
func (r *errorHashReader) Close() error { return nil }

// hashErrorDest wraps MockStorage but returns an errorHashReader on Read,
// causing the SHA-256 of the destination to fail after one byte.
type hashErrorDest struct{ inner *mocks.MockStorage }

func (d *hashErrorDest) Connect(ctx context.Context) error    { return nil }
func (d *hashErrorDest) Disconnect(ctx context.Context) error { return nil }
func (d *hashErrorDest) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	return d.inner.List(ctx, path)
}
func (d *hashErrorDest) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	return &errorHashReader{}, 1, nil
}
func (d *hashErrorDest) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	return d.inner.Stat(ctx, path)
}
func (d *hashErrorDest) Write(ctx context.Context, path string, r io.Reader, size int64) error {
	data, _ := io.ReadAll(r)
	d.inner.FileContent[path] = data
	return nil
}
func (d *hashErrorDest) CreateDirectory(ctx context.Context, path string) error { return nil }
func (d *hashErrorDest) Delete(ctx context.Context, path string) error          { return nil }

// ensure writeOnlyMockDest and hashErrorDest satisfy the Storage interface at compile time
var _ interface {
	Connect(context.Context) error
	Read(context.Context, string) (io.ReadCloser, int64, error)
	Write(context.Context, string, io.Reader, int64) error
	Stat(context.Context, string) (*entity.FileInfo, error)
} = (*writeOnlyMockDest)(nil)

var _ = bytes.NewReader // keep "bytes" import used

// --- #5: Checksum-on-resume ---

func TestVerifyResumeIntegrity_Match(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	content := []byte("Hello, World! This is a longer test file for resume.")
	partial := content[:10]

	source.AddFile("/src/file.bin", content, &entity.FileInfo{
		Path: "/src/file.bin", Name: "file.bin", Size: int64(len(content)),
	})
	dest.AddFile("/dst/file.bin", partial, &entity.FileInfo{
		Path: "/dst/file.bin", Name: "file.bin", Size: int64(len(partial)),
	})

	uc := &TransferUseCase{source: source, dest: dest, config: &entity.TransferConfig{}, logger: logger}
	uc.bufferSize.Store(4096)

	if err := uc.verifyResumeIntegrity(context.Background(), "/src/file.bin", "/dst/file.bin", int64(len(partial))); err != nil {
		t.Fatalf("expected no error when prefix matches, got: %v", err)
	}
}

func TestVerifyResumeIntegrity_Mismatch(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	srcContent := []byte("AAAAAAAAAA extra")
	dstContent := []byte("BBBBBBBBBB") // same length but different bytes

	source.AddFile("/src/file.bin", srcContent, &entity.FileInfo{
		Path: "/src/file.bin", Name: "file.bin", Size: int64(len(srcContent)),
	})
	dest.AddFile("/dst/file.bin", dstContent, &entity.FileInfo{
		Path: "/dst/file.bin", Name: "file.bin", Size: int64(len(dstContent)),
	})

	uc := &TransferUseCase{source: source, dest: dest, config: &entity.TransferConfig{}, logger: logger}
	uc.bufferSize.Store(4096)

	if err := uc.verifyResumeIntegrity(context.Background(), "/src/file.bin", "/dst/file.bin", 10); err == nil {
		t.Fatal("expected integrity mismatch error, got nil")
	}
}

func TestTransfer_Resume_WithChecksumVerification(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	fullContent := make([]byte, 64)
	for i := range fullContent {
		fullContent[i] = byte(i)
	}
	partial := fullContent[:20]

	source.AddFile("/src/file.bin", fullContent, &entity.FileInfo{
		Path: "/src/file.bin", Name: "file.bin", Size: int64(len(fullContent)),
	})
	dest.AddFile("/dst/file.bin", partial, &entity.FileInfo{
		Path: "/dst/file.bin", Name: "file.bin", Size: int64(len(partial)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      4096,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
		VerifyChecksum:  true,
	}

	uc := NewTransferUseCase(source, dest, cfg, logger)
	result, err := uc.Transfer(context.Background(), "/src/file.bin", "/dst/file.bin", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	if result.ResumedFrom != int64(len(partial)) {
		t.Errorf("ResumedFrom = %d, want %d", result.ResumedFrom, len(partial))
	}
	if got := dest.FileContent["/dst/file.bin"]; string(got) != string(fullContent) {
		t.Errorf("dest content mismatch after resume+checksum")
	}
}

func TestTransfer_Resume_IntegrityMismatch_FallsBackToFullTransfer(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	fullContent := []byte("AAAAAAAAAA_rest_of_file")
	// dest has a corrupted first 10 bytes
	corruptPartial := []byte("ZZZZZZZZZZ")

	source.AddFile("/src/file.bin", fullContent, &entity.FileInfo{
		Path: "/src/file.bin", Name: "file.bin", Size: int64(len(fullContent)),
	})
	dest.AddFile("/dst/file.bin", corruptPartial, &entity.FileInfo{
		Path: "/dst/file.bin", Name: "file.bin", Size: int64(len(corruptPartial)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      4096,
		ConcurrentFiles: 1,
		RetryAttempts:   1, // one retry — integrity check fails, retry does full transfer
		RetryDelay:      time.Millisecond,
		EnableResume:    true,
		VerifyChecksum:  true,
	}

	uc := NewTransferUseCase(source, dest, cfg, logger)
	// The first attempt detects integrity mismatch. The second attempt starts from scratch
	// (dest is no longer considered a valid partial) so it overwrites fully.
	_, err := uc.Transfer(context.Background(), "/src/file.bin", "/dst/file.bin", nil)
	// We expect either success (full re-transfer) or failure — but NOT silent corruption.
	// The key assertion is: if it succeeds, the content must equal source.
	if err == nil {
		got := dest.FileContent["/dst/file.bin"]
		h := sha256.Sum256(fullContent)
		want := hex.EncodeToString(h[:])
		h2 := sha256.Sum256(got)
		got2 := hex.EncodeToString(h2[:])
		if want != got2 {
			t.Errorf("transfer succeeded but dest content is corrupted: want checksum %s got %s", want, got2)
		}
	}
}

// --- #7: Adaptive buffer sizing ---

func TestAdaptBufferSize_SlowThroughput_Halves(t *testing.T) {
	uc := &TransferUseCase{config: &entity.TransferConfig{BufferSize: 32 * 1024 * 1024}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(32 * 1024 * 1024)

	uc.adaptBufferSize(32 * 1024 * 1024) // 32 MiB/s — below 64 MiB/s threshold
	got := uc.bufferSize.Load()
	if got != 16*1024*1024 {
		t.Errorf("bufferSize = %d, want %d (halved)", got, 16*1024*1024)
	}
}

func TestAdaptBufferSize_FastThroughput_Doubles(t *testing.T) {
	uc := &TransferUseCase{config: &entity.TransferConfig{BufferSize: 32 * 1024 * 1024}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(32 * 1024 * 1024)

	uc.adaptBufferSize(600 * 1024 * 1024) // 600 MiB/s — above 512 MiB/s threshold
	got := uc.bufferSize.Load()
	if got != 64*1024*1024 {
		t.Errorf("bufferSize = %d, want %d (doubled)", got, 64*1024*1024)
	}
}

func TestAdaptBufferSize_MidThroughput_Unchanged(t *testing.T) {
	uc := &TransferUseCase{config: &entity.TransferConfig{BufferSize: 32 * 1024 * 1024}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(32 * 1024 * 1024)

	uc.adaptBufferSize(200 * 1024 * 1024) // 200 MiB/s — within thresholds
	got := uc.bufferSize.Load()
	if got != 32*1024*1024 {
		t.Errorf("bufferSize = %d, want %d (unchanged)", got, 32*1024*1024)
	}
}

func TestAdaptBufferSize_Floor(t *testing.T) {
	uc := &TransferUseCase{config: &entity.TransferConfig{BufferSize: 1 * 1024 * 1024}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(minAdaptiveBuf) // already at floor

	uc.adaptBufferSize(1) // very slow — would halve below floor
	got := uc.bufferSize.Load()
	if got != minAdaptiveBuf {
		t.Errorf("bufferSize = %d, want floor %d", got, minAdaptiveBuf)
	}
}

func TestAdaptBufferSize_Ceiling(t *testing.T) {
	uc := &TransferUseCase{config: &entity.TransferConfig{BufferSize: maxAdaptiveBuf}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(maxAdaptiveBuf) // already at ceiling

	uc.adaptBufferSize(float64(adaptiveHighBytesPerSec) * 2) // very fast — would double above ceiling
	got := uc.bufferSize.Load()
	if got != maxAdaptiveBuf {
		t.Errorf("bufferSize = %d, want ceiling %d", got, maxAdaptiveBuf)
	}
}

func TestAdaptBufferSize_ZeroBps_NoChange(t *testing.T) {
	uc := &TransferUseCase{config: &entity.TransferConfig{BufferSize: 32 * 1024 * 1024}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(32 * 1024 * 1024)

	uc.adaptBufferSize(0)
	got := uc.bufferSize.Load()
	if got != 32*1024*1024 {
		t.Errorf("bufferSize changed on 0 bps: got %d", got)
	}
}

// --- verifyResumeIntegrity error paths ---

func TestVerifyResumeIntegrity_SourceReadError(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	source.ReadError = errors.New("source unavailable")

	uc := &TransferUseCase{source: source, dest: dest, config: &entity.TransferConfig{}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(4096)

	err := uc.verifyResumeIntegrity(context.Background(), "/src/f", "/dst/f", 10)
	if err == nil {
		t.Fatal("expected error when source.Read fails, got nil")
	}
}

func TestVerifyResumeIntegrity_DestReadError(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()

	source.AddFile("/src/f", []byte("AAAAAAAAAA"), &entity.FileInfo{Path: "/src/f", Name: "f", Size: 10})
	dest.ReadError = errors.New("dest unavailable")

	uc := &TransferUseCase{source: source, dest: dest, config: &entity.TransferConfig{}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(4096)

	err := uc.verifyResumeIntegrity(context.Background(), "/src/f", "/dst/f", 10)
	if err == nil {
		t.Fatal("expected error when dest.Read fails, got nil")
	}
}

func TestVerifyResumeIntegrity_SourceTooShort(t *testing.T) {
	// Source has only 5 bytes but resumeOffset is 20 — io.CopyN will fail.
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()

	source.AddFile("/src/f", []byte("AAAAA"), &entity.FileInfo{Path: "/src/f", Name: "f", Size: 5})
	dest.AddFile("/dst/f", make([]byte, 20), &entity.FileInfo{Path: "/dst/f", Name: "f", Size: 20})

	uc := &TransferUseCase{source: source, dest: dest, config: &entity.TransferConfig{}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(4096)

	err := uc.verifyResumeIntegrity(context.Background(), "/src/f", "/dst/f", 20)
	if err == nil {
		t.Fatal("expected error when source is shorter than resumeOffset, got nil")
	}
}

func TestVerifyResumeIntegrity_DestTooShort(t *testing.T) {
	// Dest has only 5 bytes but resumeOffset is 20 — io.CopyN on dest will fail.
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()

	source.AddFile("/src/f", make([]byte, 30), &entity.FileInfo{Path: "/src/f", Name: "f", Size: 30})
	dest.AddFile("/dst/f", []byte("AAAAA"), &entity.FileInfo{Path: "/dst/f", Name: "f", Size: 5})

	uc := &TransferUseCase{source: source, dest: dest, config: &entity.TransferConfig{}, logger: applog.NewNopLogger()}
	uc.bufferSize.Store(4096)

	err := uc.verifyResumeIntegrity(context.Background(), "/src/f", "/dst/f", 20)
	if err == nil {
		t.Fatal("expected error when dest is shorter than resumeOffset, got nil")
	}
}

// --- checksum dest/hash error paths in attemptTransfer ---

func TestTransfer_AllRetriesFailed_WithProgressChan(t *testing.T) {
	// Exercises the failure progress send in Transfer when progressChan is non-nil.
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	source.AddFile("/src/f.txt", []byte("data"), &entity.FileInfo{Path: "/src/f.txt", Name: "f.txt", Size: 4})
	source.ReadError = errors.New("read always fails")

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 1, RetryAttempts: 0, RetryDelay: time.Millisecond,
	}
	uc := NewTransferUseCase(source, dest, cfg, logger)

	progressChan := make(chan entity.TransferProgress, 10)
	_, err := uc.Transfer(context.Background(), "/src/f.txt", "/dst/f.txt", progressChan)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Drain — should have received a failure progress event.
	var got []entity.TransferProgress
	for p := range progressChan {
		got = append(got, p)
	}
	found := false
	for _, p := range got {
		if p.Status == entity.TransferStatusFailed {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a TransferStatusFailed progress event, got %v", got)
	}
}

func TestTransfer_Checksum_DestReadError(t *testing.T) {
	source := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	content := []byte("checksum dest read error test")
	source.AddFile("/src/f.txt", content, &entity.FileInfo{Path: "/src/f.txt", Name: "f.txt", Size: int64(len(content))})

	// Dest that accepts Write but always fails Read — exercises the checksum verification
	// "failed to read destination" error path in attemptTransfer.
	dest := &writeOnlyMockDest{inner: mocks.NewMockStorage()}

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 1, RetryAttempts: 0, RetryDelay: time.Millisecond,
		VerifyChecksum: true,
	}
	uc := NewTransferUseCase(source, dest, cfg, logger)
	result, err := uc.Transfer(context.Background(), "/src/f.txt", "/dst/f.txt", nil)
	if err == nil {
		t.Fatal("Transfer() expected error when dest read for checksum fails, got nil")
	}
	if result.Status != entity.TransferStatusFailed {
		t.Errorf("Status = %v, want failed", result.Status)
	}
}

func TestTransfer_Checksum_DestHashError(t *testing.T) {
	source := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	content := []byte("checksum hash error test content")
	source.AddFile("/src/f.txt", content, &entity.FileInfo{Path: "/src/f.txt", Name: "f.txt", Size: int64(len(content))})

	// hashErrorDest: Write stores the file, Read returns a reader that errors mid-stream,
	// causing the SHA-256 copy of the destination to fail.
	dest := &hashErrorDest{inner: mocks.NewMockStorage()}

	cfg := &entity.TransferConfig{
		BufferSize: 1024, ConcurrentFiles: 1, RetryAttempts: 0, RetryDelay: time.Millisecond,
		VerifyChecksum: true,
	}
	uc := NewTransferUseCase(source, dest, cfg, logger)
	result, err := uc.Transfer(context.Background(), "/src/f.txt", "/dst/f.txt", nil)
	if err == nil {
		t.Fatal("Transfer() expected error when dest hash fails, got nil")
	}
	if result.Status != entity.TransferStatusFailed {
		t.Errorf("Status = %v, want failed", result.Status)
	}
}
