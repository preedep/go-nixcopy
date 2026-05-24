// Package usecase implements the application's business logic layer.
package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
)

// defaultListParallelism is the default number of concurrent Storage.List calls.
const defaultListParallelism = 8

// PatternMatcher handles file pattern matching and expansion for wildcard-based file discovery.
// Recursive traversal uses a bounded goroutine pool to list subdirectories in parallel.
type PatternMatcher struct {
	storage      repository.StorageReader
	logger       *applog.StandardLogger
	listParallel int // max concurrent Storage.List calls; defaults to defaultListParallelism
}

// NewPatternMatcher creates a new PatternMatcher instance.
func NewPatternMatcher(storage repository.StorageReader, logger *applog.StandardLogger) *PatternMatcher {
	return &PatternMatcher{
		storage:      storage,
		logger:       logger,
		listParallel: defaultListParallelism,
	}
}

// MatchFiles finds all files matching the given pattern.
func (pm *PatternMatcher) MatchFiles(ctx context.Context, pattern string) ([]string, error) {
	filePattern := entity.NewFilePattern(pattern)

	if !filePattern.IsWildcard {
		return []string{pattern}, nil
	}

	basePath := pm.getBasePath(pattern)

	pm.logger.Info("Matching files",
		applog.F("pattern", pattern),
		applog.F("base_path", basePath),
		applog.F("is_recursive", filePattern.IsRecursive),
	)

	files, err := pm.listFilesParallel(ctx, basePath, filePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	pm.logger.Info("Pattern matching completed",
		applog.F("pattern", pattern),
		applog.F("matched_files", len(files)),
	)

	return files, nil
}

// getBasePath extracts the non-wildcard prefix from a file pattern.
func (pm *PatternMatcher) getBasePath(pattern string) string {
	parts := strings.Split(pattern, "/")
	baseParts := []string{}

	// Iterate through path parts until we hit a wildcard
	for _, part := range parts {
		if strings.ContainsAny(part, "*?[]") {
			break // Stop at first wildcard component
		}
		baseParts = append(baseParts, part)
	}

	// If no base path found (pattern starts with wildcard), use root
	if len(baseParts) == 0 {
		return "/"
	}

	return strings.Join(baseParts, "/")
}

// listFilesParallel lists and filters files using a bounded worker pool for concurrent
// subdirectory traversal. It uses a work-queue goroutine to avoid semaphore deadlocks
// that arise when workers try to enqueue their own children.
func (pm *PatternMatcher) listFilesParallel(
	ctx context.Context,
	basePath string,
	pattern *entity.FilePattern,
) ([]string, error) {
	// List the root directory first to detect hard errors early.
	rootEntries, err := pm.storage.List(ctx, basePath)
	if err != nil {
		return nil, err
	}

	type workItem struct{ path string }
	type result struct {
		matched []string
		dirs    []string
	}

	work := make(chan workItem, 256)
	results := make(chan result, 256)
	errs := make(chan error, 1)

	var wg sync.WaitGroup

	// Seed the initial directory queue and collect root-level matches.
	var initMatched []string
	for _, entry := range rootEntries {
		if entry.IsDirectory {
			if pattern.IsRecursive {
				wg.Add(1)
				work <- workItem{entry.Path}
			}
		} else if pm.matchesPattern(entry.Path, pattern) {
			initMatched = append(initMatched, entry.Path)
		}
	}

	// Workers: list one directory, publish result.
	for i := 0; i < pm.listParallel; i++ {
		go func() {
			for item := range work {
				entries, listErr := pm.storage.List(ctx, item.path)
				if listErr != nil {
					pm.logger.Warn("Failed to list subdirectory",
						applog.F("path", item.path),
						applog.FError(listErr),
					)
					select {
					case errs <- listErr:
					default:
					}
					wg.Done()
					continue
				}

				var r result
				for _, entry := range entries {
					if entry.IsDirectory {
						r.dirs = append(r.dirs, entry.Path)
					} else if pm.matchesPattern(entry.Path, pattern) {
						r.matched = append(r.matched, entry.Path)
					}
				}
				results <- r
				// wg.Done() is called by the coordinator after processing this result.
			}
		}()
	}

	// Coordinator: drain results, enqueue newly discovered dirs, call wg.Done.
	// Runs in its own goroutine so workers and coordinator never block each other.
	done := make(chan []string)
	go func() {
		var all []string
		all = append(all, initMatched...)
		for r := range results {
			wg.Done() // one work item consumed
			all = append(all, r.matched...)
			if pattern.IsRecursive {
				for _, dir := range r.dirs {
					wg.Add(1)
					work <- workItem{dir}
				}
			}
		}
		done <- all
	}()

	wg.Wait()
	close(work)
	close(results)

	matched := <-done

	select {
	case firstErr := <-errs:
		_ = firstErr // sub-directory errors are logged; we still return partial results
	default:
	}

	return matched, nil
}

// matchesPattern reports whether path matches the file pattern.
func (pm *PatternMatcher) matchesPattern(path string, pattern *entity.FilePattern) bool {
	// Fast path: exact match (no wildcards)
	if !pattern.IsWildcard {
		return path == pattern.Pattern
	}

	// Handle recursive patterns (contains **)
	if pattern.IsRecursive {
		// Split pattern by ** to get prefix and suffix
		// Example: "data/**/*.pdf" -> ["data/", "/*.pdf"]
		patternParts := strings.Split(pattern.Pattern, "**")
		if len(patternParts) == 2 {
			prefix := strings.TrimSuffix(patternParts[0], "/")
			suffix := strings.TrimPrefix(patternParts[1], "/")

			// Check prefix match (directory path)
			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			// Check suffix match (filename pattern)
			if suffix != "" {
				// Match suffix against filename only
				matched, err := filepath.Match(suffix, filepath.Base(path))
				if err != nil || !matched {
					return false
				}
			}

			return true
		}
	}

	// Handle simple wildcard patterns (*, ?, [])
	// Match against full path when pattern contains a separator, basename otherwise
	matchTarget := filepath.Base(path)
	if strings.Contains(pattern.Pattern, "/") {
		matchTarget = path
	}
	matched, err := filepath.Match(pattern.Pattern, matchTarget)
	if err != nil {
		return false
	}

	return matched
}
