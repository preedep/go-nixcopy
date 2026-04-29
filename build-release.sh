#!/bin/bash

# Build script for go-nixcopy release version
# - Removes debug symbols
# - Optimizes binary size
# - Supports multiple platforms
# - Optional UPX compression: set UPX=1 to enable (requires upx installed)

set -e

VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

BINARY_NAME="nixcopy"
OUTPUT_DIR="./dist"

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.Version=$VERSION"
LDFLAGS="$LDFLAGS -X main.BuildTime=$BUILD_TIME"
LDFLAGS="$LDFLAGS -X main.GitCommit=$GIT_COMMIT"

# Go build flags — trimpath removes embedded source paths; CGO_ENABLED=0 ensures
# a fully static binary with no libc dependency (required for scratch containers).
BUILDFLAGS="-trimpath"
export CGO_ENABLED=0

echo "================================================"
echo "Building go-nixcopy Release Version"
echo "================================================"
echo "Version:     $VERSION"
echo "Build Time:  $BUILD_TIME"
echo "Git Commit:  $GIT_COMMIT"
echo "Output Dir:  $OUTPUT_DIR"
echo "UPX:         ${UPX:-0}"
echo "================================================"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# compress_binary <path> <goos> <goarch>
# Runs UPX on the binary when UPX=1 and the target is supported.
# UPX does not support darwin/arm64 (Apple Silicon).
compress_binary() {
    local PATH_BIN=$1
    local GOOS=$2
    local GOARCH=$3

    if [ "${UPX:-0}" != "1" ]; then
        return 0
    fi

    if [ "$GOOS" = "darwin" ] && [ "$GOARCH" = "arm64" ]; then
        echo "  ↷ UPX skipped (not supported on darwin/arm64)"
        return 0
    fi

    if ! command -v upx >/dev/null 2>&1; then
        echo "  ✗ UPX=1 set but upx not found in PATH — skipping compression"
        return 0
    fi

    local BEFORE
    BEFORE=$(du -h "$PATH_BIN" | cut -f1)
    upx --best --quiet "$PATH_BIN"
    local AFTER
    AFTER=$(du -h "$PATH_BIN" | cut -f1)
    echo "  ↓ UPX: $BEFORE → $AFTER"
}

# Function to build for a specific platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2

    local OUTPUT_PATH="${OUTPUT_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_PATH="${OUTPUT_PATH}.exe"
    fi

    echo ""
    echo "Building for $GOOS/$GOARCH..."

    GOOS=$GOOS GOARCH=$GOARCH go build \
        $BUILDFLAGS \
        -ldflags="$LDFLAGS" \
        -o "$OUTPUT_PATH" \
        cmd/nixcopy/main.go

    local SIZE
    SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
    echo "✓ Built: $OUTPUT_PATH ($SIZE)"
    compress_binary "$OUTPUT_PATH" "$GOOS" "$GOARCH"
}

# Build for current platform only (default)
if [ "$1" = "current" ] || [ -z "$1" ]; then
    CURRENT_GOOS=$(go env GOOS)
    CURRENT_GOARCH=$(go env GOARCH)
    echo ""
    echo "Building for current platform ($CURRENT_GOOS/$CURRENT_GOARCH)..."
    go build \
        $BUILDFLAGS \
        -ldflags="$LDFLAGS" \
        -o "$OUTPUT_DIR/$BINARY_NAME" \
        cmd/nixcopy/main.go

    SIZE=$(du -h "$OUTPUT_DIR/$BINARY_NAME" | cut -f1)
    echo "✓ Built: $OUTPUT_DIR/$BINARY_NAME ($SIZE)"
    compress_binary "$OUTPUT_DIR/$BINARY_NAME" "$CURRENT_GOOS" "$CURRENT_GOARCH"

# Build for all platforms
elif [ "$1" = "all" ]; then
    echo ""
    echo "Building for all platforms..."

    build_platform linux amd64
    build_platform linux arm64
    build_platform linux 386

    build_platform darwin amd64
    build_platform darwin arm64

    build_platform windows amd64
    build_platform windows 386

    build_platform freebsd amd64

    echo ""
    echo "================================================"
    echo "All builds completed!"
    echo "================================================"
    ls -lh "$OUTPUT_DIR"

# Build for specific platform
else
    GOOS=$1
    GOARCH=${2:-amd64}
    build_platform "$GOOS" "$GOARCH"
fi

echo ""
echo "================================================"
echo "Build completed successfully!"
echo "================================================"
