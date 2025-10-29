#!/bin/bash

# Build script for ARM-7 Linux binary
# This script cross-compiles Peerless for ARM-7 architecture on Linux

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARY_NAME="peerless"
OUTPUT_DIR="$PROJECT_ROOT/dist/test-arm7"
TARGET_GOOS="linux"
TARGET_GOARCH="arm"
TARGET_GOARM="7"

echo -e "${GREEN}Building ARM-7 Linux binary for Peerless${NC}"
echo "Project root: $PROJECT_ROOT"
echo "Output directory: $OUTPUT_DIR"
echo "Target: $TARGET_GOOS/$TARGET_GOARCH (ARM v7)"
echo

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if we're in the right directory
if [ ! -f "$PROJECT_ROOT/main.go" ]; then
    echo -e "${RED}Error: main.go not found in project root${NC}"
    echo "Please run this script from the project root or ensure main.go exists"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version)
echo -e "${YELLOW}Go version:${NC} $GO_VERSION"

# Set environment variables for cross-compilation
export GOOS=$TARGET_GOOS
export GOARCH=$TARGET_GOARCH
export GOARM=$TARGET_GOARM
export CGO_ENABLED=0

# Build information
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

echo
echo -e "${YELLOW}Build information:${NC}"
echo "Build time: $BUILD_TIME"
echo "Git commit: $GIT_COMMIT"
echo "Git tag: $GIT_TAG"
echo

# Build flags
LDFLAGS="-s -w -X main.version=$GIT_TAG -X main.commit=$GIT_COMMIT -X main.buildTime=$BUILD_TIME"

echo -e "${YELLOW}Building binary...${NC}"

# Build the binary
cd "$PROJECT_ROOT"
go build \
    -ldflags "$LDFLAGS" \
    -o "$OUTPUT_DIR/$BINARY_NAME-linux-armv7" \
    main.go

# Check if build was successful
if [ $? -eq 0 ]; then
    echo
    echo -e "${GREEN}✅ Build successful!${NC}"

    # Get binary info
    BINARY_PATH="$OUTPUT_DIR/$BINARY_NAME-linux-armv7"
    BINARY_SIZE=$(ls -lh "$BINARY_PATH" | awk '{print $5}')

    echo -e "${YELLOW}Binary details:${NC}"
    echo "Path: $BINARY_PATH"
    echo "Size: $BINARY_SIZE"
    echo "Target: $TARGET_GOOS/$TARGET_GOARCH v$TARGET_GOARM"

    # Show file type information
    echo
    echo -e "${YELLOW}File information:${NC}"
    file "$BINARY_PATH"

    echo
    echo -e "${GREEN}🎉 ARM-7 Linux binary is ready!${NC}"
    echo "You can copy it to your ARM-7 Linux device (e.g., Raspberry Pi, Synology NAS)."
else
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi