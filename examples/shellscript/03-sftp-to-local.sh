#!/usr/bin/env bash
# 03-sftp-to-local.sh — Download files from an SFTP server to local disk
#
# Beginner scenario: pull files from a remote SFTP host to your local machine.
# Useful for scheduled backups, data ingestion, or syncing configs.
#
# Prerequisites:
#   nixcopy installed
#   An SFTP server with files you want to download
#
# Usage:
#   SFTP_HOST=sftp.example.com SFTP_USER=alice SFTP_PASS=secret \
#     bash 03-sftp-to-local.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
SFTP_HOST="${SFTP_HOST:-sftp.example.com}"
SFTP_PORT="${SFTP_PORT:-22}"
SFTP_USER="${SFTP_USER:-testuser}"
SFTP_PASS="${SFTP_PASS:-}"
SFTP_KEY="${SFTP_KEY:-${HOME}/.ssh/id_rsa}"

REMOTE_FILE="/upload/report.pdf"
LOCAL_DIR="/tmp/nixcopy-download/"

mkdir -p "$LOCAL_DIR"

# Build auth flags
if [[ -n "$SFTP_PASS" ]]; then
  AUTH_FLAGS="--source-password $SFTP_PASS"
else
  AUTH_FLAGS="--source-private-key $SFTP_KEY"
fi

# ── Example 1: Download a single file ─────────────────────────────────────────
echo "=== Download a single file ==="
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $AUTH_FLAGS \
  --source "$REMOTE_FILE" \
  --dest-type local \
  --dest "$LOCAL_DIR"

echo "Downloaded to $LOCAL_DIR:"
ls -lh "$LOCAL_DIR"

# ── Example 2: Download with resume (safe for large files over slow links) ─────
echo ""
echo "=== Download a large file with resume support ==="
# If the transfer is interrupted, nixcopy picks up where it left off.
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $AUTH_FLAGS \
  --source "/upload/archive.tar.gz" \
  --dest-type local \
  --dest "/tmp/nixcopy-download/archive.tar.gz" \
  --resume \

echo "Resumable download complete."

# ── Example 3: Download multiple files matching a pattern ─────────────────────
echo ""
echo "=== Download all .log files from /upload/logs/ ==="
# shellcheck disable=SC2086
nixcopy transfer \
  --source-type sftp \
  --source-host "$SFTP_HOST" \
  --source-port "$SFTP_PORT" \
  --source-username "$SFTP_USER" \
  $AUTH_FLAGS \
  --source "/upload/logs/*.log" \
  --dest-type local \
  --dest "/tmp/nixcopy-download/logs/" \
  --concurrent-files 4

echo "Batch download complete."
ls -lh /tmp/nixcopy-download/logs/ 2>/dev/null || true

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf /tmp/nixcopy-download
