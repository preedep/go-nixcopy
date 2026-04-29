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
