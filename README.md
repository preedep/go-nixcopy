# go-nixcopy

เครื่องมือ CLI สำหรับการถ่ายโอนไฟล์ความเร็วสูงระหว่างระบบ Storage ต่างๆ ที่พัฒนาด้วย Go

## 🚀 คุณสมบัติ

- **รองรับหลาย Storage Systems**
  - Local File System (การถ่ายโอนไฟล์ภายใน local disk)
  - SFTP (SSH File Transfer Protocol)
  - FTPS (FTP over SSL/TLS)
  - Azure Blob Storage
  - AWS S3 (และ S3-compatible storage เช่น MinIO)

- **การถ่ายโอนที่รองรับ**
  - Local ↔ Local (copy ระหว่าง directories)
  - Local ↔ SFTP
  - Local ↔ FTPS
  - Local ↔ Azure Blob Storage
  - Local ↔ AWS S3
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

- Go 1.24 หรือสูงกว่า
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
  type: sftp  # local, sftp, ftps, blob, s3
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
  verify_checksum: false     # SHA256 end-to-end integrity check
  enable_resume: false       # resume interrupted transfers (local & SFTP)
  skip_existing: false       # skip file if destination already has same size
```

> **Note:** The `logging:` block is deprecated and no longer read. All logs are now emitted in standard-app-log v1.0 JSON format to stdout automatically. See [Application Logging](#-application-logging) below.

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

#### Local File System Configuration

```yaml
local:
  base_path: /data/files  # Base directory สำหรับการเข้าถึงไฟล์
```

**Use Cases:**
- Backup ไฟล์จาก local ไปยัง cloud (S3, Azure Blob)
- Download ไฟล์จาก cloud มาเก็บที่ local
- Copy/organize ไฟล์ระหว่าง directories
- Migration ข้อมูลระหว่าง local และ cloud storage

**ข้อควรระวัง:**
- ต้องมี read permission สำหรับ source files
- ต้องมี write permission สำหรับ destination directory
- Base path จะถูกสร้างอัตโนมัติถ้ายังไม่มี
- ทุก operations จะถูกจำกัดภายใน base_path เพื่อความปลอดภัย

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

go-nixcopy รองรับการส่ง parameters ได้ 4 แบบ:

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

4. **ใช้ NIXCOPY_* Environment Variables อย่างเดียว** (แนะนำสำหรับ Kubernetes / KPO)
   ```bash
   export NIXCOPY_SOURCE_TYPE=sftp
   export NIXCOPY_SOURCE_HOST=sftp.example.com
   export NIXCOPY_SOURCE_USERNAME=user
   export NIXCOPY_SOURCE_PASSWORD=secret
   export NIXCOPY_DEST_TYPE=s3
   export NIXCOPY_DEST_REGION=ap-southeast-1
   export NIXCOPY_DEST_BUCKET=my-bucket
   export NIXCOPY_DEST_AUTH_TYPE=web_identity
   nixcopy transfer -s /data/file.csv -d processed/file.csv
   ```

**ลำดับความสำคัญ (Precedence):**
```
CLI Flags  >  NIXCOPY_* Env Vars  >  Config File  >  Defaults
```

📖 **อ่านเพิ่มเติม:** [CLI_USAGE.md](CLI_USAGE.md) - คู่มือการใช้ CLI parameters แบบละเอียด พร้อมตาราง NIXCOPY_* env vars ครบทุกตัว

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

#### 4. Backup จาก Local ไปยัง S3

```bash
nixcopy transfer \
  --config examples/local-to-s3.yaml \
  --sources "documents/**/*.pdf" \
  --dest backup/documents/ \
  --concurrent-files 8
```

#### 5. Download จาก S3 มา Local

```bash
nixcopy transfer \
  --config examples/s3-to-local.yaml \
  --sources "reports/2024/*.xlsx" \
  --dest downloads/
```

#### 6. Copy ไฟล์ระหว่าง Local Directories

```bash
nixcopy transfer \
  --config examples/local-to-local.yaml \
  --sources "temp/*.pdf" \
  --dest organized/pdfs/
```

#### 7. แสดงรายการไฟล์

```bash
# แสดงไฟล์ใน SFTP
nixcopy list -c config.yaml -p /remote/directory --source

# แสดงไฟล์ใน S3
nixcopy list -c config.yaml -p data/ --source=false
```

### Output ตัวอย่าง

Progress lines go to **stderr** (human-readable, safe to discard in pipelines):

```
[file.zip] 45.23% | 125.45 MB/s | ETA: 1m15s
[file.zip] ✓ Completed | 128.32 MB/s
```

On completion, a single JSON line is written to **stdout** (machine-readable):

```json
{"event":"transfer_summary","total_files":3,"successful":2,"skipped":1,"failed":0,"bytes_transferred":10737418240,"duration_ms":79823,"average_speed_mbps":128.32}
```

Downstream tools (Airflow log parsers, Loki, CloudWatch Insights) filter on `event="transfer_summary"` to extract metrics without parsing human text. When files fail, a `failed_files` array is included:

```json
{"event":"transfer_summary","total_files":3,"successful":2,"skipped":0,"failed":1,"bytes_transferred":5368709120,"duration_ms":45000,"average_speed_mbps":115.6,"failed_files":[{"path":"/src/bad.csv","error":"connection reset by peer"}]}
```

## 🧪 การทดสอบ

### รัน Unit Tests

```bash
# รัน tests ทั้งหมด (ตรงกับ CI)
go test -race -covermode=atomic -count=1 ./...

# รัน tests พร้อม verbose output
go test -v ./...

# รัน tests พร้อม coverage report
go test -race -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# รัน integration tests (ต้องการ MinIO และ SFTP server)
go test -tags=integration -race -count=1 -v ./internal/infrastructure/storage/...

# ใช้ Makefile
make test
make test-coverage
make test-verbose
make test-race
```

### Test Coverage

โปรเจกต์มี unit tests ครอบคลุมส่วนสำคัญ:
- ✅ Domain entities (pattern matching, transfer)
- ✅ Use cases (transfer, pattern matcher, retry, checksum, resume)
- ✅ Skip-existing / idempotent retry (5 cases)
- ✅ S3 multipart part-size algorithm (5 size scenarios)
- ✅ Azure Blob block-size algorithm (5 size scenarios)
- ✅ JSON exit summary shape and omitempty behaviour
- ✅ `NIXCOPY_*` env var loading including `NIXCOPY_SKIP_EXISTING`
- ✅ CLI flags including `--skip-existing` and `--resume`
- ✅ Configuration loading and validation

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
│   │       ├── local.go
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

### S3 Multipart Upload

ใช้ AWS SDK v2 `transfermanager` สำหรับการ upload ไปยัง S3 — ไม่มีข้อจำกัด 5 GB ของ `PutObject`:

- **Block size**: dynamic — default 16 MiB, scales up เมื่อ `ceil(size / 10,000) > 16 MiB` เพื่อไม่เกิน 10,000 parts ของ S3
- **Minimum**: 5 MiB (S3 hard limit)
- **Concurrency**: 5 concurrent part uploads ต่อไฟล์
- รองรับไฟล์ได้ถึง ~5 TiB

### Azure Blob Parallel Block Upload

ใช้ `UploadStream` พร้อม `BlockSize` และ `Concurrency` options:

- **Block size**: dynamic — default 16 MiB, scales up เมื่อ `ceil(size / 50,000) > 16 MiB` เพื่อไม่เกิน 50,000 blocks ของ Azure
- **Minimum**: 1 MiB (Azure SDK floor)
- **Concurrency**: 5 concurrent block uploads ต่อ blob
- รองรับ blob ได้ถึง ~190 TiB

### Idempotent Retry / Skip-Existing

ป้องกันการถ่ายโอนซ้ำเมื่อ Airflow DAG retry:

```bash
nixcopy transfer --skip-existing -s /data/*.csv -d processed/
# หรือ
export NIXCOPY_SKIP_EXISTING=true
```

- ก่อนถ่ายโอนแต่ละไฟล์ จะตรวจสอบ `Stat` ที่ปลายทาง
- หาก destination มีขนาดเท่ากับ source → skip (status: `skipped`)
- หากขนาดต่างกัน หรือยังไม่มีไฟล์ → transfer ปกติ
- ปลอดภัยสำหรับ batch — รายงาน `skipped` แยกจาก `successful` ใน JSON summary

### Checksum Verification

ตรวจสอบความถูกต้องของข้อมูลด้วย SHA-256 แบบ end-to-end:

```yaml
transfer:
  verify_checksum: true
```

หรือใช้ CLI flag:

```bash
nixcopy transfer --verify-checksum -s /data/file.zip -d /backup/file.zip ...
```

- คำนวณ SHA-256 ของ source ระหว่างการถ่ายโอน (ไม่มี overhead เพิ่มเติม)
- อ่านไฟล์ปลายทางกลับมาตรวจสอบหลังเขียนสำเร็จ
- หาก hash ไม่ตรง จะ retry อัตโนมัติ
- `TransferResult.Checksum` เก็บค่า hex ของ SHA-256 เมื่อสำเร็จ

### Resume Transfer

ต่อการถ่ายโอนที่หยุดกลางคันโดยไม่ต้องเริ่มใหม่จากต้น รองรับ **Local** และ **SFTP** (S3/Azure Blob ใช้การ retry แบบปกติ):

```yaml
transfer:
  enable_resume: true
```

หรือใช้ CLI flag:

```bash
nixcopy transfer --resume -s /data/large_file.tar.gz -d /backup/large_file.tar.gz ...
```

- ก่อน retry แต่ละครั้ง จะตรวจสอบขนาดไฟล์ปลายทาง
- หากมีไฟล์บางส่วนอยู่แล้ว จะอ่าน source ต่อจาก offset นั้น
- Progress แสดงเป็น % ของไฟล์ทั้งหมด (รวม bytes ที่โอนไปแล้ว)
- `TransferResult.ResumedFrom` เก็บ byte offset ที่ต่อการโอน

## 📋 Application Logging

go-nixcopy emits **structured JSON logs to stdout** conforming to [standard-app-log v1.0](https://github.com/preedep/standard-app-log). Every log line contains a fixed set of standard fields plus app-specific fields merged at the top level.

### Log Types

| `log_type` | When emitted |
|---|---|
| `APP_LOG` | Application lifecycle, retry warnings, batch start |
| `REQ_EX_LOG` | Initiating a connection or read from source storage |
| `RES_EX_LOG` | Write completion, checksum result, transfer success/failure |

### Example Output

```json
{"event_date_time":"2026-04-28T14:47:03.301Z","log_type":"REQ_EX_LOG","level":"INFO","app_id":"go-nixcopy","app_version":"1.0.0","service_id":"sftp-to-s3","service_pod_name":"nixcopy-pod-abc123","message":"Starting transfer","correlation_id":"b1753d4b-ef4e-4c84-a53b-afb377ca2cc4","request_id":"b111895b-37c8-4448-833e-3191507b5aeb","source":"/data/exports/report.csv","destination":"s3://my-bucket/processed/report.csv"}
{"event_date_time":"2026-04-28T14:47:05.812Z","log_type":"RES_EX_LOG","level":"INFO","app_id":"go-nixcopy","app_version":"1.0.0","service_id":"sftp-to-s3","service_pod_name":"nixcopy-pod-abc123","message":"Transfer completed","correlation_id":"b1753d4b-ef4e-4c84-a53b-afb377ca2cc4","request_id":"b111895b-37c8-4448-833e-3191507b5aeb","bytes":1048576,"execution_time":2511,"source":"/data/exports/report.csv","destination":"s3://my-bucket/processed/report.csv"}
```

### Context Injection via Environment Variables

These env vars are read at startup and embedded in every log line:

| Env var | Field in log | Default | Purpose |
|---|---|---|---|
| `NIXCOPY_CORRELATION_ID` | `correlation_id` | auto-generated UUID | Carry Airflow DAG run ID or upstream trace ID through all logs |
| `NIXCOPY_APP_ID` | `app_id` | `go-nixcopy` | Application identifier |
| `NIXCOPY_APP_VERSION` | `app_version` | `1.0.0` | Application version (inject from image tag) |
| `POD_NAME` | `service_pod_name` | _(empty)_ | K8s pod name via Downward API |

### Kubernetes / Airflow KubernetesPodOperator Usage

The golden image supports **config-file-free** operation — all storage credentials are injected via Kubernetes Secrets as `NIXCOPY_*` environment variables. No config file mount needed.

```python
from airflow.providers.cncf.kubernetes.operators.pod import KubernetesPodOperator
from kubernetes.client import models as k8s

KubernetesPodOperator(
    task_id="transfer_sftp_to_s3",
    image="your-registry/nixcopy:1.2.0",   # pin to a specific OCI-labeled tag
    cmds=["./nixcopy"],
    arguments=["transfer", "-s", "/data/exports/*.csv", "-d", "processed/"],
    env_vars=[
        # Observability
        k8s.V1EnvVar(name="NIXCOPY_CORRELATION_ID", value="{{ run_id }}"),
        k8s.V1EnvVar(name="NIXCOPY_APP_VERSION",    value="1.2.0"),
        k8s.V1EnvVar(name="POD_NAME",               value_from=k8s.V1EnvVarSource(
            field_ref=k8s.V1ObjectFieldSelector(field_path="metadata.name"))),
        # Source — SFTP (credentials from K8s Secret)
        k8s.V1EnvVar(name="NIXCOPY_SOURCE_TYPE",     value="sftp"),
        k8s.V1EnvVar(name="NIXCOPY_SOURCE_HOST",     value_from=k8s.V1EnvVarSource(
            secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="host"))),
        k8s.V1EnvVar(name="NIXCOPY_SOURCE_USERNAME", value_from=k8s.V1EnvVarSource(
            secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="username"))),
        k8s.V1EnvVar(name="NIXCOPY_SOURCE_PASSWORD", value_from=k8s.V1EnvVarSource(
            secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="password"))),
        # Destination — S3 via IRSA (no keys needed, role bound to service account)
        k8s.V1EnvVar(name="NIXCOPY_DEST_TYPE",      value="s3"),
        k8s.V1EnvVar(name="NIXCOPY_DEST_REGION",    value="ap-southeast-1"),
        k8s.V1EnvVar(name="NIXCOPY_DEST_BUCKET",    value="my-data-bucket"),
        k8s.V1EnvVar(name="NIXCOPY_DEST_AUTH_TYPE", value="web_identity"),
    ],
    security_context=k8s.V1PodSecurityContext(run_as_non_root=True),
    get_logs=True,
    is_delete_operator_pod=True,
)
```

Logs go to stdout and are collected automatically by the K8s logging stack (Loki, CloudWatch, etc.).

---

## 🐳 Docker & Kubernetes Golden Image

### Building the Image

```bash
# Single-arch build for local development (current platform)
make docker-build

# Multi-arch build and push to registry (linux/amd64 + linux/arm64)
# Requires: docker buildx, a builder with multi-arch support, and a registry
IMAGE_NAME=your-registry/nixcopy make docker-buildx

# Inspect OCI labels on the built image
docker inspect go-nixcopy:latest | jq '.[0].Config.Labels'
```

The image is built with:

| Property | Value |
|---|---|
| Base image | `alpine:3.21` (pinned) |
| User | Non-root (`nixcopy`, UID assigned by Alpine) |
| Architectures | `linux/amd64`, `linux/arm64` |
| OCI labels | `version`, `revision` (git SHA), `created` (ISO-8601) |

### OCI Labels (Traceability)

Every `make docker-build` or `make docker-buildx` run injects:

```
org.opencontainers.image.version   = git describe --tags (e.g. v1.2.0-3-gabcd1234)
org.opencontainers.image.revision  = git SHA short
org.opencontainers.image.created   = build timestamp (UTC)
org.opencontainers.image.title     = go-nixcopy
org.opencontainers.image.source    = https://github.com/preedep/go-nixcopy
```

### KPO Golden Image — Configuration via Env Vars

The image supports **no-config-file** mode. Inject all storage config as `NIXCOPY_*` Kubernetes Secret env vars — no YAML file mount required:

| Category | Env Var | Description |
|---|---|---|
| **Storage type** | `NIXCOPY_SOURCE_TYPE` | `sftp`, `ftps`, `s3`, `blob`, `local` |
| | `NIXCOPY_DEST_TYPE` | Same values |
| **SFTP/FTPS** | `NIXCOPY_SOURCE_HOST` / `NIXCOPY_DEST_HOST` | Hostname |
| | `NIXCOPY_SOURCE_PORT` / `NIXCOPY_DEST_PORT` | Port |
| | `NIXCOPY_SOURCE_USERNAME` / `NIXCOPY_DEST_USERNAME` | Username |
| | `NIXCOPY_SOURCE_PASSWORD` / `NIXCOPY_DEST_PASSWORD` | Password |
| | `NIXCOPY_SOURCE_PRIVATE_KEY` / `NIXCOPY_DEST_PRIVATE_KEY` | Private key path |
| **S3** | `NIXCOPY_SOURCE_REGION` / `NIXCOPY_DEST_REGION` | AWS region |
| | `NIXCOPY_SOURCE_BUCKET` / `NIXCOPY_DEST_BUCKET` | Bucket name |
| | `NIXCOPY_SOURCE_AUTH_TYPE` / `NIXCOPY_DEST_AUTH_TYPE` | `access_key`, `iam_role`, `web_identity`, `assume_role` |
| | `NIXCOPY_SOURCE_ACCESS_KEY` / `NIXCOPY_DEST_ACCESS_KEY` | Access key ID |
| | `NIXCOPY_SOURCE_SECRET_KEY` / `NIXCOPY_DEST_SECRET_KEY` | Secret key |
| | `NIXCOPY_SOURCE_ROLE_ARN` / `NIXCOPY_DEST_ROLE_ARN` | IAM role ARN |
| **Azure Blob** | `NIXCOPY_SOURCE_ACCOUNT_NAME` / `NIXCOPY_DEST_ACCOUNT_NAME` | Storage account |
| | `NIXCOPY_SOURCE_CONTAINER` / `NIXCOPY_DEST_CONTAINER` | Container name |
| | `NIXCOPY_SOURCE_AUTH_TYPE` / `NIXCOPY_DEST_AUTH_TYPE` | `shared_key`, `managed_identity`, `service_principal`, `sas_token` |
| | `NIXCOPY_SOURCE_ACCOUNT_KEY` / `NIXCOPY_DEST_ACCOUNT_KEY` | Account key |
| | `NIXCOPY_SOURCE_CLIENT_ID` / `NIXCOPY_DEST_CLIENT_ID` | Client ID (SP/MI) |
| | `NIXCOPY_SOURCE_CLIENT_SECRET` / `NIXCOPY_DEST_CLIENT_SECRET` | Client secret (SP) |
| | `NIXCOPY_SOURCE_TENANT_ID` / `NIXCOPY_DEST_TENANT_ID` | Tenant ID (SP) |
| **Transfer** | `NIXCOPY_BUFFER_SIZE` | Buffer size in bytes (default: 33554432) |
| | `NIXCOPY_CONCURRENT_FILES` | Parallel file count (default: 4) |
| | `NIXCOPY_RETRY_ATTEMPTS` | Retry count (default: 3) |
| | `NIXCOPY_RETRY_DELAY` | Retry delay e.g. `10s` (default: 5s) |
| | `NIXCOPY_VERIFY_CHECKSUM` | `true`/`false` (default: false) |
| | `NIXCOPY_ENABLE_RESUME` | `true`/`false` (default: false) |
| | `NIXCOPY_SKIP_EXISTING` | `true`/`false` — skip if dest has same size (default: false) |

📖 Full reference with every env var: [CLI_USAGE.md — NIXCOPY_* Environment Variables](CLI_USAGE.md#nixcopy-environment-variables)

---

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
- [standard-app-log](https://github.com/preedep/standard-app-log) - Application log schema
- [Zap](https://github.com/uber-go/zap) - Logging (internal infrastructure)

## 📚 เอกสารเพิ่มเติม

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Best Practices](https://golang.org/doc/effective_go)
- [SFTP Protocol](https://tools.ietf.org/html/draft-ietf-secsh-filexfer-02)
- [AWS S3 Documentation](https://docs.aws.amazon.com/s3/)
- [Azure Blob Storage Documentation](https://docs.microsoft.com/en-us/azure/storage/blobs/)

## 🗺️ Roadmap

### ✅ Completed
- [x] Local file system support
- [x] Batch transfer สำหรับหลายไฟล์
- [x] Parallel file transfer
- [x] Wildcard pattern support (`*.pdf`, `**/*.log`)
- [x] CLI parameters support
- [x] Comprehensive unit tests
- [x] Release build optimization
- [x] Professional-grade code documentation
- [x] SHA-256 checksum verification (end-to-end integrity)
- [x] Resume capability สำหรับการถ่ายโอนที่ถูกขัดจอน (Local & SFTP)
- [x] Standard application logging (standard-app-log v1.0) — structured JSON to stdout, K8s/Airflow ready
- [x] **KPO Golden Image** — non-root user, pinned base image, correct exit codes, `NIXCOPY_*` env-var config, multi-arch (`linux/amd64` + `linux/arm64`), OCI labels
- [x] **S3 multipart upload** — AWS SDK v2 `transfermanager`, dynamic part sizing, 5 concurrent parts, supports up to ~5 TiB
- [x] **Azure Blob parallel block upload** — dynamic `BlockSize` + `Concurrency=5`, supports up to ~190 TiB
- [x] **Idempotent retry** — `--skip-existing` / `NIXCOPY_SKIP_EXISTING` skips files where destination size matches source
- [x] **Structured JSON exit summary** — `event=transfer_summary` JSON line to stdout; progress to stderr; `sync.WaitGroup` prevents interleave in K8s log streams
- [x] **CI workflow** — GitHub Actions with correct Go version (`go-version-file: go.mod`), `go vet`, `go mod tidy` check, `-race -covermode=atomic`, MinIO + SFTP integration tests, `.golangci.yml` with explicit linter set

### 🚧 In Progress / Planned
- [ ] Web UI สำหรับการจัดการ
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
