# คู่มือการใช้งานตามสภาพแวดล้อม - go-nixcopy

## 🌍 ภาพรวม Authentication ตามสภาพแวดล้อม

การเลือกใช้ authentication method ขึ้นอยู่กับว่าโปรแกรมรันบน infrastructure ไหน

### ⚠️ กฎสำคัญ

**IAM Role / Instance Profile / Web Identity (AWS)**
- ✅ ใช้ได้เฉพาะบน AWS infrastructure (EC2, ECS, Lambda, EKS)
- ❌ ใช้ไม่ได้บน Azure VM, AKS, หรือ on-premise

**Managed Identity / Workload Identity (Azure)**
- ✅ ใช้ได้เฉพาะบน Azure infrastructure (VM, App Service, AKS)
- ❌ ใช้ไม่ได้บน AWS EC2, EKS, หรือ on-premise

**Access Key / Shared Key / SAS Token**
- ✅ ใช้ได้ทุกที่ (แต่ไม่ secure เท่า managed identity/IAM role)

---

## 📍 สภาพแวดล้อมต่างๆ

### 1️⃣ รันบน AWS EC2/ECS/Lambda

**S3 Authentication:**
```yaml
source:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-bucket
    auth_type: iam_role  # ✅ แนะนำ - ใช้ IAM role
```

**Azure Blob Authentication:**
```yaml
destination:
  type: blob
  blob:
    account_name: mystorageaccount
    container_name: mycontainer
    auth_type: shared_key  # ✅ ต้องใช้ credentials
    account_key: YOUR_KEY
    # หรือ
    # auth_type: sas_token
    # sas_token: "sv=2021..."
```

**ตัวอย่าง:** `examples/aws-ec2-to-azure-blob.yaml`

---

### 2️⃣ รันบน Azure VM/App Service

**Azure Blob Authentication:**
```yaml
source:
  type: blob
  blob:
    account_name: mystorageaccount
    container_name: mycontainer
    auth_type: managed_identity  # ✅ แนะนำ - ใช้ managed identity
    client_id: ""  # ระบุถ้าใช้ user-assigned
```

**S3 Authentication:**
```yaml
destination:
  type: s3
  s3:
    region: us-east-1
    bucket: my-bucket
    auth_type: access_key  # ✅ ต้องใช้ credentials
    access_key_id: AKIA...
    secret_access_key: SECRET...
```

**ตัวอย่าง:** `examples/azure-vm-to-s3.yaml`

---

### 3️⃣ รันบน Amazon EKS (Kubernetes)

**S3 Authentication (IRSA):**
```yaml
source:
  type: s3
  s3:
    region: us-east-1
    bucket: my-bucket
    auth_type: web_identity  # ✅ แนะนำ - ใช้ IRSA
    role_arn: arn:aws:iam::123456789012:role/EKSPodRole
    web_identity_token_file: /var/run/secrets/eks.amazonaws.com/serviceaccount/token
```

**Azure Blob Authentication:**
```yaml
destination:
  type: blob
  blob:
    account_name: mystorageaccount
    container_name: mycontainer
    # Option 1: Service Principal (แนะนำสำหรับ K8s)
    auth_type: service_principal
    tenant_id: your-tenant-id
    client_id: your-client-id
    client_secret: your-client-secret
    # Option 2: Shared Key
    # auth_type: shared_key
    # account_key: YOUR_KEY
```

**Prerequisites:**
- EKS cluster with IRSA enabled
- Service account annotated: `eks.amazonaws.com/role-arn: arn:aws:iam::...`
- IAM role with trust policy for OIDC provider

**ตัวอย่าง:** `examples/eks-irsa-to-azure-blob.yaml`

---

### 4️⃣ รันบน Azure AKS (Kubernetes)

**Azure Blob Authentication (Workload Identity):**
```yaml
source:
  type: blob
  blob:
    account_name: mystorageaccount
    container_name: mycontainer
    auth_type: managed_identity  # ✅ แนะนำ - ใช้ workload identity
    client_id: ""  # อ่านจาก AZURE_CLIENT_ID env var
```

**S3 Authentication:**
```yaml
destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: my-bucket
    auth_type: access_key  # ✅ ต้องใช้ credentials
    access_key_id: AKIA...
    secret_access_key: SECRET...
```

**Prerequisites:**
- AKS cluster with Workload Identity enabled
- Service account with federated identity credential
- Environment variables: `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_FEDERATED_TOKEN_FILE`

**ตัวอย่าง:** `examples/aks-workload-identity-to-s3.yaml`

---

### 5️⃣ รันบน On-Premise / Local Server

**ทุก Storage ต้องใช้ Credentials:**

```yaml
source:
  type: sftp
  sftp:
    host: sftp.company.local
    username: user
    private_key_path: /path/to/key

destination:
  type: s3
  s3:
    region: ap-southeast-1
    bucket: backup-bucket
    auth_type: access_key  # ✅ ต้องใช้ access key
    access_key_id: AKIA...
    secret_access_key: SECRET...

# หรือ Azure Blob
# destination:
#   type: blob
#   blob:
#     account_name: backupaccount
#     container_name: backups
#     auth_type: shared_key  # ✅ ต้องใช้ shared key
#     account_key: YOUR_KEY
```

**ตัวอย่าง:** `examples/on-premise-to-cloud.yaml`

---

## 🔐 ตารางสรุป Authentication Methods

| สภาพแวดล้อม | S3 Auth | Azure Blob Auth |
|-------------|---------|-----------------|
| **AWS EC2/ECS/Lambda** | ✅ `iam_role`<br>✅ `instance_profile`<br>✅ `access_key` | ✅ `shared_key`<br>✅ `sas_token`<br>✅ `service_principal`<br>❌ `managed_identity` |
| **Azure VM/App Service** | ✅ `access_key`<br>❌ `iam_role` | ✅ `managed_identity`<br>✅ `shared_key`<br>✅ `sas_token`<br>✅ `service_principal` |
| **Amazon EKS** | ✅ `web_identity` (IRSA)<br>✅ `access_key`<br>❌ `iam_role` | ✅ `service_principal`<br>✅ `shared_key`<br>✅ `sas_token`<br>❌ `managed_identity` |
| **Azure AKS** | ✅ `access_key`<br>❌ `iam_role`<br>❌ `web_identity` | ✅ `managed_identity` (Workload Identity)<br>✅ `service_principal`<br>✅ `shared_key` |
| **On-Premise/Local** | ✅ `access_key`<br>❌ `iam_role`<br>❌ `web_identity` | ✅ `shared_key`<br>✅ `sas_token`<br>✅ `service_principal`<br>❌ `managed_identity` |

---

## 🎯 Best Practices

### 1. ใช้ Managed Identity/IAM Role เมื่อเป็นไปได้

**ข้อดี:**
- ✅ ไม่ต้องจัดการ credentials
- ✅ Auto-rotation
- ✅ ปลอดภัยกว่า
- ✅ Audit trail ดีกว่า

**ตัวอย่าง:**
```yaml
# บน AWS EC2
s3:
  auth_type: iam_role  # ✅ แนะนำ

# บน Azure VM
blob:
  auth_type: managed_identity  # ✅ แนะนำ
```

### 2. ใช้ Service Principal สำหรับ Kubernetes

**สำหรับ Azure Blob บน EKS:**
```yaml
blob:
  auth_type: service_principal
  tenant_id: ${AZURE_TENANT_ID}
  client_id: ${AZURE_CLIENT_ID}
  client_secret: ${AZURE_CLIENT_SECRET}
```

### 3. ใช้ Environment Variables สำหรับ Secrets

```bash
export AWS_ACCESS_KEY_ID="AKIA..."
export AWS_SECRET_ACCESS_KEY="SECRET..."
export AZURE_STORAGE_KEY="KEY..."
```

```yaml
s3:
  access_key_id: ${AWS_ACCESS_KEY_ID}
  secret_access_key: ${AWS_SECRET_ACCESS_KEY}

blob:
  account_key: ${AZURE_STORAGE_KEY}
```

### 4. ใช้ SAS Token แทน Shared Key เมื่อเป็นไปได้

**ข้อดี:**
- ✅ จำกัดสิทธิ์ได้
- ✅ มีวันหมดอายุ
- ✅ Revoke ได้ง่าย

```yaml
blob:
  auth_type: sas_token
  sas_token: "sv=2021-06-08&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2024-12-31..."
```

---

## ❌ สิ่งที่ไม่สามารถทำได้

### 1. ใช้ IAM Role และ Managed Identity พร้อมกัน

```yaml
# ❌ ไม่สามารถทำได้
source:
  type: s3
  s3:
    auth_type: iam_role  # ต้องรันบน AWS

destination:
  type: blob
  blob:
    auth_type: managed_identity  # ต้องรันบน Azure
```

**เหตุผล:** โปรแกรมรันได้ที่ละ infrastructure เท่านั้น

### 2. ใช้ IAM Role บน Azure VM

```yaml
# ❌ ไม่สามารถทำได้
# รันบน Azure VM
s3:
  auth_type: iam_role  # ใช้ไม่ได้บน Azure
```

**แก้ไข:** ใช้ `access_key` แทน

### 3. ใช้ Managed Identity บน AWS EC2

```yaml
# ❌ ไม่สามารถทำได้
# รันบน AWS EC2
blob:
  auth_type: managed_identity  # ใช้ไม่ได้บน AWS
```

**แก้ไข:** ใช้ `shared_key` หรือ `sas_token` แทน

---

## 📚 ตัวอย่าง Config Files

- `examples/aws-ec2-to-azure-blob.yaml` - รันบน AWS EC2
- `examples/azure-vm-to-s3.yaml` - รันบน Azure VM
- `examples/eks-irsa-to-azure-blob.yaml` - รันบน Amazon EKS
- `examples/aks-workload-identity-to-s3.yaml` - รันบน Azure AKS
- `examples/on-premise-to-cloud.yaml` - รันบน On-premise
- `examples/sftp-to-s3.yaml` - SFTP to S3 (ใช้ได้ทุกที่)

---

## 🆘 Troubleshooting

### Error: "NoCredentialProviders: no valid providers in chain"

**สาเหตุ:** ใช้ `iam_role` บนสภาพแวดล้อมที่ไม่ใช่ AWS

**แก้ไข:**
```yaml
s3:
  auth_type: access_key  # เปลี่ยนเป็น access_key
  access_key_id: AKIA...
  secret_access_key: SECRET...
```

### Error: "DefaultAzureCredential authentication failed"

**สาเหตุ:** ใช้ `managed_identity` บนสภาพแวดล้อมที่ไม่ใช่ Azure

**แก้ไข:**
```yaml
blob:
  auth_type: shared_key  # เปลี่ยนเป็น shared_key
  account_key: YOUR_KEY
```

### Error: "WebIdentityErr: failed to retrieve credentials"

**สาเหตุ:** Token file ไม่พบหรือ IAM role ไม่ถูกต้อง

**ตรวจสอบ:**
1. Token file exists: `/var/run/secrets/eks.amazonaws.com/serviceaccount/token`
2. Service account annotation ถูกต้อง
3. IAM role trust policy ถูกต้อง

---

## 💡 สรุป

1. **เลือก auth method ตามสภาพแวดล้อมที่รัน**
2. **ใช้ managed identity/IAM role เมื่อเป็นไปได้**
3. **ใช้ credentials (access key/shared key) เมื่อจำเป็น**
4. **อย่าผสม AWS auth กับ Azure auth ในสภาพแวดล้อมเดียวกัน**
5. **ใช้ environment variables สำหรับ secrets**
