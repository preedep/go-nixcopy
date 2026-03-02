# คู่มือการมีส่วนร่วม (Contributing Guide)

ขอบคุณที่สนใจมีส่วนร่วมในโปรเจกต์ go-nixcopy! 🎉

## การเริ่มต้น

1. **Fork repository** นี้ไปยัง GitHub account ของคุณ
2. **Clone** fork ของคุณมาที่เครื่อง:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-nixcopy.git
   cd go-nixcopy
   ```
3. **ติดตั้ง dependencies**:
   ```bash
   make deps
   ```
4. **สร้าง branch ใหม่** สำหรับ feature ของคุณ:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## การพัฒนา

### Code Style

- ใช้ `gofmt` สำหรับ format โค้ด
- ปฏิบัติตาม [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- เขียน comments ที่ชัดเจนสำหรับ exported functions และ types
- ตั้งชื่อตัวแปรและ function ให้มีความหมาย

### Testing

- เขียน unit tests สำหรับ code ใหม่
- รัน tests ก่อน commit:
  ```bash
  make test
  ```
- ตรวจสอบ test coverage:
  ```bash
  make test-coverage
  ```

### Commit Messages

ใช้รูปแบบ commit message ที่ชัดเจน:

```
type: subject

body (optional)

footer (optional)
```

**Types:**
- `feat`: feature ใหม่
- `fix`: bug fix
- `docs`: เปลี่ยนแปลง documentation
- `style`: การเปลี่ยนแปลงที่ไม่กระทบ code logic (formatting, etc.)
- `refactor`: refactor code
- `test`: เพิ่มหรือแก้ไข tests
- `chore`: งานอื่นๆ (dependencies, build, etc.)

**ตัวอย่าง:**
```
feat: add support for Google Cloud Storage

Implement GCS storage adapter with streaming support.
Includes read, write, list, and delete operations.

Closes #123
```

## Pull Request Process

1. **Update documentation** ถ้าจำเป็น
2. **เพิ่ม tests** สำหรับ code ใหม่
3. **รัน tests และ linter**:
   ```bash
   make test
   make lint
   make fmt
   ```
4. **Update CHANGELOG.md** (ถ้ามี)
5. **Push** branch ของคุณ:
   ```bash
   git push origin feature/your-feature-name
   ```
6. **เปิด Pull Request** บน GitHub
7. **รอ review** และตอบคำถามหรือแก้ไขตาม feedback

## Architecture Guidelines

โปรเจกต์นี้ใช้ Clean Architecture:

### Domain Layer (`internal/domain/`)
- กำหนด business entities และ interfaces
- **ห้าม** depend on layers อื่น
- เขียน pure business logic

### Use Case Layer (`internal/usecase/`)
- Implement application logic
- Orchestrate domain entities
- Depend เฉพาะ domain layer

### Infrastructure Layer (`internal/infrastructure/`)
- Implement technical details
- Database, external APIs, file systems
- Implement domain interfaces

### Interface Layer (`internal/interfaces/`)
- Handle user interaction
- CLI, REST API, gRPC
- Convert external data to domain models

## Adding New Storage Provider

ถ้าต้องการเพิ่ม storage provider ใหม่:

1. **สร้าง implementation** ใน `internal/infrastructure/storage/`:
   ```go
   type NewStorage struct {
       config *config.NewStorageConfig
       client *SomeClient
   }

   func NewNewStorage(cfg *config.NewStorageConfig) repository.Storage {
       return &NewStorage{config: cfg}
   }

   // Implement repository.Storage interface
   func (n *NewStorage) Connect(ctx context.Context) error { ... }
   func (n *NewStorage) Disconnect(ctx context.Context) error { ... }
   func (n *NewStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) { ... }
   func (n *NewStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error { ... }
   // ... other methods
   ```

2. **เพิ่ม config** ใน `internal/infrastructure/config/config.go`:
   ```go
   type NewStorageConfig struct {
       Endpoint string `yaml:"endpoint" json:"endpoint"`
       APIKey   string `yaml:"api_key" json:"api_key"`
       // ... other fields
   }
   ```

3. **Update factory** ใน `internal/infrastructure/storage/factory.go`:
   ```go
   case config.StorageTypeNew:
       if cfg.NewStorage == nil {
           return nil, fmt.Errorf("NewStorage configuration is required")
       }
       return NewNewStorage(cfg.NewStorage), nil
   ```

4. **เขียน tests** สำหรับ implementation ใหม่

5. **Update documentation** และ examples

## Code Review Checklist

ก่อน submit PR ตรวจสอบ:

- [ ] Code ทำงานถูกต้อง
- [ ] Tests ผ่านทั้งหมด
- [ ] Test coverage ไม่ลดลง
- [ ] Documentation ถูก update
- [ ] Code style ถูกต้อง (run `make fmt`)
- [ ] Linter ไม่มี warnings (run `make lint`)
- [ ] Commit messages ชัดเจน
- [ ] ไม่มี sensitive data (passwords, keys) ใน code

## Questions?

หากมีคำถาม:
- เปิด [Issue](https://github.com/preedep/go-nixcopy/issues)
- ดู [README.md](README.md)
- ติดต่อ maintainers

## License

โดยการมีส่วนร่วมในโปรเจกต์นี้ คุณยอมรับว่า contributions ของคุณจะอยู่ภายใต้ MIT License เช่นเดียวกับโปรเจกต์

---

ขอบคุณสำหรับ contributions ของคุณ! 🙏
