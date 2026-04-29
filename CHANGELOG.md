# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added — FTPS TLS Mode (Enterprise)

- **`--source-tls-mode` / `--dest-tls-mode` CLI flags** (`internal/interfaces/cli/transfer.go`) — accept `explicit` (STARTTLS on port 21) or `implicit` (TLS-first on port 990). Empty string leaves any value already loaded from the config file or env var unchanged.
- **`NIXCOPY_SOURCE_TLS_MODE` / `NIXCOPY_DEST_TLS_MODE` env vars** — same values; applied via `config.ApplyBackendEnv`, consistent with all other per-backend env vars.
- **`validateTLSMode`** helper (`internal/interfaces/cli/flags.go`) — rejects any value other than `""`, `"explicit"`, `"implicit"` with a clear error at startup.
- **Unit tests** (`internal/interfaces/cli/ftps_tls_mode_test.go`) — 13 tests covering explicit/implicit source+dest, empty-no-override, `validateConfig` invalid/valid modes, env var path, and CLI-over-env precedence.

### Added — Tests (Coverage Improvements)

- **`internal/infrastructure/logger/applog_test.go`** — 22 tests: all log methods (`Info`/`Warn`/`Error`/`Debug`, `InfoReqEx`/`InfoResEx`/`WarnResEx`/`ErrorReqEx`/`ErrorResEx`), field helpers (`F`/`FError`/`FDurationMs`), `WithCorrelation`/`WithRequest`/`WithTrace`, `NopLogger`, `GenerateID` format/uniqueness, extra fields, parent immutability. Raises logger coverage from 0% → ~72%.
- **`internal/infrastructure/storage/factory_test.go`** — `NewStorageFromSourceConfig` + `NewStorageFromDestConfig` for all 5 backends, nil-config errors, unknown type errors.
- **`internal/infrastructure/storage/local_unit_test.go`** — full `LocalStorage` coverage using `t.TempDir()`: Connect, Disconnect, List, Read, Write, Delete, Stat, CreateDirectory, ReadFrom, AppendWrite — no external dependencies.
- **`internal/infrastructure/storage/ftps_nil_client_test.go`** — FTPS nil-client guards: List/Read/Stat/Delete/Disconnect/Write.
- **`internal/infrastructure/storage/blob_nil_client_test.go`** — Azure Blob nil-client guards: List/Read/Stat/Write/Delete; Disconnect/CreateDirectory no-ops.
- **`internal/infrastructure/storage/s3_nil_client_test.go`** — S3 nil-client guards: List/Read/Stat/Write/Delete; Disconnect/CreateDirectory no-ops.
- **`internal/infrastructure/storage/sftp_nil_client_test.go`** — SFTP nil-client guards: List/Read/Stat/Write/Delete/CreateDirectory/ReadFrom/AppendWrite/Disconnect.
- **`internal/interfaces/cli/validate_format_test.go`** — all `validateConfig` source/dest backend error paths, compression validation, `formatSize` boundary cases (B/KB/MB/GB).
- **`internal/usecase/transfer_usecase_branches_test.go`** — `progressReader.Close` (with and without inner Closer), resume+compression warning path, checksum skipped for resumed transfers, checksum skipped for compressed transfers, ReadFrom error exhausting retries.
- **Entity edge cases** (`internal/domain/entity/pattern_test.go`) — `Match`/`MatchFull` with invalid bracket patterns, multiple `**`, prefix mismatch, prefix-only recursive match. Raises entity coverage to 100%.
- **CLI flag coverage** (`internal/interfaces/cli/flags_test.go`) — local storage source/dest, SFTP private key source/dest, dest FTPS, source S3 access/secret key. Raises `applyCliFlags` coverage from 84% → 97%.
- **Env var + backend tests** (`internal/infrastructure/config/envloader_test.go`) — NIXCOPY_* env var loading for all storage types and transfer flags; CLI-over-env precedence.

### Added — Compression

- **`--compress gzip|zstd` CLI flag** and **`NIXCOPY_COMPRESSION` env var** — compress the data stream on-the-fly before writing to the destination. No temporary files; uses `io.Pipe` so memory stays bounded. Destination receives compressed bytes — name the dest path accordingly (`.gz`, `.zst`).
- **`Compression string` field** added to `entity.TransferConfig` and `config.TransferConfig` (`compression` YAML/JSON key). Valid values: `""` (default, no compression), `"gzip"`, `"zstd"`.
- **`newCompressWriter`** (`internal/usecase/compress.go`) — wraps an `io.Writer` with gzip (stdlib `compress/gzip`) or zstd (`github.com/klauspost/compress/zstd`). Empty algo returns a no-op passthrough. Invalid algo returns an error at transfer time.
- **Auto-disable incompatible features**: resume is disabled with a warning when compression is active (compressed chunks cannot be appended); checksum verification is disabled with a warning (destination bytes differ from source hash).
- **Progress correctness**: the progress reader wraps the uncompressed source stream, so the percentage displayed is based on raw source bytes read — not the smaller compressed wire size.
- **Unit tests** (`internal/usecase/compress_test.go`) — 6 tests: invalid algo error, passthrough identity, gzip round-trip, zstd round-trip, gzip compression ratio on compressible input, zstd compression ratio on compressible input.
- **`TestLoadFromEnv_Compression`** added to `internal/infrastructure/config/envloader_test.go`.
- **Dependency**: `github.com/klauspost/compress` promoted from indirect to direct (zstd encoder/decoder).

### Added — Bandwidth Limiting

- **`--bandwidth-limit` CLI flag** (`internal/interfaces/cli/transfer.go`) — accepts human-readable suffixes (`10MB`, `1GB`, `512KB`, `10MiB`, etc.) or raw byte integers. Parsed via `usecase.ParseBandwidth`. Empty / `0` = unlimited (default).
- **`NIXCOPY_BANDWIDTH_LIMIT` env var** (`internal/infrastructure/config/envloader.go`) — raw bytes/sec integer; read by `setEnvInt64` helper added alongside the existing `setEnvInt`.
- **`BandwidthLimit int64` field** added to `entity.TransferConfig` and `config.TransferConfig` (`bandwidth_limit` YAML/JSON key).
- **`ThrottledReader`** (`internal/usecase/throttle.go`) — cumulative-bytes approach: tracks bytes sent since an epoch and sleeps the exact deficit when ahead of the target rate. Context cancellation aborts the sleep immediately. Epoch window is normalised every `bytesPerSec` bytes to prevent int64 overflow on very long transfers. No external dependencies (stdlib `time` only).
- **`ParseBandwidth`** (`internal/usecase/throttle.go`) — parses `GiB`/`GB`/`MiB`/`MB`/`KiB`/`KB`/`G`/`M`/`K` suffixes (case-insensitive) and raw integer strings. Returns `(0, nil)` for empty or `"0"` (unlimited).
- **Unit tests** (`internal/usecase/throttle_test.go`) — 4 tests: all `ParseBandwidth` formats and error cases, zero-rate passthrough (no `ThrottledReader` allocated), data integrity over 64 KB payload, context cancel abort with 1-byte/s rate.
- **`TestLoadFromEnv_BandwidthLimit`** added to `internal/infrastructure/config/envloader_test.go`.

### Added — Speed & Reliability

- **S3 multipart upload** (`internal/infrastructure/storage/s3.go`) — replaced single `PutObject` (5 GB hard limit) with AWS SDK v2 `transfermanager`. Block size is computed dynamically: default 16 MiB, scales up when `ceil(size / 10,000) > 16 MiB` so the 10,000-part S3 limit is never exceeded; minimum 5 MiB. Five concurrent part uploads per file. Supports files up to ~5 TiB.
- **Azure Blob parallel block upload** (`internal/infrastructure/storage/blob.go`) — replaced empty `UploadStreamOptions{}` with `BlockSize` + `Concurrency`. Block size computed with same algorithm as S3 (50,000-block Azure limit, 1 MiB floor). Five concurrent block uploads per blob. Supports blobs up to ~190 TiB.
- **Idempotent retry / skip-existing** (`--skip-existing` / `NIXCOPY_SKIP_EXISTING=true`) — before each transfer, `Stat` the destination; if it exists with the same byte count as the source, mark the file `skipped` without touching the destination. Files with mismatched sizes are re-transferred. Batch transfers report `skipped` separately from `successful` in the JSON summary. Safe for Airflow DAG retries.
- **Structured JSON exit summary to stdout** (`internal/interfaces/cli/transfer.go`) — on completion, a single JSON line is written to stdout: `{"event":"transfer_summary","total_files":N,"successful":N,"skipped":N,"failed":N,"bytes_transferred":N,"duration_ms":N,"average_speed_mbps":N}`. `failed_files` array included when `failed > 0`. `average_speed_mbps` omitted (omitempty) when zero.
- **Progress to stderr** — all `\r`-based progress lines moved from stdout to stderr. A `sync.WaitGroup` ensures the progress goroutine fully drains before the JSON summary line is written, preventing interleaving in K8s log streams.

### Added — Kubernetes / KPO Golden Image

- **NIXCOPY_* env-var-only config** (`internal/infrastructure/config/envloader.go`) — all storage and transfer settings can now be injected as `NIXCOPY_SOURCE_*` / `NIXCOPY_DEST_*` / `NIXCOPY_*` environment variables; no config file mount required. Precedence: CLI flags > `NIXCOPY_*` env vars > config file > defaults. Enables clean KubernetesPodOperator deployments where credentials come from Kubernetes Secrets.
- **Multi-arch Docker image** (`Dockerfile`, `Makefile`) — build stage uses `--platform=$BUILDPLATFORM` + `GOOS`/`GOARCH` cross-compilation so a single `docker buildx` run produces both `linux/amd64` and `linux/arm64` layers. New Makefile target: `make docker-buildx` builds and pushes both platforms; `make docker-build` builds the current platform locally.
- **OCI image labels** (`Dockerfile`) — `org.opencontainers.image.version`, `revision` (git SHA), `created` (ISO-8601 UTC), `title`, `source`, and `licenses` are injected at build time via `--build-arg`. Both `make docker-build` and `make docker-buildx` pass these automatically.

### Fixed — Kubernetes / KPO Golden Image

- **Distroless runtime image** (`Dockerfile`) — runtime stage switched from `alpine:3.21` to `gcr.io/distroless/static-debian12:nonroot`. No shell, no package manager, no apk CVEs. CA certificates and tzdata included in the distroless base. Non-root user (UID 65532) enforced by the `:nonroot` image tag — no `adduser` step needed. Builder updated from `golang:1.21-alpine3.21` to `golang:1.24-alpine3.21` to match `go.mod`. Binary now built with `-trimpath -ldflags="-s -w"` (strips debug symbols and path info; ~30% smaller). CI `build` job now also builds the Docker image and runs `docker run --rm go-nixcopy:ci-test --help` as a smoke test.
- **Non-root container user** (`Dockerfile`) — runtime stage now creates a dedicated `nixcopy` system user/group and switches to it with `USER nixcopy` before the entrypoint. Pods will no longer be rejected by `runAsNonRoot: true` Pod Security Admission policies.
- **Pinned base images** (`Dockerfile`) — builder changed from `golang:1.21-alpine` to `golang:1.21-alpine3.21`; runtime changed from `alpine:latest` to `alpine:3.21`. Golden images are now reproducible and auditable.
- **Correct exit codes on partial failure** (`internal/interfaces/cli/transfer.go`) — when one or more files in a batch fail to transfer, the command now returns a non-nil error (`N of M file(s) failed to transfer`). Previously `runTransfer` returned `nil` even when `failCount > 0`, causing KPO to mark the Airflow task as successful despite data loss.

### Added — CI / Quality

- **GitHub Actions CI workflow** (`.github/workflows/ci.yml`) — four jobs: `lint` (go vet + `go mod tidy` diff check + golangci-lint), `test` (unit tests with `-race -covermode=atomic -count=1`), `integration-test` (MinIO + SFTP service containers), `build` (binary produced only after all three pass). `build` job depends on all three gates.
- **`.golangci.yml`** — explicit linter set: `errcheck`, `staticcheck`, `unused`, `gofmt`, `goimports`, `misspell`, `unconvert`, `unparam`. Prevents accidental reliance on golangci-lint's unstable default set.

### Fixed — CI

- **Go version mismatch** — `go-version: "1.21"` hardcoded in all workflow jobs replaced with `go-version-file: go.mod` (go.mod declares `go 1.24`). Applies to both `ci.yml` and `release.yml`.
- **Coverage mode** — test run now uses `-covermode=atomic` (required alongside `-race`; without it coverage numbers are unreliable under the race detector).
- **Test caching** — added `-count=1` to unit and integration test runs to prevent Go's test cache from hiding real failures on re-runs.

### Added — Tests

- `internal/infrastructure/storage/s3_multipart_test.go` — 5 table-driven tests for `s3PartSizeFor`: unknown size, 1 MiB, 100 MiB, 160 GiB, 5 TiB.
- `internal/infrastructure/storage/blob_multipart_test.go` — 5 table-driven tests for `blobBlockSizeFor`: unknown size, 1 MiB, 100 MiB, 800 GiB, 190 TiB.
- `internal/usecase/transfer_usecase_skip_test.go` — 5 tests: same-size skips, different-size re-transfers, missing destination transfers, `SkipExisting=false` always transfers, batch with 2-of-3 skipped.
- `internal/interfaces/cli/transfer_summary_test.go` — 4 tests: JSON field names, `average_speed_mbps` omitted when zero, `failed_files` array shape, `event` field always present.
- `internal/infrastructure/config/envloader_test.go` — added `TestLoadFromEnv_SkipExisting` for `NIXCOPY_SKIP_EXISTING`.
- `internal/interfaces/cli/flags_test.go` — added `TestApplyCliFlags_SkipExisting`, `TestApplyCliFlags_EnableResume`, `TestApplyCliFlags_SkipExisting_False_DoesNotOverrideConfigFile`; extracted `resetTransferFlags` helper.

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
