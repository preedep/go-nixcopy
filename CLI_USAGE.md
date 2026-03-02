# คู่มือการใช้งาน CLI Parameters - go-nixcopy

เอกสารนี้อธิบายวิธีการใช้งาน CLI parameters เพื่อ override config file หรือใช้งานโดยไม่ต้องมี config file

## 📋 สารบัญ

- [การใช้งานพื้นฐาน](#การใช้งานพื้นฐาน)
- [CLI Flags ทั้งหมด](#cli-flags-ทั้งหมด)
- [ตัวอย่างการใช้งาน](#ตัวอย่างการใช้งาน)
- [Environment Variables](#environment-variables)
- [ลำดับความสำคัญ](#ลำดับความสำคัญ)

---

## การใช้งานพื้นฐาน

### แบบที่ 1: ใช้ Config File (แนะนำ)

```bash
nixcopy transfer -c config.yaml -s /source/file.txt -d /dest/file.txt
```

### แบบที่ 2: ใช้ CLI Parameters เต็มรูปแบบ (ไม่ต้องมี config file)

```bash
nixcopy transfer \
  --source-type sftp \
  --source-host sftp.example.com \
  --source-port 22 \
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
```

### แบบที่ 3: ผสมระหว่าง Config File และ CLI Parameters

```bash
# ใช้ config file เป็นฐาน แต่ override บาง parameters
nixcopy transfer -c config.yaml \
  --source-password $SFTP_PASSWORD \
  --dest-access-key $AWS_ACCESS_KEY \
  --dest-secret-key $AWS_SECRET_KEY \
  -s /source/file.txt \
  -d /dest/file.txt
```

---

## CLI Flags ทั้งหมด

### 🔹 Path Flags (Required)

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--source` | `-s` | Source file path | `-s /data/file.zip` |
| `--dest` | `-d` | Destination file path | `-d /backup/file.zip` |

### 🔹 General Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--config` | `-c` | Config file path | `-c config.yaml` |
| `--verbose` | `-v` | Verbose output | `-v` |

### 🔹 Source Storage Flags

#### Common Source Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--source-type` | Storage type (sftp, ftps, blob, s3) | `--source-type sftp` |
| `--source-auth-type` | Authentication type | `--source-auth-type iam_role` |

#### SFTP/FTPS Source Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--source-host` | Server hostname | `--source-host sftp.example.com` |
| `--source-port` | Server port | `--source-port 22` |
| `--source-username` | Username | `--source-username user` |
| `--source-password` | Password | `--source-password pass` |
| `--source-private-key` | Private key path | `--source-private-key ~/.ssh/id_rsa` |

#### S3 Source Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--source-region` | AWS region | `--source-region ap-southeast-1` |
| `--source-bucket` | S3 bucket name | `--source-bucket my-bucket` |
| `--source-access-key` | AWS access key ID | `--source-access-key AKIA...` |
| `--source-secret-key` | AWS secret access key | `--source-secret-key wJal...` |

#### Azure Blob Source Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--source-account-name` | Storage account name | `--source-account-name mystorageaccount` |
| `--source-account-key` | Storage account key | `--source-account-key abc123...` |
| `--source-container` | Container name | `--source-container mycontainer` |

### 🔹 Destination Storage Flags

#### Common Destination Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--dest-type` | Storage type (sftp, ftps, blob, s3) | `--dest-type s3` |
| `--dest-auth-type` | Authentication type | `--dest-auth-type access_key` |

#### SFTP/FTPS Destination Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--dest-host` | Server hostname | `--dest-host sftp.example.com` |
| `--dest-port` | Server port | `--dest-port 22` |
| `--dest-username` | Username | `--dest-username user` |
| `--dest-password` | Password | `--dest-password pass` |
| `--dest-private-key` | Private key path | `--dest-private-key ~/.ssh/id_rsa` |

#### S3 Destination Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--dest-region` | AWS region | `--dest-region us-east-1` |
| `--dest-bucket` | S3 bucket name | `--dest-bucket backup-bucket` |
| `--dest-access-key` | AWS access key ID | `--dest-access-key AKIA...` |
| `--dest-secret-key` | AWS secret access key | `--dest-secret-key wJal...` |

#### Azure Blob Destination Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--dest-account-name` | Storage account name | `--dest-account-name deststorage` |
| `--dest-account-key` | Storage account key | `--dest-account-key xyz789...` |
| `--dest-container` | Container name | `--dest-container destcontainer` |

### 🔹 Transfer Configuration Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--buffer-size` | Buffer size in bytes | 32MB | `--buffer-size 67108864` |
| `--concurrent-files` | Concurrent file transfers | 4 | `--concurrent-files 8` |
| `--retry-attempts` | Number of retry attempts | 3 | `--retry-attempts 5` |

---

## ตัวอย่างการใช้งาน

### 1. SFTP to S3 (ไม่ใช้ config file)

```bash
nixcopy transfer \
  --source-type sftp \
  --source-host sftp.example.com \
  --source-port 22 \
  --source-username sftpuser \
  --source-password "MyP@ssw0rd" \
  -s /data/backup.tar.gz \
  --dest-type s3 \
  --dest-region ap-southeast-1 \
  --dest-bucket my-backup-bucket \
  --dest-auth-type access_key \
  --dest-access-key AKIAIOSFODNN7EXAMPLE \
  --dest-secret-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
  -d backups/backup.tar.gz
```

### 2. S3 to Azure Blob (ใช้ IAM Role และ Managed Identity)

```bash
# รันบน EC2 instance ที่มี IAM role
nixcopy transfer \
  --source-type s3 \
  --source-region us-east-1 \
  --source-bucket source-bucket \
  --source-auth-type iam_role \
  -s data/file.zip \
  --dest-type blob \
  --dest-account-name deststorage \
  --dest-container destination \
  --dest-auth-type managed_identity \
  -d backups/file.zip
```

### 3. SFTP to SFTP (Copy ระหว่าง servers)

```bash
nixcopy transfer \
  --source-type sftp \
  --source-host source-sftp.example.com \
  --source-port 22 \
  --source-username sourceuser \
  --source-private-key ~/.ssh/source_key \
  -s /data/file.txt \
  --dest-type sftp \
  --dest-host dest-sftp.example.com \
  --dest-port 2222 \
  --dest-username destuser \
  --dest-private-key ~/.ssh/dest_key \
  -d /backup/file.txt
```

### 4. Azure Blob to S3 (ใช้ SAS Token)

```bash
nixcopy transfer \
  --source-type blob \
  --source-account-name sourcestorage \
  --source-container sourcecontainer \
  --source-auth-type sas_token \
  --source-account-key "sv=2021-06-08&ss=bfqt&srt=sco&sp=rwdlacupiytfx..." \
  -s myfile.pdf \
  --dest-type s3 \
  --dest-region eu-west-1 \
  --dest-bucket dest-bucket \
  --dest-auth-type access_key \
  --dest-access-key AKIAI44QH8DHBEXAMPLE \
  --dest-secret-key je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY \
  -d documents/myfile.pdf
```

### 5. Override Config File Parameters

```bash
# ใช้ config.yaml แต่ override credentials
nixcopy transfer -c config.yaml \
  --source-password $SFTP_PASSWORD \
  --dest-access-key $AWS_ACCESS_KEY_ID \
  --dest-secret-key $AWS_SECRET_ACCESS_KEY \
  --buffer-size 67108864 \
  --concurrent-files 8 \
  -s /data/largefile.zip \
  -d /backup/largefile.zip
```

### 6. ปรับแต่ง Performance

```bash
nixcopy transfer -c config.yaml \
  --buffer-size 134217728 \
  --concurrent-files 16 \
  --retry-attempts 5 \
  -s /source/file.zip \
  -d /dest/file.zip
```

---

## Environment Variables

คุณสามารถใช้ environment variables แทนการส่ง parameters โดยตรง:

### ตั้งค่า Environment Variables

```bash
# Source credentials
export SFTP_PASSWORD="MySecretPassword"
export SOURCE_PRIVATE_KEY="~/.ssh/id_rsa"

# Destination credentials
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="ap-southeast-1"

# Azure credentials
export AZURE_STORAGE_ACCOUNT="mystorageaccount"
export AZURE_STORAGE_KEY="your-storage-key"
export AZURE_TENANT_ID="your-tenant-id"
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"
```

### ใช้งานกับ CLI

```bash
# ใช้ environment variables
nixcopy transfer \
  --source-type sftp \
  --source-host sftp.example.com \
  --source-username user \
  --source-password "$SFTP_PASSWORD" \
  -s /data/file.txt \
  --dest-type s3 \
  --dest-region "$AWS_REGION" \
  --dest-bucket my-bucket \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY_ID" \
  --dest-secret-key "$AWS_SECRET_ACCESS_KEY" \
  -d /backup/file.txt
```

### ใช้กับ Config File

```yaml
source:
  type: sftp
  sftp:
    host: sftp.example.com
    username: user
    password: ${SFTP_PASSWORD}

destination:
  type: s3
  s3:
    region: ${AWS_REGION}
    bucket: my-bucket
    auth_type: access_key
    access_key_id: ${AWS_ACCESS_KEY_ID}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY}
```

---

## ลำดับความสำคัญ (Precedence)

เมื่อมีการตั้งค่าหลายแหล่ง ระบบจะใช้ค่าตามลำดับความสำคัญดังนี้:

1. **CLI Flags** (สูงสุด) - Override ทุกอย่าง
2. **Environment Variables** - ใช้เมื่อไม่มี CLI flags
3. **Config File** - ใช้เมื่อไม่มี CLI flags และ env vars
4. **Default Values** (ต่ำสุด) - ใช้เมื่อไม่มีการตั้งค่าใดๆ

### ตัวอย่าง

```bash
# config.yaml มี buffer_size = 32MB
# CLI flag ระบุ --buffer-size 67108864 (64MB)
# ผลลัพธ์: ใช้ 64MB (จาก CLI flag)

nixcopy transfer -c config.yaml \
  --buffer-size 67108864 \
  -s /source/file \
  -d /dest/file
```

---

## 💡 Tips & Best Practices

### 1. ใช้ Config File สำหรับ Production

```bash
# แนะนำ: ใช้ config file
nixcopy transfer -c production.yaml -s /data/file -d /backup/file
```

### 2. ใช้ CLI Flags สำหรับ Quick Tasks

```bash
# เหมาะสำหรับ: one-time transfers, testing
nixcopy transfer --source-type sftp --source-host ... -s /file -d /file
```

### 3. ใช้ Environment Variables สำหรับ Secrets

```bash
# ดี: ไม่เก็บ passwords ใน command history
export SFTP_PASSWORD="secret"
nixcopy transfer -c config.yaml --source-password "$SFTP_PASSWORD" ...

# ไม่ดี: password ถูกเก็บใน shell history
nixcopy transfer --source-password "secret" ...
```

### 4. ใช้ Scripts สำหรับ Repeated Tasks

```bash
#!/bin/bash
# backup.sh

SOURCE_PATH="/data/$(date +%Y%m%d).tar.gz"
DEST_PATH="backups/$(date +%Y%m%d).tar.gz"

nixcopy transfer \
  -c config.yaml \
  -s "$SOURCE_PATH" \
  -d "$DEST_PATH" \
  --buffer-size 67108864 \
  --concurrent-files 8
```

### 5. Validate Configuration

```bash
# ตรวจสอบ config ก่อนรัน
nixcopy transfer -c config.yaml --help

# ใช้ verbose mode เพื่อ debug
nixcopy transfer -c config.yaml -v -s /file -d /file
```

---

## 🔍 Debugging

### แสดง Help

```bash
# Help ทั่วไป
nixcopy --help

# Help สำหรับ transfer command
nixcopy transfer --help

# List ทุก flags
nixcopy transfer --help | grep -E '^\s+--'
```

### Verbose Mode

```bash
# เปิด verbose logging
nixcopy transfer -c config.yaml -v -s /file -d /file
```

### Dry Run (ถ้ามี feature นี้)

```bash
# ตรวจสอบ configuration โดยไม่ transfer จริง
nixcopy transfer -c config.yaml --dry-run -s /file -d /file
```

---

## 📚 เอกสารเพิ่มเติม

- [README.md](README.md) - คู่มือหลัก
- [AUTHENTICATION.md](AUTHENTICATION.md) - คู่มือ authentication
- [QUICKSTART.md](QUICKSTART.md) - คู่มือเริ่มต้นใช้งาน
- [examples/](examples/) - ตัวอย่าง config files

---

## ❓ FAQ

**Q: ต้องมี config file ไหม?**  
A: ไม่จำเป็น คุณสามารถใช้ CLI flags อย่างเดียวได้ แต่แนะนำให้ใช้ config file สำหรับ production

**Q: CLI flags override config file ได้ไหม?**  
A: ได้ CLI flags มีความสำคัญสูงกว่า config file

**Q: ใช้ environment variables ได้ไหม?**  
A: ได้ ทั้งใน config file (${VAR_NAME}) และใน CLI flags

**Q: จะเก็บ passwords อย่างปลอดภัยได้อย่างไร?**  
A: ใช้ environment variables, AWS Secrets Manager, หรือ Azure Key Vault

**Q: สามารถ override เฉพาะบาง parameters ได้ไหม?**  
A: ได้ ระบุเฉพาะ flags ที่ต้องการ override ส่วนที่เหลือจะใช้จาก config file
