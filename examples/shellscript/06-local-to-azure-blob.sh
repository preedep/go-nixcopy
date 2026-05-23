#!/usr/bin/env bash
# 06-local-to-azure-blob.sh — Upload local files to Azure Blob Storage
#
# Beginner scenario: push files to an Azure Blob container.
# Covers three common auth methods:
#   1. connection_string — paste from Azure portal (easiest for beginners)
#   2. shared_key        — storage account name + key
#   3. managed_identity  — automatic on Azure VMs / AKS (no secrets needed)
#
# Prerequisites:
#   nixcopy installed
#   An Azure Blob Storage account and container
#
# Usage (connection_string):
#   BLOB_CONN="DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net" \
#   BLOB_CONTAINER=my-container \
#     bash 06-local-to-azure-blob.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
BLOB_ACCOUNT="${BLOB_ACCOUNT:-mystorageaccount}"
BLOB_CONTAINER="${BLOB_CONTAINER:-my-container}"
BLOB_CONN="${BLOB_CONN:-}"           # full connection string from Azure portal
BLOB_KEY="${BLOB_KEY:-}"             # storage account key (for shared_key auth)

LOCAL_FILE="/tmp/nixcopy-blob-demo/invoice.pdf"
BLOB_PREFIX="invoices/2024/"

# ── Setup ──────────────────────────────────────────────────────────────────────
mkdir -p "$(dirname "$LOCAL_FILE")"
echo "Sample invoice content" > "$LOCAL_FILE"

# ── Example 1: Upload with connection string (simplest) ───────────────────────
if [[ -n "$BLOB_CONN" ]]; then
  echo "=== Upload with connection_string auth ==="
  NIXCOPY_DEST_CONNECTION_STRING="$BLOB_CONN" nixcopy transfer \
    --source-type local \
    --source "$LOCAL_FILE" \
    --dest-type blob \
    --dest-container "$BLOB_CONTAINER" \
    --dest-auth-type connection_string \
    --dest "${BLOB_PREFIX}invoice.pdf"
  echo "Upload complete."
fi

# ── Example 2: Upload with shared_key (account name + key) ───────────────────
if [[ -n "$BLOB_KEY" ]]; then
  echo ""
  echo "=== Upload with shared_key auth ==="
  nixcopy transfer \
    --source-type local \
    --source "$LOCAL_FILE" \
    --dest-type blob \
    --dest-account-name "$BLOB_ACCOUNT" \
    --dest-container "$BLOB_CONTAINER" \
    --dest-auth-type shared_key \
    --dest-account-key "$BLOB_KEY" \
    --dest "${BLOB_PREFIX}invoice-key.pdf"
  echo "Upload complete."
fi

# ── Example 3: Upload on Azure VM using Managed Identity ─────────────────────
echo ""
echo "=== Upload using Managed Identity (Azure VM / AKS) ==="
echo "# nixcopy transfer \\"
echo "#   --source-type local --source $LOCAL_FILE \\"
echo "#   --dest-type blob \\"
echo "#   --dest-account-name $BLOB_ACCOUNT --dest-container $BLOB_CONTAINER \\"
echo "#   --dest-auth-type managed_identity \\"
echo "#   --dest ${BLOB_PREFIX}invoice-mi.pdf"
echo "(Skipped — requires an Azure VM/AKS environment)"

# ── Example 4: Batch upload with concurrency ─────────────────────────────────
if [[ -n "$BLOB_CONN" ]]; then
  echo ""
  echo "=== Batch upload: 5 files with 3 concurrent workers ==="
  for i in 1 2 3 4 5; do
    cp "$LOCAL_FILE" "/tmp/nixcopy-blob-demo/invoice-$i.pdf"
  done

  NIXCOPY_DEST_CONNECTION_STRING="$BLOB_CONN" nixcopy transfer \
    --source-type local \
    --source "/tmp/nixcopy-blob-demo/*.pdf" \
    --dest-type blob \
    --dest-container "$BLOB_CONTAINER" \
    --dest-auth-type connection_string \
    --dest "$BLOB_PREFIX" \
    --concurrent-files 3
  echo "Batch upload complete."
fi

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf /tmp/nixcopy-blob-demo
