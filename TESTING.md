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

### Integration tests (requires MinIO + SFTP)

```bash
# Start dependencies
docker run -d --name minio -p 9000:9000 \
  -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data

docker run -d --name sftp -p 2222:22 \
  atmoz/sftp testuser:testpass:::upload

# Create test bucket
docker run --rm --network host \
  -e MC_HOST_local=http://minioadmin:minioadmin@localhost:9000 \
  minio/mc mb local/test-bucket

# Run
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
│   ├── pattern_test.go              — FilePattern wildcard detection
│   └── transfer_test.go             — TransferConfig / TransferResult fields
├── infrastructure/config/
│   ├── config_test.go               — YAML load, ${ENV_VAR} expansion, defaults
│   └── envloader_test.go            — NIXCOPY_* env var loading (all storage types + transfer flags)
├── infrastructure/storage/
│   ├── blob_auth_test.go            — BlobStorage.Connect auth error paths
│   ├── blob_multipart_test.go       — blobBlockSizeFor: 5 size scenarios
│   ├── ftps_ext_test.go             — FTPS directory creation helpers
│   ├── ftps_test.go                 — FTPS config validation
│   ├── local_integration_test.go    — Local storage read/write/list (integration tag)
│   ├── s3_auth_test.go              — S3Storage.Connect auth error paths
│   ├── s3_integration_test.go       — S3 read/write/list against MinIO (integration tag)
│   ├── s3_multipart_test.go         — s3PartSizeFor: 5 size scenarios
│   └── sftp_integration_test.go     — SFTP read/write/list (integration tag)
├── interfaces/cli/
│   ├── flags_test.go                — applyCliFlags, validateConfig, --skip-existing, --resume
│   └── transfer_summary_test.go     — transferSummary JSON shape, omitempty, failed_files
└── usecase/
    ├── compress_test.go             — gzip/zstd round-trips, passthrough, ratio, invalid algo (6 tests)
    ├── pattern_matcher_test.go      — glob expansion, recursive **, no-match behaviour
    ├── throttle_test.go             — ParseBandwidth formats, passthrough, data integrity, context cancel (4 tests)
    ├── transfer_usecase_skip_test.go — SkipExisting: 5 cases including batch
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
| `internal/usecase` | ~91% | Core transfer logic, all feature paths |
| `internal/domain/entity` | ~85% | Entities and value objects |
| `internal/infrastructure/config` | ~52% | Config loading and env vars |
| `internal/interfaces/cli` | ~28% | CLI wiring; `runTransfer` requires real storage |
| `internal/infrastructure/storage` | ~16% | Auth error paths unit-tested; happy paths need integration tag |

Storage coverage is low in unit mode by design — the happy paths require real network endpoints and are covered by integration tests (`-tags=integration`) against MinIO and SFTP service containers.
