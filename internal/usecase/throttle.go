package usecase

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ThrottledReader wraps an io.Reader and limits read throughput to bytesPerSec.
// It uses a cumulative-bytes approach: tracks bytes read since the epoch and
// sleeps when the actual rate exceeds the target. Context cancellation aborts the sleep.
// The measurement window is normalised every bytesPerSec bytes to prevent overflow.
type ThrottledReader struct {
	r           io.Reader
	ctx         context.Context
	bytesPerSec int64
	epoch       time.Time
	epochBytes  int64
}

func newThrottledReader(ctx context.Context, r io.Reader, bytesPerSec int64) io.Reader {
	if bytesPerSec <= 0 {
		return r
	}
	return &ThrottledReader{
		r:           r,
		ctx:         ctx,
		bytesPerSec: bytesPerSec,
		epoch:       time.Now(),
	}
}

func (t *ThrottledReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	if n <= 0 {
		return n, err
	}

	t.epochBytes += int64(n)

	// How long should it take to send epochBytes at the target rate?
	expected := time.Duration(float64(t.epochBytes) / float64(t.bytesPerSec) * float64(time.Second))
	if delay := expected - time.Since(t.epoch); delay > 0 {
		select {
		case <-t.ctx.Done():
			return n, t.ctx.Err()
		case <-time.After(delay):
		}
	}

	// Advance the epoch window to prevent int64 overflow on very long transfers.
	if t.epochBytes >= t.bytesPerSec {
		t.epoch = t.epoch.Add(time.Second)
		t.epochBytes -= t.bytesPerSec
	}

	return n, err
}

// ParseBandwidth converts a human-readable bandwidth string to bytes per second.
// Accepted formats: raw integer bytes ("10485760"), or a number with a unit suffix
// ("10MB", "10MiB", "10M", "1GB", "512KB"). Case-insensitive. 0 means unlimited.
func ParseBandwidth(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0, nil
	}

	upper := strings.ToUpper(s)

	// Check longer suffixes first to avoid "M" matching before "MB".
	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"GIB", 1024 * 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"MIB", 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KIB", 1024},
		{"KB", 1024},
		{"G", 1024 * 1024 * 1024},
		{"M", 1024 * 1024},
		{"K", 1024},
	}

	for _, pair := range suffixes {
		if strings.HasSuffix(upper, pair.suffix) {
			numStr := strings.TrimSpace(s[:len(s)-len(pair.suffix)])
			f, err := strconv.ParseFloat(numStr, 64)
			if err != nil || f < 0 {
				return 0, fmt.Errorf("invalid bandwidth %q", s)
			}
			return int64(f * float64(pair.mult)), nil
		}
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid bandwidth %q: use bytes (10485760) or a unit suffix (10MB, 1GB, 512KB)", s)
	}
	return n, nil
}
