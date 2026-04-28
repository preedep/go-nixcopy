# คู่มือการทดสอบ - go-nixcopy

## การรัน Tests

### รัน All Tests
```bash
go test ./...
```

### รัน Tests พร้อม Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### รัน Specific Package
```bash
go test ./internal/domain/entity
go test ./internal/usecase
```

### รัน Verbose Mode
```bash
go test -v ./...
```

## Test Structure

```
internal/
├── domain/entity/
│   ├── pattern_test.go
│   └── transfer_test.go
├── usecase/
│   ├── transfer_usecase_test.go
│   ├── pattern_matcher_test.go
│   └── mocks/
│       └── storage_mock.go
├── infrastructure/config/
│   └── config_test.go
└── interfaces/cli/
    └── flags_test.go
```

## ตัวอย่าง Tests

### Domain Entity Tests
```go
func TestNewFilePattern(t *testing.T) {
    fp := NewFilePattern("*.pdf")
    if !fp.IsWildcard {
        t.Error("Should be wildcard")
    }
}
```

### Use Case Tests with Mocks
```go
func TestTransferUseCase_Transfer_Success(t *testing.T) {
    source := mocks.NewMockStorage()
    dest := mocks.NewMockStorage()
    
    // Setup test data
    source.AddFile("/test.txt", []byte("content"), &entity.FileInfo{...})
    
    // Execute
    result, err := useCase.Transfer(ctx, "/test.txt", "/dest.txt", nil)
    
    // Assert
    if err != nil {
        t.Fatal(err)
    }
}
```

### Checksum Verification Tests

```go
// TestTransferUseCase_Transfer_ChecksumVerification_Success
// Verifies that result.Checksum contains the correct SHA-256 hex when VerifyChecksum: true

// TestTransferUseCase_Transfer_ChecksumVerification_Mismatch
// Uses dest.CorruptWrite = true to flip a byte; expects transfer to fail with checksum error
```

### Resume Tests

```go
// TestTransferUseCase_Transfer_Resume_WithPartialDest
// Pre-seeds destination with first N bytes; verifies transfer resumes from that offset
// and result.ResumedFrom == N

// TestTransferUseCase_Transfer_Resume_NoPartialDest
// EnableResume: true but no partial file at dest; verifies full transfer runs normally
```

### MockStorage Capabilities

`MockStorage` implements both `repository.Storage` and `repository.Resumer`:
- `CorruptWrite bool` — flips first byte of written content (checksum mismatch tests)
- `ReadFrom(offset)` — returns content slice starting at offset
- `AppendWrite(offset)` — merges new content at offset into existing `FileContent`

## Make Commands

```bash
make test              # รัน tests
make test-coverage     # รัน tests พร้อม coverage
```
