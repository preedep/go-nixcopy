package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// --------------------------------------------------------------------------
// A — Batch context cancellation stops new transfers
// --------------------------------------------------------------------------

// blockingDest wraps MockStorage and makes every Write block until release is
// closed.  ctx.Done() is deliberately NOT checked inside Write so that the
// goroutine keeps holding the semaphore even after the context is cancelled —
// this gives the for-loop's ctx.Done() case time to win deterministically.
type blockingDest struct {
	*mocks.MockStorage
	once    sync.Once
	started chan struct{} // closed when the first Write begins
	release chan struct{} // close to unblock all pending writes
}

func (b *blockingDest) Write(ctx context.Context, path string, r io.Reader, size int64) error {
	b.once.Do(func() { close(b.started) })
	<-b.release // block until explicitly released
	return b.MockStorage.Write(ctx, path, r, size)
}

// TestTransferBatch_ContextCancelled_StopsNewTransfers verifies that when
// the context is cancelled while the first goroutine is blocked in Write
// (holding the semaphore), TransferBatch returns context.Canceled rather
// than launching the remaining transfers.
func TestTransferBatch_ContextCancelled_StopsNewTransfers(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := &blockingDest{
		MockStorage: mocks.NewMockStorage(),
		started:     make(chan struct{}),
		release:     make(chan struct{}),
	}

	paths := []string{"/src/a.txt", "/src/b.txt", "/src/c.txt"}
	for _, p := range paths {
		c := []byte("hello")
		source.AddFile(p, c, &entity.FileInfo{Path: p, Name: p[5:], Size: int64(len(c))})
	}

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1, // semaphore of 1 — goroutine 0 fills it while blocked in Write
		RetryAttempts:   0,
		RetryDelay:      time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		uc := NewTransferUseCase(source, dest, cfg, applog.NewNopLogger())
		_, err := uc.TransferBatch(ctx, paths, "/dst/", nil)
		errCh <- err
	}()

	// Goroutine 0 is now in Write, holding the semaphore.
	// The for-loop's next iteration is therefore blocked on "semaphore <- struct{}{}".
	<-dest.started

	// Cancelling ctx makes the for-loop's ctx.Done() case the only ready one
	// (semaphore is still full), so TransferBatch returns context.Canceled.
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("TransferBatch did not return after context cancellation (possible deadlock)")
	}

	// Unblock goroutine 0 so it can exit Write and release the semaphore cleanly.
	close(dest.release)
}

// --------------------------------------------------------------------------
// B — Bandwidth throttle + checksum verification together
// --------------------------------------------------------------------------

// TestTransfer_Throttle_And_Checksum verifies that BandwidthLimit and
// VerifyChecksum are compatible: the transfer completes, the digest is
// correct, and the destination holds the original bytes.
func TestTransfer_Throttle_And_Checksum(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()

	content := []byte("throttle + checksum combined test content")
	source.AddFile("/src/file.txt", content, &entity.FileInfo{
		Path: "/src/file.txt",
		Name: "file.txt",
		Size: int64(len(content)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		BandwidthLimit:  4 * 1024 * 1024, // 4 MB/s — high enough not to slow the test
		VerifyChecksum:  true,
	}

	uc := NewTransferUseCase(source, dest, cfg, applog.NewNopLogger())
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}

	h := sha256.Sum256(content)
	want := hex.EncodeToString(h[:])
	if result.Checksum != want {
		t.Errorf("Checksum = %q, want %q", result.Checksum, want)
	}

	if got := dest.FileContent["/dst/file.txt"]; !bytes.Equal(got, content) {
		t.Errorf("dest content mismatch after throttled+checksum transfer")
	}
}

// --------------------------------------------------------------------------
// C — Resume when dest size >= source size (no resume, full re-transfer)
// --------------------------------------------------------------------------

// TestTransfer_Resume_DestSizeEqualToSource confirms that when the destination
// file is the same size as the source, EnableResume does NOT trigger a resume
// (the condition is destSize > 0 && destSize < srcSize).  The file is
// overwritten with the source content and ResumedFrom is 0.
func TestTransfer_Resume_DestSizeEqualToSource(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()

	srcContent := []byte("source v2 content") // 17 bytes
	dstContent := []byte("old stale content") // 17 bytes — same length, different data

	if len(srcContent) != len(dstContent) {
		t.Fatal("test setup error: src and dst content must have equal length")
	}

	source.AddFile("/src/file.txt", srcContent, &entity.FileInfo{
		Path: "/src/file.txt", Name: "file.txt", Size: int64(len(srcContent)),
	})
	dest.AddFile("/dst/file.txt", dstContent, &entity.FileInfo{
		Path: "/dst/file.txt", Name: "file.txt", Size: int64(len(dstContent)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		EnableResume:    true,
		SkipExisting:    false, // ensure skip is not the reason it re-transfers
	}

	uc := NewTransferUseCase(source, dest, cfg, applog.NewNopLogger())
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	if result.ResumedFrom != 0 {
		t.Errorf("ResumedFrom = %d, want 0 (dest size == source size, no resume)", result.ResumedFrom)
	}
	if got := dest.FileContent["/dst/file.txt"]; !bytes.Equal(got, srcContent) {
		t.Errorf("dest content = %q, want %q (should be overwritten)", got, srcContent)
	}
}

// TestTransfer_Resume_DestSizeExceedsSource confirms that when the destination
// file is LARGER than the source, EnableResume also skips resume and performs
// a full overwrite.
func TestTransfer_Resume_DestSizeExceedsSource(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()

	srcContent := []byte("short") // 5 bytes
	dstContent := []byte("much longer stale content at destination") // > 5 bytes

	source.AddFile("/src/file.txt", srcContent, &entity.FileInfo{
		Path: "/src/file.txt", Name: "file.txt", Size: int64(len(srcContent)),
	})
	dest.AddFile("/dst/file.txt", dstContent, &entity.FileInfo{
		Path: "/dst/file.txt", Name: "file.txt", Size: int64(len(dstContent)),
	})

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 1,
		RetryAttempts:   0,
		EnableResume:    true,
		SkipExisting:    false,
	}

	uc := NewTransferUseCase(source, dest, cfg, applog.NewNopLogger())
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)

	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want completed", result.Status)
	}
	if result.ResumedFrom != 0 {
		t.Errorf("ResumedFrom = %d, want 0 (dest > source size, no resume)", result.ResumedFrom)
	}
	if got := dest.FileContent["/dst/file.txt"]; !bytes.Equal(got, srcContent) {
		t.Errorf("dest content = %q, want %q (should be overwritten with shorter source)", got, srcContent)
	}
}

// --------------------------------------------------------------------------
// D — Batch + SkipExisting: three-way outcome (Skipped, Completed, Failed)
// --------------------------------------------------------------------------

// TestTransferBatch_SkipExisting_ThreeWayOutcome runs a 3-file batch with
// SkipExisting enabled and asserts all three possible outcomes appear in the
// result slice:
//   - a.txt: present at dest with matching size → Skipped
//   - b.txt: present at source only → Completed
//   - c.txt: missing from source → Failed
func TestTransferBatch_SkipExisting_ThreeWayOutcome(t *testing.T) {
	source := mocks.NewMockStorage()
	dest := mocks.NewMockStorage()

	aContent := []byte("file a content")
	bContent := []byte("file b content")

	// a.txt: both source and dest, same size → will be skipped
	source.AddFile("/src/a.txt", aContent, &entity.FileInfo{Path: "/src/a.txt", Name: "a.txt", Size: int64(len(aContent))})
	dest.AddFile("/dst/a.txt", aContent, &entity.FileInfo{Path: "/dst/a.txt", Name: "a.txt", Size: int64(len(aContent))})

	// b.txt: source only, dest absent → will be transferred
	source.AddFile("/src/b.txt", bContent, &entity.FileInfo{Path: "/src/b.txt", Name: "b.txt", Size: int64(len(bContent))})

	// c.txt: not in source → Stat fails → Failed (c.txt intentionally absent)

	cfg := &entity.TransferConfig{
		BufferSize:      1024,
		ConcurrentFiles: 3,
		RetryAttempts:   0,
		SkipExisting:    true,
	}

	uc := NewTransferUseCase(source, dest, cfg, applog.NewNopLogger())
	results, err := uc.TransferBatch(
		context.Background(),
		[]string{"/src/a.txt", "/src/b.txt", "/src/c.txt"},
		"/dst/",
		nil,
	)

	if err != nil {
		t.Fatalf("TransferBatch error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("results count = %d, want 3", len(results))
	}

	var skipped, completed, failed int
	for _, r := range results {
		switch r.Status {
		case entity.TransferStatusSkipped:
			skipped++
		case entity.TransferStatusCompleted:
			completed++
		case entity.TransferStatusFailed:
			failed++
		}
	}

	if skipped != 1 {
		t.Errorf("skipped = %d, want 1 (a.txt present at dest with same size)", skipped)
	}
	if completed != 1 {
		t.Errorf("completed = %d, want 1 (b.txt transferred)", completed)
	}
	if failed != 1 {
		t.Errorf("failed = %d, want 1 (c.txt absent from source)", failed)
	}

	// a.txt must be unchanged at dest
	if got := dest.FileContent["/dst/a.txt"]; !bytes.Equal(got, aContent) {
		t.Errorf("a.txt content changed (should have been skipped): got %q, want %q", got, aContent)
	}
	// b.txt must have been written
	if got := dest.FileContent["/dst/b.txt"]; !bytes.Equal(got, bContent) {
		t.Errorf("b.txt content = %q, want %q", got, bContent)
	}
	// Write should have been called (for b.txt only)
	if !dest.WriteCalled {
		t.Error("Write was never called — b.txt should have been transferred")
	}
}
