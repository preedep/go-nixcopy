# คู่มือการถ่ายโอนไฟล์แบบ Parallel - go-nixcopy

เอกสารนี้อธิบายวิธีการถ่ายโอนหลายไฟล์พร้อมกัน (parallel transfer) รองรับ wildcard patterns และ batch operations

## 📋 สารบัญ

- [คุณสมบัติ](#คุณสมบัติ)
- [Wildcard Patterns](#wildcard-patterns)
- [การใช้งาน](#การใช้งาน)
- [ตัวอย่างการใช้งาน](#ตัวอย่างการใช้งาน)
- [Performance Tuning](#performance-tuning)
- [Best Practices](#best-practices)

---

## คุณสมบัติ

### ✨ Parallel Transfer
- ถ่ายโอนหลายไฟล์พร้อมกัน (concurrent)
- ปรับจำนวน concurrent transfers ได้
- Progress tracking แบบ real-time สำหรับแต่ละไฟล์
- Automatic retry สำหรับไฟล์ที่ล้มเหลว

### 🎯 Pattern Matching
- รองรับ wildcard patterns (`*`, `?`, `[]`)
- รองรับ recursive patterns (`**`)
- ค้นหาไฟล์อัตโนมัติตาม pattern
- Filter ตาม file extension

### 📊 Batch Operations
- ถ่ายโอนหลายไฟล์ในคำสั่งเดียว
- Summary report หลังเสร็จ
- แสดงรายการไฟล์ที่สำเร็จและล้มเหลว
- คำนวณ total speed และ duration

---

## Wildcard Patterns

### Basic Wildcards

| Pattern | Description | Example |
|---------|-------------|---------|
| `*` | Match ทุกอักษร (ยกเว้น `/`) | `*.pdf` = ทุกไฟล์ .pdf |
| `?` | Match อักษรเดียว | `file?.txt` = file1.txt, fileA.txt |
| `[abc]` | Match อักษรใน set | `file[123].txt` = file1.txt, file2.txt, file3.txt |
| `[a-z]` | Match range | `file[a-z].txt` = filea.txt ถึง filez.txt |

### Recursive Patterns

| Pattern | Description | Example |
|---------|-------------|---------|
| `**/*.pdf` | ทุกไฟล์ .pdf ใน subdirectories ทั้งหมด | รวม nested folders |
| `data/**/*.csv` | ทุกไฟล์ .csv ใน data/ และ subdirectories | |
| `**/backup/*.zip` | ทุกไฟล์ .zip ใน folder ชื่อ backup | |

### ตัวอย่าง Patterns

```bash
# ไฟล์ PDF ทั้งหมดใน directory ปัจจุบัน
*.pdf

# ไฟล์ที่ขึ้นต้นด้วย report
report*.xlsx

# ไฟล์ที่ลงท้ายด้วยตัวเลข
file[0-9].txt

# ทุกไฟล์ .log ใน subdirectories ทั้งหมด
**/*.log

# ไฟล์ backup ทั้งหมด
backup_*.tar.gz

# ไฟล์ที่มีวันที่ในชื่อ
data_2024*.csv
```

---

## การใช้งาน

### 1. ถ่ายโอนไฟล์เดียว

```bash
nixcopy transfer -c config.yaml \
  -s /data/file.pdf \
  -d /backup/file.pdf
```

### 2. ถ่ายโอนหลายไฟล์ (ระบุชื่อชัดเจน)

```bash
nixcopy transfer -c config.yaml \
  --sources /data/file1.pdf,/data/file2.pdf,/data/file3.pdf \
  -d /backup/
```

### 3. ถ่ายโอนด้วย Wildcard Pattern

```bash
# ถ่ายโอนไฟล์ PDF ทั้งหมด
nixcopy transfer -c config.yaml \
  -s "*.pdf" \
  -d /backup/

# ถ่ายโอนไฟล์ที่ขึ้นต้นด้วย report
nixcopy transfer -c config.yaml \
  -s "report*.xlsx" \
  -d /backup/
```

### 4. ถ่ายโอนแบบ Recursive

```bash
# ถ่ายโอนไฟล์ .log ทั้งหมดใน subdirectories
nixcopy transfer -c config.yaml \
  -s "**/*.log" \
  -d /backup/logs/
```

### 5. ผสม Patterns หลายแบบ

```bash
nixcopy transfer -c config.yaml \
  --sources "*.pdf,*.docx,reports/*.xlsx" \
  -d /backup/documents/
```

---

## ตัวอย่างการใช้งาน

### ตัวอย่าง 1: Backup ไฟล์ PDF ทั้งหมดจาก SFTP ไปยัง S3

```bash
nixcopy transfer \
  --source-type sftp \
  --source-host sftp.example.com \
  --source-username user \
  --source-password "$SFTP_PASSWORD" \
  -s "documents/*.pdf" \
  --dest-type s3 \
  --dest-region ap-southeast-1 \
  --dest-bucket backup-bucket \
  --dest-auth-type iam_role \
  -d backups/documents/ \
  --concurrent-files 8
```

**Output:**
```
[invoice_001.pdf] ✓ Completed | 45.32 MB/s
[invoice_002.pdf] ✓ Completed | 48.21 MB/s
[report_2024.pdf] ✓ Completed | 52.15 MB/s
...

=== Transfer Summary ===
Total Files: 150
Successful: 150
Failed: 0
Total Bytes: 2147483648 (2048.00 MB)
Total Duration: 45.234s
Average Speed: 45.28 MB/s
```

### ตัวอย่าง 2: ถ่ายโอน Log Files แบบ Recursive

```bash
nixcopy transfer -c config.yaml \
  -s "logs/**/*.log" \
  -d /backup/logs/ \
  --concurrent-files 16
```

### ตัวอย่าง 3: Backup Database Dumps

```bash
nixcopy transfer -c config.yaml \
  -s "backups/db_backup_*.sql.gz" \
  -d /s3-backup/databases/ \
  --concurrent-files 4 \
  --buffer-size 67108864
```

### ตัวอย่าง 4: ถ่ายโอนไฟล์หลาย Types

```bash
nixcopy transfer -c config.yaml \
  --sources "*.pdf,*.docx,*.xlsx,images/*.jpg" \
  -d /backup/office-files/ \
  --concurrent-files 12
```

### ตัวอย่าง 5: Azure Blob to S3 (Parallel)

```bash
nixcopy transfer \
  --source-type blob \
  --source-account-name sourcestorage \
  --source-container sourcecontainer \
  --source-auth-type managed_identity \
  -s "data/2024/**/*.csv" \
  --dest-type s3 \
  --dest-region us-east-1 \
  --dest-bucket dest-bucket \
  --dest-auth-type access_key \
  --dest-access-key "$AWS_ACCESS_KEY_ID" \
  --dest-secret-key "$AWS_SECRET_ACCESS_KEY" \
  -d analytics/2024/ \
  --concurrent-files 20 \
  --buffer-size 134217728
```

### ตัวอย่าง 6: ใช้ Config File + Pattern

**config.yaml:**
```yaml
source:
  type: sftp
  sftp:
    host: sftp.example.com
    port: 22
    username: user
    private_key_path: ~/.ssh/id_rsa

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: backup-bucket
    auth_type: iam_role

transfer:
  buffer_size: 67108864      # 64MB
  concurrent_files: 16       # 16 files พร้อมกัน
  retry_attempts: 5
  retry_delay: 10s
  timeout: 60m
```

**Command:**
```bash
nixcopy transfer -c config.yaml \
  -s "reports/**/*.xlsx" \
  -d backups/reports/
```

---

## Performance Tuning

### จำนวน Concurrent Files

| Use Case | Concurrent Files | เหมาะสำหรับ |
|----------|------------------|-------------|
| Small files (< 1MB) | 16-32 | ไฟล์เล็กจำนวนมาก |
| Medium files (1-100MB) | 8-16 | ไฟล์ขนาดกลาง |
| Large files (> 100MB) | 2-8 | ไฟล์ขนาดใหญ่ |
| Very large files (> 1GB) | 1-4 | ไฟล์ขนาดใหญ่มาก |

### Buffer Size

| File Size | Buffer Size | Setting |
|-----------|-------------|---------|
| < 10MB | 8-16 MB | `--buffer-size 16777216` |
| 10-100MB | 32-64 MB | `--buffer-size 67108864` |
| > 100MB | 64-128 MB | `--buffer-size 134217728` |

### Network Optimization

**Fast Network (> 1 Gbps):**
```bash
nixcopy transfer -c config.yaml \
  -s "*.zip" \
  -d /backup/ \
  --concurrent-files 32 \
  --buffer-size 134217728
```

**Slow Network (< 100 Mbps):**
```bash
nixcopy transfer -c config.yaml \
  -s "*.zip" \
  -d /backup/ \
  --concurrent-files 4 \
  --buffer-size 16777216
```

**Unstable Network:**
```bash
nixcopy transfer -c config.yaml \
  -s "*.zip" \
  -d /backup/ \
  --concurrent-files 2 \
  --retry-attempts 10 \
  --retry-delay 30s
```

---

## Best Practices

### 1. ใช้ Patterns อย่างระมัดระวัง

```bash
# ดี: Specific pattern
nixcopy transfer -c config.yaml -s "reports/2024/*.pdf" -d /backup/

# ไม่ดี: Too broad, อาจได้ไฟล์มากเกินไป
nixcopy transfer -c config.yaml -s "**/*" -d /backup/
```

### 2. ทดสอบ Pattern ก่อน

```bash
# ใช้ list command เพื่อดูว่า pattern match ไฟล์อะไรบ้าง
nixcopy list -c config.yaml -p "reports/*.pdf" --source
```

### 3. ปรับ Concurrent Files ตาม Resource

```bash
# เช็ค CPU และ Memory ก่อนตั้งค่า concurrent files สูง
# Rule of thumb: concurrent_files = (Available CPU cores) * 2
```

### 4. ใช้ Retry สำหรับ Network ไม่เสถียร

```bash
nixcopy transfer -c config.yaml \
  -s "*.zip" \
  -d /backup/ \
  --retry-attempts 10 \
  --retry-delay 30s
```

### 5. Monitor Progress

```bash
# ใช้ verbose mode เพื่อดู detailed logs
nixcopy transfer -c config.yaml -v \
  -s "**/*.log" \
  -d /backup/logs/
```

### 6. Batch Similar Files

```bash
# ดี: Group ไฟล์ขนาดใกล้เคียงกัน
nixcopy transfer -c config.yaml -s "small_files/*.txt" -d /backup/
nixcopy transfer -c config.yaml -s "large_files/*.zip" -d /backup/ --concurrent-files 4

# ไม่ดี: Mix ไฟล์ขนาดต่างกันมาก
nixcopy transfer -c config.yaml -s "**/*" -d /backup/
```

### 7. Use Appropriate Auth Methods

```bash
# Production: ใช้ IAM Role / Managed Identity
nixcopy transfer \
  --source-type s3 \
  --source-auth-type iam_role \
  -s "*.pdf" \
  --dest-type blob \
  --dest-auth-type managed_identity \
  -d /backup/

# Development: ใช้ Access Keys
nixcopy transfer \
  --source-type s3 \
  --source-auth-type access_key \
  --source-access-key "$AWS_KEY" \
  -s "*.pdf" \
  -d /backup/
```

---

## Advanced Examples

### ตัวอย่าง 1: Incremental Backup

```bash
#!/bin/bash
# backup_daily.sh

DATE=$(date +%Y%m%d)
PATTERN="data/daily_${DATE}_*.csv"

nixcopy transfer -c config.yaml \
  -s "$PATTERN" \
  -d "backups/daily/$DATE/" \
  --concurrent-files 16
```

### ตัวอย่าง 2: Multi-Pattern Transfer

```bash
#!/bin/bash
# backup_documents.sh

PATTERNS=(
  "documents/*.pdf"
  "spreadsheets/*.xlsx"
  "presentations/*.pptx"
  "images/**/*.jpg"
)

for pattern in "${PATTERNS[@]}"; do
  echo "Transferring: $pattern"
  nixcopy transfer -c config.yaml \
    -s "$pattern" \
    -d "backups/documents/" \
    --concurrent-files 8
done
```

### ตัวอย่าง 3: Conditional Transfer

```bash
#!/bin/bash
# backup_large_files.sh

# ถ่ายโอนเฉพาะไฟล์ที่ใหญ่กว่า 100MB
nixcopy transfer -c config.yaml \
  -s "data/*.zip" \
  -d "backups/large/" \
  --concurrent-files 4 \
  --buffer-size 134217728
```

---

## Troubleshooting

### ปัญหา: Pattern ไม่ match ไฟล์

**สาเหตุ:** Pattern syntax ผิด หรือ path ไม่ถูกต้อง

**แก้ไข:**
```bash
# ทดสอบ pattern ด้วย list command
nixcopy list -c config.yaml -p "your/pattern/*.pdf" --source

# ใช้ quotes สำหรับ patterns
nixcopy transfer -c config.yaml -s "*.pdf" -d /backup/  # ถูก
nixcopy transfer -c config.yaml -s *.pdf -d /backup/    # อาจผิด (shell expansion)
```

### ปัญหา: Out of Memory

**สาเหตุ:** Concurrent files สูงเกินไป

**แก้ไข:**
```bash
# ลด concurrent files
nixcopy transfer -c config.yaml \
  -s "*.zip" \
  -d /backup/ \
  --concurrent-files 4 \
  --buffer-size 33554432
```

### ปัญหา: Slow Transfer

**สาเหตุ:** Concurrent files ต่ำเกินไป หรือ buffer size เล็กเกินไป

**แก้ไข:**
```bash
# เพิ่ม concurrent files และ buffer size
nixcopy transfer -c config.yaml \
  -s "*.pdf" \
  -d /backup/ \
  --concurrent-files 16 \
  --buffer-size 67108864
```

### ปัญหา: Some Files Failed

**สาเหตุ:** Network issues, permissions, หรือ file locks

**แก้ไข:**
```bash
# เพิ่ม retry attempts
nixcopy transfer -c config.yaml \
  -s "*.zip" \
  -d /backup/ \
  --retry-attempts 10 \
  --retry-delay 30s
```

---

## 📚 เอกสารเพิ่มเติม

- [README.md](README.md) - คู่มือหลัก
- [CLI_USAGE.md](CLI_USAGE.md) - คู่มือ CLI parameters
- [AUTHENTICATION.md](AUTHENTICATION.md) - คู่มือ authentication
- [QUICKSTART.md](QUICKSTART.md) - คู่มือเริ่มต้นใช้งาน

---

## 💡 Tips

1. **ทดสอบก่อน:** ใช้ `list` command ทดสอบ pattern ก่อนถ่ายโอนจริง
2. **Start Small:** เริ่มด้วย concurrent files น้อยๆ แล้วค่อยเพิ่ม
3. **Monitor Resources:** ดู CPU, Memory, Network usage ขณะถ่ายโอน
4. **Use Logging:** เปิด verbose mode เพื่อ debug
5. **Batch Wisely:** แบ่งไฟล์เป็น batches ตามขนาดและประเภท
