# คู่มือการ Authentication - go-nixcopy

เอกสารนี้อธิบายวิธีการ authentication ทั้งหมดที่รองรับสำหรับแต่ละ cloud storage provider

## 📑 สารบัญ

- [AWS S3 Authentication](#aws-s3-authentication)
- [Azure Blob Storage Authentication](#azure-blob-storage-authentication)
- [ตัวอย่างการใช้งาน](#ตัวอย่างการใช้งาน)
- [Best Practices](#best-practices)

---

## AWS S3 Authentication

S3 รองรับวิธีการ authentication หลายแบบตาม use case ที่แตกต่างกัน

### 1. Access Key (Static Credentials)

**เหมาะสำหรับ:** Development, testing, หรือการรันบนเครื่อง local

```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: access_key
  access_key_id: AKIAIOSFODNN7EXAMPLE
  secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
  session_token: ""  # optional, สำหรับ temporary credentials
```

**ข้อดี:**
- ง่ายต่อการตั้งค่า
- ใช้งานได้ทุกที่

**ข้อเสีย:**
- ต้องจัดการ credentials เอง
- มีความเสี่ยงด้านความปลอดภัยถ้าเก็บไม่ดี

---

### 2. IAM Role

**เหมาะสำหรับ:** EC2 instances, ECS tasks, Lambda functions

```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: iam_role
```

**วิธีการตั้งค่า:**

1. สร้าง IAM Role ที่มี S3 permissions
2. Attach role กับ EC2 instance หรือ ECS task
3. ไม่ต้องระบุ credentials ในไฟล์ config

**ข้อดี:**
- ปลอดภัยที่สุด - ไม่ต้องเก็บ credentials
- Credentials rotate อัตโนมัติ
- ใช้ AWS IAM policies ควบคุม permissions

**ข้อเสีย:**
- ใช้ได้เฉพาะบน AWS resources

---

### 3. Instance Profile

**เหมาะสำหรับ:** EC2 instances

```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: instance_profile
```

**วิธีการตั้งค่า:**

1. สร้าง IAM Role
2. สร้าง Instance Profile และ attach role
3. Launch EC2 instance พร้อม instance profile

**ข้อดี:**
- เหมือน IAM Role แต่เฉพาะเจาะจงสำหรับ EC2
- Credentials rotate อัตโนมัติ

---

### 4. Assume Role

**เหมาะสำหรับ:** Cross-account access, temporary elevated permissions

```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: assume_role
  role_arn: arn:aws:iam::123456789012:role/MyRole
  role_session_name: nixcopy-session
  external_id: my-external-id  # optional, สำหรับ third-party access
```

**วิธีการตั้งค่า:**

1. สร้าง IAM Role ใน target account
2. ตั้งค่า trust relationship ให้ source account assume ได้
3. ระบุ role ARN ในไฟล์ config

**ข้อดี:**
- เข้าถึง resources ใน account อื่นได้
- Temporary credentials
- Audit trail ชัดเจน

**Use Cases:**
- Cross-account data transfer
- Third-party access with external ID
- Temporary elevated permissions

---

### 5. Web Identity (OIDC/IRSA)

**เหมาะสำหรับ:** Kubernetes (EKS), containerized applications

```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: web_identity
  role_arn: arn:aws:iam::123456789012:role/EKSPodRole
  web_identity_token_file: /var/run/secrets/eks.amazonaws.com/serviceaccount/token
```

**วิธีการตั้งค่า (EKS IRSA):**

1. สร้าง OIDC provider สำหรับ EKS cluster
2. สร้าง IAM Role พร้อม trust relationship กับ OIDC provider
3. สร้าง Kubernetes ServiceAccount และ annotate ด้วย role ARN
4. Pod จะได้ token อัตโนมัติ

**ข้อดี:**
- ปลอดภัยสำหรับ Kubernetes workloads
- ไม่ต้องเก็บ AWS credentials ใน cluster
- Fine-grained permissions per pod

---

### 6. AWS Profile

**เหมาะสำหรับ:** Development, multiple AWS accounts

```yaml
s3:
  region: ap-southeast-1
  bucket: my-bucket
  auth_type: access_key
  profile: production  # ใช้ profile จาก ~/.aws/credentials
```

**ไฟล์ ~/.aws/credentials:**
```ini
[production]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[staging]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
```

**ข้อดี:**
- จัดการหลาย accounts ได้ง่าย
- แยก credentials ออกจาก application config

---

## Azure Blob Storage Authentication

Azure Blob Storage รองรับวิธีการ authentication หลายแบบ

### 1. Shared Key

**เหมาะสำหรับ:** Development, testing

```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: shared_key
  account_key: YOUR_ACCOUNT_KEY
```

**วิธีหา Account Key:**
1. ไปที่ Azure Portal → Storage Account
2. Settings → Access keys
3. คัดลอก key1 หรือ key2

**ข้อดี:**
- ง่ายต่อการตั้งค่า
- Full access ทุก operations

**ข้อเสีย:**
- ความเสี่ยงด้านความปลอดภัยสูงถ้า key รั่วไหล
- ไม่มี fine-grained permissions

---

### 2. SAS Token (Shared Access Signature)

**เหมาะสำหรับ:** Temporary access, limited permissions

```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: sas_token
  sas_token: "sv=2021-06-08&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2024-12-31T23:59:59Z&st=2024-01-01T00:00:00Z&spr=https&sig=SIGNATURE"
```

**วิธีสร้าง SAS Token:**

1. Azure Portal → Storage Account → Shared access signature
2. เลือก permissions, expiry time
3. Generate SAS and connection string
4. คัดลอก SAS token (ส่วนที่ขึ้นต้นด้วย `sv=`)

**ข้อดี:**
- จำกัด permissions ได้ (read-only, write-only, etc.)
- มี expiration time
- Revoke ได้โดยการ regenerate account key

**ข้อเสีย:**
- ต้องสร้างใหม่เมื่อหมดอายุ

**SAS Token Types:**
- **Account SAS:** Access หลาย services (blob, file, queue, table)
- **Service SAS:** Access เฉพาะ service เดียว
- **User Delegation SAS:** ใช้ Azure AD credentials (ปลอดภัยที่สุด)

---

### 3. Connection String

**เหมาะสำหรับ:** Quick setup, legacy applications

```yaml
blob:
  auth_type: connection_string
  connection_string: "DefaultEndpointsProtocol=https;AccountName=mystorageaccount;AccountKey=YOUR_KEY;EndpointSuffix=core.windows.net"
  container_name: mycontainer
```

**วิธีหา Connection String:**
1. Azure Portal → Storage Account
2. Settings → Access keys
3. คัดลอก Connection string

**ข้อดี:**
- ตั้งค่าง่าย - ใส่ string เดียว
- รวม account name, key, และ endpoint

**ข้อเสีย:**
- เหมือน Shared Key - มีความเสี่ยงด้านความปลอดภัย

---

### 4. Managed Identity

**เหมาะสำหรับ:** Azure VMs, App Services, Azure Functions, AKS

```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: managed_identity
  client_id: ""  # ระบุเฉพาะถ้าใช้ user-assigned managed identity
```

**วิธีการตั้งค่า (System-assigned):**

1. Enable managed identity บน Azure resource (VM, App Service, etc.)
2. ไปที่ Storage Account → Access Control (IAM)
3. Add role assignment: "Storage Blob Data Contributor" ให้กับ managed identity
4. ไม่ต้องระบุ credentials ในไฟล์ config

**วิธีการตั้งค่า (User-assigned):**

1. สร้าง User-assigned Managed Identity
2. Assign identity ให้กับ Azure resource
3. Add role assignment ที่ Storage Account
4. ระบุ `client_id` ในไฟล์ config

**ข้อดี:**
- ปลอดภัยที่สุด - ไม่ต้องเก็บ credentials
- Credentials managed โดย Azure
- Fine-grained RBAC permissions

**ข้อเสีย:**
- ใช้ได้เฉพาะบน Azure resources

**Managed Identity Types:**
- **System-assigned:** ผูกติดกับ resource เดียว, ลบตาม resource
- **User-assigned:** แชร์ได้หลาย resources, จัดการแยกต่างหาก

---

### 5. Service Principal

**เหมาะสำหรับ:** Applications, CI/CD pipelines, cross-tenant access

```yaml
blob:
  account_name: mystorageaccount
  container_name: mycontainer
  auth_type: service_principal
  tenant_id: your-tenant-id
  client_id: your-client-id
  client_secret: your-client-secret
```

**วิธีการตั้งค่า:**

1. สร้าง App Registration ใน Azure AD
   ```bash
   az ad sp create-for-rbac --name nixcopy-sp
   ```

2. บันทึก output:
   - `appId` → client_id
   - `password` → client_secret
   - `tenant` → tenant_id

3. Add role assignment ที่ Storage Account:
   ```bash
   az role assignment create \
     --role "Storage Blob Data Contributor" \
     --assignee <client_id> \
     --scope /subscriptions/<subscription-id>/resourceGroups/<rg>/providers/Microsoft.Storage/storageAccounts/<account>
   ```

**ข้อดี:**
- ใช้ได้นอก Azure (on-premises, other clouds)
- Fine-grained RBAC permissions
- Audit trail ชัดเจน

**ข้อเสีย:**
- ต้องจัดการ client secret
- Secret มี expiration date

**Best Practices:**
- Rotate client secrets เป็นประจำ
- ใช้ Azure Key Vault เก็บ secrets
- ใช้ certificate-based authentication แทน client secret (ปลอดภัยกว่า)

---

## ตัวอย่างการใช้งาน

### Scenario 1: EC2 to Azure VM

**Source (EC2):** ใช้ IAM Role  
**Destination (Azure VM):** ใช้ Managed Identity

```yaml
source:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: source-bucket
    auth_type: iam_role

destination:
  type: blob
  blob:
    account_name: deststorage
    container_name: dest-container
    auth_type: managed_identity
```

---

### Scenario 2: On-premises to Cloud

**Source (SFTP):** Username/Password  
**Destination (S3):** Access Key

```yaml
source:
  type: sftp
  sftp:
    host: onprem-sftp.company.com
    port: 22
    username: transfer-user
    password: ${SFTP_PASSWORD}  # จาก environment variable

destination:
  type: s3
  s3:
    region: us-east-1
    bucket: cloud-backup
    auth_type: access_key
    access_key_id: ${AWS_ACCESS_KEY_ID}
    secret_access_key: ${AWS_SECRET_ACCESS_KEY}
```

---

### Scenario 3: Kubernetes (EKS) to Azure

**Source (S3):** Web Identity (IRSA)  
**Destination (Blob):** Service Principal

```yaml
source:
  type: s3
  s3:
    region: us-west-2
    bucket: k8s-data
    auth_type: web_identity
    role_arn: arn:aws:iam::123456789012:role/EKSPodRole
    web_identity_token_file: /var/run/secrets/eks.amazonaws.com/serviceaccount/token

destination:
  type: blob
  blob:
    account_name: azurestorage
    container_name: k8s-backup
    auth_type: service_principal
    tenant_id: ${AZURE_TENANT_ID}
    client_id: ${AZURE_CLIENT_ID}
    client_secret: ${AZURE_CLIENT_SECRET}
```

---

### Scenario 4: Cross-Account S3 Transfer

**Source:** Account A (Assume Role)  
**Destination:** Account B (IAM Role)

```yaml
source:
  type: s3
  s3:
    region: us-east-1
    bucket: account-a-bucket
    auth_type: assume_role
    role_arn: arn:aws:iam::111111111111:role/SourceRole
    role_session_name: nixcopy-source
    external_id: shared-secret

destination:
  type: s3
  s3:
    region: us-east-1
    bucket: account-b-bucket
    auth_type: iam_role
```

---

## Best Practices

### 🔐 Security

1. **ใช้ IAM Roles/Managed Identity เมื่อเป็นไปได้**
   - ไม่ต้องเก็บ credentials
   - Automatic rotation
   - Audit trail ดีกว่า

2. **ใช้ Environment Variables สำหรับ Secrets**
   ```yaml
   access_key_id: ${AWS_ACCESS_KEY_ID}
   secret_access_key: ${AWS_SECRET_ACCESS_KEY}
   ```

3. **ใช้ Least Privilege Principle**
   - ให้ permissions เท่าที่จำเป็นเท่านั้น
   - ใช้ SAS token แทน shared key เมื่อเป็นไปได้

4. **Rotate Credentials เป็นประจำ**
   - Access keys: ทุก 90 วัน
   - Service principal secrets: ทุก 6-12 เดือน

5. **เก็บ Config Files อย่างปลอดภัย**
   ```bash
   chmod 600 config.yaml
   ```

6. **ใช้ Azure Key Vault / AWS Secrets Manager**
   - เก็บ secrets แบบ centralized
   - Automatic rotation
   - Access logging

---

### 🚀 Performance

1. **เลือก Region ใกล้กัน**
   - ลด latency
   - ลดค่าใช้จ่าย data transfer

2. **ปรับ Buffer Size**
   - ไฟล์เล็ก: 8-16 MB
   - ไฟล์ใหญ่: 32-64 MB

3. **ใช้ Concurrent Transfers**
   - Network ดี: 8-16 concurrent files
   - Network ปานกลาง: 4-8 concurrent files

---

### 📊 Monitoring

1. **Enable Logging**
   ```yaml
   logging:
     level: info
     format: json
     output_path: /var/log/nixcopy.log
   ```

2. **Monitor Costs**
   - AWS: CloudWatch, Cost Explorer
   - Azure: Cost Management

3. **Track Transfer Metrics**
   - Transfer speed
   - Success/failure rates
   - Data volume

---

### 🔄 Automation

1. **CI/CD Integration**
   ```yaml
   # GitHub Actions example
   - name: Transfer files
     env:
       AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
       AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
     run: |
       nixcopy transfer -c config.yaml -s /data -d /backup
   ```

2. **Scheduled Transfers**
   ```bash
   # Cron job
   0 2 * * * /usr/local/bin/nixcopy transfer -c /etc/nixcopy/config.yaml
   ```

---

## 🆘 Troubleshooting

### AWS S3 Issues

**Error: "Access Denied"**
- ตรวจสอบ IAM permissions
- ตรวจสอบ bucket policy
- ตรวจสอบ VPC endpoint (ถ้าใช้)

**Error: "No credentials found"**
- ตรวจสอบ environment variables
- ตรวจสอบ ~/.aws/credentials
- ตรวจสอบ IAM role attachment

---

### Azure Blob Issues

**Error: "Authentication failed"**
- ตรวจสอบ RBAC role assignments
- ตรวจสอบ managed identity status
- ตรวจสอบ service principal credentials

**Error: "SAS token expired"**
- สร้าง SAS token ใหม่
- ตรวจสอบ system time

---

## 📚 เอกสารอ้างอิง

- [AWS IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [Azure Managed Identities](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/)
- [AWS STS AssumeRole](https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html)
- [Azure Storage SAS](https://docs.microsoft.com/en-us/azure/storage/common/storage-sas-overview)
