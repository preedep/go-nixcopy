# Quick Start Guide - go-nixcopy

## การติดตั้งและรันครั้งแรก

### 1. ติดตั้ง nixcopy

```bash
# Homebrew (macOS / Linux) — แนะนำ
brew tap preedep/tap && brew install nixcopy

# APT (Debian / Ubuntu)
echo "deb [trusted=yes] https://apt.fury.io/preedep/ /" | sudo tee /etc/apt/sources.list.d/nixcopy.list
sudo apt-get update && sudo apt-get install nixcopy

# YUM (RHEL / CentOS / Fedora)
sudo yum install --repofrompath nixcopy,https://yum.fury.io/preedep/ nixcopy

# Go install
go install github.com/preedep/go-nixcopy/cmd/nixcopy@latest

# Build จาก source
make deps && make build   # binary อยู่ที่ ./bin/nixcopy
```

### 2. สร้างไฟล์ Config

คัดลอก example config และแก้ไขตามต้องการ:

```bash
cp config.example.yaml config.yaml
```

แก้ไข `config.yaml` ให้ตรงกับ storage ของคุณ

### 3. ทดสอบการเชื่อมต่อ

```bash
# ดูรายการไฟล์ใน source storage
./bin/nixcopy list -c config.yaml -p / --source

# ดูรายการไฟล์ใน destination storage
./bin/nixcopy list -c config.yaml -p / --source=false
```

### 4. ถ่ายโอนไฟล์

```bash
./bin/nixcopy transfer \
  -c config.yaml \
  -s /path/to/source/file.zip \
  -d /path/to/dest/file.zip
```

## ตัวอย่างการใช้งานจริง

### SFTP → S3

```yaml
# config.yaml
source:
  type: sftp
  sftp:
    host: your-sftp-server.com
    port: 22
    username: your-username
    private_key_path: ~/.ssh/id_rsa
    timeout: 30s
    max_packet_size: 32768

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: your-bucket
    access_key_id: YOUR_KEY
    secret_access_key: YOUR_SECRET

transfer:
  buffer_size: 33554432
  concurrent_files: 4
  retry_attempts: 3
  retry_delay: 5s
  timeout: 30m
```

> **Note:** The `logging:` block is deprecated and no longer read. Logs are always emitted as structured JSON to stdout.

```bash
./bin/nixcopy transfer -c config.yaml -s /data/backup.tar.gz -d backups/backup.tar.gz
```

### Azure Blob → FTPS

ใช้ config จาก `examples/blob-to-ftps.yaml`

```bash
./bin/nixcopy transfer \
  -c examples/blob-to-ftps.yaml \
  -s myfile.pdf \
  -d /upload/myfile.pdf
```

## คำสั่งที่มีประโยชน์

```bash
# Build
make build              # Build binary
make build-all          # Build สำหรับทุก platform
make install            # ติดตั้งไปยัง $GOPATH/bin

# Testing
make test               # รัน tests
make test-coverage      # รัน tests พร้อม coverage

# Code Quality
make fmt                # Format code
make lint               # รัน linter

# Docker
make docker-build       # Build Docker image for current platform (with OCI labels)
make docker-buildx      # Build + push multi-arch image (linux/amd64, linux/arm64)
make docker-run         # รัน Docker container

# Cleanup
make clean              # ลบไฟล์ build
```

## KPO Quick Start (ไม่ต้องมี config file)

รันจาก Docker โดยใช้ env vars แทน config file:

```bash
docker run --rm \
  -e NIXCOPY_SOURCE_TYPE=sftp \
  -e NIXCOPY_SOURCE_HOST=sftp.example.com \
  -e NIXCOPY_SOURCE_PORT=22 \
  -e NIXCOPY_SOURCE_USERNAME=user \
  -e NIXCOPY_SOURCE_PASSWORD=secret \
  -e NIXCOPY_DEST_TYPE=s3 \
  -e NIXCOPY_DEST_REGION=ap-southeast-1 \
  -e NIXCOPY_DEST_BUCKET=my-bucket \
  -e NIXCOPY_DEST_AUTH_TYPE=access_key \
  -e NIXCOPY_DEST_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE \
  -e NIXCOPY_DEST_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  go-nixcopy:latest \
  transfer -s /data/file.csv -d processed/file.csv
```

ดู [README.md — KPO Golden Image](../../README.md#-docker--kubernetes-golden-image) สำหรับตัวอย่าง KubernetesPodOperator แบบเต็ม

## Performance Tips

### ไฟล์ขนาดเล็ก (< 10MB)
```yaml
transfer:
  buffer_size: 8388608      # 8MB
  concurrent_files: 8
```

### ไฟล์ขนาดกลาง (10-100MB)
```yaml
transfer:
  buffer_size: 33554432     # 32MB
  concurrent_files: 4
```

### ไฟล์ขนาดใหญ่ (> 100MB)
```yaml
transfer:
  buffer_size: 67108864     # 64MB
  concurrent_files: 2
```

## Troubleshooting

### ปัญหา: Connection timeout
```yaml
source:
  sftp:
    timeout: 60s  # เพิ่ม timeout
```

### ปัญหา: Out of memory
```yaml
transfer:
  buffer_size: 16777216      # ลดเหลือ 16MB
  concurrent_files: 2        # ลดจำนวน concurrent
```

### ปัญหา: Authentication failed
- ตรวจสอบ username/password
- ตรวจสอบ permissions ของ private key (`chmod 600 ~/.ssh/id_rsa`)
- ตรวจสอบ firewall rules

## Next Steps

1. อ่าน [README.md](../../README.md) สำหรับข้อมูลเพิ่มเติม
2. ดู [examples/](examples/) สำหรับ config ตัวอย่าง
3. อ่าน [CONTRIBUTING.md](../development/contributing.md) ถ้าต้องการมีส่วนร่วม

## Support

- GitHub Issues: https://github.com/preedep/go-nixcopy/issues
- Documentation: [README.md](../../README.md)
