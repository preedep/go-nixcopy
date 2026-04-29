# go-nixcopy

[🇬🇧 English](README.md) | 🇹🇭 ภาษาไทย

CLI สำหรับถ่ายโอนไฟล์แบบรวดเร็ว รองรับหลาย storage — ย้ายข้อมูลระหว่าง SFTP, FTPS, Azure Blob, AWS S3 และ local disk ด้วยคำสั่งเดียว

[![CI](https://github.com/preedep/go-nixcopy/actions/workflows/ci.yml/badge.svg)](https://github.com/preedep/go-nixcopy/actions/workflows/ci.yml)
[![Docker Hub](https://img.shields.io/docker/v/nickmsft/gonixcopy?label=Docker%20Hub)](https://hub.docker.com/r/nickmsft/gonixcopy)
[![Go Version](https://img.shields.io/github/go-mod/go-version/preedep/go-nixcopy)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## สารบัญ

- [go-nixcopy คืออะไร?](#go-nixcopy-คืออะไร)
- [ฟีเจอร์](#ฟีเจอร์)
- [สถาปัตยกรรม](#สถาปัตยกรรม)
- [วิธีใช้งาน](#วิธีใช้งาน)
- [การเชื่อมต่อกับ Airflow](#การเชื่อมต่อกับ-airflow)
- [แผนงาน](#แผนงาน)

---

## go-nixcopy คืออะไร?

go-nixcopy คือ CLI ขนาดเล็กและรวดเร็วที่เขียนด้วย Go สำหรับถ่ายโอนไฟล์ระหว่าง storage ต่าง ๆ ออกแบบมาให้รันได้ทุกที่ ไม่ว่าจะเป็น bare metal, Kubernetes pod (KPO) หรือ CI/CD pipeline โดยใช้หน่วยความจำน้อยและไม่มี runtime dependency

**Storage backend ที่รองรับ:**

| Backend | อ่าน | เขียน |
|---|---|---|
| Local file system | ✅ | ✅ |
| SFTP | ✅ | ✅ |
| FTPS | ✅ | ✅ |
| Azure Blob Storage | ✅ | ✅ |
| AWS S3 / MinIO | ✅ | ✅ |

**ติดตั้ง:**

```bash
# Homebrew (macOS / Linux)
brew tap preedep/tap && brew install nixcopy

# APT (Debian / Ubuntu)
echo "deb [trusted=yes] https://apt.fury.io/preedep/ /" | sudo tee /etc/apt/sources.list.d/nixcopy.list
sudo apt-get update && sudo apt-get install nixcopy

# YUM (RHEL / CentOS / Fedora)
sudo yum install --repofrompath nixcopy,https://yum.fury.io/preedep/ nixcopy

# ติดตั้งผ่าน Go
go install github.com/preedep/go-nixcopy/cmd/nixcopy@latest

# ติดตั้งผ่าน Docker
docker pull nickmsft/gonixcopy:latest
```

---

## ฟีเจอร์

- **Parallel transfer** — กำหนดจำนวนไฟล์ที่โอนพร้อมกันได้ด้วย semaphore
- **Wildcard patterns** — รองรับ `*.pdf`, `**/*.log`, `data/2024/**/*.csv`
- **On-the-fly compression** — บีบอัดแบบ gzip / zstd แบบ streaming ไม่สร้าง temp file (`--compress gzip`)
- **Bandwidth limiting** — จำกัดความเร็วต่อไฟล์ รองรับ `10MB`, `1GiB` หรือ bytes (`--bandwidth-limit 10MB`)
- **Resume** — ต่อการโอนที่ค้างไว้จาก byte offset ได้ (รองรับ Local & SFTP)
- **Checksum verification** — ตรวจสอบความถูกต้องด้วย SHA-256 แบบ end-to-end (`--verify-checksum`)
- **Skip existing** — ข้ามไฟล์ที่ปลายทางมีอยู่แล้วและขนาดเท่ากัน ทำให้ retry ได้อย่าง idempotent (`--skip-existing`)
- **FTPS explicit & implicit TLS** — `--source-tls-mode explicit` (STARTTLS, port 21) หรือ `implicit` (TLS-first, port 990)
- **Structured JSON logging** — ส่ง log ในรูปแบบ [standard-app-log v1.0](https://github.com/preedep/standard-app-log) ไปยัง stdout รองรับ K8s / Airflow
- **Config-free operation** — กำหนดค่าทั้งหมดผ่าน `NIXCOPY_*` env vars ได้ ไม่ต้องมีไฟล์ YAML ใน container
- **Distroless Docker image** — รันแบบ non-root, ไม่มี shell, รองรับ multi-arch (`linux/amd64` + `linux/arm64`)

---

## สถาปัตยกรรม

Clean Architecture — dependency ไหลเข้าด้านในเสมอ:

```
cmd/nixcopy/main.go
    └─> internal/interfaces/cli/        ← Cobra commands, Viper config binding
            └─> internal/usecase/       ← orchestration, pattern matching, retry
                    └─> internal/domain/      ← entities + repository interfaces
            └─> internal/infrastructure/      ← storage impls, config, logger
```

**โครงสร้างไดเรกทอรี:**

```
go-nixcopy/
├── cmd/nixcopy/               # Entry point
├── internal/
│   ├── domain/                # Entities, repository & service interfaces
│   ├── usecase/               # Transfer orchestration, pattern matcher, mocks
│   ├── infrastructure/
│   │   ├── config/            # YAML/JSON loader with ${ENV_VAR} expansion
│   │   ├── logger/            # standard-app-log JSON emitter
│   │   └── storage/           # local, sftp, ftps, blob, s3, factory
│   └── interfaces/cli/        # Cobra subcommands: transfer, list
└── examples/                  # ตัวอย่าง config ต่อ backend pair
```

---

## วิธีใช้งาน

### เริ่มต้นอย่างรวดเร็ว

```bash
# โอนไฟล์เดี่ยว (ใช้ config file)
nixcopy transfer -c config.yaml -s /remote/file.csv -d backup/file.csv

# โอนไฟล์โดยไม่ใช้ config file (กำหนดทุกอย่างผ่าน CLI flags)
nixcopy transfer \
  --source-type sftp --source-host sftp.example.com \
  --source-username user --source-password pass \
  -s /remote/file.csv \
  --dest-type s3 --dest-region ap-southeast-1 \
  --dest-bucket my-bucket --dest-auth-type iam_role \
  -d processed/file.csv

# โอนไฟล์หลายไฟล์ด้วย wildcard
nixcopy transfer -c config.yaml -s "reports/*.pdf" -d backup/ --concurrent-files 8

# แสดงรายการไฟล์ใน storage
nixcopy list -c config.yaml -p /remote/path --source
```

### Config file

```yaml
source:
  type: sftp          # local | sftp | ftps | blob | s3
  sftp:
    host: sftp.example.com
    port: 22
    username: user
    password: ${SFTP_PASSWORD}

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-bucket
    auth_type: iam_role   # access_key | iam_role | web_identity | assume_role

transfer:
  buffer_size: 33554432    # 32 MB
  concurrent_files: 4
  retry_attempts: 3
  retry_delay: 5s
  verify_checksum: false
  compression: ""          # "gzip" | "zstd" | "" (ไม่บีบอัด)
  bandwidth_limit: 0       # bytes/sec ต่อไฟล์; 0 = ไม่จำกัด
  skip_existing: false
  enable_resume: false
```

**ตัวอย่าง FTPS** (enterprise FTP server มักต้องการ explicit TLS บน port 21):

```yaml
source:
  type: ftps
  ftps:
    host: ftps.enterprise.com
    port: 21
    username: ${FTPS_USER}
    password: ${FTPS_PASSWORD}
    tls_mode: explicit      # explicit (STARTTLS) | implicit (TLS-first, default)
    skip_verify: false      # ตั้งเป็น true เฉพาะใน dev/test เท่านั้น

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: landing-zone
    auth_type: iam_role
```

หรือใช้ CLI flags / env vars โดยไม่ต้องมี config file:

```bash
nixcopy transfer \
  --source-type ftps --source-host ftps.enterprise.com --source-port 21 \
  --source-username user --source-password pass \
  --source-tls-mode explicit \
  --dest-type s3 --dest-region ap-southeast-1 --dest-bucket landing-zone \
  --dest-auth-type iam_role \
  -s "/data/exports/*.csv" -d processed/

# หรือใช้ env vars (เหมาะสำหรับ container / CI)
export NIXCOPY_SOURCE_TYPE=ftps
export NIXCOPY_SOURCE_HOST=ftps.enterprise.com
export NIXCOPY_SOURCE_TLS_MODE=explicit
export NIXCOPY_SOURCE_USERNAME=user
export NIXCOPY_SOURCE_PASSWORD=secret
```

**ลำดับความสำคัญของ config:** `CLI flags > NIXCOPY_* env vars > config file > defaults`

### Docker / Kubernetes

```bash
# รันผ่าน env vars — ไม่ต้องมี config file (แนะนำสำหรับ KPO)
docker run --rm \
  -e NIXCOPY_SOURCE_TYPE=sftp \
  -e NIXCOPY_SOURCE_HOST=sftp.example.com \
  -e NIXCOPY_SOURCE_USERNAME=user \
  -e NIXCOPY_SOURCE_PASSWORD=secret \
  -e NIXCOPY_DEST_TYPE=s3 \
  -e NIXCOPY_DEST_REGION=ap-southeast-1 \
  -e NIXCOPY_DEST_BUCKET=my-bucket \
  -e NIXCOPY_DEST_AUTH_TYPE=web_identity \
  nickmsft/gonixcopy:latest transfer -s /data/file.csv -d processed/file.csv
```

### ผลลัพธ์การโอน

ความคืบหน้าจะแสดงที่ **stderr** (อ่านได้โดยมนุษย์):

```
[file.csv] 45.23% | 125.45 MB/s | ETA: 1m15s
[file.csv] ✓ Completed | 128.32 MB/s
```

สรุปผลในรูปแบบ JSON จะส่งไปที่ **stdout** เมื่อเสร็จสิ้น (อ่านโดยเครื่อง):

```json
{"event":"transfer_summary","total_files":3,"successful":2,"skipped":1,"failed":0,"bytes_transferred":10737418240,"duration_ms":79823,"average_speed_mbps":128.32}
```

### เอกสารเพิ่มเติม

| หัวข้อ | ไฟล์ |
|---|---|
| เริ่มต้น, troubleshooting และ performance tips | [QUICKSTART.md](docs/guides/quickstart.md) |
| CLI flags และ `NIXCOPY_*` env vars ทั้งหมด | [CLI_USAGE.md](docs/guides/cli-usage.md) |
| การตั้งค่า authentication (AWS, Azure) | [AUTHENTICATION.md](docs/guides/authentication.md) |
| Parallel transfer และ wildcard patterns | [PARALLEL_TRANSFER.md](docs/guides/parallel-transfer.md) |
| คู่มือการทดสอบ | [TESTING.md](docs/development/testing.md) |
| Release build flags และขนาด binary | [BUILD.md](docs/development/build.md) |
| สภาพแวดล้อม deployment (EKS, AKS, on-prem) | [ENVIRONMENT_GUIDE.md](docs/guides/environment-guide.md) |
| การ contribute storage backend ใหม่ | [CONTRIBUTING.md](docs/development/contributing.md) |

---

## การเชื่อมต่อกับ Airflow

go-nixcopy ใช้งานได้ทันทีใน Apache Airflow ผ่าน `KubernetesPodOperator` (KPO)  
ไม่จำเป็นต้อง mount config file — ฉีด credential ทั้งหมดผ่าน `NIXCOPY_*` env vars จาก Kubernetes Secret

### การทำงาน

```
Airflow Scheduler
    └─> KubernetesPodOperator
            └─> nickmsft/gonixcopy pod
                    ├── NIXCOPY_* env vars  ← จาก K8s Secret
                    ├── transfer (source → destination)
                    └── stdout: JSON summary  ← parse โดย Airflow log / Loki
```

### ตัวอย่าง DAG — SFTP ไปยัง S3

```python
from datetime import datetime
from airflow import DAG
from airflow.providers.cncf.kubernetes.operators.pod import KubernetesPodOperator
from kubernetes.client import models as k8s

with DAG(
    dag_id="sftp_to_s3_transfer",
    schedule="0 2 * * *",       # ทุกวันเวลา 02:00
    start_date=datetime(2024, 1, 1),
    catchup=False,
    tags=["nixcopy", "sftp", "s3"],
) as dag:

    transfer = KubernetesPodOperator(
        task_id="transfer_sftp_to_s3",
        image="nickmsft/gonixcopy:latest",   # ใช้ tag เฉพาะใน production
        cmds=["./nixcopy"],
        arguments=[
            "transfer",
            "-s", "/data/exports/*.csv",
            "-d", "processed/{{ ds }}/",     # Airflow date partition
            "--concurrent-files", "4",
            "--skip-existing",
        ],
        env_vars=[
            # Observability
            k8s.V1EnvVar(name="NIXCOPY_CORRELATION_ID", value="{{ run_id }}"),
            k8s.V1EnvVar(name="NIXCOPY_APP_VERSION",    value="latest"),
            k8s.V1EnvVar(name="POD_NAME", value_from=k8s.V1EnvVarSource(
                field_ref=k8s.V1ObjectFieldSelector(field_path="metadata.name"))),

            # Source — SFTP credentials จาก K8s Secret
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_TYPE", value="sftp"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_HOST", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="host"))),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_USERNAME", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="username"))),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_PASSWORD", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="password"))),

            # Destination — S3 ผ่าน IRSA (ไม่ต้องใช้ access key)
            k8s.V1EnvVar(name="NIXCOPY_DEST_TYPE",      value="s3"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_REGION",    value="ap-southeast-1"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_BUCKET",    value="my-data-bucket"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_AUTH_TYPE", value="web_identity"),
        ],
        security_context=k8s.V1PodSecurityContext(run_as_non_root=True),
        namespace="airflow",
        service_account_name="nixcopy-sa",   # ผูกกับ IAM role ผ่าน IRSA
        get_logs=True,
        is_delete_operator_pod=True,
        in_cluster=True,
    )
```

### ตัวอย่าง DAG — Azure Blob ไปยัง SFTP

```python
    blob_to_sftp = KubernetesPodOperator(
        task_id="transfer_blob_to_sftp",
        image="nickmsft/gonixcopy:latest",
        cmds=["./nixcopy"],
        arguments=[
            "transfer",
            "-s", "reports/{{ ds }}/*.pdf",
            "-d", "/upload/reports/{{ ds }}/",
        ],
        env_vars=[
            k8s.V1EnvVar(name="NIXCOPY_CORRELATION_ID", value="{{ run_id }}"),

            # Source — Azure Blob ด้วย Managed Identity
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_TYPE",         value="blob"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_ACCOUNT_NAME", value="mystorageaccount"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_CONTAINER",    value="reports"),
            k8s.V1EnvVar(name="NIXCOPY_SOURCE_AUTH_TYPE",    value="managed_identity"),

            # Destination — SFTP
            k8s.V1EnvVar(name="NIXCOPY_DEST_TYPE", value="sftp"),
            k8s.V1EnvVar(name="NIXCOPY_DEST_HOST", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="host"))),
            k8s.V1EnvVar(name="NIXCOPY_DEST_USERNAME", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="username"))),
            k8s.V1EnvVar(name="NIXCOPY_DEST_PASSWORD", value_from=k8s.V1EnvVarSource(
                secret_key_ref=k8s.V1SecretKeySelector(name="sftp-creds", key="password"))),
        ],
        security_context=k8s.V1PodSecurityContext(run_as_non_root=True),
        namespace="airflow",
        get_logs=True,
        is_delete_operator_pod=True,
        in_cluster=True,
    )
```

### ผลลัพธ์ใน Airflow logs

บรรทัดความคืบหน้าจะแสดงที่ **stderr** (เห็นได้ใน Airflow task logs):

```
[report_jan.pdf] 72.10% | 98.32 MB/s | ETA: 0m12s
[report_jan.pdf] ✓ Completed | 101.45 MB/s
```

สรุปผล JSON จะส่งไปที่ **stdout** — parse ด้วย Loki / CloudWatch Insights โดยใช้ `event="transfer_summary"`:

```json
{"event":"transfer_summary","total_files":5,"successful":4,"skipped":1,"failed":0,"bytes_transferred":524288000,"duration_ms":5120,"average_speed_mbps":98.32}
```

> **Tip:** ส่ง `{{ run_id }}` เป็น `NIXCOPY_CORRELATION_ID` เพื่อ correlate log ทุกบรรทัดจาก DAG run เดียวกันใน Loki / CloudWatch

---

## แผนงาน

### เสร็จแล้ว
- [x] รองรับ Local, SFTP, FTPS, Azure Blob, AWS S3 / MinIO
- [x] Parallel transfer และ wildcard patterns (`*.pdf`, `**/*.log`)
- [x] ตรวจสอบ checksum ด้วย SHA-256
- [x] Resume การโอนที่ค้างไว้ (Local & SFTP)
- [x] จำกัด bandwidth ต่อไฟล์
- [x] บีบอัดแบบ on-the-fly ด้วย gzip / zstd
- [x] Idempotent retry (`--skip-existing`)
- [x] Structured JSON logging (standard-app-log v1.0)
- [x] Distroless Docker image, multi-arch (amd64 + arm64), รองรับ config-free KPO
- [x] S3 multipart upload (สูงสุด ~5 TiB) และ Azure Blob parallel block upload (สูงสุด ~190 TiB)
- [x] GitHub Actions CI พร้อม integration tests (MinIO + SFTP)

### แผนงานในอนาคต
- [ ] รองรับ Google Cloud Storage
- [ ] Web UI สำหรับจัดการการโอน
- [ ] Scheduled transfers
- [ ] แจ้งเตือนผ่าน Email / webhook
- [ ] Incremental backup

---

## การมีส่วนร่วม

ยินดีรับ bug report และ pull request ที่ [GitHub Issues](https://github.com/preedep/go-nixcopy/issues)  
ดู [CONTRIBUTING.md](docs/development/contributing.md) สำหรับ workflow การพัฒนาและ code template

## License

MIT — ดู [LICENSE](LICENSE)

---

สร้างด้วย ❤️ ในประเทศไทย
