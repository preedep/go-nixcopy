# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added — Kubernetes / KPO Golden Image

- **NIXCOPY_* env-var-only config** (`internal/infrastructure/config/envloader.go`) — all storage and transfer settings can now be injected as `NIXCOPY_SOURCE_*` / `NIXCOPY_DEST_*` / `NIXCOPY_*` environment variables; no config file mount required. Precedence: CLI flags > `NIXCOPY_*` env vars > config file > defaults. Enables clean KubernetesPodOperator deployments where credentials come from Kubernetes Secrets.
- **Multi-arch Docker image** (`Dockerfile`, `Makefile`) — build stage uses `--platform=$BUILDPLATFORM` + `GOOS`/`GOARCH` cross-compilation so a single `docker buildx` run produces both `linux/amd64` and `linux/arm64` layers. New Makefile target: `make docker-buildx` builds and pushes both platforms; `make docker-build` builds the current platform locally.
- **OCI image labels** (`Dockerfile`) — `org.opencontainers.image.version`, `revision` (git SHA), `created` (ISO-8601 UTC), `title`, `source`, and `licenses` are injected at build time via `--build-arg`. Both `make docker-build` and `make docker-buildx` pass these automatically.

### Fixed — Kubernetes / KPO Golden Image

- **Non-root container user** (`Dockerfile`) — runtime stage now creates a dedicated `nixcopy` system user/group and switches to it with `USER nixcopy` before the entrypoint. Pods will no longer be rejected by `runAsNonRoot: true` Pod Security Admission policies.
- **Pinned base images** (`Dockerfile`) — builder changed from `golang:1.21-alpine` to `golang:1.21-alpine3.21`; runtime changed from `alpine:latest` to `alpine:3.21`. Golden images are now reproducible and auditable.
- **Correct exit codes on partial failure** (`internal/interfaces/cli/transfer.go`) — when one or more files in a batch fail to transfer, the command now returns a non-nil error (`N of M file(s) failed to transfer`). Previously `runTransfer` returned `nil` even when `failCount > 0`, causing KPO to mark the Airflow task as successful despite data loss.

### Added — Previous Release

- **Standard application logging (standard-app-log v1.0)** — all logs are emitted as structured JSON to stdout, conforming to the [standard-app-log v1.0](https://github.com/preedep/standard-app-log) schema. Log types used: `APP_LOG` (lifecycle, retry), `REQ_EX_LOG` (initiating reads from source), `RES_EX_LOG` (write completion, checksum result). Every entry carries `correlation_id` and `request_id` for distributed tracing. Inject context via env vars: `NIXCOPY_CORRELATION_ID`, `NIXCOPY_APP_ID`, `NIXCOPY_APP_VERSION`, `POD_NAME`. Replaces the previous Zap-based configurable logger; the `logging:` config block is deprecated and no longer read.
- **SHA-256 checksum verification** — set `verify_checksum: true` or `--verify-checksum`; hash is computed in-flight on the source stream then compared against a re-read of the destination. A mismatch triggers automatic retry. `TransferResult.Checksum` carries the hex digest on success.
- **Resume capability for interrupted transfers** — set `enable_resume: true` or `--resume`; before each retry attempt the destination is stat'd and, if a partial file exists, the source is read from that offset and the destination is appended rather than overwritten. Supported backends: Local, SFTP. S3 / Azure Blob / FTPS fall back to full re-transfer with a warning log. `TransferResult.ResumedFrom` records the byte offset used.
- `repository.Resumer` interface — optional interface that storage backends implement to opt into resume support (`ReadFrom` + `AppendWrite`).

### Initial release of go-nixcopy
- Support for SFTP storage
- Support for FTPS storage
- Support for Azure Blob Storage
- Support for AWS S3 storage
- Streaming file transfer with minimal memory usage
- Progress tracking with real-time updates
- Concurrent file transfer support
- Configurable retry mechanism
- CLI interface with transfer and list commands
- Comprehensive configuration via YAML files
- Structured logging with Zap
- Clean Architecture implementation
- Docker support
- Example configurations for common use cases

### Features
- **Transfer Operations**
  - SFTP ↔ FTPS
  - SFTP ↔ Azure Blob Storage
  - SFTP ↔ AWS S3
  - FTPS ↔ Azure Blob Storage
  - FTPS ↔ AWS S3
  - Azure Blob Storage ↔ AWS S3

- **Performance**
  - Streaming I/O for memory efficiency
  - Configurable buffer sizes
  - Concurrent file transfers
  - Automatic retry on failures

- **Configuration**
  - YAML-based configuration
  - Support for environment variables
  - Flexible timeout settings
  - TLS/SSL support for secure connections

- **Monitoring**
  - Real-time progress tracking
  - Transfer speed calculation
  - ETA estimation
  - Structured JSON logging

## [0.1.0] - 2024-XX-XX

### Added
- Initial project structure
- Core domain entities and interfaces
- Storage implementations (SFTP, FTPS, Blob, S3)
- Transfer use case with streaming support
- CLI commands (transfer, list)
- Configuration management
- Logging infrastructure
- Example configurations
- Comprehensive Thai documentation

[Unreleased]: https://github.com/preedep/go-nixcopy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/preedep/go-nixcopy/releases/tag/v0.1.0
