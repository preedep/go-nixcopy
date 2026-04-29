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
| `--verify-checksum` | SHA-256 end-to-end integrity check | false | `--verify-checksum` |
| `--resume` | Resume interrupted transfer from partial destination file (local & SFTP) | false | `--resume` |
| `--skip-existing` | Skip file if destination already has same size (idempotent retries) | false | `--skip-existing` |
| `--bandwidth-limit` | Max bandwidth per file — accepts suffixes: `KB`, `MB`, `GB`, `KiB`, `MiB`, `GiB`, or raw bytes; `0`/empty = unlimited | unlimited | `--bandwidth-limit 10MB` |

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

## NIXCOPY_* Environment Variables

`NIXCOPY_*` variables are **first-class configuration** — they are read automatically and overlay the config file. No `${VAR}` YAML syntax needed; no CLI flags needed either. This is the recommended approach for Kubernetes / KPO deployments where credentials are stored in Kubernetes Secrets.

### Source Storage

| Env Var | Applies to | Description |
|---|---|---|
| `NIXCOPY_SOURCE_TYPE` | all | Storage type: `local`, `sftp`, `ftps`, `s3`, `blob` |
| `NIXCOPY_SOURCE_BASE_PATH` | local | Base directory |
| `NIXCOPY_SOURCE_HOST` | sftp, ftps | Hostname or IP |
| `NIXCOPY_SOURCE_PORT` | sftp, ftps | Port number |
| `NIXCOPY_SOURCE_USERNAME` | sftp, ftps | Username |
| `NIXCOPY_SOURCE_PASSWORD` | sftp, ftps | Password |
| `NIXCOPY_SOURCE_PRIVATE_KEY` | sftp | Path to private key file |
| `NIXCOPY_SOURCE_PRIVATE_KEY_PASS` | sftp | Private key passphrase |
| `NIXCOPY_SOURCE_TIMEOUT` | sftp, ftps | Connection timeout (e.g. `30s`, `2m`) |
| `NIXCOPY_SOURCE_MAX_PACKET_SIZE` | sftp | Max SFTP packet size (default: 32768) |
| `NIXCOPY_SOURCE_TLS_MODE` | ftps | `explicit` or `implicit` |
| `NIXCOPY_SOURCE_SKIP_VERIFY` | ftps | Skip TLS cert verify: `true`/`false` |
| `NIXCOPY_SOURCE_REGION` | s3 | AWS region (e.g. `ap-southeast-1`) |
| `NIXCOPY_SOURCE_BUCKET` | s3 | S3 bucket name |
| `NIXCOPY_SOURCE_ENDPOINT` | s3, blob | Custom endpoint URL |
| `NIXCOPY_SOURCE_USE_PATH_STYLE` | s3 | Path-style URLs (MinIO): `true`/`false` |
| `NIXCOPY_SOURCE_AUTH_TYPE` | s3, blob | Auth type (see tables below) |
| `NIXCOPY_SOURCE_ACCESS_KEY` | s3 | AWS access key ID |
| `NIXCOPY_SOURCE_SECRET_KEY` | s3 | AWS secret access key |
| `NIXCOPY_SOURCE_SESSION_TOKEN` | s3 | AWS session token (STS) |
| `NIXCOPY_SOURCE_ROLE_ARN` | s3 | IAM role ARN (`assume_role` / `web_identity`) |
| `NIXCOPY_SOURCE_ROLE_SESSION_NAME` | s3 | Role session name |
| `NIXCOPY_SOURCE_EXTERNAL_ID` | s3 | External ID for cross-account roles |
| `NIXCOPY_SOURCE_WEB_IDENTITY_TOKEN_FILE` | s3 | Token file path (IRSA: `/var/run/secrets/eks.amazonaws.com/serviceaccount/token`) |
| `NIXCOPY_SOURCE_PROFILE` | s3 | AWS credentials profile name |
| `NIXCOPY_SOURCE_ACCOUNT_NAME` | blob | Azure storage account name |
| `NIXCOPY_SOURCE_CONTAINER` | blob | Azure container name |
| `NIXCOPY_SOURCE_ACCOUNT_KEY` | blob | Azure account key (`shared_key`) |
| `NIXCOPY_SOURCE_SAS_TOKEN` | blob | SAS token (`sas_token`) |
| `NIXCOPY_SOURCE_CONNECTION_STRING` | blob | Connection string (`connection_string`) |
| `NIXCOPY_SOURCE_TENANT_ID` | blob | Azure tenant ID (`service_principal`) |
| `NIXCOPY_SOURCE_CLIENT_ID` | blob | Azure client ID (SP / user-assigned MI) |
| `NIXCOPY_SOURCE_CLIENT_SECRET` | blob | Azure client secret (`service_principal`) |
| `NIXCOPY_SOURCE_USE_MANAGED_IDENTITY` | blob | Enable managed identity: `true`/`false` |

### Destination Storage

Same set of variables with `NIXCOPY_DEST_` prefix:

| Env Var | Applies to | Description |
|---|---|---|
| `NIXCOPY_DEST_TYPE` | all | Storage type: `local`, `sftp`, `ftps`, `s3`, `blob` |
| `NIXCOPY_DEST_BASE_PATH` | local | Base directory |
| `NIXCOPY_DEST_HOST` | sftp, ftps | Hostname or IP |
| `NIXCOPY_DEST_PORT` | sftp, ftps | Port number |
| `NIXCOPY_DEST_USERNAME` | sftp, ftps | Username |
| `NIXCOPY_DEST_PASSWORD` | sftp, ftps | Password |
| `NIXCOPY_DEST_PRIVATE_KEY` | sftp | Path to private key file |
| `NIXCOPY_DEST_PRIVATE_KEY_PASS` | sftp | Private key passphrase |
| `NIXCOPY_DEST_TIMEOUT` | sftp, ftps | Connection timeout |
| `NIXCOPY_DEST_MAX_PACKET_SIZE` | sftp | Max SFTP packet size |
| `NIXCOPY_DEST_TLS_MODE` | ftps | `explicit` or `implicit` |
| `NIXCOPY_DEST_SKIP_VERIFY` | ftps | Skip TLS cert verify |
| `NIXCOPY_DEST_REGION` | s3 | AWS region |
| `NIXCOPY_DEST_BUCKET` | s3 | S3 bucket name |
| `NIXCOPY_DEST_ENDPOINT` | s3, blob | Custom endpoint URL |
| `NIXCOPY_DEST_USE_PATH_STYLE` | s3 | Path-style URLs |
| `NIXCOPY_DEST_AUTH_TYPE` | s3, blob | Auth type |
| `NIXCOPY_DEST_ACCESS_KEY` | s3 | AWS access key ID |
| `NIXCOPY_DEST_SECRET_KEY` | s3 | AWS secret access key |
| `NIXCOPY_DEST_SESSION_TOKEN` | s3 | AWS session token |
| `NIXCOPY_DEST_ROLE_ARN` | s3 | IAM role ARN |
| `NIXCOPY_DEST_ROLE_SESSION_NAME` | s3 | Role session name |
| `NIXCOPY_DEST_EXTERNAL_ID` | s3 | External ID |
| `NIXCOPY_DEST_WEB_IDENTITY_TOKEN_FILE` | s3 | Token file path (IRSA) |
| `NIXCOPY_DEST_PROFILE` | s3 | AWS credentials profile |
| `NIXCOPY_DEST_ACCOUNT_NAME` | blob | Azure storage account name |
| `NIXCOPY_DEST_CONTAINER` | blob | Azure container name |
| `NIXCOPY_DEST_ACCOUNT_KEY` | blob | Azure account key |
| `NIXCOPY_DEST_SAS_TOKEN` | blob | SAS token |
| `NIXCOPY_DEST_CONNECTION_STRING` | blob | Connection string |
| `NIXCOPY_DEST_TENANT_ID` | blob | Azure tenant ID |
| `NIXCOPY_DEST_CLIENT_ID` | blob | Azure client ID |
| `NIXCOPY_DEST_CLIENT_SECRET` | blob | Azure client secret |
| `NIXCOPY_DEST_USE_MANAGED_IDENTITY` | blob | Enable managed identity |

### Transfer Settings

| Env Var | Default | Description |
|---|---|---|
| `NIXCOPY_BUFFER_SIZE` | `33554432` | Transfer buffer in bytes (32 MB) |
| `NIXCOPY_CONCURRENT_FILES` | `4` | Parallel file transfers |
| `NIXCOPY_RETRY_ATTEMPTS` | `3` | Max retry count per file |
| `NIXCOPY_RETRY_DELAY` | `5s` | Delay between retries (Go duration: `5s`, `1m`) |
| `NIXCOPY_TIMEOUT` | `30m` | Per-transfer timeout |
| `NIXCOPY_VERIFY_CHECKSUM` | `false` | SHA-256 end-to-end integrity check |
| `NIXCOPY_ENABLE_RESUME` | `false` | Resume partial transfers (local & SFTP) |
| `NIXCOPY_SKIP_EXISTING` | `false` | Skip file if destination size matches source |
| `NIXCOPY_BANDWIDTH_LIMIT` | `0` | Max bytes/sec per file (raw integer; `0` = unlimited) |

### Observability / Context

| Env Var | Log field | Default | Purpose |
|---|---|---|---|
| `NIXCOPY_CORRELATION_ID` | `correlation_id` | auto UUID | Carry Airflow DAG run ID through all log lines |
| `NIXCOPY_APP_ID` | `app_id` | `go-nixcopy` | Application identifier |
| `NIXCOPY_APP_VERSION` | `app_version` | `1.0.0` | Inject from image tag |
| `POD_NAME` | `service_pod_name` | _(empty)_ | K8s Downward API pod name |

### Example — Env-var-only (no config file, ideal for KPO)

```bash
export NIXCOPY_SOURCE_TYPE=sftp
export NIXCOPY_SOURCE_HOST=sftp.example.com
export NIXCOPY_SOURCE_PORT=22
export NIXCOPY_SOURCE_USERNAME=sftpuser
export NIXCOPY_SOURCE_PASSWORD=secret

export NIXCOPY_DEST_TYPE=s3
export NIXCOPY_DEST_REGION=ap-southeast-1
export NIXCOPY_DEST_BUCKET=my-bucket
export NIXCOPY_DEST_AUTH_TYPE=web_identity   # IRSA — no keys needed

nixcopy transfer -s "/data/exports/*.csv" -d processed/
```

### ${VAR} Substitution in YAML (alternative)

If you prefer a config file but want to keep secrets out of it, you can still use `${VAR}` syntax inside YAML values — these are expanded at load time:

```yaml
source:
  type: sftp
  sftp:
    host: sftp.example.com
    username: user
    password: ${SFTP_PASSWORD}         # expanded from shell env

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

1. **CLI Flags** (สูงสุด) — override ทุกอย่าง
2. **`NIXCOPY_*` Environment Variables** — overlay บน config file
3. **Config File** (`-c config.yaml`) — ค่าที่โหลดจากไฟล์ YAML
4. **Default Values** (ต่ำสุด) — ใช้เมื่อไม่มีการตั้งค่าใดๆ

### ตัวอย่าง

```bash
# config.yaml มี buffer_size = 32MB
# NIXCOPY_BUFFER_SIZE=67108864 (64MB)
# CLI flag --buffer-size 134217728 (128MB)
# ผลลัพธ์: ใช้ 128MB (จาก CLI flag)

NIXCOPY_BUFFER_SIZE=67108864 nixcopy transfer -c config.yaml \
  --buffer-size 134217728 \
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

### 6. SHA-256 Checksum Verification

ตรวจสอบความสมบูรณ์ของข้อมูลหลังการถ่ายโอน:

```bash
# เปิด checksum ผ่าน CLI flag
nixcopy transfer \
  --source-type local --dest-type sftp \
  --dest-host sftp.example.com --dest-username user --dest-password pass \
  --verify-checksum \
  -s /data/archive.tar.gz -d /backup/archive.tar.gz
```

ผ่าน config file:

```yaml
transfer:
  verify_checksum: true
```

> **หมายเหตุ**: checksum จะถูกข้ามเมื่อใช้ `--resume` และมีไฟล์บางส่วนอยู่แล้ว

### 7. Resume Interrupted Transfer

ต่อการถ่ายโอนที่หยุดกลางคัน (รองรับ Local และ SFTP เท่านั้น):

```bash
# ต่อการโอนที่หยุดกลางคัน
nixcopy transfer \
  --source-type local --dest-type sftp \
  --dest-host sftp.example.com --dest-username user --dest-password pass \
  --resume \
  -s /data/10gb_file.tar.gz -d /backup/10gb_file.tar.gz
```

ผ่าน config file:

```yaml
transfer:
  enable_resume: true
  retry_attempts: 5
```

> **หมายเหตุ**: หาก backend ปลายทางเป็น S3 หรือ Azure Blob จะ fallback เป็น full transfer โดยอัตโนมัติ

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
A: ได้ 2 รูปแบบ: (1) `NIXCOPY_*` env vars อ่านโดยตรงโดย nixcopy (แนะนำสำหรับ K8s/KPO); (2) `${VAR}` ใน YAML config file ที่ถูก expand ตอน load

**Q: จะเก็บ passwords อย่างปลอดภัยได้อย่างไร?**  
A: ใช้ environment variables, AWS Secrets Manager, หรือ Azure Key Vault

**Q: สามารถ override เฉพาะบาง parameters ได้ไหม?**  
A: ได้ ระบุเฉพาะ flags ที่ต้องการ override ส่วนที่เหลือจะใช้จาก config file
