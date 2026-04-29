# go-nixcopy

Fast universal file-transfer CLI — move data between SFTP, FTPS, Azure Blob, AWS S3, and local disk with a single command.

[![CI](https://github.com/preedep/go-nixcopy/actions/workflows/ci.yml/badge.svg)](https://github.com/preedep/go-nixcopy/actions/workflows/ci.yml)
[![Docker Hub](https://img.shields.io/docker/v/nickmsft/gonixcopy?label=Docker%20Hub)](https://hub.docker.com/r/nickmsft/gonixcopy)
[![Go Version](https://img.shields.io/github/go-mod/go-version/preedep/go-nixcopy)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Table of Contents

- [What is go-nixcopy?](#what-is-go-nixcopy)
- [Features](#features)
- [Architecture](#architecture)
- [How to Use](#how-to-use)
- [Airflow Integration](#airflow-integration)
- [Roadmap](#roadmap)

---

## What is go-nixcopy?

go-nixcopy is a small, fast Go CLI for transferring files between storage systems. It is designed to run anywhere — bare metal, Kubernetes pods (KPO), or CI/CD pipelines — with minimal memory footprint and no runtime dependencies.

**Supported storage backends:**

| Backend | Read | Write |
|---|---|---|
| Local file system | ✅ | ✅ |
| SFTP | ✅ | ✅ |
| FTPS | ✅ | ✅ |
| Azure Blob Storage | ✅ | ✅ |
| AWS S3 / MinIO | ✅ | ✅ |

**Install:**

```bash
# Go install
go install github.com/preedep/go-nixcopy/cmd/nixcopy@latest

# Docker
docker pull nickmsft/gonixcopy:latest
```

---

## Features

- **Parallel transfer** — configurable concurrent file count with semaphore control
- **Wildcard patterns** — `*.pdf`, `**/*.log`, `data/2024/**/*.csv`
- **On-the-fly compression** — gzip / zstd streaming, no temp files (`--compress gzip`)
- **Bandwidth limiting** — per-file throttle, accepts `10MB`, `1GiB`, or raw bytes (`--bandwidth-limit 10MB`)
- **Resume** — continue interrupted transfers from byte offset (Local & SFTP)
- **Checksum verification** — SHA-256 end-to-end integrity check (`--verify-checksum`)
- **Skip existing** — idempotent retry; skip if destination already has the same size (`--skip-existing`)
- **Structured JSON logging** — [standard-app-log v1.0](https://github.com/preedep/standard-app-log) to stdout, K8s / Airflow ready
- **Config-free operation** — all settings injectable via `NIXCOPY_*` env vars; no YAML file needed in containers
- **Distroless Docker image** — non-root, no shell, multi-arch (`linux/amd64` + `linux/arm64`)

---

## Architecture

Clean Architecture — dependencies flow strictly inward:

```
cmd/nixcopy/main.go
    └─> internal/interfaces/cli/        ← Cobra commands, Viper config binding
            └─> internal/usecase/       ← orchestration, pattern matching, retry
                    └─> internal/domain/      ← entities + repository interfaces
            └─> internal/infrastructure/      ← storage impls, config, logger
```

**Directory layout:**

```
go-nixcopy/
├── cmd/nixcopy/               # Entry point
├── internal/
│   ├── domain/                # Entities, repository & service interfaces
│   ├── usecase/               # Transfer orchestration, pattern matcher, mocks
│   ├── infrastructure/
│   │   ├── config/            # YAML/JSON loader with ${ENV_VAR} expansion
│   │   ├── logger/            # standard-app-log JSON emitter
│   │   └── storage/           # local, sftp, ftps, blob, s3, factory
│   └── interfaces/cli/        # Cobra subcommands: transfer, list
└── examples/                  # Sample config files per backend pair
```

---

## How to Use

### Quick start

```bash
# Transfer a single file (config file)
nixcopy transfer -c config.yaml -s /remote/file.csv -d backup/file.csv

# Transfer without a config file (all via CLI flags)
nixcopy transfer \
  --source-type sftp --source-host sftp.example.com \
  --source-username user --source-password pass \
  -s /remote/file.csv \
  --dest-type s3 --dest-region ap-southeast-1 \
  --dest-bucket my-bucket --dest-auth-type iam_role \
  -d processed/file.csv

# Wildcard batch transfer
nixcopy transfer -c config.yaml -s "reports/*.pdf" -d backup/ --concurrent-files 8

# List files in a storage backend
nixcopy list -c config.yaml -p /remote/path --source
```

### Config file

```yaml
source:
  type: sftp          # local | sftp | ftps | blob | s3
  sftp:
    host: sftp.example.com
    port: 22
    username: user
    password: ${SFTP_PASSWORD}

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-bucket
    auth_type: iam_role   # access_key | iam_role | web_identity | assume_role

transfer:
  buffer_size: 33554432    # 32 MB
  concurrent_files: 4
  retry_attempts: 3
  retry_delay: 5s
  verify_checksum: false
  compression: ""          # "gzip" | "zstd" | "" (no compression)
  bandwidth_limit: 0       # bytes/sec per file; 0 = unlimited
  skip_existing: false
  enable_resume: false
```

**Config precedence:** `CLI flags > NIXCOPY_* env vars > config file > defaults`

### Docker / Kubernetes

```bash
# Run via env vars — no config file needed (recommended for KPO)
docker run --rm \
  -e NIXCOPY_SOURCE_TYPE=sftp \
  -e NIXCOPY_SOURCE_HOST=sftp.example.com \
  -e NIXCOPY_SOURCE_USERNAME=user \
  -e NIXCOPY_SOURCE_PASSWORD=secret \
  -e NIXCOPY_DEST_TYPE=s3 \
  -e NIXCOPY_DEST_REGION=ap-southeast-1 \
  -e NIXCOPY_DEST_BUCKET=my-bucket \
  -e NIXCOPY_DEST_AUTH_TYPE=web_identity \
  nickmsft/gonixcopy:latest transfer -s /data/file.csv -d processed/file.csv
```

### Transfer output

Progress goes to **stderr** (human-readable):

```
[file.csv] 45.23% | 125.45 MB/s | ETA: 1m15s
[file.csv] ✓ Completed | 128.32 MB/s
```

A JSON summary goes to **stdout** on completion (machine-readable):

```json
{"event":"transfer_summary","total_files":3,"successful":2,"skipped":1,"failed":0,"bytes_transferred":10737418240,"duration_ms":79823,"average_speed_mbps":128.32}
```

### Detailed documentation

| Topic | File |
|---|---|
| Quick start & troubleshooting & performance tips | [QUICKSTART.md](QUICKSTART.md) |
| All CLI flags & `NIXCOPY_*` env vars | [CLI_USAGE.md](CLI_USAGE.md) |
| Authentication setup (AWS, Azure) | [AUTHENTICATION.md](AUTHENTICATION.md) |
| Parallel transfer & wildcard patterns | [PARALLEL_TRANSFER.md](PARALLEL_TRANSFER.md) |
| Testing guide | [TESTING.md](TESTING.md) |
| Release build flags & binary sizes | [BUILD.md](BUILD.md) |
| Deployment environments (EKS, AKS, on-prem) | [ENVIRONMENT_GUIDE.md](ENVIRONMENT_GUIDE.md) |
| Contributing a new storage backend | [CONTRIBUTING.md](CONTRIBUTING.md) |

---

## Airflow Integration

go-nixcopy works as a drop-in task in Apache Airflow via `KubernetesPodOperator` (KPO).  
No config file mount is needed — all storage credentials are injected as `NIXCOPY_*` env vars from Kubernetes Secrets.

### How it works

```
Airflow Scheduler
    └─> KubernetesPodOperator
            └─> nickmsft/gonixcopy pod
                    ├── NIXCOPY_* env vars  ← from K8s Secret
                    ├── transfer (source → destination)
                    └── stdout: JSON summary  ← parsed by Airflow log / Loki
```

### DAG example — SFTP to S3

```python
from datetime import datetime
from airflow import DAG
from airflow.providers.cncf.kubernetes.operators.pod import KubernetesPodOperator
from kubernetes.client import models as k8s

with DAG(
    dag_id="sftp_to_s3_transfer",
    schedule="0 2 * * *",       # daily at 02:00
    start_date=datetime(2024, 1, 1),
    catchup=False,
    tags=["nixcopy", "sftp", "s3"],
) as dag:

    transfer = KubernetesPodOperator(
        task_id="transfer_sftp_to_s3",
        image="nickmsft/gonixcopy:latest",   # pin to a specific tag in production
        cmds=["./nixcopy"],
        arguments=[
            "transfer",
            "-s", "/data/exports/*.csv",
            "-d", "processed/{{ ds }}/",     # Airflow date partition
            "--concurrent-files", "4",
            "--skip-existing",
        ],
        env_vars=[
            # Observability
            k8s.V1EnvVar(name="NIXCOPY_CORRELATION_ID", value="{{ run_id }}"),
            k8s.V1EnvVar(name="NIXCOPY_APP_VERSION",    value="latest"),
            k8s.V1EnvVar(name="POD_NAME", value_from=k8s.V1EnvVarSource(
                field_ref=k8s.V1ObjectFieldSelector(field_path="metadata.name"))),

            # Source — SFTP credentials from K8s Secret
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_TYPE", value="sftp"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_HOST", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="host"))),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_USERNAME", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="username"))),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_PASSWORD", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="password"))),

            # Destination — S3 via IRSA (no keys needed)
            k8s.V1EnvVar(name="NIXCOPY_DEST_TYPE",      value="s3"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_REGION",    value="ap-southeast-1"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_BUCKET",    value="my-data-bucket"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_AUTH_TYPE", value="web_identity"),
        ],
        security_context=k8s.V1PodSecurityContext(run_as_non_root=True),
        namespace="airflow",
        service_account_name="nixcopy-sa",   # bound to IAM role via IRSA
        get_logs=True,
        is_delete_operator_pod=True,
        in_cluster=True,
    )
```

### DAG example — Azure Blob to SFTP

```python
    blob_to_sftp = KubernetesPodOperator(
        task_id="transfer_blob_to_sftp",
        image="nickmsft/gonixcopy:latest",
        cmds=["./nixcopy"],
        arguments=[
            "transfer",
            "-s", "reports/{{ ds }}/*.pdf",
            "-d", "/upload/reports/{{ ds }}/",
        ],
        env_vars=[
            k8s.V1EnvVar(name="NIXCOPY_CORRELATION_ID", value="{{ run_id }}"),

            # Source — Azure Blob with Managed Identity
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_TYPE",         value="blob"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_ACCOUNT_NAME", value="mystorageaccount"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_CONTAINER",    value="reports"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_AUTH_TYPE",    value="managed_identity"),

            # Destination — SFTP
            k8s.V1EnvVar(name="NIXCOPY_DEST_TYPE", value="sftp"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_HOST", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="host"))),
            k8s.V1EnvVar(name="NIXCOPY_DEST_USERNAME", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="username"))),
            k8s.V1EnvVar(name="NIXCOPY_DEST_PASSWORD", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="password"))),
        ],
        security_context=k8s.V1PodSecurityContext(run_as_non_root=True),
        namespace="airflow",
        get_logs=True,
        is_delete_operator_pod=True,
        in_cluster=True,
    )
```

### Transfer output in Airflow logs

Progress lines go to **stderr** (visible in Airflow task logs):

```
[report_jan.pdf] 72.10% | 98.32 MB/s | ETA: 0m12s
[report_jan.pdf] ✓ Completed | 101.45 MB/s
```

The JSON summary goes to **stdout** — parse it with Loki / CloudWatch Insights using `event="transfer_summary"`:

```json
{"event":"transfer_summary","total_files":5,"successful":4,"skipped":1,"failed":0,"bytes_transferred":524288000,"duration_ms":5120,"average_speed_mbps":98.32}
```

> **Tip:** Pass `{{ run_id }}` as `NIXCOPY_CORRELATION_ID` to correlate all log lines from a DAG run across Loki / CloudWatch.

---

## Roadmap

### Done
- [x] Local, SFTP, FTPS, Azure Blob, AWS S3 / MinIO support
- [x] Parallel transfer & wildcard patterns (`*.pdf`, `**/*.log`)
- [x] SHA-256 checksum verification
- [x] Resume interrupted transfers (Local & SFTP)
- [x] Bandwidth limiting per file
- [x] On-the-fly gzip / zstd compression
- [x] Idempotent retry (`--skip-existing`)
- [x] Structured JSON logging (standard-app-log v1.0)
- [x] Distroless Docker image, multi-arch (amd64 + arm64), config-free KPO mode
- [x] S3 multipart upload (up to ~5 TiB) & Azure Blob parallel block upload (up to ~190 TiB)
- [x] GitHub Actions CI with integration tests (MinIO + SFTP)

### Planned
- [ ] Google Cloud Storage support
- [ ] Web UI for transfer management
- [ ] Scheduled transfers
- [ ] Email / webhook notifications
- [ ] Incremental backup

---

## Contributing

Bug reports and pull requests are welcome on [GitHub Issues](https://github.com/preedep/go-nixcopy/issues).
See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow and code template.

## License

MIT — see [LICENSE](LICENSE).

---

Made with ❤️ in Thailand
