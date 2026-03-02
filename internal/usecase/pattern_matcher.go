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

type PatternMatcher struct {
	storage repository.StorageReader
	logger  *zap.Logger
}

func NewPatternMatcher(storage repository.StorageReader, logger *zap.Logger) *PatternMatcher {
	return &PatternMatcher{
		storage: storage,
		logger:  logger,
	}
}

func (pm *PatternMatcher) MatchFiles(ctx context.Context, pattern string) ([]string, error) {
	filePattern := entity.NewFilePattern(pattern)

	if !filePattern.IsWildcard {
		return []string{pattern}, nil
	}

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

func (pm *PatternMatcher) getBasePath(pattern string) string {
	parts := strings.Split(pattern, "/")
	baseParts := []string{}

	for _, part := range parts {
		if strings.ContainsAny(part, "*?[]") {
			break
		}
		baseParts = append(baseParts, part)
	}

	if len(baseParts) == 0 {
		return "/"
	}

	return strings.Join(baseParts, "/")
}

func (pm *PatternMatcher) listFilesRecursive(
	ctx context.Context,
	basePath string,
	pattern *entity.FilePattern,
) ([]string, error) {
	var matchedFiles []string

	files, err := pm.storage.List(ctx, basePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDirectory {
			if pattern.IsRecursive {
				subFiles, err := pm.listFilesRecursive(ctx, file.Path, pattern)
				if err != nil {
					pm.logger.Warn("Failed to list subdirectory",
						zap.String("path", file.Path),
						zap.Error(err),
					)
					continue
				}
				matchedFiles = append(matchedFiles, subFiles...)
			}
		} else {
			if pm.matchesPattern(file.Path, pattern) {
				matchedFiles = append(matchedFiles, file.Path)
			}
		}
	}

	return matchedFiles, nil
}

func (pm *PatternMatcher) matchesPattern(path string, pattern *entity.FilePattern) bool {
	if !pattern.IsWildcard {
		return path == pattern.Pattern
	}

	if pattern.IsRecursive {
		patternParts := strings.Split(pattern.Pattern, "**")
		if len(patternParts) == 2 {
			prefix := strings.TrimSuffix(patternParts[0], "/")
			suffix := strings.TrimPrefix(patternParts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			if suffix != "" {
				matched, err := filepath.Match(suffix, filepath.Base(path))
				if err != nil || !matched {
					return false
				}
			}

			return true
		}
	}

	matched, err := filepath.Match(pattern.Pattern, filepath.Base(path))
	if err != nil {
		return false
	}

	return matched
}
