package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	applog "github.com/preedep/go-nixcopy/internal/infrastructure/logger"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/usecase/mocks"
)

func TestPatternMatcher_MatchFiles_ExactMatch(t *testing.T) {
	// Setup
	storage := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

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
	logger := applog.NewNopLogger()

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
	logger := applog.NewNopLogger()
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
	logger := applog.NewNopLogger()
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

func TestPatternMatcher_MatchFiles_Recursive(t *testing.T) {
	storage := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	// Simulate nested directory structure via ListFunc:
	//   /logs/app.log
	//   /logs/errors/2024-01.log
	//   /logs/errors/2024-02.log
	//   /logs/archive/2023.log
	storage.ListFunc = func(ctx context.Context, path string) ([]entity.FileInfo, error) {
		switch path {
		case "/logs":
			return []entity.FileInfo{
				{Path: "/logs/app.log", Name: "app.log", IsDirectory: false},
				{Path: "/logs/errors", Name: "errors", IsDirectory: true},
				{Path: "/logs/archive", Name: "archive", IsDirectory: true},
			}, nil
		case "/logs/errors":
			return []entity.FileInfo{
				{Path: "/logs/errors/2024-01.log", Name: "2024-01.log", IsDirectory: false},
				{Path: "/logs/errors/2024-02.log", Name: "2024-02.log", IsDirectory: false},
			}, nil
		case "/logs/archive":
			return []entity.FileInfo{
				{Path: "/logs/archive/2023.log", Name: "2023.log", IsDirectory: false},
			}, nil
		default:
			return nil, nil
		}
	}

	matcher := NewPatternMatcher(storage, logger)
	files, err := matcher.MatchFiles(context.Background(), "/logs/**/*.log")
	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}
	if len(files) != 4 {
		t.Errorf("matched %d files, want 4: %v", len(files), files)
	}
}

func TestPatternMatcher_MatchFiles_ListError(t *testing.T) {
	storage := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	storage.ListError = errors.New("storage unavailable")

	matcher := NewPatternMatcher(storage, logger)
	_, err := matcher.MatchFiles(context.Background(), "/data/*.pdf")

	if err == nil {
		t.Fatal("MatchFiles() expected error when List fails, got nil")
	}
}

func TestPatternMatcher_MatchFiles_NoMatches(t *testing.T) {
	storage := mocks.NewMockStorage()
	logger := applog.NewNopLogger()

	storage.AddFile("/data/readme.txt", []byte("content"), &entity.FileInfo{
		Path:        "/data/readme.txt",
		Name:        "readme.txt",
		ModifiedTime: time.Now(),
		IsDirectory: false,
	})
	storage.AddFile("/data/notes.md", []byte("notes"), &entity.FileInfo{
		Path:        "/data/notes.md",
		Name:        "notes.md",
		ModifiedTime: time.Now(),
		IsDirectory: false,
	})

	matcher := NewPatternMatcher(storage, logger)
	files, err := matcher.MatchFiles(context.Background(), "/data/*.pdf")

	if err != nil {
		t.Fatalf("MatchFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 matches, got %d: %v", len(files), files)
	}
}
