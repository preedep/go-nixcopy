#!/usr/bin/env bash
# 07-local-to-gcs.sh — Upload local files to Google Cloud Storage
#
# Beginner scenario: push files to a GCS bucket.
# Covers three common auth methods:
#   1. application_default — uses `gcloud auth application-default login` (local dev)
#   2. service_account     — JSON key file (CI, servers, cross-project)
#   3. workload_identity   — automatic on GKE (no keys needed)
#
# Prerequisites:
#   nixcopy installed
#   A GCS bucket you have write access to
#
# Usage (application_default — run gcloud auth first):
#   GCS_BUCKET=my-bucket bash 07-local-to-gcs.sh
#
# Usage (service_account):
#   GCS_BUCKET=my-bucket GCS_SA_FILE=/path/to/key.json bash 07-local-to-gcs.sh
#
# Usage against fake-gcs-server (local emulator):
#   GCS_BUCKET=test-bucket NIXCOPY_DEST_ENDPOINT=http://localhost:4443/storage/v1/ \
#     bash 07-local-to-gcs.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
GCS_BUCKET="${GCS_BUCKET:-my-gcs-bucket}"
GCS_SA_FILE="${GCS_SA_FILE:-}"       # path to service account JSON key

LOCAL_FILE="/tmp/nixcopy-gcs-demo/model.pkl"
GCS_PREFIX="ml-models/v2/"

# ── Setup ──────────────────────────────────────────────────────────────────────
mkdir -p "$(dirname "$LOCAL_FILE")"
dd if=/dev/urandom bs=1K count=32 2>/dev/null > "$LOCAL_FILE"
echo "Created sample file: $LOCAL_FILE ($(du -sh "$LOCAL_FILE" | cut -f1))"

# ── Example 1: Upload using Application Default Credentials (local dev) ────────
echo "=== Upload with Application Default Credentials ==="
echo "    (run 'gcloud auth application-default login' first)"
nixcopy transfer \
  --source-type local \
  --source "$LOCAL_FILE" \
  --dest-type gcs \
  --dest-bucket "$GCS_BUCKET" \
  --dest-auth-type application_default \
  --dest "${GCS_PREFIX}model.pkl"
echo "Upload complete."

# ── Example 2: Upload with a Service Account JSON key ─────────────────────────
if [[ -n "$GCS_SA_FILE" && -f "$GCS_SA_FILE" ]]; then
  echo ""
  echo "=== Upload with service_account key file ==="
  nixcopy transfer \
    --source-type local \
    --source "$LOCAL_FILE" \
    --dest-type gcs \
    --dest-bucket "$GCS_BUCKET" \
    --dest-auth-type service_account \
    --dest-credentials-file "$GCS_SA_FILE" \
    --dest "${GCS_PREFIX}model-sa.pkl"
  echo "Upload complete."
fi

# ── Example 3: Upload on GKE using Workload Identity ─────────────────────────
echo ""
echo "=== Upload using Workload Identity (GKE) ==="
echo "# nixcopy transfer \\"
echo "#   --source-type local --source $LOCAL_FILE \\"
echo "#   --dest-type gcs --dest-bucket $GCS_BUCKET \\"
echo "#   --dest-auth-type workload_identity \\"
echo "#   --dest ${GCS_PREFIX}model-wi.pkl"
echo "(Skipped — requires a GKE environment)"

# ── Example 4: Batch upload ───────────────────────────────────────────────────
echo ""
echo "=== Batch upload: all model checkpoints ==="
for i in 1 2 3; do
  cp "$LOCAL_FILE" "/tmp/nixcopy-gcs-demo/checkpoint-epoch$i.pkl"
done

nixcopy transfer \
  --source-type local \
  --source "/tmp/nixcopy-gcs-demo/*.pkl" \
  --dest-type gcs \
  --dest-bucket "$GCS_BUCKET" \
  --dest-auth-type application_default \
  --dest "$GCS_PREFIX" \
  --concurrent-files 3
echo "Batch upload complete."

# ── Cleanup ────────────────────────────────────────────────────────────────────
rm -rf /tmp/nixcopy-gcs-demo
