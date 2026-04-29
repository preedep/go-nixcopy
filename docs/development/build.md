# คู่มือการ Build - go-nixcopy

## การ Build Release Version

### Build สำหรับ Platform ปัจจุบัน

```bash
./build-release.sh
# หรือ
./build-release.sh current
```

Output: `dist/nixcopy`

### Build สำหรับทุก Platforms

```bash
./build-release.sh all
```

Output:
- `dist/nixcopy-linux-amd64`
- `dist/nixcopy-linux-arm64`
- `dist/nixcopy-linux-386`
- `dist/nixcopy-darwin-amd64` (macOS Intel)
- `dist/nixcopy-darwin-arm64` (macOS Apple Silicon)
- `dist/nixcopy-windows-amd64.exe`
- `dist/nixcopy-windows-386.exe`
- `dist/nixcopy-freebsd-amd64`

### Build สำหรับ Platform เฉพาะ

```bash
# Linux AMD64
./build-release.sh linux amd64

# macOS ARM64 (Apple Silicon)
./build-release.sh darwin arm64

# Windows AMD64
./build-release.sh windows amd64
```

### Build พร้อม Version

```bash
VERSION=1.2.3 ./build-release.sh all
```

## Build Flags

Script ใช้ build flags ต่อไปนี้เพื่อ optimize binary:

### `-ldflags="-s -w"`
- `-s`: ลบ symbol table (ลด binary size ~30%)
- `-w`: ลบ DWARF debugging information (ลด binary size เพิ่มเติม)

### `-trimpath`
- ลบ absolute file paths จาก binary
- ทำให้ builds reproducible
- เพิ่ม security (ไม่เปิดเผย local paths)

### Version Information
- `main.Version`: เวอร์ชันของโปรแกรม
- `main.BuildTime`: เวลาที่ build
- `main.GitCommit`: Git commit hash

## การใช้งาน Makefile

เพิ่ม targets ใน Makefile:

```bash
# Build release version
make release

# Build สำหรับทุก platforms
make release-all

# Clean build artifacts
make clean-dist
```

## ขนาด Binary

### Development Build (with debug symbols)
```bash
go build -o nixcopy cmd/nixcopy/main.go
# Size: ~25-30 MB
```

### Release Build (optimized)
```bash
./build-release.sh
# Size: ~15-18 MB (ลดลง ~40-50%)
```

### Release Build + UPX Compression (optional)
```bash
./build-release.sh
upx --best --lzma dist/nixcopy
# Size: ~5-7 MB (ลดลง ~70-80%)
```

## Advanced Optimization

### 1. ใช้ UPX Compression

```bash
# Install UPX
# macOS
brew install upx

# Linux
sudo apt-get install upx-ucl

# Compress binary
upx --best --lzma dist/nixcopy
```

### 2. Build สำหรับ Static Binary (Linux)

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags="-s -w -extldflags '-static'" \
  -o dist/nixcopy-linux-amd64-static \
  cmd/nixcopy/main.go
```

### 3. Build พร้อม Compiler Optimizations

```bash
go build \
  -trimpath \
  -ldflags="-s -w" \
  -gcflags="all=-l -B" \
  -o dist/nixcopy \
  cmd/nixcopy/main.go
```

## Verification

### ตรวจสอบ Binary Size

```bash
ls -lh dist/
du -h dist/nixcopy
```

### ตรวจสอบ Debug Symbols

```bash
# macOS
nm dist/nixcopy | wc -l
# ถ้า = 0 แสดงว่าไม่มี symbols

# Linux
file dist/nixcopy
readelf -S dist/nixcopy | grep debug
# ถ้าไม่มี output แสดงว่าไม่มี debug info
```

### ตรวจสอบ Dependencies

```bash
# macOS
otool -L dist/nixcopy

# Linux
ldd dist/nixcopy
```

### ทดสอบ Binary

```bash
dist/nixcopy --version
dist/nixcopy --help
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Release Build

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build Release
        run: |
          VERSION=${GITHUB_REF#refs/tags/v} ./build-release.sh all
      
      - name: Upload Artifacts
        uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: dist/*
```

## Troubleshooting

### ปัญหา: Binary ใหญ่เกินไป

**แก้ไข:**
1. ตรวจสอบว่าใช้ `-ldflags="-s -w"` แล้ว
2. ใช้ `-trimpath` flag
3. ลองใช้ UPX compression
4. ตรวจสอบ dependencies ที่ไม่จำเป็น

### ปัญหา: Build ล้มเหลวบน Cross-compilation

**แก้ไข:**
1. ตรวจสอบว่า CGO_ENABLED=0 (ถ้าไม่ใช้ CGO)
2. ตรวจสอบ dependencies ที่ต้องการ CGO
3. ใช้ Docker สำหรับ cross-compilation

### ปัญหา: Binary ไม่รันบน Target Platform

**แก้ไข:**
1. ตรวจสอบ GOOS และ GOARCH ให้ถูกต้อง
2. สำหรับ Linux: พิจารณาใช้ static build
3. ตรวจสอบ dependencies ที่ต้องการ

## Best Practices

1. **Always use release builds for production**
   - ขนาดเล็กกว่า
   - เร็วกว่า
   - ปลอดภัยกว่า (ไม่มี debug info)

2. **Tag releases with version numbers**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   VERSION=1.0.0 ./build-release.sh all
   ```

3. **Test builds before releasing**
   ```bash
   ./build-release.sh current
   ./dist/nixcopy --version
   ./dist/nixcopy transfer --help
   ```

4. **Keep build artifacts organized**
   ```bash
   dist/
   ├── nixcopy-linux-amd64
   ├── nixcopy-darwin-arm64
   └── nixcopy-windows-amd64.exe
   ```

5. **Document build requirements**
   - Go version
   - Required tools (UPX, etc.)
   - Platform-specific notes
