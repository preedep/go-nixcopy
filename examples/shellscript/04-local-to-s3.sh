#!/usr/bin/env bash
# 04-local-to-s3.sh — Upload local files to AWS S3
#
# Beginner scenario: push files from your machine to an S3 bucket.
# Covers three common auth methods:
#   1. access_key  — explicit key + secret (common for scripts / CI)
#   2. iam_role    — automatic on EC2/ECS/Lambda (no keys needed)
#   3. profile     — uses a named AWS credentials profile from ~/.aws/
#
# Prerequisites:
#   nixcopy installed
#   An S3 bucket you have write access to
#
# Usage (access_key):
#   AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=secret... S3_BUCKET=my-bucket \
#     bash 04-local-to-s3.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
S3_BUCKET="${S3_BUCKET:-my-backup-bucket}"
S3_REGION="${S3_REGION:-ap-southeast-1}"
AWS_ACCESS_KEY="${AWS_ACCESS_KEY:-}"
AWS_SECRET_KEY="${AWS_SECRET_KEY:-}"
AWS_PROFILE="${AWS_PROFILE:-}"       # set to use a named profile from ~/.aws/

LOCAL_FILE="/tmp/nixcopy-s3-demo/data.csv"
S3_PREFIX="uploads/2024/"            # destination prefix inside the bucket

# ── Setup ──────────────────────────────────────────────────────────────────────
mkdir -p "$(dirname "$LOCAL_FILE")"
printf "id,name,value\n1,alice,100\n2,bob,200\n" > "$LOCAL_FILE"

# ── Example 1: Upload with explicit access key ─────────────────────────────────
if [[ -n "$AWS_ACCESS_KEY" ]]; then
  echo "=== Upload with access_key auth ==="
  nixcopy transfer \
    --source-type local \
    --source "$LOCAL_FILE" \
    --dest-type s3 \
    --dest-region "$S3_REGION" \
    --dest-bucket "$S3_BUCKET" \
    --dest-auth-type access_key \
    --dest-access-key "$AWS_ACCESS_KEY" \
    --dest-secret-key "$AWS_SECRET_KEY" \
    --dest "${S3_PREFIX}data.csv"
  echo "Upload complete."
fi

# ── Example 2: Upload on EC2 / ECS using the attached IAM role ────────────────
# No key flags needed — credentials come from the EC2 metadata service.
echo ""
echo "=== Upload using IAM role (EC2/ECS/Lambda) ==="
echo "# nixcopy transfer \\"
echo "#   --source-type local --source $LOCAL_FILE \\"
echo "#   --dest-type s3 --dest-region $S3_REGION --dest-bucket $S3_BUCKET \\"
echo "#   --dest-auth-type iam_role \\"
echo "#   --dest ${S3_PREFIX}data.csv"
echo "(Skipped — requires an EC2/ECS environment)"

# ── Example 3: Upload using a named AWS profile ───────────────────────────────
if [[ -n "$AWS_PROFILE" ]]; then
  echo ""
  echo "=== Upload using AWS profile '$AWS_PROFILE' ==="
  nixcopy transfer \
    --source-type local \
    --source "$LOCAL_FILE" \
    --dest-type s3 \
    --dest-region "$S3_REGION" \
    --dest-bucket "$S3_BUCKET" \
    --dest-auth-type profile \
    --dest-profile "$AWS_PROFILE" \
    --dest "${S3_PREFIX}data-profile.csv"
  echo "Upload complete."
fi

# ── Example 4: Upload an entire folder with concurrency ──────────────────────
if [[ -n "$AWS_ACCESS_KEY" ]]; then
  echo ""
  echo "=== Batch upload: all CSV files with 4 concurrent workers ==="
  for i in 1 2 3 4; do
    cp "$LOCAL_FILE" "/tmp/nixcopy-s3-demo/shard-$i.csv"
  done

  nixcopy transfer \
    --source-type local \
    --source "/tmp/nixcopy-s3-demo/*.csv" \
    --dest-type s3 \
    --dest-region "$S3_REGION" \
    --dest-bucket "$S3_BUCKET" \
    --dest-auth-type access_key \
    --dest-access-key "$AWS_ACCESS_KEY" \
    --dest-secret-key "$AWS_SECRET_KEY" \
    --dest "${S3_PREFIX}" \
    --concurrent-files 4 \
  echo "Batch upload complete."
fi

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf /tmp/nixcopy-s3-demo
