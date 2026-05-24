# CLAUDE.md

## Project Identity

`go-nixcopy` is Golang CLI Project , for fast universal copy several source/sink (ex. sftp server to azure blob storage)  
 `go-nixcop` is focus small binary , fast , low memory consumption

## Architecture Decision Records

Non-obvious design and testing tradeoffs are documented as ADRs in [`docs/adr/`](docs/adr/). Check there before re-proposing a rejected alternative.

| ADR | Decision |
|-----|----------|
| [ADR-0001](docs/adr/0001-sftp-auth-unit-test-strategy.md) | SFTP auth unit test strategy: extract helper over in-process SSH server |
| [ADR-0002](docs/adr/0002-gcs-emulator-transport.md) | GCS emulator HTTP transport: custom `RoundTripper` rewrites XML API paths for fake-gcs-server compatibility |
## Commands

```bash
# Dependencies
make deps               # download Go modules

# Build
make build              # ./bin/nixcopy (dev)
make build-all          # cross-compile (Linux, macOS, Windows, FreeBSD)
make release            # optimized release binary → dist/nixcopy
make release-all        # release for all platforms
make install            # install to $GOPATH/bin

# Test
make test               # all tests
make test-verbose       # with -v flag
make test-coverage      # generates coverage.html
make test-race          # with race detector
go test ./internal/usecase/... -run TestTransferUseCase_Transfer_Success  # single test

# Lint / Format
make lint               # golangci-lint (must be installed separately)
make fmt                # go fmt + gofmt

# Docker
make docker-build && make docker-run

# Integration tests (local — requires Docker + Azure CLI)
make integration-test-local   # spin up SFTP + FTPS + MinIO + Azurite + fake-gcs-server, run all suites, tear down
make integration-test-down    # tear down containers without running tests
./test-integration.sh ftps    # run only the FTPS suite
./test-integration.sh blob    # run only the Azure Blob / Azurite suite
./test-integration.sh gcs     # run only the GCS / fake-gcs-server suite

# Shell script examples (no build required — uses installed nixcopy)
bash examples/shellscript/01-local-to-local.sh            # local copy, no credentials needed
bash examples/shellscript/10-advanced-options.sh          # resume, skip-existing, bandwidth limit
bash examples/shellscript/11-s3-to-blob.sh                # S3 ↔ Azure Blob (set env vars first)
bash examples/shellscript/12-performance-tuning.sh        # buffer size, concurrency, upload-concurrency tuning

# Clean
make clean              # build artifacts (bin/)
make clean-dist         # release artifacts (dist/)

# Release (GoReleaser)
make goreleaser-check     # validate .goreleaser.yaml
make goreleaser-snapshot  # local test build — no publish, no tag required
git tag v1.x.x && git push origin v1.x.x  # trigger full release in CI
```

See [BUILD.md](docs/development/build.md) for release build flags, binary size benchmarks, and CI/CD integration examples.

## CI Pitfalls

Three non-obvious issues that have already burned us — don't repeat them:

1. **golangci-lint `install-mode: goinstall`** (`.github/workflows/ci.yml`): golangci-lint pre-built binaries are compiled with the Go version available at release time. When `go.mod` targets a newer Go minor version (e.g. `go 1.25`), the binary refuses to run with *"Go language version used to build golangci-lint is lower than targeted"*. The CI uses `install-mode: goinstall` so golangci-lint is compiled from source with the currently-installed Go toolchain. Do not change this back to `binary`.

2. **Never commit the `toolchain` directive in `go.mod`**: `go mod tidy` auto-inserts `toolchain goX.Y.Z` based on the developer's local patch version (e.g. `toolchain go1.25.4`). CI runners typically only have the `.0` patch release (`go1.25.0`), causing *"version go1.25.4 does not match go tool version go1.25.0"*. After running `go mod tidy` locally, remove the `toolchain` line before committing.

3. **Dockerfile builder image must match `go.mod` Go version**: The builder stage (`golang:X.Y-alpineZ`) must be kept in sync with the `go` directive in `go.mod`. A mismatch causes `go mod download` to fail with *"go.mod requires go >= X.Y (running go A.B)"*.

## Distribution Channels

Releases are fully automated via GoReleaser (`.goreleaser.yaml`) triggered by a `v*` tag push on `main`.

| Channel | How users install | Secret required |
|---|---|---|
| GitHub Releases | Download binary from releases page | — |
| Homebrew | `brew tap preedep/tap && brew install nixcopy` | `HOMEBREW_TAP_TOKEN` |
| APT (Debian/Ubuntu) | `apt-get install nixcopy` via Fury.io | `FURY_TOKEN` |
| YUM (RHEL/CentOS) | `yum install nixcopy` via Fury.io | `FURY_TOKEN` |
| Docker Hub | `docker pull nickmsft/gonixcopy:latest` | `DOCKERHUB_USERNAME/TOKEN` |

---

## Architecture

Clean Architecture — dependencies flow strictly inward:

```
cmd/nixcopy/main.go
    └─> internal/interfaces/cli/        ← Cobra commands, Viper config binding
            └─> internal/usecase/       ← orchestration, pattern matching, retry logic
                    └─> internal/domain/      ← entities + repository/service interfaces
            └─> internal/infrastructure/      ← storage impls, config loader, logger
```

### Domain (`internal/domain/`)
- **entity/** — `TransferConfig` (includes `VerifyChecksum`, `EnableResume`), `TransferResult` (includes `Checksum`, `ResumedFrom`), `FileInfo`, `TransferProgress`, `FilePattern`
- **repository/** — `Storage` interface (composes `StorageReader` + `StorageWriter`); optional `Resumer` interface (`ReadFrom` + `AppendWrite`) for resume support; no concrete implementations here
- **service/** — `TransferService` interface

### Use Case (`internal/usecase/`)
- `transfer_usecase.go` — single-file and batch transfer; decomposed into `Transfer` (orchestrator), `runWithRetry` (retry loop), `attemptTransfer` (single attempt), `buildPipeline` (I/O wiring); semaphore-based concurrency; SHA-256 checksum verification; resume from partial destination files; `transferBufPool` (256 KiB `sync.Pool`) used in the compression pipeline goroutine
- `pattern_matcher.go` — expands glob patterns (`*.pdf`, `**/*.log`) by calling `Storage.List` recursively
- `mocks/storage_mock.go` — `MockStorage` satisfies the `Storage` and `Resumer` interfaces; used in all unit tests

### Infrastructure (`internal/infrastructure/`)
- **storage/** — factory + one file per backend: `local.go`, `sftp.go`, `ftps.go`, `blob.go` (Azure), `s3.go` (AWS), `gcs.go` (GCP); `pool.go` provides a package-level 256 KiB `sync.Pool` used by `local.go` and `sftp.go` via `io.CopyBuffer` to eliminate per-transfer heap allocations
- **config/** — YAML/JSON loader with `${ENV_VAR}` expansion; precedence: CLI flags > env vars > config file > defaults
- **logger/** — `StandardLogger` in `applog.go` emits one JSON line per call conforming to [standard-app-log v1.0](https://github.com/preedep/standard-app-log). Log types: `APP_LOG`, `REQ_EX_LOG`, `RES_EX_LOG`. Default minimum level is `INFO`; set `WithMinLevel(LogLevelDebug)` to emit `DEBUG` entries. The CLI wires `--verbose` / `-v` to `WithMinLevel(LogLevelDebug)` automatically. Use `NewNopLogger()` in tests (replaces `zap.NewNop()`). Zap is still in `logger.go` as a dead stub; the CLI no longer calls it. Context injection at startup: `NIXCOPY_CORRELATION_ID`, `NIXCOPY_APP_ID`, `NIXCOPY_APP_VERSION`, `POD_NAME` env vars.

### CLI (`internal/interfaces/cli/`)
- `root.go` — global `--config` / `--verbose` flags; `--verbose` / `-v` sets the logger to `DEBUG` level, emitting resolved config, per-file queued paths, per-attempt error detail, and connection diagnostics
- `transfer.go` — `transfer` subcommand; mirrors `TransferConfig` with ~50 source/dest flags; calls `expandTransferPaths()` before using `--source`, `--sources`, and `--dest` so scripts can pass `${PWD}` or other env vars in path arguments; calls `resolveDestPath()` for single-file transfers — if `--dest` ends with `/` the source filename is appended automatically (same as Unix `cp file /dir/`)
- `list.go` — `list` subcommand for browsing storage
- `flags.go` — shared flag definitions and parsing helpers; `os.ExpandEnv` applied to all file-path flags (`--source-private-key`, `--dest-private-key`, `--source-credentials-file`, `--dest-credentials-file`) inside `applyCliFlags`

**Flag description convention**: every flag that has a corresponding `NIXCOPY_*` env var must include `(env: NIXCOPY_VAR)` at the end of its description string. The canonical mapping is in `internal/infrastructure/config/envloader.go`.

**Settings that have no CLI flag** — must be set via env var or config file:

| Setting | Env var | CLI flag |
|---|---|---|
| S3 / GCS / Blob custom endpoint | `NIXCOPY_SOURCE_ENDPOINT` / `NIXCOPY_DEST_ENDPOINT` | none |
| S3 path-style addressing (MinIO) | `NIXCOPY_SOURCE_USE_PATH_STYLE` / `NIXCOPY_DEST_USE_PATH_STYLE` | none |
| S3 named profile | `NIXCOPY_SOURCE_PROFILE` / `NIXCOPY_DEST_PROFILE` | none |
| Azure Blob connection string | `NIXCOPY_SOURCE_CONNECTION_STRING` / `NIXCOPY_DEST_CONNECTION_STRING` | none |
| SHA-256 checksum verification | `NIXCOPY_VERIFY_CHECKSUM=true` | none |
| Retry delay | `NIXCOPY_RETRY_DELAY` | none |

Shell scripts must use env-var prefixing for these: `NIXCOPY_DEST_CONNECTION_STRING="$CONN" nixcopy transfer ...`

**Blob `connection_string` auth** does not require `--dest-account-name` / `--source-account-name` — the account name is embedded in the connection string. The CLI validator skips the account-name check when `auth_type == connection_string`.

---

## Adding a New Storage Backend

Follow these three steps (see [CONTRIBUTING.md](docs/development/contributing.md) for a full code template):

1. **Implement** `repository.Storage` in `internal/infrastructure/storage/<name>.go`
2. **Add config struct** in `internal/infrastructure/config/config.go`
3. **Register** the new type in `internal/infrastructure/storage/factory.go`

Then write unit tests using `MockStorage` as a reference and add an entry to `examples/` (YAML config) and `examples/shellscript/` (shell script).

**Optionally** implement `repository.Resumer` (`ReadFrom` + `AppendWrite`) to enable `--resume` support. The use case detects this via type assertion at runtime — backends that don't implement it fall back to full re-transfer silently.

**For backends with multipart/block upload** (like S3 and Blob): add an `UploadConcurrency int` field to the backend's config struct and honour it in `Write`. Wire it from `--upload-concurrency` CLI flag via `applyCliFlags` in `flags.go`. Use the package-level `copyBufPool` from `storage/pool.go` in `Write` and `AppendWrite` instead of plain `io.Copy`.

---

## Commit Message Format

```
type: subject

body (optional)
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example: `feat: add Google Cloud Storage support`

---

## Configuration & Credentials

Config file structure: `config.example.yaml`. Use `${ENV_VAR}` syntax inside YAML for secrets:

```yaml
source:
  sftp:
    password: ${SFTP_PASSWORD}
destination:
  s3:
    access_key_id: ${AWS_ACCESS_KEY_ID}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY}
```

Protect config files: `chmod 600 config.yaml`.

For full CLI flag reference and precedence rules, see [CLI_USAGE.md](docs/guides/cli-usage.md).

---

## Authentication — Choose by Environment

Auth methods that require cloud metadata (IAM roles, Managed Identity) only work on the matching infrastructure. Mixing them silently fails.

| Environment | AWS S3 auth | Azure Blob auth | GCS auth |
|---|---|---|---|
| AWS EC2 / ECS / Lambda | `iam_role` ✅ recommended | `shared_key` / `sas_token` / `service_principal` | `service_account` |
| Azure VM / App Service | `access_key` | `managed_identity` ✅ recommended | `service_account` |
| GCE VM | `access_key` | `shared_key` / `sas_token` | `application_default` ✅ recommended |
| Amazon EKS | `web_identity` (IRSA) ✅ | `service_principal` | `service_account` |
| Azure AKS | `access_key` | `managed_identity` (Workload Identity) ✅ | `service_account` |
| GKE | `access_key` | `shared_key` / `sas_token` | `application_default` (Workload Identity) ✅ |
| On-premise / local | `access_key` | `shared_key` / `sas_token` | `service_account` / `access_token` |

For detailed setup steps, cross-account and Kubernetes scenarios, see [AUTHENTICATION.md](docs/guides/authentication.md) and [ENVIRONMENT_GUIDE.md](docs/guides/environment-guide.md).

---

## Performance Tuning

Memory is bounded by `bufferSize × concurrentFiles`. Tune all three levers together:

| File size | `--buffer-size` | `--concurrent-files` | `--upload-concurrency` |
|---|---|---|---|
| < 10 MB | 4–8 MB | 8–16 | 5 (default) |
| 10–100 MB | 32–64 MB | 4–8 | 5–8 |
| > 100 MB | 64–128 MB | 2–4 | 8–10 |

- **`--upload-concurrency`** controls how many parts are uploaded in parallel *within a single file* for S3 and Azure Blob. Has no effect on local or SFTP. Default is 5; raise to 8–10 on high-bandwidth links for large files.
- **`--buffer-size`** is in bytes: `4194304` = 4 MB, `33554432` = 32 MB, `134217728` = 128 MB.
- On an unstable network, increase `--retry-attempts` before increasing concurrency.
- `--retry-delay` is config-file / `NIXCOPY_RETRY_DELAY` env var only — no CLI flag.

**Implementation note:** local and SFTP backends use a shared `sync.Pool` of 256 KiB buffers (`storage/pool.go`) via `io.CopyBuffer`, eliminating per-transfer heap allocations. The compression pipeline goroutine uses a separate pool (`transferBufPool` in `transfer_usecase.go`).

See `examples/shellscript/12-performance-tuning.sh` for ready-to-run tuning examples.

---

## Wildcard Pattern Syntax

Always quote patterns in the shell to prevent shell expansion:

| Pattern | Matches |
|---|---|
| `*.pdf` | All `.pdf` files in the current directory |
| `report*.xlsx` | Files starting with `report` |
| `**/*.log` | All `.log` files recursively |
| `data/2024/**/*.csv` | All `.csv` files under `data/2024/` recursively |

Use `nixcopy list` to validate a pattern before a real transfer:

```bash
nixcopy list -c config.yaml -p "reports/*.pdf" --source
```

See [PARALLEL_TRANSFER.md](docs/guides/parallel-transfer.md) for batch, multi-pattern, and scripting examples.

---

## Testing

Unit tests use `MockStorage` (no real credentials needed). See [TESTING.md](docs/development/testing.md) for full examples.

CLI flag tests (`internal/interfaces/cli/flags_test.go`) operate on package-level flag vars. Call `resetTransferFlags()` at the start of every CLI test to prevent state bleed between tests.

```bash
go test ./internal/usecase/... -run TestPatternMatcher   # specific test
go test -race ./...                                       # race detector
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### Integration Tests

Integration tests use the `//go:build integration` tag and require real service containers. There are two ways to run them:

**Docker Compose (recommended for local dev)** — single command, tears down automatically:
```bash
make integration-test-local
```
Starts: SFTP (port 2222), FTPS/pure-ftpd (port 21, explicit TLS, self-signed cert), MinIO/S3 (port 9000), Azurite/Azure Blob (port 10000), fake-gcs-server/GCS (port 4443). All env vars are set automatically. Requires `az` CLI on the host for Azurite container creation (`brew install azure-cli`).

**Shell script (CI-style, individual docker run)** — matches the CI workflow:
```bash
./test-integration.sh               # all suites
./test-integration.sh ftps          # FTPS only
./test-integration.sh sftp s3       # specific suites
./test-integration.sh blob gcs      # Azure Blob + GCS only
./test-integration.sh -v --no-clean # verbose, keep containers
```

Integration test files live alongside unit tests, gated by `-tags=integration`. Each file's helper skips automatically when its required env vars are absent, so `go test ./...` (no tag) always runs cleanly.

| File | Suite | Skip guard |
|---|---|---|
| `local_integration_test.go` | LocalStorage | never skipped (uses t.TempDir) |
| `sftp_integration_test.go` | SFTPStorage | `SFTP_HOST` |
| `ftps_integration_test.go` | FTPSStorage | `FTPS_HOST` |
| `s3_integration_test.go` | S3Storage | `S3_ENDPOINT` |
| `blob_integration_test.go` | BlobStorage | `BLOB_CONNECTION_STRING` |
| `gcs_integration_test.go` | GCSStorage | `GCS_ENDPOINT` |

### Benchmarks

Benchmarks cover local transfer (1 MB–128 MB), batch concurrency, buffer pool throughput, pattern matching, and SHA-256 checksum throughput. No credentials or build tags required.

```bash
go test -bench=. -benchmem ./internal/usecase/           # transfer + batch + buffer pool benchmarks
go test -bench=. -benchmem ./internal/infrastructure/storage/  # (future storage-level benchmarks)
```

Run results are machine-specific — do not commit numbers to docs. Use the benchmarks to validate tuning changes locally.

| Benchmark | What it measures |
|---|---|
| `BenchmarkTransfer` | Single-file local transfer at 1 MB / 16 MB / 128 MB |
| `BenchmarkTransfer_WithChecksum` | SHA-256 overhead vs baseline |
| `BenchmarkTransferBatch` | Batch throughput at concurrency 1 / 4 / 8 |
| `BenchmarkTransfer_BufferPool` | Buffer pool path at 1 MB / 16 MB / 64 MB |

---

## Code Comments

- **Exported types and functions**: one-line godoc only (`// Foo does X`). No multi-paragraph blocks, no `Parameters:` / `Returns:` / `Example:` sections.
- **Inline comments**: only when the *why* is non-obvious — a hidden constraint, a subtle invariant, or a workaround for a specific bug. If removing the comment wouldn't confuse a future reader, don't write it.
- **Never** explain what the code does; well-named identifiers already do that.

The existing files (`local.go`, `transfer_usecase.go`, `pattern_matcher.go`) predate this rule and are over-commented — treat them as legacy, not as the style to follow.

---

## Go Patterns

Patterns established in this codebase. Follow them when adding or modifying code.

### Interface injection for testability

Never call package-level constructors directly inside functions that need to be unit-tested. Wrap them behind an interface stored in a package-level variable so tests can swap the implementation:

```go
// production code
type storageFactory interface {
    NewSource(cfg *config.SourceConfig) (repository.Storage, error)
    NewDest(cfg *config.DestinationConfig) (repository.Storage, error)
}

type realStorageFactory struct{}
func (realStorageFactory) NewSource(cfg *config.SourceConfig) (repository.Storage, error) {
    return storage.NewStorageFromSourceConfig(cfg)
}

var activeStorageFactory storageFactory = realStorageFactory{}

// test code
type mockFactory struct{ src, dst *mocks.MockStorage }
func (f *mockFactory) NewSource(_ *config.SourceConfig) (repository.Storage, error) { return f.src, nil }
func (f *mockFactory) NewDest(_ *config.DestinationConfig) (repository.Storage, error) { return f.dst, nil }

func withMockFactory(t *testing.T, src, dst *mocks.MockStorage) {
    orig := activeStorageFactory
    activeStorageFactory = &mockFactory{src: src, dst: dst}
    t.Cleanup(func() { activeStorageFactory = orig })
}
```

Applied in: `internal/interfaces/cli/transfer.go` (`activeStorageFactory`).

### t.Cleanup over defer in test helpers

Use `t.Cleanup` (not `defer`) to register teardown in test helpers. `defer` runs at the end of the helper function, not the test; `t.Cleanup` runs at the end of the test:

```go
// correct
func withMockFactory(t *testing.T, ...) {
    orig := activeStorageFactory
    activeStorageFactory = ...
    t.Cleanup(func() { activeStorageFactory = orig })
}

// wrong — restores too early
func withMockFactory(t *testing.T, ...) {
    orig := activeStorageFactory
    activeStorageFactory = ...
    defer func() { activeStorageFactory = orig }()  // runs when helper returns, not when test ends
}
```

### resetXxxFlags() in CLI tests

Package-level flag vars bleed state between tests because `init()` sets them once at startup. Every CLI test must call `resetTransferFlags()` as its first line:

```go
func TestRunTransfer_Something(t *testing.T) {
    resetTransferFlags()
    // ... set only the flags this test needs
}
```

Never add state cleanup at the end of a test — use `t.Cleanup` or `resetXxx` at the start.

### Verify mock state, not output

CLI tests check observable side effects on the mock (`.WriteCalled`, `.FileContent[path]`) rather than parsing stdout/stderr. This keeps tests fast and decoupled from formatting:

```go
// correct
if !dst.WriteCalled {
    t.Error("expected dest.Write to be called")
}
if got := dst.FileContent["/dst/out.txt"]; string(got) != string(want) {
    t.Errorf("content = %q, want %q", got, want)
}

// avoid — brittle, slow, tied to output format
out := captureStdout(...)
json.Unmarshal(out, &summary)
```

### Decompose long functions into named steps

When a function exceeds ~50 lines or mixes more than two concerns, extract named private methods rather than adding more inline blocks. Name each method after what it does, not when it runs:

```go
// good — each method name is self-documenting
func (t *TransferUseCase) Transfer(...) { ... }          // orchestrator, ~50 lines
func (t *TransferUseCase) runWithRetry(...) error { ... } // retry loop
func (t *TransferUseCase) attemptTransfer(...) error { ... } // one attempt
func (t *TransferUseCase) buildPipeline(...) (io.Reader, ...) { ... } // I/O wiring

// avoid — one giant function with section comments
func (t *TransferUseCase) Transfer(...) {
    // --- retry loop ---
    // --- build pipeline ---
    // --- verify checksum ---
}
```

Applied in: `internal/usecase/transfer_usecase.go`.

### Deduplicate symmetric flag registration

When source and destination share identical flag sets, extract a helper that takes a prefix string and a struct of pointers. This keeps `init()` short and makes it trivial to add a flag to both sides at once:

```go
type storageVarSet struct {
    storageType *string
    host        *string
    // ...
}

func registerStorageFlags(cmd *cobra.Command, prefix string, v *storageVarSet) {
    up := strings.ToUpper(prefix)
    cmd.Flags().StringVar(v.storageType, prefix+"-type", "", "... (env: NIXCOPY_"+up+"_TYPE)")
    // ...
}

func init() {
    registerStorageFlags(transferCmd, "source", &storageVarSet{storageType: &sourceType, ...})
    registerStorageFlags(transferCmd, "dest",   &storageVarSet{storageType: &destType,   ...})
}
```

Applied in: `internal/interfaces/cli/transfer.go`.
