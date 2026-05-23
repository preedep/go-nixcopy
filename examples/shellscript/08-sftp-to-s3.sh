#!/usr/bin/env bash
# 08-sftp-to-s3.sh — Transfer files directly from SFTP to S3 (no local disk)
#
# Beginner scenario: migrate or sync files from a legacy SFTP server to AWS S3
# without touching local disk. nixcopy streams the data directly.
#
# Run 02-local-to-sftp.sh first to seed the SFTP server with test files.
#
# Prerequisites:
#   nixcopy installed
#   SFTP server credentials + S3 bucket with write access
#
# Usage:
#   SFTP_HOST=sftp.example.com SFTP_USER=alice SFTP_PASS=secret \
#   AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=secret... S3_BUCKET=my-bucket \
#     bash 08-sftp-to-s3.sh
#
# Usage against local emulators (SFTP container + MinIO):
#   SFTP_HOST=localhost SFTP_PORT=2222 SFTP_USER=testuser SFTP_PASS=testpass \
#   AWS_ACCESS_KEY=minioadmin AWS_SECRET_KEY=minioadmin S3_BUCKET=test-bucket \
#   NIXCOPY_DEST_ENDPOINT=http://localhost:9000 NIXCOPY_DEST_USE_PATH_STYLE=true \
#     bash 08-sftp-to-s3.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
SFTP_HOST="${SFTP_HOST:-sftp.example.com}"
SFTP_PORT="${SFTP_PORT:-22}"
SFTP_USER="${SFTP_USER:-alice}"
SFTP_PASS="${SFTP_PASS:-}"
SFTP_KEY="${SFTP_KEY:-${HOME}/.ssh/id_rsa}"

S3_BUCKET="${S3_BUCKET:-my-backup-bucket}"
S3_REGION="${S3_REGION:-ap-southeast-1}"
AWS_ACCESS_KEY="${AWS_ACCESS_KEY:-}"
AWS_SECRET_KEY="${AWS_SECRET_KEY:-}"

# Build SFTP auth flags
if [[ -n "$SFTP_PASS" ]]; then
  SFTP_AUTH="--source-password $SFTP_PASS"
else
  SFTP_AUTH="--source-private-key $SFTP_KEY"
fi

# ── Example 1: Transfer a single file SFTP → S3 ──────────────────────────────
echo "=== Transfer single file: SFTP → S3 ==="
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $SFTP_AUTH \
  --source "/upload/report.pdf" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "archive/report.pdf"
echo "Transfer complete."

# ── Example 2: Batch transfer — all PDFs from SFTP to S3 ─────────────────────
echo ""
echo "=== Batch transfer: all PDFs from /upload/ → S3 ==="
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $SFTP_AUTH \
  --source "/upload/*.pdf" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "archive/batch/" \
  --concurrent-files 4 \
  --retry-attempts 3
echo "Batch transfer complete."

# ── Example 3: Nightly backup with bandwidth cap ──────────────────────────────
echo ""
echo "=== Nightly backup: SFTP → S3 with retry and 20 MB/s cap ==="
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $SFTP_AUTH \
  --source "/upload/report.pdf" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "backups/nightly/report.pdf" \
  --retry-attempts 5 \
  --bandwidth-limit 20971520 \
  --resume
echo "Nightly backup complete."
