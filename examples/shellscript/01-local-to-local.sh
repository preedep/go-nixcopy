#!/usr/bin/env bash
# 01-local-to-local.sh — Copy files between local directories
#
# Beginner scenario: copy a single file or a folder of files from one
# local directory to another. No credentials required.
#
# Prerequisites:
#   nixcopy installed (make install  or  go install ./cmd/nixcopy)
#
# Usage:
#   bash 01-local-to-local.sh

set -euo pipefail

# ── Configuration ─────────────────────────────────────────────────────────────
SOURCE_FILE="/tmp/nixcopy-demo/input/hello.txt"
DEST_DIR="/tmp/nixcopy-demo/output/"   # trailing slash = place file inside dir

# ── Setup: create a sample file to copy ───────────────────────────────────────
mkdir -p "$(dirname "$SOURCE_FILE")" "$DEST_DIR"
echo "Hello from nixcopy!" > "$SOURCE_FILE"

echo "=== Example 1: Copy a single file ==="
nixcopy transfer \
  --source-type local \
  --source "$SOURCE_FILE" \
  --dest-type local \
  --dest "$DEST_DIR"

echo "Result:"
ls -lh "$DEST_DIR"

# ── Copy multiple files with a glob pattern ────────────────────────────────────
echo ""
echo "=== Example 2: Copy all .txt files using a glob pattern ==="

# Create a few more sample files
for i in 1 2 3; do
  echo "File $i" > "/tmp/nixcopy-demo/input/file$i.txt"
done

nixcopy transfer \
  --source-type local \
  --source "/tmp/nixcopy-demo/input/*.txt" \
  --dest-type local \
  --dest "/tmp/nixcopy-demo/output/"

echo "Result:"
ls -lh "$DEST_DIR"

# ── Verify checksum (via env var — not a CLI flag) ────────────────────────────
echo ""
echo "=== Example 3: Copy with SHA-256 checksum verification ==="
# verify_checksum is set in config.yaml or via the NIXCOPY_VERIFY_CHECKSUM env var.

NIXCOPY_VERIFY_CHECKSUM=true nixcopy transfer \
  --source-type local \
  --source "$SOURCE_FILE" \
  --dest-type local \
  --dest "/tmp/nixcopy-demo/output/hello-verified.txt"

echo "Checksum-verified copy complete."

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf /tmp/nixcopy-demo
echo ""
echo "Done. Temp files cleaned up."
