package usecase

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"
)

func TestParseBandwidth(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"", 0, false},
		{"0", 0, false},
		{"1024", 1024, false},
		{"512KB", 512 * 1024, false},
		{"512kb", 512 * 1024, false},
		{"10MB", 10 * 1024 * 1024, false},
		{"10mb", 10 * 1024 * 1024, false},
		{"10MiB", 10 * 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"1GiB", 1024 * 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"1M", 1024 * 1024, false},
		{"1K", 1024, false},
		{"0.5MB", 512 * 1024, false},
		{"invalid", 0, true},
		{"-1MB", 0, true},
		{"MB", 0, true},
	}

	for _, tc := range tests {
		got, err := ParseBandwidth(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseBandwidth(%q): expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseBandwidth(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseBandwidth(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestNewThrottledReader_ZeroRate_Passthrough(t *testing.T) {
	data := []byte("hello world")
	r := newThrottledReader(context.Background(), bytes.NewReader(data), 0)
	// Should be the original reader, not a ThrottledReader.
	if _, ok := r.(*ThrottledReader); ok {
		t.Fatal("expected passthrough reader when bytesPerSec=0, got ThrottledReader")
	}
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("data mismatch: got %q want %q", got, data)
	}
}

func TestThrottledReader_DataIntegrity(t *testing.T) {
	data := make([]byte, 64*1024) // 64 KB
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Rate = 1 MB/s — fast enough the test finishes quickly with no real sleep.
	r := newThrottledReader(context.Background(), bytes.NewReader(data), 1024*1024)
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("data corrupted by ThrottledReader")
	}
}

func TestThrottledReader_ContextCancel(t *testing.T) {
	// Very low rate (1 byte/s) so the first real read triggers a long sleep.
	// Cancel the context immediately and expect the read to return quickly.
	data := make([]byte, 1024)
	ctx, cancel := context.WithCancel(context.Background())

	r := newThrottledReader(ctx, bytes.NewReader(data), 1) // 1 byte/s
	cancel()

	start := time.Now()
	_, err := io.ReadAll(r)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	// Should abort well within 1 second despite the 1-byte/s rate.
	if elapsed > 2*time.Second {
		t.Fatalf("ThrottledReader did not respect context cancel quickly enough: %v", elapsed)
	}
}
