#!/usr/bin/env bash
# 05-s3-to-local.sh — Download files from AWS S3 to local disk
#
# Beginner scenario: pull an object (or a set of objects) from S3.
# Useful for restoring backups, fetching ML datasets, or archiving logs.
#
# Run 04-local-to-s3.sh first to seed the bucket with test files.
#
# Usage:
#   AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=secret... S3_BUCKET=my-bucket \
#     bash 05-s3-to-local.sh
#
# Usage against MinIO (local emulator):
#   AWS_ACCESS_KEY=minioadmin AWS_SECRET_KEY=minioadmin S3_BUCKET=test-bucket \
#   NIXCOPY_SOURCE_ENDPOINT=http://localhost:9000 NIXCOPY_SOURCE_USE_PATH_STYLE=true \
#     bash 05-s3-to-local.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
S3_BUCKET="${S3_BUCKET:-my-backup-bucket}"
S3_REGION="${S3_REGION:-ap-southeast-1}"
AWS_ACCESS_KEY="${AWS_ACCESS_KEY:-}"
AWS_SECRET_KEY="${AWS_SECRET_KEY:-}"

# S3_PREFIX must match what 04-local-to-s3.sh uploaded under
S3_PREFIX="uploads/2024/"

LOCAL_DIR="/tmp/nixcopy-s3-download/"
mkdir -p "$LOCAL_DIR"

# ── Example 1: Download a single file ─────────────────────────────────────────
echo "=== Download a single object ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "${S3_PREFIX}data.csv" \
  --dest-type local \
  --dest "$LOCAL_DIR"

echo "Downloaded:"
ls -lh "$LOCAL_DIR"

# ── Example 2: Download again to the same path (idempotent) ──────────────────
echo ""
echo "=== Download same file again (skip-existing) ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "${S3_PREFIX}data.csv" \
  --dest-type local \
  --dest "${LOCAL_DIR}data.csv" \
  --skip-existing

echo "Skip-existing download complete."

# ── Example 3: Download multiple objects (prefix/glob) ────────────────────────
echo ""
echo "=== Download all objects under ${S3_PREFIX} ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "${S3_PREFIX}**" \
  --dest-type local \
  --dest "${LOCAL_DIR}archive/" \
  --concurrent-files 4

echo "Batch download complete."
ls -lh "${LOCAL_DIR}archive/" 2>/dev/null || true

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf /tmp/nixcopy-s3-download
