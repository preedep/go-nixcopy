#!/usr/bin/env bash
# 02-local-to-sftp.sh — Upload local files to an SFTP server
#
# Beginner scenario: push a file from your laptop/server to a remote SFTP host.
# Supports password auth or SSH private key auth.
#
# Prerequisites:
#   nixcopy installed
#   An SFTP server you can reach (host, port, user, password or key)
#
# Usage:
#   SFTP_HOST=sftp.example.com SFTP_USER=alice SFTP_PASS=secret \
#     bash 02-local-to-sftp.sh

set -euo pipefail

# ── Configuration — override with environment variables ────────────────────────
SFTP_HOST="${SFTP_HOST:-sftp.example.com}"
SFTP_PORT="${SFTP_PORT:-22}"
SFTP_USER="${SFTP_USER:-testuser}"
SFTP_PASS="${SFTP_PASS:-}"                       # leave empty to use a key
SFTP_KEY="${SFTP_KEY:-${HOME}/.ssh/id_rsa}"      # used only when SFTP_PASS is empty

LOCAL_FILE="/tmp/nixcopy-sftp-demo/report.pdf"
REMOTE_DIR="/upload/"                            # trailing slash = place file in dir

# ── Setup: create a dummy file ─────────────────────────────────────────────────
mkdir -p "$(dirname "$LOCAL_FILE")"
dd if=/dev/urandom bs=1K count=64 2>/dev/null | base64 > "$LOCAL_FILE"
echo "Created sample file: $LOCAL_FILE ($(du -sh "$LOCAL_FILE" | cut -f1))"

# ── Example 1: Upload with password auth ──────────────────────────────────────
if [[ -n "$SFTP_PASS" ]]; then
  echo ""
  echo "=== Upload with password auth ==="
  nixcopy transfer \
    --source-type local \
    --source "$LOCAL_FILE" \
    --dest-type sftp \
    --dest-host "$SFTP_HOST" \
    --dest-port "$SFTP_PORT" \
    --dest-username "$SFTP_USER" \
    --dest-password "$SFTP_PASS" \
    --dest "$REMOTE_DIR"
  echo "Upload complete."
fi

# ── Example 2: Upload with SSH private key ────────────────────────────────────
if [[ -z "$SFTP_PASS" && -f "$SFTP_KEY" ]]; then
  echo ""
  echo "=== Upload with SSH private key ==="
  nixcopy transfer \
    --source-type local \
    --source "$LOCAL_FILE" \
    --dest-type sftp \
    --dest-host "$SFTP_HOST" \
    --dest-port "$SFTP_PORT" \
    --dest-username "$SFTP_USER" \
    --dest-private-key "$SFTP_KEY" \
    --dest "$REMOTE_DIR"
  echo "Upload complete."
fi

# ── Example 3: Upload multiple files (glob) ────────────────────────────────────
echo ""
echo "=== Upload all PDF files in a directory ==="

for i in 1 2 3; do
  cp "$LOCAL_FILE" "/tmp/nixcopy-sftp-demo/report-$i.pdf"
done

# Choose the auth flag dynamically
if [[ -n "$SFTP_PASS" ]]; then
  AUTH_FLAGS="--dest-password $SFTP_PASS"
else
  AUTH_FLAGS="--dest-private-key $SFTP_KEY"
fi

# shellcheck disable=SC2086
nixcopy transfer \
  --source-type local \
  --source "/tmp/nixcopy-sftp-demo/*.pdf" \
  --dest-type sftp \
  --dest-host "$SFTP_HOST" \
  --dest-port "$SFTP_PORT" \
  --dest-username "$SFTP_USER" \
  $AUTH_FLAGS \
  --dest "/upload/" \
  --concurrent-files 3

echo "Batch upload complete."

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf /tmp/nixcopy-sftp-demo
