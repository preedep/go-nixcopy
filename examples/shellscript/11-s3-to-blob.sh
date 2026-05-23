#!/usr/bin/env bash
# 11-s3-to-blob.sh — Transfer files between AWS S3 and Azure Blob Storage
#
# Covers two directions:
#   Part A: S3 → Azure Blob  (migrate from AWS to Azure)
#   Part B: Azure Blob → S3  (migrate from Azure to AWS, or dual-cloud backup)
#
# nixcopy streams data directly between the two clouds — no local disk used.
#
# Prerequisites:
#   nixcopy installed
#   AWS S3 credentials (access_key, or IAM role on EC2)
#   Azure Blob credentials (connection_string or shared_key)
#
# Usage:
#   # S3 → Blob
#   AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=... S3_BUCKET=src-bucket \
#   BLOB_CONN="DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net" \
#   BLOB_CONTAINER=dest-container \
#     bash 11-s3-to-blob.sh
#
#   # Blob → S3 (swap the env-var roles)
#   BLOB_CONN="..." BLOB_CONTAINER=src-container \
#   AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=... S3_BUCKET=dest-bucket \
#     bash 11-s3-to-blob.sh

set -euo pipefail

# ── AWS S3 configuration ───────────────────────────────────────────────────────
S3_BUCKET="${S3_BUCKET:-my-s3-bucket}"
S3_REGION="${S3_REGION:-ap-southeast-1}"
AWS_ACCESS_KEY="${AWS_ACCESS_KEY:-}"
AWS_SECRET_KEY="${AWS_SECRET_KEY:-}"

# ── Azure Blob configuration ───────────────────────────────────────────────────
BLOB_ACCOUNT="${BLOB_ACCOUNT:-mystorageaccount}"
BLOB_CONTAINER="${BLOB_CONTAINER:-my-container}"
BLOB_CONN="${BLOB_CONN:-}"    # connection string from Azure portal (easiest)
BLOB_KEY="${BLOB_KEY:-}"      # storage account key (alternative to conn string)

# ─────────────────────────────────────────────────────────────────────────────
# PART A: S3 → Azure Blob
# ─────────────────────────────────────────────────────────────────────────────

echo "════════════════════════════════════════"
echo " PART A: AWS S3  →  Azure Blob Storage"
echo "════════════════════════════════════════"

# ── A1: Single file ───────────────────────────────────────────────────────────
echo ""
echo "=== A1. Copy one object from S3 to Blob ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "data/report.csv" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest-connection-string "$BLOB_CONN" \
  --dest "migrated/report.csv"
echo "Done."

# ── A2: Batch — migrate an entire prefix ─────────────────────────────────────
echo ""
echo "=== A2. Migrate all objects under data/2024/ from S3 to Blob ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "data/2024/**" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest-connection-string "$BLOB_CONN" \
  --dest "migrated/2024/" \
  --concurrent-files 6 \
  --retry-attempts 3
echo "Batch migration complete."

# ── A3: Incremental sync (skip objects already copied) ────────────────────────
echo ""
echo "=== A3. Incremental sync: skip objects already present in Blob ==="
nixcopy transfer \
  --source-type s3 \
  --source-region "$S3_REGION" \
  --source-bucket "$S3_BUCKET" \
  --source-auth-type access_key \
  --source-access-key "$AWS_ACCESS_KEY" \
  --source-secret-key "$AWS_SECRET_KEY" \
  --source "data/2024/**" \
  --dest-type blob \
  --dest-container "$BLOB_CONTAINER" \
  --dest-auth-type connection_string \
  --dest-connection-string "$BLOB_CONN" \
  --dest "migrated/2024/" \
  --concurrent-files 6 \
  --skip-existing
echo "Incremental sync complete."

# ── A4: Using IAM role on EC2 and Managed Identity on Azure ──────────────────
echo ""
echo "=== A4. EC2 IAM role → Azure Managed Identity (cloud-native, no keys) ==="
echo "# nixcopy transfer \\"
echo "#   --source-type s3 \\"
echo "#   --source-region $S3_REGION --source-bucket $S3_BUCKET \\"
echo "#   --source-auth-type iam_role \\"
echo "#   --source 'data/**' \\"
echo "#   --dest-type blob \\"
echo "#   --dest-account $BLOB_ACCOUNT --dest-container $BLOB_CONTAINER \\"
echo "#   --dest-auth-type managed_identity \\"
echo "#   --dest 'migrated/' \\"
echo "#   --concurrent-files 8"
echo "(Skipped — requires AWS EC2 + Azure VM environment)"

# ─────────────────────────────────────────────────────────────────────────────
# PART B: Azure Blob → S3
# ─────────────────────────────────────────────────────────────────────────────

echo ""
echo "════════════════════════════════════════"
echo " PART B: Azure Blob Storage  →  AWS S3"
echo "════════════════════════════════════════"

# ── B1: Single file ───────────────────────────────────────────────────────────
echo ""
echo "=== B1. Copy one blob to S3 ==="
nixcopy transfer \
  --source-type blob \
  --source-container "$BLOB_CONTAINER" \
  --source-auth-type connection_string \
  --source-connection-string "$BLOB_CONN" \
  --source "migrated/report.csv" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "restored/report.csv"
echo "Done."

# ── B2: Batch — copy an entire container prefix to S3 ────────────────────────
echo ""
echo "=== B2. Batch copy all blobs under migrated/2024/ to S3 ==="
nixcopy transfer \
  --source-type blob \
  --source-container "$BLOB_CONTAINER" \
  --source-auth-type connection_string \
  --source-connection-string "$BLOB_CONN" \
  --source "migrated/2024/**" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "restored/2024/" \
  --concurrent-files 6 \
  --retry-attempts 3
echo "Batch copy complete."

# ── B3: Dual-cloud daily backup (cron-friendly) ───────────────────────────────
echo ""
echo "=== B3. Daily backup snapshot: Blob → S3 with bandwidth cap ==="
TODAY=$(date +%Y-%m-%d)
nixcopy transfer \
  --source-type blob \
  --source-container "$BLOB_CONTAINER" \
  --source-auth-type connection_string \
  --source-connection-string "$BLOB_CONN" \
  --source "production/**" \
  --dest-type s3 \
  --dest-region "$S3_REGION" \
  --dest-bucket "$S3_BUCKET" \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY" \
  --dest-secret-key "$AWS_SECRET_KEY" \
  --dest "backup/${TODAY}/" \
  --concurrent-files 4 \
  --retry-attempts 5 \
  --bandwidth-limit 52428800   # 50 MB/s — avoid saturating the network
echo "Daily backup complete."
