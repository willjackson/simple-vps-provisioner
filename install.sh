#!/bin/bash
# install.sh - Build and install Simple VPS Provisioner from source
# For installing pre-built binaries, use install-from-github.sh instead

set -e

echo "==================================="
echo "Simple VPS Provisioner - Build from Source"
echo "==================================="
echo ""

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing Go 1.21..."
    cd /tmp
    wget -q https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    echo "Go installed successfully"
else
    echo "Go is already installed: $(go version)"
fi

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "Building Simple VPS Provisioner (svp)..."
cd "$SCRIPT_DIR"

# Initialize Go module if needed
if [ ! -f "go.sum" ]; then
    go mod tidy
fi

# Determine version from git tags, or use dev
if command -v git &> /dev/null && [ -d ".git" ]; then
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    # Remove 'v' prefix if present
    VERSION=${VERSION#v}
    echo "Building version: $VERSION"
else
    VERSION="dev"
    echo "Building development version (no git repository found)"
fi

# Build the binary with version injection
go build -ldflags="-X main.version=${VERSION}" -o svp

# Install to /usr/local/bin
echo "Installing to /usr/local/bin..."
mv svp /usr/local/bin/
chmod +x /usr/local/bin/svp

echo ""
echo "==================================="
echo "Installation Complete!"
echo "==================================="
echo ""
echo "Usage examples:"
echo "  # Install Drupal:"
echo "  svp setup -cms drupal -domain example.com"
echo ""
echo "  # Install WordPress:"
echo "  svp setup -cms wordpress -domain example.com"
echo ""
echo "  # Verify configuration:"
echo "  svp verify"
echo ""
echo "For more information, see: README.md"
echo ""
