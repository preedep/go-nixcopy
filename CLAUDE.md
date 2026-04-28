# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

# Clean
make clean              # build artifacts (bin/)
make clean-dist         # release artifacts (dist/)
```

See [BUILD.md](BUILD.md) for release build flags, binary size benchmarks, and CI/CD integration examples.

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
- `transfer_usecase.go` — single-file and batch transfer; semaphore-based concurrency, non-blocking progress channel updates, configurable retry with backoff; SHA-256 checksum verification; resume from partial destination files
- `pattern_matcher.go` — expands glob patterns (`*.pdf`, `**/*.log`) by calling `Storage.List` recursively
- `mocks/storage_mock.go` — `MockStorage` satisfies the `Storage` and `Resumer` interfaces; used in all unit tests

### Infrastructure (`internal/infrastructure/`)
- **storage/** — factory + one file per backend: `local.go`, `sftp.go`, `ftps.go`, `blob.go` (Azure), `s3.go` (AWS)
- **config/** — YAML/JSON loader with `${ENV_VAR}` expansion; precedence: CLI flags > env vars > config file > defaults
- **logger/** — Zap wrapper; JSON or console output, level-configurable

### CLI (`internal/interfaces/cli/`)
- `root.go` — global `--config` / `--verbose` flags
- `transfer.go` — `transfer` subcommand; mirrors `TransferConfig` with ~50 source/dest flags
- `list.go` — `list` subcommand for browsing storage
- `flags.go` — shared flag definitions and parsing helpers

---

## Adding a New Storage Backend

Follow these three steps (see [CONTRIBUTING.md](CONTRIBUTING.md) for a full code template):

1. **Implement** `repository.Storage` in `internal/infrastructure/storage/<name>.go`
2. **Add config struct** in `internal/infrastructure/config/config.go`
3. **Register** the new type in `internal/infrastructure/storage/factory.go`

Then write unit tests using `MockStorage` as a reference and add an entry to `examples/`.

**Optionally** implement `repository.Resumer` (`ReadFrom` + `AppendWrite`) to enable `--resume` support. The use case detects this via type assertion at runtime — backends that don't implement it fall back to full re-transfer silently.

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

For full CLI flag reference and precedence rules, see [CLI_USAGE.md](CLI_USAGE.md).

---

## Authentication — Choose by Environment

Auth methods that require cloud metadata (IAM roles, Managed Identity) only work on the matching infrastructure. Mixing them silently fails.

| Environment | AWS S3 auth | Azure Blob auth |
|---|---|---|
| AWS EC2 / ECS / Lambda | `iam_role` ✅ recommended | `shared_key` / `sas_token` / `service_principal` |
| Azure VM / App Service | `access_key` | `managed_identity` ✅ recommended |
| Amazon EKS | `web_identity` (IRSA) ✅ | `service_principal` |
| Azure AKS | `access_key` | `managed_identity` (Workload Identity) ✅ |
| On-premise / local | `access_key` | `shared_key` / `sas_token` |

For detailed setup steps, cross-account and Kubernetes scenarios, see [AUTHENTICATION.md](AUTHENTICATION.md) and [ENVIRONMENT_GUIDE.md](ENVIRONMENT_GUIDE.md).

---

## Performance Tuning

Memory is bounded by `bufferSize × concurrentFiles`. Tune together:

| File size | `buffer_size` | `concurrent_files` |
|---|---|---|
| < 10 MB | 8–16 MB | 8–16 |
| 10–100 MB | 32–64 MB | 4–8 |
| > 100 MB | 64–128 MB | 2–4 |

On an unstable network, increase `--retry-attempts` and `--retry-delay` before increasing concurrency.

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

See [PARALLEL_TRANSFER.md](PARALLEL_TRANSFER.md) for batch, multi-pattern, and scripting examples.

---

## Testing

Unit tests use `MockStorage` (no real credentials needed). See [TESTING.md](TESTING.md) for full examples.

```bash
go test ./internal/usecase/... -run TestPatternMatcher   # specific test
go test -race ./...                                       # race detector
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```
