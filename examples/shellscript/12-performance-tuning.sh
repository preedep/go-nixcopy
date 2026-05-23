#!/usr/bin/env bash
# 12-performance-tuning.sh — Tune nixcopy for maximum throughput
#
# Demonstrates --concurrent-files, --buffer-size, --upload-concurrency,
# --skip-existing for incremental runs, and bandwidth cap.
#
# Uses local storage so no credentials are needed.
# Run from the repo root or any directory with nixcopy in PATH.
#
# Flags demonstrated:
#   --concurrent-files    number of parallel file workers
#   --buffer-size         I/O buffer per worker in bytes (affects local/SFTP writes)
#   --upload-concurrency  concurrent part uploads per file for S3 / Azure Blob
#   --skip-existing       incremental sync — skip files already at destination
#   --bandwidth-limit     cap throughput (bytes/s) to avoid saturating a network link
#
# Memory rule of thumb:
#   peak_memory ≈ buffer_size × concurrent_files
#
# Usage:
#   bash 12-performance-tuning.sh

set -euo pipefail

SRC="/tmp/nixcopy-perf-src"
DST_SMALL="/tmp/nixcopy-perf-dst-small"
DST_LARGE="/tmp/nixcopy-perf-dst-large"
DST_INCR="/tmp/nixcopy-perf-dst-incr"

cleanup() {
  rm -rf "$SRC" "$DST_SMALL" "$DST_LARGE" "$DST_INCR"
}
trap cleanup EXIT

mkdir -p "$SRC" "$DST_SMALL" "$DST_LARGE" "$DST_INCR"

# ── Generate sample files ──────────────────────────────────────────────────────
echo "Generating sample files…"
for i in $(seq 1 16); do
  dd if=/dev/urandom bs=512K count=1 2>/dev/null > "$SRC/small-$i.bin"
done
for i in $(seq 1 4); do
  dd if=/dev/urandom bs=32M count=1 2>/dev/null > "$SRC/large-$i.bin"
done
echo "Created 16 × 512 KB + 4 × 32 MB files in $SRC"

# ── Scenario 1: Many small files — maximise concurrency, small buffer ──────────
echo ""
echo "=== 1. Small files: high concurrency, 4 MB buffer ==="
# 16 concurrent workers × 4 MB buffer ≈ 64 MB peak memory
nixcopy transfer \
  --source-type local --source "$SRC/small-*.bin" \
  --dest-type   local --dest   "$DST_SMALL/" \
  --concurrent-files 16 \
  --buffer-size 4194304
echo "Small-file transfer complete. Files in $DST_SMALL:"
ls -lh "$DST_SMALL" | tail -5

# ── Scenario 2: Large files — fewer workers, larger buffer, more parts ─────────
echo ""
echo "=== 2. Large files: low concurrency, 128 MB buffer, 10-part uploads ==="
# 2 concurrent workers × 128 MB buffer ≈ 256 MB peak memory.
# --upload-concurrency applies to S3/Blob multipart uploads.
# For local storage it is accepted but has no effect.
nixcopy transfer \
  --source-type local --source "$SRC/large-*.bin" \
  --dest-type   local --dest   "$DST_LARGE/" \
  --concurrent-files 2 \
  --buffer-size 134217728 \
  --upload-concurrency 10
echo "Large-file transfer complete. Files in $DST_LARGE:"
ls -lh "$DST_LARGE"

# ── Scenario 3: Incremental sync — skip unchanged files ───────────────────────
echo ""
echo "=== 3. Incremental sync: skip files already at destination ==="
# First pass: copy all small files.
nixcopy transfer \
  --source-type local --source "$SRC/small-*.bin" \
  --dest-type   local --dest   "$DST_INCR/" \
  --concurrent-files 8

# Second pass: only new/changed files are copied; existing ones are skipped.
# Add one new file to demonstrate what gets transferred.
dd if=/dev/urandom bs=512K count=1 2>/dev/null > "$SRC/small-new.bin"
nixcopy transfer \
  --source-type local --source "$SRC/small-*.bin" \
  --dest-type   local --dest   "$DST_INCR/" \
  --concurrent-files 8 \
  --skip-existing
echo "Second pass done — only small-new.bin was transferred."

# ── Notes for cloud backends ───────────────────────────────────────────────────
echo ""
echo "For S3 / Azure Blob, replace local flags with cloud-specific ones, e.g.:"
echo ""
echo "  nixcopy transfer \\"
echo "    --source-type local  --source /data/*.bin \\"
echo "    --dest-type   s3     --dest   s3-key-prefix/ \\"
echo "    --dest-region us-east-1 \\"
echo "    --dest-bucket my-bucket \\"
echo "    --concurrent-files 4 \\"
echo "    --buffer-size 67108864 \\"
echo "    --upload-concurrency 8"
echo ""
echo "All performance-tuning examples finished successfully."
