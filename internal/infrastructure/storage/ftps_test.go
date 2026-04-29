package storage

import (
	"fmt"
	"testing"
)

// ---- ftpMkdirAll unit tests (white-box, no real server needed) ----

func TestFtpMkdirAll_NestedAbsolutePath(t *testing.T) {
	var called []string
	makeDir := func(path string) error {
		called = append(called, path)
		return nil
	}

	ftpMkdirAll("/backup/2024/jan", makeDir)

	want := []string{"/backup", "/backup/2024", "/backup/2024/jan"}
	if len(called) != len(want) {
		t.Fatalf("MakeDir called %d times, want %d: %v", len(called), len(want), called)
	}
	for i, w := range want {
		if called[i] != w {
			t.Errorf("call[%d] = %q, want %q", i, called[i], w)
		}
	}
}

func TestFtpMkdirAll_SingleComponent(t *testing.T) {
	var called []string
	makeDir := func(path string) error {
		called = append(called, path)
		return nil
	}

	ftpMkdirAll("/upload", makeDir)

	if len(called) != 1 || called[0] != "/upload" {
		t.Errorf("MakeDir calls = %v, want [/upload]", called)
	}
}

func TestFtpMkdirAll_ErrorsIgnored(t *testing.T) {
	// Even when makeDir returns an error for every call (e.g. dir already
	// exists), all components must still be attempted.
	calls := 0
	makeDir := func(path string) error {
		calls++
		return fmt.Errorf("directory already exists")
	}

	// Should not panic or stop early.
	ftpMkdirAll("/a/b/c/d", makeDir)

	if calls != 4 {
		t.Errorf("expected 4 MakeDir calls, got %d", calls)
	}
}

func TestFtpMkdirAll_DeepPath(t *testing.T) {
	var called []string
	makeDir := func(path string) error {
		called = append(called, path)
		return nil
	}

	ftpMkdirAll("/data/reports/2024/q1/raw", makeDir)

	want := []string{
		"/data",
		"/data/reports",
		"/data/reports/2024",
		"/data/reports/2024/q1",
		"/data/reports/2024/q1/raw",
	}
	if len(called) != len(want) {
		t.Fatalf("MakeDir called %d times, want %d: %v", len(called), len(want), called)
	}
	for i, w := range want {
		if called[i] != w {
			t.Errorf("call[%d] = %q, want %q", i, called[i], w)
		}
	}
}
