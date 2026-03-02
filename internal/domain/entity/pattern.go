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
		pattern := strings.ReplaceAll(fp.Pattern, "**", "*")
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return false
		}
		return matched
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
