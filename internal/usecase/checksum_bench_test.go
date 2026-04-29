package usecase

import (
	"bytes"
	"io"
	"testing"
)

// BenchmarkChecksumReader measures SHA-256 throughput through checksumReader at
// different payload sizes. Use b.SetBytes to get MB/s in the benchmark output.
func BenchmarkChecksumReader(b *testing.B) {
	for _, tc := range []struct {
		name string
		size int
	}{
		{"1MB", 1 << 20},
		{"16MB", 16 << 20},
		{"64MB", 64 << 20},
	} {
		b.Run(tc.name, func(b *testing.B) {
			data := make([]byte, tc.size)
			b.SetBytes(int64(tc.size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cr := newChecksumReader(bytes.NewReader(data))
				if _, err := io.Copy(io.Discard, cr); err != nil {
					b.Fatal(err)
				}
				_ = cr.sum()
			}
		})
	}
}
