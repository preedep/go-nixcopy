package entity

import (
	"testing"
)

func TestNewFilePattern(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		wantWildcard  bool
		wantRecursive bool
	}{
		{
			name:          "simple file",
			pattern:       "file.txt",
			wantWildcard:  false,
			wantRecursive: false,
		},
		{
			name:          "wildcard pattern",
			pattern:       "*.pdf",
			wantWildcard:  true,
			wantRecursive: false,
		},
		{
			name:          "recursive pattern",
			pattern:       "**/*.log",
			wantWildcard:  true,
			wantRecursive: true,
		},
		{
			name:          "question mark wildcard",
			pattern:       "file?.txt",
			wantWildcard:  true,
			wantRecursive: false,
		},
		{
			name:          "bracket wildcard",
			pattern:       "file[0-9].txt",
			wantWildcard:  true,
			wantRecursive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := NewFilePattern(tt.pattern)

			if fp.Pattern != tt.pattern {
				t.Errorf("Pattern = %v, want %v", fp.Pattern, tt.pattern)
			}

			if fp.IsWildcard != tt.wantWildcard {
				t.Errorf("IsWildcard = %v, want %v", fp.IsWildcard, tt.wantWildcard)
			}

			if fp.IsRecursive != tt.wantRecursive {
				t.Errorf("IsRecursive = %v, want %v", fp.IsRecursive, tt.wantRecursive)
			}
		})
	}
}

func TestFilePattern_Match(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "file.txt",
			path:    "file.txt",
			want:    true,
		},
		{
			name:    "exact no match",
			pattern: "file.txt",
			path:    "other.txt",
			want:    false,
		},
		{
			name:    "wildcard match",
			pattern: "*.pdf",
			path:    "/path/to/document.pdf",
			want:    true,
		},
		{
			name:    "wildcard no match",
			pattern: "*.pdf",
			path:    "/path/to/document.txt",
			want:    false,
		},
		{
			name:    "prefix wildcard match",
			pattern: "report*.xlsx",
			path:    "/data/report_2024.xlsx",
			want:    true,
		},
		{
			name:    "question mark match",
			pattern: "file?.txt",
			path:    "/data/file1.txt",
			want:    true,
		},
		{
			name:    "question mark no match",
			pattern: "file?.txt",
			path:    "/data/file12.txt",
			want:    false,
		},
		{
			name:    "bracket match",
			pattern: "file[0-9].txt",
			path:    "/data/file5.txt",
			want:    true,
		},
		{
			name:    "bracket no match",
			pattern: "file[0-9].txt",
			path:    "/data/fileA.txt",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := NewFilePattern(tt.pattern)
			got := fp.Match(tt.path)

			if got != tt.want {
				t.Errorf("Match() = %v, want %v for pattern %q and path %q", got, tt.want, tt.pattern, tt.path)
			}
		})
	}
}

func TestFilePattern_MatchFull(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "/data/file.txt",
			path:    "/data/file.txt",
			want:    true,
		},
		{
			name:    "wildcard full path",
			pattern: "/data/*.pdf",
			path:    "/data/document.pdf",
			want:    true,
		},
		{
			name:    "recursive pattern match",
			pattern: "**/*.log",
			path:    "/var/log/app/error.log",
			want:    true,
		},
		{
			name:    "recursive pattern no match",
			pattern: "**/*.log",
			path:    "/var/log/app/error.txt",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := NewFilePattern(tt.pattern)
			got := fp.MatchFull(tt.path)

			if got != tt.want {
				t.Errorf("MatchFull() = %v, want %v for pattern %q and path %q", got, tt.want, tt.pattern, tt.path)
			}
		})
	}
}
