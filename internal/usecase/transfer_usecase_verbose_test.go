package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// decodeAllLogLines splits buf on newlines and parses each non-empty JSON line.
func decodeAllLogLines(t *testing.T, buf *bytes.Buffer) []map[string]interface{} {
	t.Helper()
	var result []map[string]interface{}
	for _, line := range strings.Split(buf.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("failed to decode log line: %v\nraw: %s", err, line)
		}
		result = append(result, m)
	}
	return result
}

func countLogEntries(logs []map[string]interface{}, msg, level string) int {
	n := 0
	for _, m := range logs {
		if m["message"] == msg && m["level"] == level {
			n++
		}
	}
	return n
}

func hasLogEntry(logs []map[string]interface{}, msg, level string) bool {
	return countLogEntries(logs, msg, level) > 0
}

// verboseTestSource returns a MockStorage pre-loaded with one file at path.
func verboseTestSource(path string, content []byte) *mocks.MockStorage {
	src := mocks.NewMockStorage()
	src.AddFile(path, content, &entity.FileInfo{
		Path:         path,
		Name:         "test.txt",
		Size:         int64(len(content)),
		ModifiedTime: time.Now(),
		IsDirectory:  false,
	})
	return src
}

// TestVerbose_Transfer_Success_EmitsDebugLines verifies that in verbose (DEBUG) mode a
// successful transfer emits a "transfer attempt" debug line and a "Transfer completed" info line.
func TestVerbose_Transfer_Success_EmitsDebugLines(t *testing.T) {
	var buf bytes.Buffer
	log := applog.NewStandardLogger(
		applog.WithOutput(&buf),
		applog.WithMinLevel(applog.LogLevelDebug),
	)

	content := []byte("hello verbose world")
	src := verboseTestSource("/src/file.txt", content)
	dst := mocks.NewMockStorage()

	cfg := &entity.TransferConfig{
		BufferSize:    1024,
		RetryAttempts: 0,
	}

	uc := NewTransferUseCase(src, dst, cfg, log)
	result, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)

	if err != nil {
		t.Fatalf("unexpected transfer error: %v", err)
	}
	if result.Status != entity.TransferStatusCompleted {
		t.Errorf("status = %v, want Completed", result.Status)
	}

	logs := decodeAllLogLines(t, &buf)

	if !hasLogEntry(logs, "transfer attempt", string(applog.LogLevelDebug)) {
		t.Error(`expected debug log "transfer attempt" in verbose mode — not found`)
	}
	if !hasLogEntry(logs, "Transfer completed", string(applog.LogLevelInfo)) {
		t.Error(`expected info log "Transfer completed" — not found`)
	}
}

// TestVerbose_Transfer_Failure_EmitsAttemptDebugAndFinalError verifies that in verbose (DEBUG)
// mode an exhausted retry sequence emits one "attempt failed" debug line per attempt and one
// always-on "Transfer failed after all attempts" error line.
func TestVerbose_Transfer_Failure_EmitsAttemptDebugAndFinalError(t *testing.T) {
	var buf bytes.Buffer
	log := applog.NewStandardLogger(
		applog.WithOutput(&buf),
		applog.WithMinLevel(applog.LogLevelDebug),
	)

	content := []byte("hello")
	src := verboseTestSource("/src/file.txt", content)
	dst := mocks.NewMockStorage()
	dst.WriteError = errors.New("disk full")

	const retries = 1 // 2 total attempts (attempt 0 and attempt 1)
	cfg := &entity.TransferConfig{
		BufferSize:    1024,
		RetryAttempts: retries,
		RetryDelay:    0,
	}

	uc := NewTransferUseCase(src, dst, cfg, log)
	_, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)

	if err == nil {
		t.Fatal("expected transfer to fail, got nil error")
	}

	logs := decodeAllLogLines(t, &buf)

	// Each failed attempt emits one "attempt failed" DEBUG line.
	n := countLogEntries(logs, "attempt failed", string(applog.LogLevelDebug))
	if n != retries+1 {
		t.Errorf(`"attempt failed" DEBUG count = %d, want %d`, n, retries+1)
	}

	// The always-on final error must appear regardless of log level.
	if !hasLogEntry(logs, "Transfer failed after all attempts", string(applog.LogLevelError)) {
		t.Error(`expected ERROR log "Transfer failed after all attempts" — not found`)
	}
}

// TestNonVerbose_Transfer_Failure_SuppressesDebugButEmitsFinalError verifies that at the
// default INFO level the per-attempt "attempt failed" debug lines are suppressed while the
// final "Transfer failed after all attempts" error line is still emitted.
func TestNonVerbose_Transfer_Failure_SuppressesDebugButEmitsFinalError(t *testing.T) {
	var buf bytes.Buffer
	log := applog.NewStandardLogger(applog.WithOutput(&buf)) // default INFO

	content := []byte("hello")
	src := verboseTestSource("/src/file.txt", content)
	dst := mocks.NewMockStorage()
	dst.WriteError = errors.New("disk full")

	cfg := &entity.TransferConfig{
		BufferSize:    1024,
		RetryAttempts: 1,
		RetryDelay:    0,
	}

	uc := NewTransferUseCase(src, dst, cfg, log)
	_, err := uc.Transfer(context.Background(), "/src/file.txt", "/dst/file.txt", nil)

	if err == nil {
		t.Fatal("expected transfer to fail, got nil error")
	}

	logs := decodeAllLogLines(t, &buf)

	// DEBUG lines must be absent at INFO minLevel.
	if hasLogEntry(logs, "attempt failed", string(applog.LogLevelDebug)) {
		t.Error(`"attempt failed" DEBUG must be suppressed at INFO minLevel`)
	}
	if hasLogEntry(logs, "transfer attempt", string(applog.LogLevelDebug)) {
		t.Error(`"transfer attempt" DEBUG must be suppressed at INFO minLevel`)
	}

	// Final error is ERROR severity — must always appear.
	if !hasLogEntry(logs, "Transfer failed after all attempts", string(applog.LogLevelError)) {
		t.Error(`expected ERROR log "Transfer failed after all attempts" — not found`)
	}
}
