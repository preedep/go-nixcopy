package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
	"go.uber.org/zap"
)

func TestPatternMatcher_MatchFiles_ExactMatch(t *testing.T) {
	// Setup
	storage := mocks.NewMockStorage()
	logger := zap.NewNop()

	storage.AddFile("/data/file.txt", []byte("content"), &entity.FileInfo{
		Path:         "/data/file.txt",
		Name:         "file.txt",
		Size:         7,
		ModifiedTime: time.Now(),
		IsDirectory:  false,
	})

	matcher := NewPatternMatcher(storage, logger)

	// Execute
	ctx := context.Background()
	files, err := matcher.MatchFiles(ctx, "/data/file.txt")

	// Assert
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Matched files count = %v, want 1", len(files))
	}

	if files[0] != "/data/file.txt" {
		t.Errorf("Matched file = %v, want /data/file.txt", files[0])
	}
}

func TestPatternMatcher_MatchFiles_Wildcard(t *testing.T) {
	// Setup
	storage := mocks.NewMockStorage()
	logger := zap.NewNop()

	// Add multiple PDF files
	pdfFiles := []string{"report1.pdf", "report2.pdf", "document.pdf"}
	for _, filename := range pdfFiles {
		path := "/data/" + filename
		storage.AddFile(path, []byte("content"), &entity.FileInfo{
			Path:         path,
			Name:         filename,
			Size:         7,
			ModifiedTime: time.Now(),
			IsDirectory:  false,
		})
	}

	// Add non-PDF file
	storage.AddFile("/data/readme.txt", []byte("content"), &entity.FileInfo{
		Path:         "/data/readme.txt",
		Name:         "readme.txt",
		Size:         7,
		ModifiedTime: time.Now(),
		IsDirectory:  false,
	})

	matcher := NewPatternMatcher(storage, logger)

	// Execute
	ctx := context.Background()
	files, err := matcher.MatchFiles(ctx, "/data/*.pdf")

	// Assert
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Matched files count = %v, want 3", len(files))
	}

	// Verify all matched files are PDFs
	for _, file := range files {
		if len(file) < 4 || file[len(file)-4:] != ".pdf" {
			t.Errorf("Non-PDF file matched: %v", file)
		}
	}
}

func TestPatternMatcher_GetBasePath(t *testing.T) {
	storage := mocks.NewMockStorage()
	logger := zap.NewNop()
	matcher := NewPatternMatcher(storage, logger)

	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{
			name:    "simple wildcard",
			pattern: "/data/*.pdf",
			want:    "/data",
		},
		{
			name:    "recursive pattern",
			pattern: "/data/**/*.log",
			want:    "/data",
		},
		{
			name:    "nested path with wildcard",
			pattern: "/data/2024/reports/*.xlsx",
			want:    "/data/2024/reports",
		},
		{
			name:    "root wildcard",
			pattern: "*.txt",
			want:    "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matcher.getBasePath(tt.pattern)
			if got != tt.want {
				t.Errorf("getBasePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatternMatcher_MatchesPattern(t *testing.T) {
	storage := mocks.NewMockStorage()
	logger := zap.NewNop()
	matcher := NewPatternMatcher(storage, logger)

	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{
			name:    "exact match",
			path:    "/data/file.txt",
			pattern: "/data/file.txt",
			want:    true,
		},
		{
			name:    "wildcard match",
			path:    "/data/report.pdf",
			pattern: "*.pdf",
			want:    true,
		},
		{
			name:    "wildcard no match",
			path:    "/data/report.txt",
			pattern: "*.pdf",
			want:    false,
		},
		{
			name:    "prefix wildcard match",
			path:    "/data/report_2024.xlsx",
			pattern: "report*.xlsx",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePattern := entity.NewFilePattern(tt.pattern)
			got := matcher.matchesPattern(tt.path, filePattern)
			if got != tt.want {
				t.Errorf("matchesPattern() = %v, want %v for path %q and pattern %q", got, tt.want, tt.path, tt.pattern)
			}
		})
	}
}
