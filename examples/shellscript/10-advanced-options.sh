#!/usr/bin/env bash
# 10-advanced-options.sh — Checksum, resume, bandwidth limit, retry, verbose
#
# Beginner scenario: learn the reliability and performance flags.
# All examples use local→local so you can run them without any credentials.
#
# Flags demonstrated:
#   NIXCOPY_VERIFY_CHECKSUM=true   SHA-256 integrity check (env var, not a CLI flag)
#   --resume             pick up interrupted transfers from where they stopped
#   --skip-existing      skip files already present at the destination
#   --bandwidth-limit    cap transfer speed in bytes/second
#   --retry-attempts     how many times to retry a failed transfer
#   --buffer-size        I/O buffer per worker in bytes
#   --concurrent-files   number of parallel workers
#   --verbose / -v       detailed debug output
#
# Usage:
#   bash 10-advanced-options.sh

set -euo pipefail

SRC="/tmp/nixcopy-adv-src"
DST="/tmp/nixcopy-adv-dst"

# ── Setup ──────────────────────────────────────────────────────────────────────
mkdir -p "$SRC" "$DST"
for i in $(seq 1 5); do
  dd if=/dev/urandom bs=1M count=2 2>/dev/null > "$SRC/file-$i.bin"
done
echo "Created 5 × 2 MB sample files in $SRC"

# ── Example 1: Verify integrity with SHA-256 checksum ─────────────────────────
# verify_checksum is set via the NIXCOPY_VERIFY_CHECKSUM env var (no CLI flag for it).
echo ""
echo "=== 1. Transfer with checksum verification ==="
NIXCOPY_VERIFY_CHECKSUM=true nixcopy transfer \
  --source-type local --source "$SRC/file-1.bin" \
  --dest-type local   --dest   "$DST/file-1.bin"
echo "Checksum passed — file arrived intact."

# ── Example 2: Resume an interrupted transfer ─────────────────────────────────
echo ""
echo "=== 2. Resume a partial transfer ==="
# Simulate a partial file at the destination
truncate -s 512K "$DST/file-2.bin"
echo "Simulated partial file: $(du -sh "$DST/file-2.bin" | cut -f1) (full = 2 MB)"

NIXCOPY_VERIFY_CHECKSUM=true nixcopy transfer \
  --source-type local --source "$SRC/file-2.bin" \
  --dest-type local   --dest   "$DST/file-2.bin" \
  --resume
echo "Resume complete. Final size: $(du -sh "$DST/file-2.bin" | cut -f1)"

# ── Example 3: Skip files that already exist ──────────────────────────────────
echo ""
echo "=== 3. Skip existing files (incremental sync) ==="
# Copy file-1 first so it already exists at the destination
cp "$SRC/file-1.bin" "$DST/file-1.bin"

nixcopy transfer \
  --source-type local --source "$SRC/*.bin" \
  --dest-type local   --dest   "$DST/" \
  --skip-existing
echo "Only new files were transferred; file-1.bin was skipped."

# ── Example 4: Bandwidth limiting ─────────────────────────────────────────────
echo ""
echo "=== 4. Transfer capped at 4 MB/s (useful on shared network links) ==="
# 4 MB/s = 4194304 bytes/s
nixcopy transfer \
  --source-type local --source "$SRC/file-3.bin" \
  --dest-type local   --dest   "$DST/file-3-slow.bin" \
  --bandwidth-limit 4194304
echo "Bandwidth-limited transfer complete."

# ── Example 5: Retry on failure ───────────────────────────────────────────────
echo ""
echo "=== 5. Retry up to 3 times with 5-second back-off ==="
nixcopy transfer \
  --source-type local --source "$SRC/file-4.bin" \
  --dest-type local   --dest   "$DST/file-4.bin" \
  --retry-attempts 3 \
echo "Transfer with retry complete."

# ── Example 6: Parallel + buffer tuning ──────────────────────────────────────
echo ""
echo "=== 6. Four concurrent workers with a 32 MB buffer each ==="
# Memory used ≈ 4 × 32 MB = 128 MB
nixcopy transfer \
  --source-type local --source "$SRC/*.bin" \
  --dest-type local   --dest   "$DST/parallel/" \
  --concurrent-files 4 \
  --buffer-size 33554432
echo "Parallel transfer complete."
ls -lh "$DST/parallel/"

# ── Example 7: Verbose output ─────────────────────────────────────────────────
echo ""
echo "=== 7. Verbose mode — shows resolved config and per-file progress ==="
nixcopy transfer \
  --source-type local --source "$SRC/file-5.bin" \
  --dest-type local   --dest   "$DST/file-5-verbose.bin" \
  --verbose
echo "Verbose transfer complete."

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf "$SRC" "$DST"
echo ""
echo "All advanced examples finished successfully."
