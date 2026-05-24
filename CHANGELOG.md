# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.5] - 2026-05-24

### Added
- Shell script examples for all storage backends (`examples/shellscript/01` through `12`), covering local, SFTP, S3, Azure Blob, GCS, batch, advanced options, and performance tuning.
- `--upload-concurrency` CLI flag for S3 and Azure Blob — controls parallel part uploads within a single file (default 5; raise to 8–10 on high-bandwidth links for large files).
- Shared `sync.Pool` of 256 KiB copy buffers (`storage/pool.go`) used by local and SFTP backends via `io.CopyBuffer`, eliminating per-transfer heap allocations.
- Separate `transferBufPool` (`sync.Pool`) in the compression pipeline goroutine to avoid allocations under concurrent transfers.
- Benchmarks for buffer pool throughput (`BenchmarkTransfer_BufferPool`) at 1 MB / 16 MB / 64 MB.
- `TestTransfer_HighConcurrency` integration test — verifies 8 concurrent 4 MB local→local transfers complete without data corruption under `copyBufPool` contention.

### Fixed
- Azure Blob CLI validator no longer requires `--dest-account-name` / `--source-account-name` when `auth_type == connection_string` (account name is embedded in the connection string).
- Shell script examples repaired: removed dangling backslashes, corrected file paths, fixed `list` subcommand flags.
- `TestTransfer_HighConcurrency` read the wrong destination paths (`hc-dst-N.bin`) — corrected to `hc-src-N.bin` to match `TransferBatch` behaviour of preserving the source filename under `destBase`.

### Changed
- `TransferUseCase.Transfer` decomposed into `runWithRetry`, `attemptTransfer`, and `buildPipeline` sub-functions for clarity.
- `registerStorageFlags` helper deduplicates symmetric source/dest flag registration in `transfer.go`.
- `storageFactory` interface added in CLI layer to enable unit testing without real storage constructors.
- Fury.io APT/YUM publishing disabled in `.goreleaser.yaml` and `release.yml` pending account re-provisioning.

### Internal
- `CLAUDE.md` updated with Go Patterns section (interface injection, `t.Cleanup`, `resetXxxFlags`, mock-state assertions, function decomposition) and performance tuning guidance.

## [1.3.4] - 2026-05-23

### Added
- Azure Blob (Azurite) and GCS (fake-gcs-server) integration test suites, runnable via `make integration-test-local` or `./test-integration.sh blob gcs`.

### Fixed
- `goimports` struct field alignment in `gcs.go`.

## [1.3.3] - 2026-05-22

### Fixed
- Race condition on `progressChan` close in `TransferBatch`; added progress channel tests.
- Double-close panic on progress channel in CLI layer.

### Docs
- Added local storage CLI section; clarified usage suppression behaviour.

[Unreleased]: https://github.com/preedep/go-nixcopy/compare/v1.3.5...HEAD
[1.3.5]: https://github.com/preedep/go-nixcopy/compare/v1.3.4...v1.3.5
[1.3.4]: https://github.com/preedep/go-nixcopy/compare/v1.3.3...v1.3.4
[1.3.3]: https://github.com/preedep/go-nixcopy/compare/v1.3.2...v1.3.3
