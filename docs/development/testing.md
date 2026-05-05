# Testing Guide — go-nixcopy

## Running Tests

### Match CI exactly

```bash
go test -race -covermode=atomic -count=1 ./...
```

### Coverage report

```bash
go test -race -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Verbose / single test

```bash
go test -v ./internal/usecase/...
go test ./internal/usecase/... -run TestTransferUseCase_Transfer_Success
```

### Integration tests (requires Docker — SFTP + FTPS + MinIO)

The recommended way is a single Make target that spins up all containers, runs the full suite, and tears everything down automatically:

```bash
make integration-test-local
```

To run only a specific suite, or to keep containers running after tests:

```bash
./test-integration.sh ftps            # FTPS only
./test-integration.sh sftp s3         # SFTP + S3/MinIO
./test-integration.sh -v --no-clean   # verbose, keep containers
```

To run the Go test command directly against already-running containers:

```bash
S3_ENDPOINT=http://localhost:9000 S3_ACCESS_KEY=minioadmin S3_SECRET_KEY=minioadmin \
S3_BUCKET=test-bucket S3_REGION=us-east-1 \
SFTP_HOST=localhost SFTP_PORT=2222 SFTP_USERNAME=testuser SFTP_PASSWORD=testpass \
FTPS_HOST=localhost FTPS_PORT=21 FTPS_USERNAME=testuser FTPS_PASSWORD=testpass \
FTPS_TLS_MODE=explicit FTPS_SKIP_VERIFY=true \
go test -tags=integration -race -count=1 -v ./internal/infrastructure/storage/...
```

### Makefile shortcuts

```bash
make test              # all unit tests
make test-verbose      # -v flag
make test-coverage     # generates coverage.html
make test-race         # race detector
```

---

## Test File Map

```
internal/
├── domain/entity/
│   ├── pattern_test.go              — FilePattern wildcard detection, Match/MatchFull edge cases
│   │                                  (invalid patterns, multiple **, prefix mismatch) → 100% coverage
│   └── transfer_test.go             — TransferConfig / TransferResult fields
├── infrastructure/config/
│   ├── config_test.go               — DefaultConfig values, StorageType/S3AuthType/BlobAuthType/GCSAuthType
│   │                                  constants, SFTPConfig/S3Config/GCSConfig/BlobConfig struct fields
│   ├── envloader_test.go            — NIXCOPY_* env var loading (all storage types + transfer flags),
│   │                                  ApplyBackendEnv idempotency, env-overrides-config-file precedence
│   └── yamlloader_test.go           — LoadFile: all six storage backends (SFTP/FTPS/S3/Blob/GCS/local),
│                                      all auth subtypes (shared_key, sas_token, connection_string,
│                                      service_principal, managed_identity, access_key, web_identity,
│                                      assume_role, service_account, impersonate, access_token,
│                                      application_default), transfer settings (all fields including
│                                      duration strings), logging config, ${ENV_VAR} expansion,
│                                      defaults preserved when sections omitted, partial override,
│                                      file-not-found error, malformed YAML error, config.example.yaml
├── infrastructure/logger/
│   └── applog_test.go               — all log methods, field helpers, WithCorrelation/WithRequest/WithTrace,
│                                      NopLogger, GenerateID, extra fields, parent immutability,
│                                      WithMinLevel filtering (debug filtered at INFO, all levels at DEBUG,
│                                      child logger inherits minLevel, InfoReqEx passes at INFO)
├── infrastructure/storage/
│   ├── blob_auth_test.go            — BlobStorage.Connect auth error paths
│   ├── blob_multipart_test.go       — blobBlockSizeFor: 5 size scenarios
│   ├── blob_nil_client_test.go      — nil-client guards: List/Read/Stat/Write/Delete; Disconnect/CreateDirectory no-ops
│   ├── factory_test.go              — NewStorageFromSourceConfig + NewStorageFromDestConfig: all 6 backends,
│                                      nil-config errors, unknown type errors
│   ├── ftps_ext_test.go             — FTPS nil-client guards: List/Read/Stat/Delete/Disconnect/Write
│   ├── ftps_integration_test.go     — FTPS WriteAndRead/Stat/List/Delete/CreateDirectory (integration tag, requires FTPS_HOST)
│   ├── ftps_nil_client_test.go      — FTPS nil-client guards: List/Read/Stat/Delete/Disconnect/Write
│   ├── ftps_test.go                 — ftpMkdirAll: nested paths, single component, errors ignored, deep path
│   ├── gcs_auth_test.go             — GCSStorage.Connect: all auth types, missing credentials, unknown type, endpoint option
│   ├── gcs_nil_client_test.go       — GCS nil-client guards: List/Read/Stat/Write/Delete; Disconnect/CreateDirectory no-ops
│   ├── local_integration_test.go    — Local storage read/write/list (integration tag)
│   ├── local_unit_test.go           — LocalStorage: Connect, Disconnect, List, Read, Write, Delete, Stat,
│                                      CreateDirectory, ReadFrom (with offset), AppendWrite — no external deps
│   ├── s3_auth_test.go              — S3Storage.Connect auth error paths
│   ├── s3_integration_test.go       — S3 read/write/list against MinIO (integration tag)
│   ├── s3_multipart_test.go         — s3PartSizeFor: 5 size scenarios
│   ├── s3_nil_client_test.go        — nil-client guards: List/Read/Stat/Write/Delete; Disconnect/CreateDirectory no-ops
│   ├── sftp_auth_test.go            — SFTPStorage.Connect auth error paths (no auth, password, private key, passphrase)
│   ├── sftp_integration_test.go     — SFTP WriteAndRead/List/Delete/ReadFrom/AppendWrite/Resume (integration tag)
│   └── sftp_nil_client_test.go      — SFTP nil-client guards: List/Read/Stat/Write/Delete/CreateDirectory,
│                                      ReadFrom/AppendWrite (Resumer), Disconnect no-op
├── interfaces/cli/
│   ├── flags_test.go                — applyCliFlags (SFTP/FTPS/S3/Blob/local source+dest, private key, all transfer flags),
│                                      validateConfig, --skip-existing, --resume, CLI-over-env precedence;
│                                      ${ENV_VAR} expansion: private key paths, credentials file paths,
│                                      expandTransferPaths (--source, --sources slice, --dest, ${PWD})
│   ├── ftps_tls_mode_test.go        — --source-tls-mode / --dest-tls-mode: explicit, implicit, empty-no-override,
│                                      invalid mode rejected, env var path, CLI-over-env precedence
│   ├── transfer_summary_test.go     — transferSummary JSON shape, omitempty, failed_files
│   └── validate_format_test.go      — validateConfig all source/dest backend error paths, compression validation,
│                                      formatSize boundary cases
└── usecase/
    ├── compress_test.go             — gzip/zstd round-trips, passthrough, ratio, invalid algo (6 tests)
    ├── pattern_matcher_test.go      — glob expansion, recursive **, no-match behaviour
    ├── throttle_test.go             — ParseBandwidth formats, passthrough, data integrity, context cancel (4 tests)
    ├── transfer_usecase_branches_test.go — progressReader.Close (Closer/non-Closer), resume+compression warning,
    │                                  checksum skipped when resumed, checksum skipped when compressed,
    │                                  ReadFrom error exhausting retries
    ├── transfer_usecase_skip_test.go — SkipExisting: 5 cases including batch
    ├── transfer_usecase_verbose_test.go — verbose/non-verbose log output: success emits "transfer attempt" DEBUG,
    │                                  failure emits per-attempt DEBUG + always-on ERROR; DEBUG absent at INFO level
    └── transfer_usecase_test.go     — success, checksum, resume, retry, batch partial failure
```

---

## Key Test Patterns

### MockStorage

`mocks.NewMockStorage()` implements both `repository.Storage` and `repository.Resumer`:

```go
source := mocks.NewMockStorage()
source.AddFile("/src/file.txt", []byte("content"), &entity.FileInfo{
    Path: "/src/file.txt", Name: "file.txt", Size: 7,
})

// Trigger a checksum mismatch
dest.CorruptWrite = true

// Verify Write was (or wasn't) called
if dest.WriteCalled { ... }
```

### Verbose logging tests

Capture structured JSON log output with `WithOutput` and `WithMinLevel`:

```go
var buf bytes.Buffer
log := applog.NewStandardLogger(
    applog.WithOutput(&buf),
    applog.WithMinLevel(applog.LogLevelDebug), // verbose
)

dest.WriteError = errors.New("disk full")
cfg := &entity.TransferConfig{BufferSize: 1024, RetryAttempts: 1, RetryDelay: 0}
uc := NewTransferUseCase(src, dest, cfg, log)
uc.Transfer(ctx, "/src/file.txt", "/dst/file.txt", nil)

// Parse each newline-terminated JSON line from buf, then assert:
// - "attempt failed" DEBUG appears once per failed attempt
// - "Transfer failed after all attempts" ERROR always appears
// - At default INFO level, "attempt failed" DEBUG is absent
```

### Skip-existing tests

```go
cfg := &entity.TransferConfig{SkipExisting: true, BufferSize: 1024}
uc := NewTransferUseCase(source, dest, cfg, applog.NewNopLogger())

result, _ := uc.Transfer(ctx, "/src/file.txt", "/dst/file.txt", nil)
// result.Status == entity.TransferStatusSkipped when sizes match
```

### Multipart sizing tests

```go
// s3PartSizeFor — table-driven
got := s3PartSizeFor(5 * 1024 * 1024 * 1024 * 1024) // 5 TiB
// got >= s3MinPartSize, ceil(size/got) <= s3MaxParts

// blobBlockSizeFor — same structure
got := blobBlockSizeFor(190 * 1024 * 1024 * 1024 * 1024) // 190 TiB
// got >= blobMinBlockSize, ceil(size/got) <= blobMaxBlocks
```

### JSON summary tests

```go
s := transferSummary{Event: "transfer_summary", TotalFiles: 3, ...}
var buf bytes.Buffer
json.NewEncoder(&buf).Encode(s)
// Assert field names, omitempty on average_speed_mbps and failed_files
```

### Bandwidth throttle tests

```go
// ParseBandwidth — table-driven
got, err := ParseBandwidth("10MB")   // 10 * 1024 * 1024
got, err  = ParseBandwidth("1GiB")   // 1 * 1024 * 1024 * 1024
got, err  = ParseBandwidth("0")      // 0 (unlimited)
got, err  = ParseBandwidth("invalid") // error

// ThrottledReader — context cancel
ctx, cancel := context.WithCancel(context.Background())
r := newThrottledReader(ctx, bytes.NewReader(data), 1) // 1 byte/s
cancel()
_, err = io.ReadAll(r) // returns context error quickly
```

### Compression tests

```go
// Round-trip: compress → decompress → compare
var buf bytes.Buffer
cw, _ := newCompressWriter(&buf, CompressionGzip)
cw.Write(original)
cw.Close()
gr, _ := gzip.NewReader(&buf)
decoded, _ := io.ReadAll(gr)
// bytes.Equal(decoded, original) == true

// Invalid algo
_, err := newCompressWriter(io.Discard, "bzip2") // error

// Compression ratio check (compressible input)
var compressed bytes.Buffer
cw, _ = newCompressWriter(&compressed, CompressionZstd)
cw.Write(make([]byte, 1024*1024)) // all zeros — highly compressible
cw.Close()
// compressed.Len() < 1024*1024
```

### YAML file loading tests

`LoadFile` is in the `config` package. Tests write a temp YAML, call `LoadFile`, and assert struct fields:

```go
func writeTempYAML(t *testing.T, content string) string { ... } // helper in yamlloader_test.go

path := writeTempYAML(t, `
source:
  type: sftp
  sftp:
    host: sftp.example.com
    port: 2222
    password: ${MY_SECRET}   # ${ENV_VAR} is expanded via os.ExpandEnv before parsing
destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-bucket
    auth_type: access_key
    access_key_id: AKIAIOSFODNN7EXAMPLE
    secret_access_key: wJalrXUtnFEMI
transfer:
  buffer_size: 67108864
  retry_delay: 10s    # Duration strings ("30s", "5m", "2h") parsed automatically
`)
cfg, err := config.LoadFile(path)
// cfg.Source.SFTP.Host == "sftp.example.com"
// cfg.Transfer.BufferSize == 67108864
// cfg.Transfer.RetryDelay == 10*time.Second
```

Fields absent from the YAML retain `DefaultConfig()` values. The `${ENV_VAR}` syntax is expanded before
YAML parsing, so secrets can be stored in environment variables and referenced by name in the config file.

### Env var tests

```go
t.Setenv("NIXCOPY_SKIP_EXISTING", "true")
t.Setenv("NIXCOPY_BANDWIDTH_LIMIT", "10485760")
t.Setenv("NIXCOPY_COMPRESSION", "gzip")
cfg := DefaultConfig()
LoadFromEnv(cfg)
// cfg.Transfer.SkipExisting == true
// cfg.Transfer.BandwidthLimit == 10485760
// cfg.Transfer.Compression == "gzip"
```

---

## Coverage by Package

| Package | Coverage | Notes |
|---|---|---|
| `internal/domain/entity` | **100%** | All pattern matching branches including error paths |
| `internal/infrastructure/config` | ~98% | YAML file loading, env var loading, all storage types and auth subtypes |
| `internal/usecase` | ~92% | Core transfer logic, all feature paths including progressReader.Close and warning branches |
| `internal/infrastructure/logger` | ~85% | All log methods, field helpers, child loggers, level filtering; `logger.go` Zap stub excluded (dead code) |
| `internal/interfaces/cli` | ~58% | Flag wiring, validateConfig, TLS mode; `runTransfer`/`runList` require real storage |
| `internal/infrastructure/storage` | ~49% | Local fully unit-tested; FTPS/Blob/S3/SFTP nil-client guards + auth paths; happy paths need integration tag |

Storage coverage in unit mode reflects the nil-client guard pattern — every backend's error paths are covered without credentials. Happy paths (List, Read, Write against real endpoints) are covered by integration tests (`-tags=integration`) against MinIO, SFTP, and FTPS containers (`make integration-test-local`).

---

## Integration Test Roadmap

Planned scenarios beyond per-backend happy paths. All require the Docker Compose stack (`make integration-test-local`).

| # | Scenario | File | Status |
|---|---|---|---|
| 1 | **Cross-backend transfers via TransferUseCase** — SFTP→S3, S3→SFTP, Local→SFTP, Local→S3, SFTP→FTPS, S3→FTPS | `transfer_integration_test.go` | ✅ done |
| 2 | **Checksum verification against real backends** — `VerifyChecksum: true`; assert `TransferResult.Checksum` is non-empty SHA-256 hex | `transfer_integration_test.go` | ✅ done |
| 3 | **SkipExisting across backends** — write a file to dest, run Transfer again, assert `TransferStatusSkipped` | `transfer_integration_test.go` | ✅ done |
| 4 | **Resume across backends** — write partial dest, resume, assert full content and `ResumedFrom > 0` | `transfer_integration_test.go` | ✅ done |
| 5 | **Batch transfer** — `TransferBatch` with 5+ files across SFTP→S3 and S3→SFTP | `transfer_integration_test.go` | ✅ done |
| 6 | **Bandwidth throttle end-to-end** — `BandwidthLimit` set to 1 MB/s; assert transfer completes and duration is within expected range | `transfer_integration_test.go` | ✅ done |
