package usecase

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func TestNewCompressWriter_InvalidAlgo(t *testing.T) {
	_, err := newCompressWriter(io.Discard, "bzip2")
	if err == nil {
		t.Fatal("expected error for unsupported algorithm, got nil")
	}
	if !strings.Contains(err.Error(), "bzip2") {
		t.Errorf("error should mention the bad algorithm: %v", err)
	}
}

func TestNewCompressWriter_Passthrough(t *testing.T) {
	var buf bytes.Buffer
	cw, err := newCompressWriter(&buf, "")
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("hello passthrough")
	_, _ = cw.Write(data)
	_ = cw.Close()
	if !bytes.Equal(buf.Bytes(), data) {
		t.Fatalf("passthrough mangled data: got %q want %q", buf.Bytes(), data)
	}
}

func TestNewCompressWriter_Gzip_RoundTrip(t *testing.T) {
	original := make([]byte, 64*1024)
	for i := range original {
		original[i] = byte(i % 256)
	}

	var compressed bytes.Buffer
	cw, err := newCompressWriter(&compressed, CompressionGzip)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cw.Write(original); err != nil {
		t.Fatal(err)
	}
	if err = cw.Close(); err != nil {
		t.Fatal(err)
	}

	gr, err := gzip.NewReader(&compressed)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(gr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, original) {
		t.Fatal("gzip round-trip: decoded bytes differ from original")
	}
}

func TestNewCompressWriter_Zstd_RoundTrip(t *testing.T) {
	original := make([]byte, 64*1024)
	for i := range original {
		original[i] = byte(i % 256)
	}

	var compressed bytes.Buffer
	cw, err := newCompressWriter(&compressed, CompressionZstd)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cw.Write(original); err != nil {
		t.Fatal(err)
	}
	if err = cw.Close(); err != nil {
		t.Fatal(err)
	}

	dec, err := zstd.NewReader(&compressed)
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()
	decoded, err := io.ReadAll(dec)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, original) {
		t.Fatal("zstd round-trip: decoded bytes differ from original")
	}
}

func TestNewCompressWriter_Gzip_SmallerThanOriginal(t *testing.T) {
	// Highly compressible input — all zeros.
	original := make([]byte, 1024*1024)

	var compressed bytes.Buffer
	cw, _ := newCompressWriter(&compressed, CompressionGzip)
	_, _ = cw.Write(original)
	_ = cw.Close()

	if compressed.Len() >= len(original) {
		t.Errorf("gzip output (%d bytes) not smaller than input (%d bytes)", compressed.Len(), len(original))
	}
}

func TestNewCompressWriter_Zstd_SmallerThanOriginal(t *testing.T) {
	original := make([]byte, 1024*1024)

	var compressed bytes.Buffer
	cw, _ := newCompressWriter(&compressed, CompressionZstd)
	_, _ = cw.Write(original)
	_ = cw.Close()

	if compressed.Len() >= len(original) {
		t.Errorf("zstd output (%d bytes) not smaller than input (%d bytes)", compressed.Len(), len(original))
	}
}
