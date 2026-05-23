#!/usr/bin/env bash
# 09-batch-transfer.sh — Transfer many files in parallel with glob patterns
#
# Beginner scenario: move a large number of files efficiently.
# Demonstrates --sources (multiple explicit patterns), concurrency tuning,
# skip-existing, and dry-run preview via `nixcopy list`.
#
# This example uses S3 as source and Azure Blob as destination, but the
# same flags work for any combination of backends.
#
# Prerequisites:
#   nixcopy installed
#   Source S3 bucket + destination Azure Blob credentials
#
# Usage:
#   AWS_ACCESS_KEY=... AWS_SECRET_KEY=... S3_BUCKET=src-bucket \
#   BLOB_CONN="..." BLOB_CONTAINER=dest-container \
#     bash 09-batch-transfer.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
S3_REGION="${S3_REGION:-ap-southeast-1}"
S3_BUCKET="${S3_BUCKET:-source-bucket}"
AWS_ACCESS_KEY="${AWS_ACCESS_KEY:-}"
AWS_SECRET_KEY="${AWS_SECRET_KEY:-}"

BLOB_CONTAINER="${BLOB_CONTAINER:-dest-container}"
BLOB_CONN="${BLOB_CONN:-}"

# ── Tip: Preview files before transferring ────────────────────────────────────
echo "=== Preview: list files that match the pattern ==="
echo "    Use 'nixcopy list' to verify what will be transferred."
echo ""
nixcopy list \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "reports/2024/**" \
  --source
echo ""

# ── Example 1: Single glob pattern — all 2024 reports ────────────────────────
echo "=== Transfer all 2024 reports (single glob) ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "reports/2024/**" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest-connection-string "$BLOB_CONN" \
  --dest "archive/reports/2024/" \
  --concurrent-files 8
echo "Done."

# ── Example 2: Multiple patterns via --sources ────────────────────────────────
echo ""
echo "=== Transfer CSVs and PDFs simultaneously (--sources) ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --sources "reports/2024/**/*.csv,reports/2024/**/*.pdf" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest-connection-string "$BLOB_CONN" \
  --dest "archive/mixed/" \
  --concurrent-files 6 \
  --retry-attempts 3
echo "Done."

# ── Example 3: Skip files that already exist at the destination ───────────────
echo ""
echo "=== Incremental sync: skip files already present at destination ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "reports/2024/**" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest-connection-string "$BLOB_CONN" \
  --dest "archive/reports/2024/" \
  --concurrent-files 8 \
  --skip-existing
echo "Incremental sync complete."

# ── Concurrency tuning guide ─────────────────────────────────────────────────
cat <<'EOF'

Concurrency tuning:
  --concurrent-files 2-4   Large files (> 100 MB) — avoid saturating memory
  --concurrent-files 8-16  Small files (< 10 MB) — maximise throughput
  --buffer-size 67108864   64 MB buffer per worker (adjust with concurrency)

  Memory used ≈ buffer_size × concurrent_files
EOF
