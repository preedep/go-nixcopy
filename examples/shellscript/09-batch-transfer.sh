#!/usr/bin/env bash
# 09-batch-transfer.sh — Transfer many files in parallel with glob patterns
#
# Beginner scenario: move a large number of files efficiently.
# Demonstrates --sources (multiple explicit patterns), concurrency tuning,
# skip-existing, and preview via `nixcopy list`.
#
# Uses S3 as source and Azure Blob as destination.
# Run 04-local-to-s3.sh first to seed the source S3 bucket.
#
# Prerequisites:
#   nixcopy installed
#   Source S3 bucket + destination Azure Blob credentials
#
# Usage:
#   AWS_ACCESS_KEY=... AWS_SECRET_KEY=... S3_BUCKET=src-bucket \
#   BLOB_CONN="..." BLOB_CONTAINER=dest-container \
#     bash 09-batch-transfer.sh
#
# Usage against local emulators (MinIO + Azurite):
#   AWS_ACCESS_KEY=minioadmin AWS_SECRET_KEY=minioadmin S3_BUCKET=test-bucket \
#   NIXCOPY_SOURCE_ENDPOINT=http://localhost:9000 NIXCOPY_SOURCE_USE_PATH_STYLE=true \
#   BLOB_CONN="DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;..." \
#   BLOB_CONTAINER=test-container \
#     bash 09-batch-transfer.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
S3_REGION="${S3_REGION:-ap-southeast-1}"
S3_BUCKET="${S3_BUCKET:-source-bucket}"
AWS_ACCESS_KEY="${AWS_ACCESS_KEY:-}"
AWS_SECRET_KEY="${AWS_SECRET_KEY:-}"

BLOB_CONTAINER="${BLOB_CONTAINER:-dest-container}"
BLOB_CONN="${BLOB_CONN:-}"

# S3_PREFIX must match what 04-local-to-s3.sh uploaded under
S3_PREFIX="uploads/2024/"

# ── Tip: Preview files before transferring ────────────────────────────────────
# nixcopy list reads backend config from a config file (-c). To preview files:
#
#   nixcopy list -c config.yaml -p "uploads/2024/**" --source
#
# The list subcommand does not accept inline --source-type / --source-bucket
# flags — use a config file or the transfer subcommand with --dry-run (if added).
echo "=== Skipping list preview (requires -c config.yaml) ==="
echo "    Tip: nixcopy list -c config.yaml -p '${S3_PREFIX}**' --source"
echo ""

# ── Example 1: Single glob pattern ────────────────────────────────────────────
echo "=== Transfer all files under ${S3_PREFIX} (single glob) ==="
NIXCOPY_DEST_CONNECTION_STRING="$BLOB_CONN" nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "${S3_PREFIX}**" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest "archive/${S3_PREFIX}" \
  --concurrent-files 4
echo "Done."

# ── Example 2: Multiple patterns via --sources ────────────────────────────────
echo ""
echo "=== Transfer CSVs and PDFs simultaneously (--sources) ==="
NIXCOPY_DEST_CONNECTION_STRING="$BLOB_CONN" nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --sources "${S3_PREFIX}*.csv,${S3_PREFIX}*.pdf" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest "archive/mixed/" \
  --concurrent-files 4 \
  --retry-attempts 3
echo "Done."

# ── Example 3: Skip files already at destination (incremental sync) ───────────
echo ""
echo "=== Incremental sync: skip files already present at destination ==="
NIXCOPY_DEST_CONNECTION_STRING="$BLOB_CONN" nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "${S3_PREFIX}**" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest "archive/${S3_PREFIX}" \
  --concurrent-files 4 \
  --skip-existing
echo "Incremental sync complete."

# ── Concurrency tuning guide ─────────────────────────────────────────────────
cat <<'EOF'

Concurrency tuning:
  --concurrent-files 2-4   Large files (> 100 MB) — avoid saturating memory
  --concurrent-files 8-16  Small files (< 10 MB) — maximise throughput
  --buffer-size 67108864   64 MB buffer per worker (adjust with concurrency)

  Memory used ≈ buffer_size × concurrent_files

Parallel directory listing (automatic):
  Recursive patterns like "uploads/2024/**" traverse subdirectories using a
  parallel worker pool (8 concurrent Storage.List calls by default). Deep
  bucket hierarchies are scanned significantly faster than serial traversal
  with no extra flags required.
EOF
