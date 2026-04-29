package storage

import "testing"

func TestBlobBlockSizeFor(t *testing.T) {
	tests := []struct {
		name      string
		size      int64
		wantMin   int64
		wantMax   int64
		maxBlocks int64
	}{
		{
			name:    "unknown size returns default",
			size:    0,
			wantMin: blobDefaultBlockSize,
			wantMax: blobDefaultBlockSize,
		},
		{
			name:    "small file uses default",
			size:    1 * 1024 * 1024, // 1 MiB
			wantMin: blobDefaultBlockSize,
			wantMax: blobDefaultBlockSize,
		},
		{
			name:    "100 MB file uses default",
			size:    100 * 1024 * 1024,
			wantMin: blobDefaultBlockSize,
			wantMax: blobDefaultBlockSize,
		},
		{
			name:      "800 GB file stays within 50000 blocks",
			size:      800 * 1024 * 1024 * 1024,
			wantMin:   blobMinBlockSize,
			wantMax:   1<<62 - 1,
			maxBlocks: blobMaxBlocks,
		},
		{
			name:      "190 TiB file stays within 50000 blocks",
			size:      190 * 1024 * 1024 * 1024 * 1024,
			wantMin:   blobMinBlockSize,
			wantMax:   1<<62 - 1,
			maxBlocks: blobMaxBlocks,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blobBlockSizeFor(tt.size)

			if got < tt.wantMin {
				t.Errorf("blobBlockSizeFor(%d) = %d, want >= %d", tt.size, got, tt.wantMin)
			}
			if got > tt.wantMax {
				t.Errorf("blobBlockSizeFor(%d) = %d, want <= %d", tt.size, got, tt.wantMax)
			}
			if got < blobMinBlockSize {
				t.Errorf("blobBlockSizeFor(%d) = %d, violates Azure minimum block size %d", tt.size, got, blobMinBlockSize)
			}
			if tt.size > 0 && tt.maxBlocks > 0 {
				blocks := (tt.size + got - 1) / got
				if blocks > tt.maxBlocks {
					t.Errorf("blobBlockSizeFor(%d) = %d, results in %d blocks (max %d)", tt.size, got, blocks, tt.maxBlocks)
				}
			}
		})
	}
}
