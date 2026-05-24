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
| [09-batch-transfer.sh](09-batch-transfer.sh) | Many files in parallel — glob, `--sources`, skip-existing, concurrency tuning; parallel directory listing for `**` patterns |
| [10-advanced-options.sh](10-advanced-options.sh) | Checksum, resume with integrity check, bandwidth limit, retry, verbose — local only, no credentials needed |
| [12-performance-tuning.sh](12-performance-tuning.sh) | Buffer size, concurrency, upload-concurrency, incremental sync, adaptive buffer auto-tuning |
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

## Testing locally with Docker emulators (no cloud account needed)

Scripts 04, 05, 06, 09, and 11 can be run against local emulators:

```bash
# Start MinIO (S3) + Azurite (Azure Blob)
docker compose -f ../../docker-compose.integration.yml up -d minio azurite --wait

# Create bucket and container
docker run --rm --network host \
  -e MC_HOST_local=http://minioadmin:minioadmin@127.0.0.1:9000 \
  minio/mc mb --ignore-existing local/test-bucket

AZURITE_CONN="DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;\
AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;\
BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
az storage container create --name test-container \
  --connection-string "$AZURITE_CONN" --output none

# Run 11-s3-to-blob.sh against local emulators
NIXCOPY_SOURCE_ENDPOINT=http://localhost:9000 NIXCOPY_SOURCE_USE_PATH_STYLE=true \
NIXCOPY_DEST_ENDPOINT=http://localhost:9000   NIXCOPY_DEST_USE_PATH_STYLE=true \
S3_BUCKET=test-bucket S3_REGION=us-east-1 \
AWS_ACCESS_KEY=minioadmin AWS_SECRET_KEY=minioadmin \
BLOB_CONTAINER=test-container BLOB_CONN="$AZURITE_CONN" \
  bash 11-s3-to-blob.sh

# Tear down
docker compose -f ../../docker-compose.integration.yml down -v
```

## Settings that have no CLI flag

Some nixcopy settings are only available via environment variable or config file — there is no `--flag` for them. Pass them inline before the nixcopy command:

| Setting | Env var |
|---------|---------|
| Custom S3/GCS/Blob endpoint | `NIXCOPY_SOURCE_ENDPOINT` / `NIXCOPY_DEST_ENDPOINT` |
| S3 path-style (required for MinIO) | `NIXCOPY_SOURCE_USE_PATH_STYLE=true` / `NIXCOPY_DEST_USE_PATH_STYLE=true` |
| Azure Blob connection string | `NIXCOPY_SOURCE_CONNECTION_STRING` / `NIXCOPY_DEST_CONNECTION_STRING` |
| AWS named profile | `NIXCOPY_SOURCE_PROFILE` / `NIXCOPY_DEST_PROFILE` |
| SHA-256 checksum verification | `NIXCOPY_VERIFY_CHECKSUM=true` |

Example — Azure Blob connection string with inline env var:
```bash
NIXCOPY_DEST_CONNECTION_STRING="DefaultEndpointsProtocol=https;..." nixcopy transfer \
  --dest-type blob --dest-auth-type connection_string --dest-container my-container ...
```

When using `connection_string` auth, `--dest-account-name` is not required (the account name is embedded in the connection string).

## Auth method quick reference

| Backend | Flag | When to use |
|---------|------|-------------|
| S3 | `--*-auth-type access_key` | CI, scripts, cross-account |
| S3 | `--*-auth-type iam_role` | EC2 / ECS / Lambda |
| S3 | `--*-auth-type profile` + `NIXCOPY_*_PROFILE` | Local dev with `~/.aws/` |
| Azure | `--*-auth-type connection_string` + `NIXCOPY_*_CONNECTION_STRING` | Quickest for beginners |
| Azure | `--*-auth-type shared_key` + `--*-account-name` + `--*-account-key` | Account name + key |
| Azure | `--*-auth-type managed_identity` | Azure VMs / AKS |
| GCS | `--*-auth-type application_default` | Local dev (`gcloud auth`) |
| GCS | `--*-auth-type service_account` + `--*-credentials-file` | CI, servers, cross-project |
| GCS | `--*-auth-type workload_identity` | GKE |
| SFTP | `--*-password` | Password auth |
| SFTP | `--*-private-key` | SSH key auth |
