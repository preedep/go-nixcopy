package usecase

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	appconfig "github.com/preedep/go-nixcopy/internal/infrastructure/config"
	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"
	"github.com/preedep/go-nixcopy/internal/infrastructure/storage"
)

func benchLocalStore(b *testing.B, dir string) repository.Storage {
	b.Helper()
	s, err := storage.NewLocalStorage(&appconfig.LocalConfig{BasePath: dir})
	if err != nil {
		b.Fatal(err)
	}
	if err := s.Connect(context.Background()); err != nil {
		b.Fatal(err)
	}
	return s
}

func BenchmarkTransfer(b *testing.B) {
	for _, tc := range []struct {
		name string
		size int64
	}{
		{"1MB", 1 << 20},
		{"16MB", 16 << 20},
		{"128MB", 128 << 20},
	} {
		b.Run(tc.name, func(b *testing.B) {
			src := benchLocalStore(b, b.TempDir())
			dst := benchLocalStore(b, b.TempDir())

			data := make([]byte, tc.size)
			ctx := context.Background()
			if err := src.Write(ctx, "src.bin", bytes.NewReader(data), tc.size); err != nil {
				b.Fatal(err)
			}

			cfg := &entity.TransferConfig{
				BufferSize:      32 * 1024 * 1024,
				ConcurrentFiles: 1,
				RetryAttempts:   0,
				RetryDelay:      time.Millisecond,
			}
			uc := NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())

			b.SetBytes(tc.size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := uc.Transfer(ctx, "src.bin", "dst.bin", nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTransfer_WithChecksum(b *testing.B) {
	for _, tc := range []struct {
		name string
		size int64
	}{
		{"1MB", 1 << 20},
		{"16MB", 16 << 20},
	} {
		b.Run(tc.name, func(b *testing.B) {
			src := benchLocalStore(b, b.TempDir())
			dst := benchLocalStore(b, b.TempDir())

			data := make([]byte, tc.size)
			ctx := context.Background()
			if err := src.Write(ctx, "src.bin", bytes.NewReader(data), tc.size); err != nil {
				b.Fatal(err)
			}

			cfg := &entity.TransferConfig{
				BufferSize:      32 * 1024 * 1024,
				ConcurrentFiles: 1,
				RetryAttempts:   0,
				RetryDelay:      time.Millisecond,
				VerifyChecksum:  true,
			}
			uc := NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())

			b.SetBytes(tc.size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := uc.Transfer(ctx, "src.bin", "dst.bin", nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkTransferBatch measures concurrent batch throughput at varying concurrency levels.
// Each run transfers 8 × 1 MB files, so b.SetBytes reports total bytes per op.
func BenchmarkTransferBatch(b *testing.B) {
	const (
		fileSize = 1 << 20 // 1 MB per file
		numFiles = 8
	)
	for _, concurrency := range []int{1, 4, 8} {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			src := benchLocalStore(b, b.TempDir())
			dst := benchLocalStore(b, b.TempDir())

			data := make([]byte, fileSize)
			ctx := context.Background()

			sourcePaths := make([]string, numFiles)
			for i := range sourcePaths {
				name := fmt.Sprintf("file%02d.bin", i)
				if err := src.Write(ctx, name, bytes.NewReader(data), fileSize); err != nil {
					b.Fatal(err)
				}
				sourcePaths[i] = name
			}

			cfg := &entity.TransferConfig{
				BufferSize:      8 * 1024 * 1024,
				ConcurrentFiles: concurrency,
				RetryAttempts:   0,
				RetryDelay:      time.Millisecond,
			}
			uc := NewTransferUseCase(src, dst, cfg, applog.NewNopLogger())

			b.SetBytes(int64(numFiles) * fileSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := uc.TransferBatch(ctx, sourcePaths, ".", nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
