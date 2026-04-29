package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

// BenchmarkPatternMatcher_ExactMatch measures the fast path: no wildcard, no List call.
func BenchmarkPatternMatcher_ExactMatch(b *testing.B) {
	mock := mocks.NewMockStorage()
	mock.AddFile("/data/report.pdf", []byte("x"), &entity.FileInfo{
		Path: "/data/report.pdf",
		Name: "report.pdf",
		Size: 1,
	})
	matcher := NewPatternMatcher(mock, applog.NewNopLogger())
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := matcher.MatchFiles(ctx, "/data/report.pdf"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPatternMatcher_Flat measures wildcard matching (*.pdf) against a flat directory.
// ListFunc returns N files synchronously — all memory, no I/O — so this isolates the
// pattern-matching and directory-scan logic.
func BenchmarkPatternMatcher_Flat(b *testing.B) {
	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("%d_files", n), func(b *testing.B) {
			mock := mocks.NewMockStorage()
			mock.ListFunc = func(_ context.Context, _ string) ([]entity.FileInfo, error) {
				files := make([]entity.FileInfo, n)
				for i := range files {
					name := fmt.Sprintf("report%04d.pdf", i)
					files[i] = entity.FileInfo{
						Path:         "/data/" + name,
						Name:         name,
						Size:         1024,
						ModifiedTime: time.Time{},
						IsDirectory:  false,
					}
				}
				return files, nil
			}
			matcher := NewPatternMatcher(mock, applog.NewNopLogger())
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := matcher.MatchFiles(ctx, "/data/*.pdf"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkPatternMatcher_Recursive measures recursive traversal (**/*.log) across a
// simulated multi-level directory tree, isolating the recursion and match overhead.
func BenchmarkPatternMatcher_Recursive(b *testing.B) {
	for _, tc := range []struct {
		dirs        int
		filesPerDir int
	}{
		{10, 100},
		{50, 100},
	} {
		b.Run(fmt.Sprintf("%ddirs_%dfiles", tc.dirs, tc.filesPerDir), func(b *testing.B) {
			mock := mocks.NewMockStorage()
			mock.ListFunc = func(_ context.Context, path string) ([]entity.FileInfo, error) {
				if path == "/logs" {
					dirs := make([]entity.FileInfo, tc.dirs)
					for i := range dirs {
						dirs[i] = entity.FileInfo{
							Path:        fmt.Sprintf("/logs/dir%04d", i),
							Name:        fmt.Sprintf("dir%04d", i),
							IsDirectory: true,
						}
					}
					return dirs, nil
				}
				files := make([]entity.FileInfo, tc.filesPerDir)
				for i := range files {
					name := fmt.Sprintf("app%04d.log", i)
					files[i] = entity.FileInfo{
						Path:        path + "/" + name,
						Name:        name,
						Size:        1024,
						IsDirectory: false,
					}
				}
				return files, nil
			}
			matcher := NewPatternMatcher(mock, applog.NewNopLogger())
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := matcher.MatchFiles(ctx, "/logs/**/*.log"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
