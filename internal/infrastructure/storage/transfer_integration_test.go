//go:build integration

package storage_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"
	"github.com/preedep/go-nixcopy/internal/usecase"
)

// sha256Re matches a lower-case hex SHA-256 digest (64 chars).
var sha256Re = regexp.MustCompile(`^[0-9a-f]{64}$`)

// baseTransferConfig returns a minimal TransferConfig suitable for integration tests.
func baseTransferConfig() *entity.TransferConfig {
	return &entity.TransferConfig{
		BufferSize:    1 * 1024 * 1024, // 1 MB
		RetryAttempts: 1,
		RetryDelay:    100 * time.Millisecond,
	}
}

// doTransfer runs a single Transfer and fails the test on error or non-completed status.
func doTransfer(t *testing.T, src repository.StorageReader, dst repository.Storage, srcPath, dstPath string, cfg *entity.TransferConfig) *entity.TransferResult {
	t.Helper()
	uc := usecase.NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())
	result, err := uc.Transfer(context.Background(), srcPath, dstPath, nil)
	if err != nil {
		t.Fatalf("Transfer(%q → %q): %v", srcPath, dstPath, err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Fatalf("Transfer status = %v, want %v (error: %v)", result.Status, entity.TransferStatusCompleted, result.Error)
	}
	return result
}

// putFile writes content to storage and registers a cleanup to delete it.
func putFile(t *testing.T, s repository.Storage, path string, content []byte) {
	t.Helper()
	if err := s.Write(context.Background(), path, bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("putFile %q: %v", path, err)
	}
	t.Cleanup(func() { s.Delete(context.Background(), path) })
}

// getFile reads the full content of a file from storage.
func getFile(t *testing.T, s repository.Storage, path string) []byte {
	t.Helper()
	rc, _, err := s.Read(context.Background(), path)
	if err != nil {
		t.Fatalf("getFile %q: %v", path, err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("getFile ReadAll %q: %v", path, err)
	}
	return data
}

// sha256Hex returns the lower-case hex SHA-256 of b.
func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// skipUnless skips the test when any of the listed env vars is empty.
func skipUnless(t *testing.T, vars ...string) {
	t.Helper()
	for _, v := range vars {
		if os.Getenv(v) == "" {
			t.Skipf("%s not set — skipping", v)
		}
	}
}

// --------------------------------------------------------------------------
// #1 — Cross-backend transfers via TransferUseCase
// --------------------------------------------------------------------------

func TestTransfer_LocalToSFTP(t *testing.T) {
	skipUnless(t, "SFTP_HOST")
	local := newLocalStore(t)
	sftp := newSFTPStore(t)

	content := []byte("hello from local to sftp")
	srcPath := "src.txt"
	dstPath := fmt.Sprintf("/upload/local-to-sftp-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { sftp.Delete(context.Background(), dstPath) })

	putFile(t, local, srcPath, content)
	doTransfer(t, local, sftp, srcPath, dstPath, baseTransferConfig())

	if got := getFile(t, sftp, dstPath); !bytes.Equal(got, content) {
		t.Errorf("content mismatch after Local→SFTP: got %q, want %q", got, content)
	}
}

func TestTransfer_LocalToS3(t *testing.T) {
	skipUnless(t, "S3_ENDPOINT")
	local := newLocalStore(t)
	s3 := newS3Store(t)

	content := []byte("hello from local to s3")
	srcPath := "src.txt"
	dstPath := fmt.Sprintf("integration-test/%d/local-to-s3.txt", time.Now().UnixNano())
	t.Cleanup(func() { s3.Delete(context.Background(), dstPath) })

	putFile(t, local, srcPath, content)
	doTransfer(t, local, s3, srcPath, dstPath, baseTransferConfig())

	if got := getFile(t, s3, dstPath); !bytes.Equal(got, content) {
		t.Errorf("content mismatch after Local→S3: got %q, want %q", got, content)
	}
}

func TestTransfer_SFTPToS3(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	content := []byte("sftp to s3 cross-backend transfer")
	srcPath := fmt.Sprintf("/upload/sftp-to-s3-src-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("integration-test/%d/sftp-to-s3.txt", time.Now().UnixNano())
	t.Cleanup(func() { s3.Delete(context.Background(), dstPath) })

	putFile(t, sftp, srcPath, content)
	doTransfer(t, sftp, s3, srcPath, dstPath, baseTransferConfig())

	if got := getFile(t, s3, dstPath); !bytes.Equal(got, content) {
		t.Errorf("content mismatch after SFTP→S3: got %q, want %q", got, content)
	}
}

func TestTransfer_S3ToSFTP(t *testing.T) {
	skipUnless(t, "S3_ENDPOINT", "SFTP_HOST")
	s3 := newS3Store(t)
	sftp := newSFTPStore(t)

	content := []byte("s3 to sftp cross-backend transfer")
	srcPath := fmt.Sprintf("integration-test/%d/s3-to-sftp-src.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("/upload/s3-to-sftp-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { sftp.Delete(context.Background(), dstPath) })

	putFile(t, s3, srcPath, content)
	doTransfer(t, s3, sftp, srcPath, dstPath, baseTransferConfig())

	if got := getFile(t, sftp, dstPath); !bytes.Equal(got, content) {
		t.Errorf("content mismatch after S3→SFTP: got %q, want %q", got, content)
	}
}

func TestTransfer_SFTPToFTPS(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "FTPS_HOST")
	sftp := newSFTPStore(t)
	ftps := newFTPSStore(t)

	content := []byte("sftp to ftps cross-backend transfer")
	srcPath := fmt.Sprintf("/upload/sftp-to-ftps-src-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("/sftp-to-ftps-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { ftps.Delete(context.Background(), dstPath) })

	putFile(t, sftp, srcPath, content)
	doTransfer(t, sftp, ftps, srcPath, dstPath, baseTransferConfig())

	if got := getFile(t, ftps, dstPath); !bytes.Equal(got, content) {
		t.Errorf("content mismatch after SFTP→FTPS: got %q, want %q", got, content)
	}
}

func TestTransfer_S3ToFTPS(t *testing.T) {
	skipUnless(t, "S3_ENDPOINT", "FTPS_HOST")
	s3 := newS3Store(t)
	ftps := newFTPSStore(t)

	content := []byte("s3 to ftps cross-backend transfer")
	srcPath := fmt.Sprintf("integration-test/%d/s3-to-ftps-src.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("/s3-to-ftps-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { ftps.Delete(context.Background(), dstPath) })

	putFile(t, s3, srcPath, content)
	doTransfer(t, s3, ftps, srcPath, dstPath, baseTransferConfig())

	if got := getFile(t, ftps, dstPath); !bytes.Equal(got, content) {
		t.Errorf("content mismatch after S3→FTPS: got %q, want %q", got, content)
	}
}

// --------------------------------------------------------------------------
// #2 — Checksum verification against real backends
// --------------------------------------------------------------------------

func TestTransfer_Checksum_LocalToSFTP(t *testing.T) {
	skipUnless(t, "SFTP_HOST")
	local := newLocalStore(t)
	sftp := newSFTPStore(t)

	content := []byte("checksum verification: local → sftp")
	srcPath := "chk-src.txt"
	dstPath := fmt.Sprintf("/upload/chk-local-sftp-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { sftp.Delete(context.Background(), dstPath) })

	putFile(t, local, srcPath, content)

	cfg := baseTransferConfig()
	cfg.VerifyChecksum = true

	result := doTransfer(t, local, sftp, srcPath, dstPath, cfg)

	if !sha256Re.MatchString(result.Checksum) {
		t.Errorf("Checksum = %q; want 64-char lowercase hex SHA-256", result.Checksum)
	}
	if want := sha256Hex(content); result.Checksum != want {
		t.Errorf("Checksum = %q, want %q", result.Checksum, want)
	}
}

func TestTransfer_Checksum_SFTPToS3(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	// 256 KB of deterministic non-zero data.
	content := make([]byte, 256*1024)
	for i := range content {
		content[i] = byte(i % 251)
	}

	srcPath := fmt.Sprintf("/upload/chk-sftp-s3-src-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("integration-test/%d/chk-sftp-to-s3.bin", time.Now().UnixNano())
	t.Cleanup(func() { s3.Delete(context.Background(), dstPath) })

	putFile(t, sftp, srcPath, content)

	cfg := baseTransferConfig()
	cfg.VerifyChecksum = true

	result := doTransfer(t, sftp, s3, srcPath, dstPath, cfg)

	if !sha256Re.MatchString(result.Checksum) {
		t.Errorf("Checksum = %q; want 64-char lowercase hex SHA-256", result.Checksum)
	}
	if want := sha256Hex(content); result.Checksum != want {
		t.Errorf("Checksum mismatch for SFTP→S3 transfer: got %q, want %q", result.Checksum, want)
	}
}

func TestTransfer_Checksum_S3ToLocal(t *testing.T) {
	skipUnless(t, "S3_ENDPOINT")
	s3 := newS3Store(t)
	local := newLocalStore(t)

	content := []byte("checksum verification: s3 → local")
	srcPath := fmt.Sprintf("integration-test/%d/chk-s3-local-src.txt", time.Now().UnixNano())
	dstPath := "chk-dst.txt"

	putFile(t, s3, srcPath, content)

	cfg := baseTransferConfig()
	cfg.VerifyChecksum = true

	result := doTransfer(t, s3, local, srcPath, dstPath, cfg)

	if !sha256Re.MatchString(result.Checksum) {
		t.Errorf("Checksum = %q; want 64-char lowercase hex SHA-256", result.Checksum)
	}
	if want := sha256Hex(content); result.Checksum != want {
		t.Errorf("Checksum mismatch for S3→Local transfer: got %q, want %q", result.Checksum, want)
	}
}

// --------------------------------------------------------------------------
// #3 — SkipExisting across backends
// --------------------------------------------------------------------------

// TestTransfer_SkipExisting_SameSize asserts that a file already present on the
// destination with the same size is not re-transferred.
func TestTransfer_SkipExisting_SameSize_S3ToSFTP(t *testing.T) {
	skipUnless(t, "S3_ENDPOINT", "SFTP_HOST")
	s3 := newS3Store(t)
	sftp := newSFTPStore(t)

	content := []byte("skip-existing payload: s3 → sftp")
	srcPath := fmt.Sprintf("integration-test/%d/skip-src.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("/upload/skip-dst-%d.txt", time.Now().UnixNano())

	putFile(t, s3, srcPath, content)
	// Pre-populate dest with the identical content so sizes match.
	putFile(t, sftp, dstPath, content)

	cfg := baseTransferConfig()
	cfg.SkipExisting = true

	result := doTransferSkip(t, s3, sftp, srcPath, dstPath, cfg)

	if result.Status != entity.TransferStatusSkipped {
		t.Errorf("Status = %v, want %v", result.Status, entity.TransferStatusSkipped)
	}
	// Dest must still hold the original content.
	if got := getFile(t, sftp, dstPath); !bytes.Equal(got, content) {
		t.Errorf("dest content changed after skip: got %q, want %q", got, content)
	}
}

func TestTransfer_SkipExisting_SameSize_SFTPToS3(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	content := []byte("skip-existing payload: sftp → s3")
	srcPath := fmt.Sprintf("/upload/skip-src-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("integration-test/%d/skip-dst.txt", time.Now().UnixNano())

	putFile(t, sftp, srcPath, content)
	putFile(t, s3, dstPath, content)

	cfg := baseTransferConfig()
	cfg.SkipExisting = true

	result := doTransferSkip(t, sftp, s3, srcPath, dstPath, cfg)

	if result.Status != entity.TransferStatusSkipped {
		t.Errorf("Status = %v, want %v", result.Status, entity.TransferStatusSkipped)
	}
	if got := getFile(t, s3, dstPath); !bytes.Equal(got, content) {
		t.Errorf("dest content changed after skip: got %q, want %q", got, content)
	}
}

// TestTransfer_SkipExisting_DifferentSize asserts that when the destination file
// has a different size the transfer proceeds normally (not skipped).
func TestTransfer_SkipExisting_DifferentSize_SFTPToS3(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	srcContent := []byte("full source content for size mismatch test")
	dstStale := []byte("shorter") // different size → should NOT be skipped

	srcPath := fmt.Sprintf("/upload/skip-size-src-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("integration-test/%d/skip-size-dst.txt", time.Now().UnixNano())

	putFile(t, sftp, srcPath, srcContent)
	putFile(t, s3, dstPath, dstStale)

	cfg := baseTransferConfig()
	cfg.SkipExisting = true

	result := doTransferSkip(t, sftp, s3, srcPath, dstPath, cfg)

	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("Status = %v, want %v (size mismatch should trigger transfer)", result.Status, entity.TransferStatusCompleted)
	}
	if got := getFile(t, s3, dstPath); !bytes.Equal(got, srcContent) {
		t.Errorf("dest content = %q, want %q after size-mismatch transfer", got, srcContent)
	}
}

// doTransferSkip is like doTransfer but expects either Skipped or Completed —
// it does not fatal on Skipped status.
func doTransferSkip(t *testing.T, src repository.StorageReader, dst repository.Storage, srcPath, dstPath string, cfg *entity.TransferConfig) *entity.TransferResult {
	t.Helper()
	uc := usecase.NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())
	result, err := uc.Transfer(context.Background(), srcPath, dstPath, nil)
	if err != nil {
		t.Fatalf("Transfer(%q → %q): %v", srcPath, dstPath, err)
	}
	if result.Status != entity.TransferStatusSkipped && result.Status != entity.TransferStatusCompleted {
		t.Fatalf("Transfer status = %v, want Skipped or Completed (error: %v)", result.Status, result.Error)
	}
	return result
}

// doBatch runs TransferBatch and fatals if the call itself errors or any
// individual result has a non-Completed status.
func doBatch(t *testing.T, src repository.StorageReader, dst repository.Storage, srcPaths []string, destBase string, cfg *entity.TransferConfig) []*entity.TransferResult {
	t.Helper()
	uc := usecase.NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())
	results, err := uc.TransferBatch(context.Background(), srcPaths, destBase, nil)
	if err != nil {
		t.Fatalf("TransferBatch: %v", err)
	}
	for i, r := range results {
		if r.Status != entity.TransferStatusCompleted {
			t.Errorf("result[%d] status = %v (error: %v)", i, r.Status, r.Error)
		}
	}
	return results
}

// --------------------------------------------------------------------------
// #4 — Resume across backends (src and dst must both implement Resumer)
// --------------------------------------------------------------------------

// TestTransfer_Resume_LocalToSFTP writes a partial file to the SFTP destination,
// then runs Transfer with EnableResume — the use case must pick up from the partial
// offset rather than re-transferring from the beginning.
func TestTransfer_Resume_LocalToSFTP(t *testing.T) {
	skipUnless(t, "SFTP_HOST")
	local := newLocalStore(t)
	sftp := newSFTPStore(t)

	fullContent := []byte("The quick brown fox jumps over the lazy dog. Resume test payload.")
	partial := fullContent[:20]

	srcPath := "resume-src.txt"
	dstPath := fmt.Sprintf("/upload/resume-local-sftp-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { sftp.Delete(context.Background(), dstPath) })

	putFile(t, local, srcPath, fullContent)
	// Seed the destination with a partial file to trigger resume.
	if err := sftp.Write(context.Background(), dstPath, bytes.NewReader(partial), int64(len(partial))); err != nil {
		t.Fatalf("seed partial dest: %v", err)
	}

	cfg := baseTransferConfig()
	cfg.EnableResume = true

	result := doTransfer(t, local, sftp, srcPath, dstPath, cfg)

	if result.ResumedFrom != int64(len(partial)) {
		t.Errorf("ResumedFrom = %d, want %d", result.ResumedFrom, len(partial))
	}
	if got := getFile(t, sftp, dstPath); !bytes.Equal(got, fullContent) {
		t.Errorf("resumed content = %q, want %q", got, fullContent)
	}
}

func TestTransfer_Resume_SFTPToLocal(t *testing.T) {
	skipUnless(t, "SFTP_HOST")
	sftp := newSFTPStore(t)
	local := newLocalStore(t)

	fullContent := []byte("Resume test: sftp source to local destination. Extra payload bytes.")
	partial := fullContent[:15]

	srcPath := fmt.Sprintf("/upload/resume-sftp-local-src-%d.txt", time.Now().UnixNano())
	dstPath := "resume-dst.txt"

	putFile(t, sftp, srcPath, fullContent)
	// Seed the local destination with a partial file.
	if err := local.Write(context.Background(), dstPath, bytes.NewReader(partial), int64(len(partial))); err != nil {
		t.Fatalf("seed partial dest: %v", err)
	}

	cfg := baseTransferConfig()
	cfg.EnableResume = true

	result := doTransfer(t, sftp, local, srcPath, dstPath, cfg)

	if result.ResumedFrom != int64(len(partial)) {
		t.Errorf("ResumedFrom = %d, want %d", result.ResumedFrom, len(partial))
	}
	if got := getFile(t, local, dstPath); !bytes.Equal(got, fullContent) {
		t.Errorf("resumed content = %q, want %q", got, fullContent)
	}
}

func TestTransfer_Resume_SFTPToSFTP(t *testing.T) {
	skipUnless(t, "SFTP_HOST")
	src := newSFTPStore(t)
	dst := newSFTPStore(t)

	fullContent := []byte("Resume test: sftp to sftp. Both sides implement Resumer.")
	partial := fullContent[:25]

	srcPath := fmt.Sprintf("/upload/resume-sftp-sftp-src-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("/upload/resume-sftp-sftp-dst-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { dst.Delete(context.Background(), dstPath) })

	putFile(t, src, srcPath, fullContent)
	// Seed dest with a partial file.
	if err := dst.Write(context.Background(), dstPath, bytes.NewReader(partial), int64(len(partial))); err != nil {
		t.Fatalf("seed partial dest: %v", err)
	}

	cfg := baseTransferConfig()
	cfg.EnableResume = true

	result := doTransfer(t, src, dst, srcPath, dstPath, cfg)

	if result.ResumedFrom != int64(len(partial)) {
		t.Errorf("ResumedFrom = %d, want %d", result.ResumedFrom, len(partial))
	}
	if got := getFile(t, dst, dstPath); !bytes.Equal(got, fullContent) {
		t.Errorf("resumed content = %q, want %q", got, fullContent)
	}
}

// TestTransfer_Resume_FallbackWhenNotSupported confirms that when a backend does
// not implement Resumer (S3), Transfer still succeeds — it falls back silently
// to a full re-transfer and leaves ResumedFrom == 0.
func TestTransfer_Resume_FallbackWhenNotSupported(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	fullContent := []byte("Resume fallback test: S3 dest does not implement Resumer.")
	srcPath := fmt.Sprintf("/upload/resume-fallback-src-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("integration-test/%d/resume-fallback-dst.txt", time.Now().UnixNano())
	t.Cleanup(func() { s3.Delete(context.Background(), dstPath) })

	putFile(t, sftp, srcPath, fullContent)

	cfg := baseTransferConfig()
	cfg.EnableResume = true

	result := doTransfer(t, sftp, s3, srcPath, dstPath, cfg)

	if result.ResumedFrom != 0 {
		t.Errorf("ResumedFrom = %d, want 0 (S3 does not support resume)", result.ResumedFrom)
	}
	if got := getFile(t, s3, dstPath); !bytes.Equal(got, fullContent) {
		t.Errorf("fallback transfer content = %q, want %q", got, fullContent)
	}
}

// --------------------------------------------------------------------------
// #5 — Batch transfer (TransferBatch) across backends
// --------------------------------------------------------------------------

// TestTransfer_Batch_SFTPToS3 writes 5 files to SFTP and batch-transfers them
// to an S3 prefix, verifying every file arrives with the correct content.
func TestTransfer_Batch_SFTPToS3(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	ts := time.Now().UnixNano()
	srcDir := fmt.Sprintf("/upload/batch-sftp-s3-%d", ts)
	dstBase := fmt.Sprintf("integration-test/%d/batch-sftp-s3/", ts)

	const n = 5
	srcPaths := make([]string, n)
	fileContents := make([][]byte, n)

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("file%d.txt", i+1)
		srcPaths[i] = fmt.Sprintf("%s/%s", srcDir, name)
		fileContents[i] = fmt.Appendf(nil, "batch sftp→s3 file %d content", i+1)
		putFile(t, sftp, srcPaths[i], fileContents[i])
	}

	cfg := baseTransferConfig()
	cfg.ConcurrentFiles = 3

	results := doBatch(t, sftp, s3, srcPaths, dstBase, cfg)

	if len(results) != n {
		t.Fatalf("got %d results, want %d", len(results), n)
	}
	for i, r := range results {
		dstPath := dstBase + fmt.Sprintf("file%d.txt", i+1)
		t.Cleanup(func() { s3.Delete(context.Background(), dstPath) })
		got := getFile(t, s3, dstPath)
		if !bytes.Equal(got, fileContents[i]) {
			t.Errorf("file%d content = %q, want %q", i+1, got, fileContents[i])
		}
		_ = r
	}
}

// TestTransfer_Batch_S3ToSFTP transfers 5 S3 objects to an SFTP directory.
func TestTransfer_Batch_S3ToSFTP(t *testing.T) {
	skipUnless(t, "S3_ENDPOINT", "SFTP_HOST")
	s3 := newS3Store(t)
	sftp := newSFTPStore(t)

	ts := time.Now().UnixNano()
	srcPrefix := fmt.Sprintf("integration-test/%d/batch-s3-sftp-src/", ts)
	dstBase := fmt.Sprintf("/upload/batch-s3-sftp-%d", ts)

	const n = 5
	srcPaths := make([]string, n)
	fileContents := make([][]byte, n)

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("obj%d.txt", i+1)
		srcPaths[i] = srcPrefix + name
		fileContents[i] = fmt.Appendf(nil, "batch s3→sftp object %d content", i+1)
		putFile(t, s3, srcPaths[i], fileContents[i])
	}

	cfg := baseTransferConfig()
	cfg.ConcurrentFiles = 3

	results := doBatch(t, s3, sftp, srcPaths, dstBase, cfg)

	if len(results) != n {
		t.Fatalf("got %d results, want %d", len(results), n)
	}
	for i, r := range results {
		dstPath := fmt.Sprintf("%s/obj%d.txt", dstBase, i+1)
		t.Cleanup(func() { sftp.Delete(context.Background(), dstPath) })
		got := getFile(t, sftp, dstPath)
		if !bytes.Equal(got, fileContents[i]) {
			t.Errorf("obj%d content = %q, want %q", i+1, got, fileContents[i])
		}
		_ = r
	}
}

// TestTransfer_Batch_PartialFailure confirms that a missing source file yields
// TransferStatusFailed for that slot while all other files complete successfully.
func TestTransfer_Batch_PartialFailure_SFTPToS3(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	ts := time.Now().UnixNano()
	dstBase := fmt.Sprintf("integration-test/%d/batch-partial/", ts)

	good1Content := []byte("good file 1")
	good2Content := []byte("good file 2")

	good1 := fmt.Sprintf("/upload/batch-partial-good1-%d.txt", ts)
	missing := fmt.Sprintf("/upload/batch-partial-missing-%d.txt", ts) // intentionally not written
	good2 := fmt.Sprintf("/upload/batch-partial-good2-%d.txt", ts)

	putFile(t, sftp, good1, good1Content)
	putFile(t, sftp, good2, good2Content)
	t.Cleanup(func() {
		s3.Delete(context.Background(), dstBase+"batch-partial-good1-"+fmt.Sprint(ts)+".txt")
		s3.Delete(context.Background(), dstBase+"batch-partial-good2-"+fmt.Sprint(ts)+".txt")
	})

	cfg := baseTransferConfig()
	cfg.ConcurrentFiles = 3

	uc := usecase.NewTransferUseCase(sftp, s3, cfg, applog.NewNopLogger())
	results, err := uc.TransferBatch(context.Background(), []string{good1, missing, good2}, dstBase, nil)
	if err != nil {
		t.Fatalf("TransferBatch: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	if results[0].Status != entity.TransferStatusCompleted {
		t.Errorf("results[0] (good1) status = %v, want Completed", results[0].Status)
	}
	if results[1].Status != entity.TransferStatusFailed {
		t.Errorf("results[1] (missing) status = %v, want Failed", results[1].Status)
	}
	if results[2].Status != entity.TransferStatusCompleted {
		t.Errorf("results[2] (good2) status = %v, want Completed", results[2].Status)
	}
}

// --------------------------------------------------------------------------
// #6 — Bandwidth throttle end-to-end
// --------------------------------------------------------------------------

// TestTransfer_Throttle_LocalToSFTP sends a 256 KB file capped at 64 KB/s and
// asserts the transfer took at least 3 seconds, confirming throttle is active.
// The upper bound is generous (30 s) to avoid flakiness on slow CI runners.
func TestTransfer_Throttle_LocalToSFTP(t *testing.T) {
	skipUnless(t, "SFTP_HOST")
	local := newLocalStore(t)
	sftp := newSFTPStore(t)

	const fileSize = 256 * 1024          // 256 KB
	const limitBytesPerSec = 64 * 1024   // 64 KB/s → expect ~4 s
	const minDuration = 3 * time.Second  // conservative lower bound
	const maxDuration = 30 * time.Second // generous upper bound for CI

	content := make([]byte, fileSize)
	for i := range content {
		content[i] = byte(i % 127)
	}

	srcPath := "throttle-src.txt"
	dstPath := fmt.Sprintf("/upload/throttle-%d.txt", time.Now().UnixNano())
	t.Cleanup(func() { sftp.Delete(context.Background(), dstPath) })

	putFile(t, local, srcPath, content)

	cfg := baseTransferConfig()
	cfg.BandwidthLimit = limitBytesPerSec

	start := time.Now()
	result := doTransfer(t, local, sftp, srcPath, dstPath, cfg)
	elapsed := time.Since(start)

	if elapsed < minDuration {
		t.Errorf("transfer took %v, want >= %v (throttle not active?)", elapsed, minDuration)
	}
	if elapsed > maxDuration {
		t.Errorf("transfer took %v, want <= %v (too slow)", elapsed, maxDuration)
	}
	if got := getFile(t, sftp, dstPath); !bytes.Equal(got, content) {
		t.Errorf("throttled transfer content mismatch")
	}
	_ = result
}

// TestTransfer_Throttle_SFTPToS3 verifies throttle works across backends that
// differ in their I/O model (streaming SFTP read → S3 multipart write).
func TestTransfer_Throttle_SFTPToS3(t *testing.T) {
	skipUnless(t, "SFTP_HOST", "S3_ENDPOINT")
	sftp := newSFTPStore(t)
	s3 := newS3Store(t)

	const fileSize = 256 * 1024
	const limitBytesPerSec = 64 * 1024
	const minDuration = 3 * time.Second
	const maxDuration = 30 * time.Second

	content := make([]byte, fileSize)
	for i := range content {
		content[i] = byte(i % 251)
	}

	srcPath := fmt.Sprintf("/upload/throttle-sftp-s3-%d.txt", time.Now().UnixNano())
	dstPath := fmt.Sprintf("integration-test/%d/throttle-sftp-s3.bin", time.Now().UnixNano())
	t.Cleanup(func() { s3.Delete(context.Background(), dstPath) })

	putFile(t, sftp, srcPath, content)

	cfg := baseTransferConfig()
	cfg.BandwidthLimit = limitBytesPerSec

	start := time.Now()
	result := doTransfer(t, sftp, s3, srcPath, dstPath, cfg)
	elapsed := time.Since(start)

	if elapsed < minDuration {
		t.Errorf("transfer took %v, want >= %v (throttle not active?)", elapsed, minDuration)
	}
	if elapsed > maxDuration {
		t.Errorf("transfer took %v, want <= %v (too slow)", elapsed, maxDuration)
	}
	if got := getFile(t, s3, dstPath); !bytes.Equal(got, content) {
		t.Errorf("throttled transfer content mismatch")
	}
	_ = result
}
