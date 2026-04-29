package usecase

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

const (
	CompressionGzip = "gzip"
	CompressionZstd = "zstd"
)

// newCompressWriter wraps w with a compressor for the given algorithm.
// The caller must Close() the returned WriteCloser to flush the compressor trailer.
// algo must be CompressionGzip, CompressionZstd, or "" (passthrough).
func newCompressWriter(w io.Writer, algo string) (io.WriteCloser, error) {
	switch algo {
	case CompressionGzip:
		return gzip.NewWriter(w), nil
	case CompressionZstd:
		enc, err := zstd.NewWriter(w)
		if err != nil {
			return nil, fmt.Errorf("zstd encoder: %w", err)
		}
		return enc, nil
	case "":
		return nopWriteCloser{w}, nil
	default:
		return nil, fmt.Errorf("unsupported compression algorithm %q: use gzip or zstd", algo)
	}
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
