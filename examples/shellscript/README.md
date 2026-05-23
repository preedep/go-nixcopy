# Shell Script Examples

Beginner-friendly scripts that demonstrate common nixcopy scenarios.
Each script is self-contained and heavily commented.

## Prerequisites

Install nixcopy:
```bash
make install   # or: go install ./cmd/nixcopy
```

## Scripts

| Script | Scenario |
|--------|----------|
| [01-local-to-local.sh](01-local-to-local.sh) | Copy files between local directories; glob patterns; checksum |
| [02-local-to-sftp.sh](02-local-to-sftp.sh) | Upload to SFTP — password auth or SSH private key |
| [03-sftp-to-local.sh](03-sftp-to-local.sh) | Download from SFTP — single file, resume, batch |
| [04-local-to-s3.sh](04-local-to-s3.sh) | Upload to AWS S3 — access_key, IAM role, named profile |
| [05-s3-to-local.sh](05-s3-to-local.sh) | Download from S3 — single file, checksum, batch |
| [06-local-to-azure-blob.sh](06-local-to-azure-blob.sh) | Upload to Azure Blob — connection_string, shared_key, Managed Identity |
| [07-local-to-gcs.sh](07-local-to-gcs.sh) | Upload to GCS — Application Default, service account key, Workload Identity |
| [08-sftp-to-s3.sh](08-sftp-to-s3.sh) | Direct SFTP → S3 (no local disk); includes nightly backup pattern |
| [09-batch-transfer.sh](09-batch-transfer.sh) | Many files in parallel — glob, `--sources`, skip-existing, concurrency tuning |
| [10-advanced-options.sh](10-advanced-options.sh) | Checksum, resume, bandwidth limit, retry, verbose — local only, no credentials needed |
| [11-s3-to-blob.sh](11-s3-to-blob.sh) | S3 ↔ Azure Blob — single file, batch migration, incremental sync, dual-cloud backup |

## Quick start (no credentials)

Scripts 01 and 10 run entirely on local disk — good for exploring the CLI:

```bash
bash 01-local-to-local.sh
bash 10-advanced-options.sh
```

## Running with real credentials

Each script reads credentials from environment variables so secrets never
appear in the script files:

```bash
# SFTP upload
SFTP_HOST=sftp.example.com SFTP_USER=alice SFTP_PASS=secret \
  bash 02-local-to-sftp.sh

# S3 upload
AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=... S3_BUCKET=my-bucket \
  bash 04-local-to-s3.sh

# Azure Blob upload
BLOB_CONN="DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net" \
BLOB_CONTAINER=my-container \
  bash 06-local-to-azure-blob.sh

# GCS upload (Application Default — run gcloud auth first)
GCS_BUCKET=my-bucket bash 07-local-to-gcs.sh

# S3 → Azure Blob migration
AWS_ACCESS_KEY=AKIA... AWS_SECRET_KEY=... S3_BUCKET=src-bucket \
BLOB_CONN="DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net" \
BLOB_CONTAINER=dest-container \
  bash 11-s3-to-blob.sh
```

## Auth method quick reference

| Backend | Flag | When to use |
|---------|------|-------------|
| S3 | `--*-auth-type access_key` | CI, scripts, cross-account |
| S3 | `--*-auth-type iam_role` | EC2 / ECS / Lambda |
| S3 | `--*-auth-type profile` | Local dev with `~/.aws/` |
| Azure | `--*-auth-type connection_string` | Quickest for beginners |
| Azure | `--*-auth-type shared_key` | Account name + key |
| Azure | `--*-auth-type managed_identity` | Azure VMs / AKS |
| GCS | `--*-auth-type application_default` | Local dev (`gcloud auth`) |
| GCS | `--*-auth-type service_account` | CI, servers, cross-project |
| GCS | `--*-auth-type workload_identity` | GKE |
| SFTP | `--*-password` | Password auth |
| SFTP | `--*-private-key` | SSH key auth |
