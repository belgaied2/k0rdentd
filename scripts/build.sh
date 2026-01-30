#!/bin/bash

# Build script for k0rdentd
# Usage: ./scripts/build.sh [version] [arch]
#   version: Version to embed in the binary (default: dev)
#   arch: Target architecture - amd64, arm64, or armv7 (default: host arch)

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to project root
cd "${PROJECT_ROOT}"

# Parse arguments
VERSION="${1:-dev}"
TARGET_ARCH="${2:-}"

# Determine target architecture if not specified
if [ -z "$TARGET_ARCH" ]; then
    HOST_ARCH=$(uname -m)
    case "$HOST_ARCH" in
        x86_64)
            TARGET_ARCH="amd64"
            ;;
        aarch64|arm64)
            TARGET_ARCH="arm64"
            ;;
        armv7l|armhf)
            TARGET_ARCH="armv7"
            ;;
        *)
            echo "Unsupported architecture: $HOST_ARCH"
            exit 1
            ;;
    esac
fi

# Set GOARCH and GOARM based on target architecture
case "$TARGET_ARCH" in
    amd64)
        GOARCH="amd64"
        GOARM=""
        ;;
    arm64)
        GOARCH="arm64"
        GOARM=""
        ;;
    armv7)
        GOARCH="arm"
        GOARM="7"
        ;;
    *)
        echo "Invalid target architecture: $TARGET_ARCH"
        echo "Supported architectures: amd64, arm64, armv7"
        exit 1
        ;;
esac

# Set output name
OUTPUT_NAME="k0rdentd"
if [ "$VERSION" != "dev" ]; then
    OUTPUT_NAME="k0rdentd-${VERSION}-linux-${TARGET_ARCH}"
fi

echo "Building k0rdentd..."
echo "  Version: $VERSION"
echo "  OS: linux"
echo "  Arch: $GOARCH"
if [ -n "$GOARM" ]; then
    echo "  ARM: v$GOARM"
fi
echo "  Output: $OUTPUT_NAME"

# Build flags
LDFLAGS="-X github.com/belgaied2/k0rdentd/pkg/cli.Version=${VERSION}"

# Build
export GOOS=linux
export GOARCH="$GOARCH"
if [ -n "$GOARM" ]; then
    export GOARM="$GOARM"
fi

go build -ldflags "${LDFLAGS}" -o "${OUTPUT_NAME}" ./cmd/k0rdentd

echo "Build complete: ${OUTPUT_NAME}"

# Verify version if not dev build
if [ "$VERSION" != "dev" ]; then
    echo ""
    echo "Verifying version..."
    "./${OUTPUT_NAME}" version
fi
