#!/bin/bash

# Build script for go-nixcopy release version
# - Removes debug symbols
# - Optimizes binary size
# - Supports multiple platforms

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

# Go build flags
BUILDFLAGS="-trimpath"

echo "================================================"
echo "Building go-nixcopy Release Version"
echo "================================================"
echo "Version:     $VERSION"
echo "Build Time:  $BUILD_TIME"
echo "Git Commit:  $GIT_COMMIT"
echo "Output Dir:  $OUTPUT_DIR"
echo "================================================"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to build for a specific platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT_NAME="${BINARY_NAME}"
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${BINARY_NAME}.exe"
    fi
    
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
    
    if [ $? -eq 0 ]; then
        local SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
        echo "✓ Built successfully: $OUTPUT_PATH ($SIZE)"
    else
        echo "✗ Build failed for $GOOS/$GOARCH"
        return 1
    fi
}

# Build for current platform only (default)
if [ "$1" = "current" ] || [ -z "$1" ]; then
    echo ""
    echo "Building for current platform..."
    go build \
        $BUILDFLAGS \
        -ldflags="$LDFLAGS" \
        -o "$OUTPUT_DIR/$BINARY_NAME" \
        cmd/nixcopy/main.go
    
    SIZE=$(du -h "$OUTPUT_DIR/$BINARY_NAME" | cut -f1)
    echo "✓ Built successfully: $OUTPUT_DIR/$BINARY_NAME ($SIZE)"

# Build for all platforms
elif [ "$1" = "all" ]; then
    echo ""
    echo "Building for all platforms..."
    
    # Linux
    build_platform linux amd64
    build_platform linux arm64
    build_platform linux 386
    
    # macOS
    build_platform darwin amd64
    build_platform darwin arm64
    
    # Windows
    build_platform windows amd64
    build_platform windows 386
    
    # FreeBSD
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
    build_platform $GOOS $GOARCH
fi

echo ""
echo "================================================"
echo "Build completed successfully!"
echo "================================================"
