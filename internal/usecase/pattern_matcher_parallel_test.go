package usecase

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"testing"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

func TestPatternMatcher_Parallel_RecursiveListing(t *testing.T) {
	storage := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	// Build a 3-level directory tree: /root/{a,b,c}/{1,2}/*.log + /root/{a,b,c}/app.log
	storage.ListFunc = func(ctx context.Context, path string) ([]entity.FileInfo, error) {
		switch path {
		case "/root":
			return []entity.FileInfo{
				{Path: "/root/a", Name: "a", IsDirectory: true},
				{Path: "/root/b", Name: "b", IsDirectory: true},
				{Path: "/root/c", Name: "c", IsDirectory: true},
			}, nil
		case "/root/a", "/root/b", "/root/c":
			return []entity.FileInfo{
				{Path: path + "/1", Name: "1", IsDirectory: true},
				{Path: path + "/2", Name: "2", IsDirectory: true},
				{Path: path + "/app.log", Name: "app.log", IsDirectory: false},
				{Path: path + "/readme.txt", Name: "readme.txt", IsDirectory: false},
			}, nil
		}
		// leaf dirs
		for _, letter := range []string{"a", "b", "c"} {
			for _, num := range []string{"1", "2"} {
				if path == fmt.Sprintf("/root/%s/%s", letter, num) {
					return []entity.FileInfo{
						{Path: path + "/out.log", Name: "out.log", IsDirectory: false},
						{Path: path + "/data.csv", Name: "data.csv", IsDirectory: false},
					}, nil
				}
			}
		}
		return nil, nil
	}

	// Expected matches: /root/{a,b,c}/app.log + /root/{a,b,c}/{1,2}/out.log = 3 + 6 = 9
	matcher := &PatternMatcher{storage: storage, logger: logger, listParallel: 4}
	files, err := matcher.MatchFiles(context.Background(), "/root/**/*.log")
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}
	if len(files) != 9 {
		t.Errorf("matched %d files, want 9: %v", len(files), files)
	}
}

func TestPatternMatcher_Parallel_ConcurrencyBounded(t *testing.T) {
	// Verify that concurrent List calls never exceed listParallel.
	const parallelism = 3
	var active atomic.Int32
	var maxSeen atomic.Int32

	storage := mocks.NewMockStorage()
	storage.ListFunc = func(ctx context.Context, path string) ([]entity.FileInfo, error) {
		cur := active.Add(1)
		if cur > maxSeen.Load() {
			maxSeen.Store(cur)
		}
		defer active.Add(-1)

		if path == "/root" {
			dirs := make([]entity.FileInfo, 10)
			for i := range dirs {
				dirs[i] = entity.FileInfo{Path: fmt.Sprintf("/root/d%d", i), Name: fmt.Sprintf("d%d", i), IsDirectory: true}
			}
			return dirs, nil
		}
		// leaf — return one file
		return []entity.FileInfo{
			{Path: path + "/x.log", Name: "x.log", IsDirectory: false},
		}, nil
	}

	logger := applog.NewNopLogger()
	matcher := &PatternMatcher{storage: storage, logger: logger, listParallel: parallelism}
	files, err := matcher.MatchFiles(context.Background(), "/root/**/*.log")
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}
	if len(files) != 10 {
		t.Errorf("matched %d files, want 10", len(files))
	}
	// The root dir is listed serially before the pool starts, so maxSeen reflects
	// pool goroutines only. With 10 leaf dirs and parallelism=3 the pool may peak at 3.
	if maxSeen.Load() > int32(parallelism)+1 { // +1: root listing is outside the pool
		t.Errorf("max concurrent List calls = %d, want <= %d", maxSeen.Load(), parallelism+1)
	}
}

func TestPatternMatcher_Parallel_SubdirListError_Skipped(t *testing.T) {
	// A subdirectory that fails to list should be skipped; other matches still returned.
	storage := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	storage.ListFunc = func(ctx context.Context, path string) ([]entity.FileInfo, error) {
		switch path {
		case "/root":
			return []entity.FileInfo{
				{Path: "/root/ok", Name: "ok", IsDirectory: true},
				{Path: "/root/bad", Name: "bad", IsDirectory: true},
				{Path: "/root/top.log", Name: "top.log", IsDirectory: false},
			}, nil
		case "/root/ok":
			return []entity.FileInfo{
				{Path: "/root/ok/a.log", Name: "a.log", IsDirectory: false},
			}, nil
		case "/root/bad":
			return nil, fmt.Errorf("permission denied")
		default:
			return nil, nil
		}
	}

	matcher := &PatternMatcher{storage: storage, logger: logger, listParallel: 4}
	files, err := matcher.MatchFiles(context.Background(), "/root/**/*.log")
	if err != nil {
		t.Fatalf("MatchFiles() unexpected error: %v", err)
	}
	// top.log + ok/a.log = 2 (bad/ is skipped, not fatal)
	if len(files) != 2 {
		t.Errorf("matched %d files, want 2 (bad/ skipped): %v", len(files), files)
	}
}

func TestPatternMatcher_Parallel_ResultsComplete(t *testing.T) {
	// Build 20 files across 4 directories; verify all are returned.
	storage := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	all := []string{}
	storage.ListFunc = func(ctx context.Context, path string) ([]entity.FileInfo, error) {
		if path == "/data" {
			dirs := []entity.FileInfo{}
			for i := 0; i < 4; i++ {
				dirs = append(dirs, entity.FileInfo{Path: fmt.Sprintf("/data/d%d", i), Name: fmt.Sprintf("d%d", i), IsDirectory: true})
			}
			return dirs, nil
		}
		for i := 0; i < 4; i++ {
			if path == fmt.Sprintf("/data/d%d", i) {
				files := []entity.FileInfo{}
				for j := 0; j < 5; j++ {
					p := fmt.Sprintf("/data/d%d/f%d.txt", i, j)
					files = append(files, entity.FileInfo{Path: p, Name: fmt.Sprintf("f%d.txt", j), IsDirectory: false})
				}
				return files, nil
			}
		}
		return nil, nil
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 5; j++ {
			all = append(all, fmt.Sprintf("/data/d%d/f%d.txt", i, j))
		}
	}

	matcher := &PatternMatcher{storage: storage, logger: logger, listParallel: 8}
	files, err := matcher.MatchFiles(context.Background(), "/data/**/*.txt")
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}
	if len(files) != len(all) {
		t.Errorf("matched %d files, want %d", len(files), len(all))
	}
	sort.Strings(files)
	sort.Strings(all)
	for i := range all {
		if files[i] != all[i] {
			t.Errorf("files[%d] = %q, want %q", i, files[i], all[i])
		}
	}
}
