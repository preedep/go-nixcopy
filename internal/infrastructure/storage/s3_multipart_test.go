package storage

import "testing"

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
