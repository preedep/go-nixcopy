# go-nixcopy

เครื่องมือ CLI สำหรับการถ่ายโอนไฟล์ความเร็วสูงระหว่างระบบ Storage ต่างๆ ที่พัฒนาด้วย Go

## 🚀 คุณสมบัติ

- **รองรับหลาย Storage Systems**
  - SFTP (SSH File Transfer Protocol)
  - FTPS (FTP over SSL/TLS)
  - Azure Blob Storage
  - AWS S3 (และ S3-compatible storage เช่น MinIO)

- **การถ่ายโอนที่รองรับ**
  - SFTP ↔ FTPS
  - SFTP ↔ Azure Blob Storage
  - SFTP ↔ AWS S3
  - FTPS ↔ Azure Blob Storage
  - FTPS ↔ AWS S3
  - Azure Blob Storage ↔ AWS S3

- **ประสิทธิภาพสูง**
  - ใช้ Streaming I/O เพื่อประหยัดหน่วยความจำ
  - รองรับการถ่ายโอนหลายไฟล์พร้อมกัน (Parallel Transfer)
  - รองรับ Wildcard Patterns (`*.pdf`, `**/*.log`)
  - Buffer size ที่ปรับแต่งได้
  - Retry mechanism อัตโนมัติเมื่อเกิดข้อผิดพลาด
  - Progress tracking แบบ real-time สำหรับแต่ละไฟล์

- **สถาปัตยกรรม**
  - ออกแบบตาม Clean Architecture principles
  - แยก Layer ชัดเจน (Domain, Use Case, Infrastructure, Interface)
  - ง่ายต่อการขยายและบำรุงรักษา
  - Test-friendly design

## 📋 ความต้องการของระบบ

- Go 1.21 หรือสูงกว่า
- การเข้าถึง Storage systems ที่ต้องการใช้งาน
- Credentials ที่จำเป็นสำหรับแต่ละ storage

## 🔧 การติดตั้ง

### ติดตั้งจาก Source

```bash
# Clone repository
git clone https://github.com/preedep/go-nixcopy.git
cd go-nixcopy

# ดาวน์โหลด dependencies
go mod download

# Build
go build -o nixcopy cmd/nixcopy/main.go

# ติดตั้งไปยัง $GOPATH/bin
go install cmd/nixcopy/main.go
```

### ติดตั้งด้วย go install

```bash
go install github.com/preedep/go-nixcopy/cmd/nixcopy@latest
```

## ⚙️ การตั้งค่า

สร้างไฟล์ `config.yaml` โดยอ้างอิงจาก `config.example.yaml`

### โครงสร้างไฟล์ Config

```yaml
source:
  type: sftp  # sftp, ftps, blob, s3
  sftp:
    host: sftp.example.com
    port: 22
    username: user
    password: password
    # หรือใช้ private key
    # private_key_path: /path/to/key
    # private_key_passphrase: passphrase
    timeout: 30s
    max_packet_size: 32768

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-bucket
    access_key_id: YOUR_ACCESS_KEY
    secret_access_key: YOUR_SECRET_KEY

transfer:
  buffer_size: 33554432      # 32MB
  concurrent_files: 4
  retry_attempts: 3
  retry_delay: 5s
  timeout: 30m
  verify_checksum: false

logging:
  level: info                # debug, info, warn, error
  format: json               # json, console
  output_path: stdout        # stdout หรือ path ของไฟล์
```

## 🔐 Authentication

go-nixcopy รองรับวิธีการ authentication หลายแบบสำหรับแต่ละ cloud provider เพื่อความยืดหยุนและความปลอดภัย

### AWS S3 Authentication Methods

- **Access Key** - Static credentials (development/testing)
- **IAM Role** - สำหรับ EC2, ECS, Lambda (แนะนำ)
- **Instance Profile** - สำหรับ EC2 instances
- **Assume Role** - Cross-account access
- **Web Identity** - สำหรับ Kubernetes (EKS IRSA)
- **AWS Profile** - ใช้ credentials จาก ~/.aws/credentials

### Azure Blob Storage Authentication Methods

- **Shared Key** - Account key (development/testing)
- **SAS Token** - Temporary access with limited permissions
- **Connection String** - Quick setup
- **Managed Identity** - สำหรับ Azure VMs, App Services (แนะนำ)
- **Service Principal** - สำหรับ applications, CI/CD

📖 **อ่านเพิ่มเติม:** [AUTHENTICATION.md](AUTHENTICATION.md) - คู่มือการ authentication แบบละเอียดพร้อมตัวอย่าง

### การตั้งค่าแต่ละ Storage Type

#### SFTP Configuration

```yaml
sftp:
  host: sftp.example.com
  port: 22
  username: user
  password: password                    # ใช้ password
  # หรือ
  private_key_path: ~/.ssh/id_rsa      # ใช้ private key
  private_key_passphrase: passphrase   # ถ้า key มี passphrase
  timeout: 30s
  max_packet_size: 32768               # ขนาด packet สูงสุด
```

#### FTPS Configuration

```yaml
ftps:
  host: ftps.example.com
  port: 21
  username: user
  password: password
  timeout: 30s
  tls_mode: explicit        # explicit หรือ implicit
  skip_verify: false        # ข้าม TLS certificate verification
```

#### Azure Blob Storage Configuration

**ตัวอย่าง 1: Shared Key**
```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: shared_key
  account_key: YOUR_ACCOUNT_KEY
  endpoint: https://mystorageaccount.blob.core.windows.net/  # optional
```

**ตัวอย่าง 2: Managed Identity (แนะนำสำหรับ Azure VMs)**
```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: managed_identity
  client_id: ""  # ระบุเฉพาะถ้าใช้ user-assigned managed identity
```

**ตัวอย่าง 3: Service Principal**
```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: service_principal
  tenant_id: your-tenant-id
  client_id: your-client-id
  client_secret: your-client-secret
```

**ตัวอย่าง 4: SAS Token**
```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: sas_token
  sas_token: "sv=2021-06-08&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2024-12-31T23:59:59Z..."
```

**ตัวอย่าง 5: Connection String**
```yaml
blob:
  auth_type: connection_string
  connection_string: "DefaultEndpointsProtocol=https;AccountName=mystorageaccount;AccountKey=YOUR_KEY;EndpointSuffix=core.windows.net"
  container_name: mycontainer
```

#### AWS S3 Configuration

**ตัวอย่าง 1: Access Key (Static Credentials)**
```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: access_key
  access_key_id: YOUR_ACCESS_KEY
  secret_access_key: YOUR_SECRET_KEY
  endpoint: ""              # ใช้สำหรับ S3-compatible (เช่น MinIO)
  use_path_style: false     # ใช้ path-style URLs
```

**ตัวอย่าง 2: IAM Role (แนะนำสำหรับ EC2/ECS)**
```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: iam_role       # ใช้ IAM role ที่ attach กับ instance
```

**ตัวอย่าง 3: Assume Role (Cross-account)**
```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: assume_role
  role_arn: arn:aws:iam::123456789012:role/MyRole
  role_session_name: nixcopy-session
  external_id: my-external-id  # optional
```

**ตัวอย่าง 4: Web Identity (EKS IRSA)**
```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: web_identity
  role_arn: arn:aws:iam::123456789012:role/EKSPodRole
  web_identity_token_file: /var/run/secrets/eks.amazonaws.com/serviceaccount/token
```

## 📖 วิธีการใช้งาน

### คำสั่งพื้นฐาน

```bash
# แสดงความช่วยเหลือ
nixcopy --help

# ถ่ายโอนไฟล์ (ใช้ config file)
nixcopy transfer --config config.yaml --source /path/to/source/file --dest /path/to/dest/file

# ถ่ายโอนไฟล์ (ใช้ CLI parameters - ไม่ต้องมี config file)
nixcopy transfer \
  --source-type sftp \
  --source-host sftp.example.com \
  --source-username user \
  --source-password pass \
  --source /remote/file.txt \
  --dest-type s3 \
  --dest-region ap-southeast-1 \
  --dest-bucket my-bucket \
  --dest-auth-type access_key \
  --dest-access-key AKIAIOSFODNN7EXAMPLE \
  --dest-secret-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  --dest /backup/file.txt

# แสดงรายการไฟล์ใน source storage
nixcopy list --config config.yaml --path /remote/path --source

# แสดงรายการไฟล์ใน destination storage
nixcopy list --config config.yaml --path /remote/path --source=false
```

### 🎯 การใช้ CLI Parameters

go-nixcopy รองรับการส่ง parameters ผ่าน command line ได้ 3 แบบ:

1. **ใช้ Config File อย่างเดียว** (แนะนำสำหรับ production)
   ```bash
   nixcopy transfer -c config.yaml -s /source/file -d /dest/file
   ```

2. **ใช้ CLI Parameters อย่างเดียว** (ไม่ต้องมี config file)
   ```bash
   nixcopy transfer --source-type sftp --source-host ... -s /file -d /file
   ```

3. **ผสมกัน - Config File + CLI Parameters** (CLI override config)
   ```bash
   nixcopy transfer -c config.yaml \
     --source-password $SFTP_PASSWORD \
     --dest-access-key $AWS_ACCESS_KEY \
     -s /source/file -d /dest/file
   ```

📖 **อ่านเพิ่มเติม:** [CLI_USAGE.md](CLI_USAGE.md) - คู่มือการใช้ CLI parameters แบบละเอียด

### 🚀 Parallel Transfer & Wildcard Patterns

go-nixcopy รองรับการถ่ายโอนหลายไฟล์พร้อมกัน (parallel) และ wildcard patterns

#### Wildcard Patterns ที่รองรับ

- `*.pdf` - ไฟล์ PDF ทั้งหมดใน directory ปัจจุบัน
- `report*.xlsx` - ไฟล์ที่ขึ้นต้นด้วย "report"
- `**/*.log` - ไฟล์ .log ทั้งหมดรวม subdirectories
- `data/2024/**/*.csv` - ไฟล์ CSV ทั้งหมดใน data/2024 และ subdirectories

#### ตัวอย่างการใช้งาน

**1. ถ่ายโอนไฟล์ PDF ทั้งหมด:**
```bash
nixcopy transfer -c config.yaml \
  -s "documents/*.pdf" \
  -d /backup/documents/ \
  --concurrent-files 8
```

**2. ถ่ายโอนหลายไฟล์ (ระบุชื่อชัดเจน):**
```bash
nixcopy transfer -c config.yaml \
  --sources file1.pdf,file2.pdf,file3.pdf \
  -d /backup/
```

**3. ถ่ายโอนแบบ Recursive:**
```bash
nixcopy transfer -c config.yaml \
  -s "logs/**/*.log" \
  -d /backup/logs/ \
  --concurrent-files 16
```

**4. ผสม Patterns หลายแบบ:**
```bash
nixcopy transfer -c config.yaml \
  --sources "*.pdf,*.docx,reports/*.xlsx" \
  -d /backup/documents/ \
  --concurrent-files 12
```

📖 **อ่านเพิ่มเติม:** [PARALLEL_TRANSFER.md](PARALLEL_TRANSFER.md) - คู่มือการถ่ายโอนแบบ parallel แบบละเอียด

#### ลำดับความสำคัญ (Precedence)

1. **CLI Flags** (สูงสุด) - Override ทุกอย่าง
2. **Environment Variables** - ใช้เมื่อไม่มี CLI flags
3. **Config File** - ใช้เมื่อไม่มี CLI flags และ env vars
4. **Default Values** (ต่ำสุด)

### ตัวอย่างการใช้งาน

#### 1. ถ่ายโอนไฟล์จาก SFTP ไปยัง AWS S3

**แบบที่ 1: ใช้ Config File**

```bash
# สร้างไฟล์ config
cat > config.yaml << EOF
source:
  type: sftp
  sftp:
    host: sftp.example.com
    port: 22
    username: user
    private_key_path: ~/.ssh/id_rsa
    timeout: 30s
    max_packet_size: 32768

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-s3-bucket
    access_key_id: AKIAIOSFODNN7EXAMPLE
    secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

transfer:
  buffer_size: 33554432
  concurrent_files: 4
  retry_attempts: 3
  retry_delay: 5s
  timeout: 30m

logging:
  level: info
  format: json
  output_path: stdout
EOF

# ถ่ายโอนไฟล์
nixcopy transfer -c config.yaml -s /remote/data/file.zip -d backup/file.zip
```

**แบบที่ 2: ใช้ CLI Parameters (ไม่ต้องมี config file)**

```bash
nixcopy transfer \
  --source-type sftp \
  --source-host sftp.example.com \
  --source-port 22 \
  --source-username user \
  --source-private-key ~/.ssh/id_rsa \
  -s /remote/data/file.zip \
  --dest-type s3 \
  --dest-region ap-southeast-1 \
  --dest-bucket my-s3-bucket \
  --dest-auth-type access_key \
  --dest-access-key AKIAIOSFODNN7EXAMPLE \
  --dest-secret-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  -d backup/file.zip
```

**แบบที่ 3: ใช้ Environment Variables**

```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export SFTP_PASSWORD="your-password"

nixcopy transfer \
  --source-type sftp \
  --source-host sftp.example.com \
  --source-username user \
  --source-password "$SFTP_PASSWORD" \
  -s /remote/data/file.zip \
  --dest-type s3 \
  --dest-region ap-southeast-1 \
  --dest-bucket my-s3-bucket \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY_ID" \
  --dest-secret-key "$AWS_SECRET_ACCESS_KEY" \
  -d backup/file.zip
```

#### 2. ถ่ายโอนจาก Azure Blob Storage ไปยัง FTPS

```bash
nixcopy transfer \
  --config examples/blob-to-ftps.yaml \
  --source myfile.pdf \
  --dest /upload/myfile.pdf
```

#### 3. ถ่ายโอนจาก S3 ไปยัง Azure Blob Storage

```bash
nixcopy transfer \
  --config examples/s3-to-blob.yaml \
  --source data/archive.tar.gz \
  --dest backups/archive.tar.gz
```

#### 4. แสดงรายการไฟล์

```bash
# แสดงไฟล์ใน SFTP
nixcopy list -c config.yaml -p /remote/directory --source

# แสดงไฟล์ใน S3
nixcopy list -c config.yaml -p data/ --source=false
```

### Output ตัวอย่าง

```
[file.zip] 45.23% | 125.45 MB/s | ETA: 0h2m15s
[file.zip] ✓ Completed | 128.32 MB/s

Transfer Summary:
  Source: /remote/data/file.zip
  Destination: backup/file.zip
  Bytes Transferred: 10737418240 (10240.00 MB)
  Duration: 1m19.823s
  Average Speed: 128.32 MB/s
  Status: completed
```

## 🧪 การทดสอบ

### รัน Unit Tests

```bash
# รัน tests ทั้งหมด
go test ./...

# รัน tests พร้อม verbose output
go test -v ./...

# รัน tests พร้อม coverage report
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# ใช้ Makefile
make test
make test-coverage
make test-verbose
```

### Test Coverage

โปรเจกต์มี unit tests ครอบคลุมส่วนสำคัญ:
- ✅ Domain entities (pattern matching, transfer)
- ✅ Use cases (transfer, pattern matcher)
- ✅ Configuration และ validation
- ✅ CLI flags และ parameter handling
- ✅ Mock storage สำหรับ testing

📖 **อ่านเพิ่มเติม:** [TESTING.md](TESTING.md) - คู่มือการทดสอบแบบละเอียด

## 📦 Build & Release

### Build Release Version

```bash
# Build สำหรับ platform ปัจจุบัน (optimized, no debug symbols)
./build-release.sh
# หรือ
make release

# Build สำหรับทุก platforms
./build-release.sh all
make release-all

# Build platform เฉพาะ
./build-release.sh linux amd64
./build-release.sh darwin arm64
./build-release.sh windows amd64
```

### Build Optimization

Release builds ใช้ optimization flags:
- **`-ldflags="-s -w"`** - ลบ debug symbols (ลด size ~40-50%)
- **`-trimpath`** - ลบ absolute paths (reproducible builds)
- **Version injection** - เพิ่ม version, build time, git commit

**ผลลัพธ์:**
- Development build: ~25-30 MB
- Release build: ~15-18 MB (ลดลง 40-50%)
- Release + UPX: ~5-7 MB (ลดลง 70-80%)

📖 **อ่านเพิ่มเติม:** [BUILD.md](BUILD.md) - คู่มือการ build แบบละเอียด

## 🏗️ สถาปัตยกรรม

โปรเจกต์นี้ใช้ Clean Architecture แบ่งเป็น 4 layers หลัก:

```
go-nixcopy/
├── cmd/
│   └── nixcopy/           # Entry point ของแอปพลิเคชัน
│       └── main.go
├── internal/
│   ├── domain/            # Domain Layer (Business Logic)
│   │   ├── entity/        # Entities และ Value Objects
│   │   ├── repository/    # Repository Interfaces
│   │   └── service/       # Service Interfaces
│   ├── usecase/           # Use Case Layer (Application Logic)
│   │   └── transfer_usecase.go
│   ├── infrastructure/    # Infrastructure Layer
│   │   ├── config/        # Configuration
│   │   ├── logger/        # Logging
│   │   └── storage/       # Storage Implementations
│   │       ├── sftp.go
│   │       ├── ftps.go
│   │       ├── blob.go
│   │       ├── s3.go
│   │       └── factory.go
│   └── interfaces/        # Interface Layer
│       └── cli/           # CLI Commands
├── examples/              # ตัวอย่าง Config Files
├── config.example.yaml
├── go.mod
└── README.md
```

### Layer Responsibilities

1. **Domain Layer**: กำหนด business entities, interfaces และ business rules
2. **Use Case Layer**: ประมวลผล application logic และ orchestrate การทำงาน
3. **Infrastructure Layer**: implement technical details (database, external services, etc.)
4. **Interface Layer**: handle user interaction (CLI, API, etc.)

## 🔍 คุณสมบัติเพิ่มเติม

### Streaming Transfer

โปรแกรมใช้ streaming I/O เพื่อประหยัดหน่วยความจำ:
- ไม่โหลดไฟล์ทั้งหมดเข้า memory
- อ่านและเขียนแบบ chunk-by-chunk
- เหมาะสำหรับไฟล์ขนาดใหญ่

### Progress Tracking

- แสดงความคืบหน้าแบบ real-time
- คำนวณความเร็วการถ่ายโอน
- ประมาณเวลาที่เหลือ (ETA)

### Error Handling

- Retry อัตโนมัติเมื่อเกิดข้อผิดพลาด
- Configurable retry attempts และ delay
- Graceful shutdown เมื่อได้รับ interrupt signal

### Concurrent Transfer

- รองรับการถ่ายโอนหลายไฟล์พร้อมกัน
- จำกัดจำนวน concurrent connections ได้
- เพิ่มประสิทธิภาพการถ่ายโอน

## 🔐 Security Best Practices

1. **ไม่ hardcode credentials** ในโค้ด
2. **ใช้ environment variables** สำหรับข้อมูลสำคัญ
3. **เก็บ config files** ในที่ปลอดภัย (chmod 600)
4. **ใช้ private keys** แทน passwords เมื่อเป็นไปได้
5. **Enable TLS/SSL** สำหรับการเชื่อมต่อทั้งหมด

### ตัวอย่างการใช้ Environment Variables

```bash
# ตั้งค่า environment variables
export SFTP_PASSWORD="your-password"
export AWS_ACCESS_KEY="your-access-key"
export AWS_SECRET_KEY="your-secret-key"

# ใช้ใน config.yaml
source:
  type: sftp
  sftp:
    password: ${SFTP_PASSWORD}

destination:
  type: s3
  s3:
    access_key_id: ${AWS_ACCESS_KEY}
    secret_access_key: ${AWS_SECRET_KEY}
```

## 📊 Performance Tuning

### Buffer Size

- **Small files (< 10MB)**: 4-8 MB buffer
- **Medium files (10-100MB)**: 16-32 MB buffer
- **Large files (> 100MB)**: 32-64 MB buffer

```yaml
transfer:
  buffer_size: 33554432  # 32MB
```

### Concurrent Files

- **Slow network**: 1-2 concurrent files
- **Normal network**: 4-8 concurrent files
- **Fast network**: 8-16 concurrent files

```yaml
transfer:
  concurrent_files: 4
```

### Timeout Settings

ปรับ timeout ตามขนาดไฟล์และความเร็ว network:

```yaml
transfer:
  timeout: 30m  # สำหรับไฟล์ขนาดใหญ่
```

## 🐛 Troubleshooting

### Connection Timeout

```
Error: failed to connect to source: dial tcp: i/o timeout
```

**แก้ไข**: เพิ่ม timeout ใน config

```yaml
source:
  sftp:
    timeout: 60s  # เพิ่มจาก 30s
```

### Authentication Failed

```
Error: failed to login: ssh: unable to authenticate
```

**แก้ไข**: ตรวจสอบ credentials และ permissions

### Out of Memory

```
Error: runtime: out of memory
```

**แก้ไข**: ลด buffer size หรือ concurrent files

```yaml
transfer:
  buffer_size: 16777216      # ลดเหลือ 16MB
  concurrent_files: 2        # ลดเหลือ 2 files
```

## 🤝 การมีส่วนร่วม

ยินดีรับ contributions! กรุณา:

1. Fork repository
2. สร้าง feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push ไปยัง branch (`git push origin feature/amazing-feature`)
5. เปิด Pull Request

## 📝 License

โปรเจกต์นี้อยู่ภายใต้ MIT License - ดูรายละเอียดใน [LICENSE](LICENSE) file

## 👨‍💻 ผู้พัฒนา

- **Preedep** - [GitHub](https://github.com/preedep)

## 🙏 Acknowledgments

- [pkg/sftp](https://github.com/pkg/sftp) - SFTP implementation
- [jlaffaye/ftp](https://github.com/jlaffaye/ftp) - FTP/FTPS client
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) - Azure Blob Storage
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS S3
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Zap](https://github.com/uber-go/zap) - Logging

## 📚 เอกสารเพิ่มเติม

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Best Practices](https://golang.org/doc/effective_go)
- [SFTP Protocol](https://tools.ietf.org/html/draft-ietf-secsh-filexfer-02)
- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [Azure Blob Storage Documentation](https://docs.microsoft.com/en-us/azure/storage/blobs/)

## 🗺️ Roadmap

### ✅ Completed
- [x] Batch transfer สำหรับหลายไฟล์
- [x] Parallel file transfer
- [x] Wildcard pattern support (`*.pdf`, `**/*.log`)
- [x] CLI parameters support
- [x] Comprehensive unit tests
- [x] Release build optimization

### 🚧 In Progress / Planned
- [ ] รองรับ checksum verification (MD5, SHA256)
- [ ] Resume capability สำหรับการถ่ายโอนที่ถูกขัดจอน
- [ ] Web UI สำหรับการจัดการ
- [ ] Docker image
- [ ] รองรับ Google Cloud Storage
- [ ] Bandwidth limiting
- [ ] Scheduling transfers
- [ ] Email notifications
- [ ] Compression support (gzip, zstd)
- [ ] Incremental backup

## 💬 ติดต่อ

หากมีคำถามหรือข้อเสนอแนะ กรุณาเปิด [Issue](https://github.com/preedep/go-nixcopy/issues) บน GitHub

---

Made with ❤️ in Thailand
