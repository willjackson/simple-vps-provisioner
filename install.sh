#!/bin/bash
# Installation script for Simple VPS Provisioner (svp)

set -e

echo "==================================="
echo "Simple VPS Provisioner - Installation"
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

# Build the binary
go build -o svp

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
echo "  svp -mode setup -cms drupal -domain example.com"
echo ""
echo "  # Install WordPress:"
echo "  svp -mode setup -cms wordpress -domain example.com"
echo ""
echo "  # Verify configuration:"
echo "  svp -mode verify"
echo ""
echo "For more information, see: README.md"
echo ""
