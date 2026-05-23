package storage

import (
	"testing"

	appconfig "github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

func TestS3PartSizeFor(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		wantMin  int64 // result must be >= wantMin
		wantMax  int64 // result must be <= wantMax
		maxParts int64 // size/result must be <= maxParts (when size > 0)
	}{
		{
			name:    "unknown size returns default",
			size:    0,
			wantMin: s3DefaultPartSize,
			wantMax: s3DefaultPartSize,
		},
		{
			name:    "small file uses default",
			size:    1 * 1024 * 1024, // 1 MB
			wantMin: s3DefaultPartSize,
			wantMax: s3DefaultPartSize,
		},
		{
			name:    "100 MB file uses default",
			size:    100 * 1024 * 1024,
			wantMin: s3DefaultPartSize,
			wantMax: s3DefaultPartSize,
		},
		{
			name:     "160 GB file stays within 10000 parts",
			size:     160 * 1024 * 1024 * 1024,
			wantMin:  s3MinPartSize,
			wantMax:  1<<62 - 1, // no hard upper limit tested
			maxParts: s3MaxParts,
		},
		{
			name:     "5 TB file stays within 10000 parts",
			size:     5 * 1024 * 1024 * 1024 * 1024,
			wantMin:  s3MinPartSize,
			wantMax:  1<<62 - 1,
			maxParts: s3MaxParts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s3PartSizeFor(tt.size)

			if got < tt.wantMin {
				t.Errorf("s3PartSizeFor(%d) = %d, want >= %d", tt.size, got, tt.wantMin)
			}
			if got > tt.wantMax {
				t.Errorf("s3PartSizeFor(%d) = %d, want <= %d", tt.size, got, tt.wantMax)
			}
			if got < s3MinPartSize {
				t.Errorf("s3PartSizeFor(%d) = %d, violates S3 minimum part size %d", tt.size, got, s3MinPartSize)
			}
			if tt.size > 0 && tt.maxParts > 0 {
				parts := (tt.size + got - 1) / got
				if parts > tt.maxParts {
					t.Errorf("s3PartSizeFor(%d) = %d, results in %d parts (max %d)", tt.size, got, parts, tt.maxParts)
				}
			}
		})
	}
}

// TestS3Storage_UploadConcurrency verifies that the uploadConcurrency field is
// propagated from S3Config to the S3Storage struct during construction.
func TestS3Storage_UploadConcurrency(t *testing.T) {
	tests := []struct {
		name            string
		cfgConcurrency  int
		wantConcurrency int
		expectDefault   bool
	}{
		{
			name:           "zero uses default",
			cfgConcurrency: 0,
			expectDefault:  true,
		},
		{
			name:            "positive value is stored",
			cfgConcurrency:  10,
			wantConcurrency: 10,
		},
		{
			name:            "one is stored",
			cfgConcurrency:  1,
			wantConcurrency: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &appconfig.S3Config{
				Region:            "us-east-1",
				Bucket:            "test-bucket",
				AuthType:          appconfig.S3AuthAccessKey,
				UploadConcurrency: tt.cfgConcurrency,
			}
			s := NewS3Storage(cfg).(*S3Storage)

			if tt.expectDefault {
				if s.uploadConcurrency != 0 {
					t.Errorf("uploadConcurrency = %d, want 0 (uses runtime default)", s.uploadConcurrency)
				}
			} else {
				if s.uploadConcurrency != tt.wantConcurrency {
					t.Errorf("uploadConcurrency = %d, want %d", s.uploadConcurrency, tt.wantConcurrency)
				}
			}
		})
	}
}
