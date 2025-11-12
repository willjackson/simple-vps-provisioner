#!/bin/bash
# build-release.sh - Build binaries for all platforms

set -e

VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="dist"

echo "==================================="
echo "Simple VPS Provisioner - Build Release"
echo "Version: ${VERSION}"
echo "==================================="
echo ""

# Verify Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go 1.21+ from https://go.dev/dl/"
    exit 1
fi

echo "Go version: $(go version)"
echo ""

# Ensure module mode is enabled
export GO111MODULE=on

# Tidy and verify module
echo "Tidying Go module..."
go mod tidy
go mod verify

# Clean previous builds
if [ -d "${BUILD_DIR}" ]; then
    echo "Cleaning previous builds..."
    rm -rf ${BUILD_DIR}
fi
mkdir -p ${BUILD_DIR}

# Build for Linux AMD64 (primary target - Debian VPS)
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w -X main.version=${VERSION}" -o ${BUILD_DIR}/svp-linux-amd64

# Build for Linux ARM64 (for ARM-based VPS)
echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -v -ldflags="-s -w -X main.version=${VERSION}" -o ${BUILD_DIR}/svp-linux-arm64

# Build for macOS AMD64 (for development/testing on Intel Macs)
echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -v -ldflags="-s -w -X main.version=${VERSION}" -o ${BUILD_DIR}/svp-darwin-amd64

# Build for macOS ARM64 (for development/testing on Apple Silicon)
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -v -ldflags="-s -w -X main.version=${VERSION}" -o ${BUILD_DIR}/svp-darwin-arm64

# Make all binaries executable
chmod +x ${BUILD_DIR}/*

# Generate checksums
echo ""
echo "Generating checksums..."
cd ${BUILD_DIR}
sha256sum svp-* > checksums.txt
cd ..

echo ""
echo "==================================="
echo "Build Complete!"
echo "==================================="
echo ""
echo "Binaries built:"
ls -lh ${BUILD_DIR}/
echo ""
echo "Checksums:"
cat ${BUILD_DIR}/checksums.txt
echo ""
echo "To create a GitHub release:"
echo "  git tag -a v${VERSION} -m 'Release v${VERSION}'"
echo "  git push origin v${VERSION}"
echo "  gh release create v${VERSION} --title 'Simple VPS Provisioner v${VERSION}' --notes 'Release notes here' dist/*"
echo ""
