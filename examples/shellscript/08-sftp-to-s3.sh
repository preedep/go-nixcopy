#!/usr/bin/env bash
# 08-sftp-to-s3.sh — Transfer files directly from SFTP to S3 (no local disk)
#
# Beginner scenario: migrate or sync files from a legacy SFTP server to AWS S3
# without touching local disk. nixcopy streams the data directly.
#
# Prerequisites:
#   nixcopy installed
#   SFTP server credentials + S3 bucket with write access
#
# Usage:
#   SFTP_HOST=sftp.example.com SFTP_USER=alice SFTP_PASS=secret \
#   AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=secret... S3_BUCKET=my-bucket \
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

# ── Example 1: Move a single file from SFTP to S3 ────────────────────────────
echo "=== Transfer single file: SFTP → S3 ==="
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $SFTP_AUTH \
  --source "/upload/daily-report.csv" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "archive/daily-report.csv"
echo "Transfer complete."

# ── Example 2: Batch transfer — all CSVs from SFTP to S3 ─────────────────────
echo ""
echo "=== Batch transfer: all CSVs from /upload/reports/ → S3 ==="
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $SFTP_AUTH \
  --source "/upload/reports/*.csv" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "archive/reports/" \
  --concurrent-files 4 \
  --retry-attempts 3 \
echo "Batch transfer complete."

# ── Example 3: Nightly cron-style transfer with retry and bandwidth cap ────────
echo ""
echo "=== Nightly backup: SFTP → S3 with retry and 20 MB/s cap ==="
TODAY=$(date +%Y-%m-%d)
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $SFTP_AUTH \
  --source "/upload/backup-${TODAY}.tar.gz" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "backups/${TODAY}/backup.tar.gz" \
  --retry-attempts 5 \
  --bandwidth-limit 20971520 \
  --resume \
echo "Nightly backup complete."
