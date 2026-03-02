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

## Make Commands

```bash
make test              # รัน tests
make test-coverage     # รัน tests พร้อม coverage
```
