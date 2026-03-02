// Package usecase implements the application's business logic layer.
package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	"go.uber.org/zap"
)

// PatternMatcher handles file pattern matching and expansion for wildcard-based file discovery.
//
// It supports various wildcard patterns including:
//   - Simple wildcards: *.pdf, report*.xlsx, file?.txt
//   - Character classes: file[0-9].txt, [a-z]*.doc
//   - Recursive patterns: **/*.log, data/**/*.csv
//
// The matcher works by:
//  1. Extracting the base path (non-wildcard prefix) from the pattern
//  2. Listing files from the base path
//  3. Recursively traversing subdirectories if pattern contains **
//  4. Matching each file against the pattern using filepath.Match
//
// Performance Considerations:
// For patterns like "**/*.log", the matcher will traverse the entire directory tree
// starting from the base path. This can be slow for large directory structures.
// Consider using more specific base paths when possible (e.g., "logs/**/*.log" instead of "**/*.log").
//
// Thread Safety:
// PatternMatcher is safe for concurrent use. Multiple goroutines can call
// MatchFiles() simultaneously on the same instance.
type PatternMatcher struct {
	storage repository.StorageReader // Storage reader for listing and accessing files
	logger  *zap.Logger              // Structured logger for debugging pattern matching
}

// NewPatternMatcher creates a new PatternMatcher instance.
//
// Parameters:
//   - storage: Storage reader for listing files and directories
//   - logger: Structured logger for tracking pattern matching operations
//
// The returned matcher is ready to use and safe for concurrent access.
//
// Example:
//
//	matcher := NewPatternMatcher(storageReader, logger)
//	files, err := matcher.MatchFiles(ctx, "logs/**/*.log")
func NewPatternMatcher(storage repository.StorageReader, logger *zap.Logger) *PatternMatcher {
	return &PatternMatcher{
		storage: storage,
		logger:  logger,
	}
}

// MatchFiles finds all files matching the given pattern.
//
// The method supports various wildcard patterns:
//   - "*.pdf" - all PDF files in base directory
//   - "report*.xlsx" - files starting with "report" and ending with .xlsx
//   - "**/*.log" - all .log files in base directory and subdirectories (recursive)
//   - "data/2024/**/*.csv" - all CSV files under data/2024/ and its subdirectories
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - pattern: File pattern with optional wildcards (*, ?, [], **)
//
// Returns:
//   - []string: Slice of matched file paths (absolute paths from storage root)
//   - error: Non-nil if listing fails or pattern is invalid
//
// Behavior:
//   - If pattern contains no wildcards, returns the pattern as-is (single file)
//   - Extracts base path (non-wildcard prefix) for efficient directory traversal
//   - Recursively traverses subdirectories if pattern contains **
//   - Skips subdirectories that fail to list (logs warning, continues)
//   - Returns empty slice if no files match (not an error)
//
// Performance:
// Recursive patterns (**) can be slow on large directory trees.
// Time complexity: O(n) where n is the number of files/directories traversed.
// For better performance, use specific base paths (e.g., "logs/**/*.log" vs "**/*.log").
//
// Example:
//
//	// Match all PDF files in documents directory
//	files, err := matcher.MatchFiles(ctx, "documents/*.pdf")
//	// files = ["/documents/report.pdf", "/documents/invoice.pdf"]
//
//	// Match all log files recursively
//	files, err := matcher.MatchFiles(ctx, "**/*.log")
//	// files = ["/app.log", "/logs/error.log", "/logs/2024/access.log"]
func (pm *PatternMatcher) MatchFiles(ctx context.Context, pattern string) ([]string, error) {
	filePattern := entity.NewFilePattern(pattern)

	// Fast path: if no wildcards, return pattern as-is
	// This avoids unnecessary directory listing for exact file paths
	if !filePattern.IsWildcard {
		return []string{pattern}, nil
	}

	// Extract base path (non-wildcard prefix) for efficient traversal
	// Example: "data/2024/**/*.csv" -> base path = "data/2024"
	basePath := pm.getBasePath(pattern)

	pm.logger.Info("Matching files",
		zap.String("pattern", pattern),
		zap.String("base_path", basePath),
		zap.Bool("is_recursive", filePattern.IsRecursive),
	)

	files, err := pm.listFilesRecursive(ctx, basePath, filePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	pm.logger.Info("Pattern matching completed",
		zap.String("pattern", pattern),
		zap.Int("matched_files", len(files)),
	)

	return files, nil
}

// getBasePath extracts the non-wildcard prefix from a file pattern.
//
// This method identifies the longest path prefix that doesn't contain wildcards,
// which serves as the starting point for directory traversal.
//
// Examples:
//   - "data/2024/**/*.csv" -> "data/2024"
//   - "*.pdf" -> "/" (root)
//   - "logs/app*.log" -> "logs"
//   - "/var/log/**/*.log" -> "/var/log"
//
// Algorithm:
// Splits the pattern by "/" and iterates through parts until finding one
// containing wildcards (*, ?, []). Returns the path up to that point.
//
// Returns:
//   - Base path string (always starts with / if absolute, or relative path)
//   - "/" if pattern starts with a wildcard
//
// Performance: O(n) where n is the number of path components.
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

// listFilesRecursive recursively lists and filters files matching the pattern.
//
// This is the core recursive traversal function that:
//  1. Lists all files/directories in the current path
//  2. For each directory: recursively descends if pattern is recursive (**)
//  3. For each file: checks if it matches the pattern
//  4. Accumulates all matching file paths
//
// Parameters:
//   - ctx: Context for cancellation
//   - basePath: Current directory path to list
//   - pattern: File pattern to match against
//
// Returns:
//   - []string: All matching file paths found in this directory and subdirectories
//   - error: Non-nil if listing the base path fails
//
// Error Handling:
// If a subdirectory fails to list, the error is logged and that subdirectory
// is skipped. The function continues processing other directories.
// This prevents a single permission error from stopping the entire traversal.
//
// Performance:
// Time complexity: O(n) where n is total files/directories traversed.
// Space complexity: O(d) where d is maximum directory depth (recursion stack).
//
// Example traversal for pattern "**/*.log":
//
//	/app/
//	  app.log        -> matched
//	  data/
//	    file.txt     -> not matched
//	  logs/
//	    error.log    -> matched
//	    2024/
//	      access.log -> matched
func (pm *PatternMatcher) listFilesRecursive(
	ctx context.Context,
	basePath string,
	pattern *entity.FilePattern,
) ([]string, error) {
	var matchedFiles []string

	// List all files and directories in current path
	files, err := pm.storage.List(ctx, basePath)
	if err != nil {
		return nil, err
	}

	// Process each entry (file or directory)
	for _, file := range files {
		if file.IsDirectory {
			// Recursively traverse subdirectories if pattern is recursive (**)
			if pattern.IsRecursive {
				subFiles, err := pm.listFilesRecursive(ctx, file.Path, pattern)
				if err != nil {
					// Log warning but continue processing other directories
					// This handles permission errors gracefully
					pm.logger.Warn("Failed to list subdirectory",
						zap.String("path", file.Path),
						zap.Error(err),
					)
					continue
				}
				// Accumulate matches from subdirectory
				matchedFiles = append(matchedFiles, subFiles...)
			}
		} else {
			// Check if file matches the pattern
			if pm.matchesPattern(file.Path, pattern) {
				matchedFiles = append(matchedFiles, file.Path)
			}
		}
	}

	return matchedFiles, nil
}

// matchesPattern checks if a file path matches the given pattern.
//
// This method handles three types of patterns:
//  1. Exact match: No wildcards, direct string comparison
//  2. Recursive pattern: Contains **, matches across directory levels
//  3. Simple wildcard: Contains *, ?, or [], matches filename only
//
// Parameters:
//   - path: Full file path to check (e.g., "/data/2024/report.pdf")
//   - pattern: File pattern to match against
//
// Returns:
//   - bool: true if path matches pattern, false otherwise
//
// Pattern Matching Logic:
//
// For recursive patterns (e.g., "data/**/*.pdf"):
//  1. Split pattern by ** into prefix and suffix
//  2. Check if path starts with prefix (if prefix exists)
//  3. Match suffix against filename using filepath.Match
//
// For simple wildcards (e.g., "*.pdf"):
//  1. Extract filename from path using filepath.Base
//  2. Match pattern against filename using filepath.Match
//
// Examples:
//   - matchesPattern("/data/report.pdf", "*.pdf") -> true
//   - matchesPattern("/data/2024/report.pdf", "**/*.pdf") -> true
//   - matchesPattern("/data/2024/report.pdf", "data/**/*.pdf") -> true
//   - matchesPattern("/other/report.pdf", "data/**/*.pdf") -> false
//
// Edge Cases:
//   - Returns false if filepath.Match returns an error (invalid pattern)
//   - Empty prefix in recursive pattern matches any directory
//   - Empty suffix in recursive pattern matches any filename
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
	// Match against filename only, not full path
	matched, err := filepath.Match(pattern.Pattern, filepath.Base(path))
	if err != nil {
		// Invalid pattern, return false
		return false
	}

	return matched
}
