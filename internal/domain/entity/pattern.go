package entity

import (
	"path/filepath"
	"strings"
	"time"
)

type FilePattern struct {
	Pattern     string
	IsWildcard  bool
	IsRecursive bool
}

func NewFilePattern(pattern string) *FilePattern {
	return &FilePattern{
		Pattern:     pattern,
		IsWildcard:  strings.ContainsAny(pattern, "*?[]"),
		IsRecursive: strings.Contains(pattern, "**"),
	}
}

func (fp *FilePattern) Match(path string) bool {
	if !fp.IsWildcard {
		return path == fp.Pattern
	}

	matched, err := filepath.Match(fp.Pattern, filepath.Base(path))
	if err != nil {
		return false
	}

	return matched
}

func (fp *FilePattern) MatchFull(path string) bool {
	if !fp.IsWildcard {
		return path == fp.Pattern
	}

	if fp.IsRecursive {
		patternParts := strings.Split(fp.Pattern, "**")
		if len(patternParts) != 2 {
			return false
		}
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

	matched, err := filepath.Match(fp.Pattern, path)
	if err != nil {
		return false
	}

	return matched
}

type BatchTransferItem struct {
	SourcePath      string
	DestinationPath string
	Size            int64
}

type BatchTransferResult struct {
	Items           []*TransferResult
	TotalFiles      int
	SuccessfulFiles int
	FailedFiles     int
	TotalBytes      int64
	TotalDuration   time.Duration
	AverageSpeed    float64
}
