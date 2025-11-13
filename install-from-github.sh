#!/bin/bash
# install-from-github.sh - Install Simple VPS Provisioner from GitHub releases
# This script downloads and installs the latest pre-built binary

set -e

REPO_OWNER="willjackson"
REPO_NAME="simple-vps-provisioner"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="svp"

echo "==================================="
echo "Simple VPS Provisioner - Quick Install"
echo "==================================="
echo ""

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: sudo bash install-from-github.sh"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH"
        echo "Supported: x86_64 (amd64), aarch64 (arm64)"
        exit 1
        ;;
esac

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$OS" != "linux" ]; then
    echo "Error: This installer is for Linux only"
    echo "Detected OS: $OS"
    echo "For other platforms, download from: https://github.com/${REPO_OWNER}/${REPO_NAME}/releases"
    exit 1
fi

# Note: We'll determine the actual binary filename after getting the version

echo "System Information:"
echo "  OS: $OS"
echo "  Architecture: $ARCH"
echo ""

# Check for required commands
if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
    echo "Error: Neither curl nor wget is installed"
    echo "Please install one of them:"
    echo "  apt-get install curl"
    echo "  or"
    echo "  apt-get install wget"
    exit 1
fi

# Function to download file
download_file() {
    local url="$1"
    local output="$2"
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$url" -o "$output"
    else
        wget -q "$url" -O "$output"
    fi
}

# Get latest release version from GitHub API
echo "Fetching latest release information..."
RELEASE_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"

if command -v curl &> /dev/null; then
    RELEASE_JSON=$(curl -fsSL "$RELEASE_URL")
else
    RELEASE_JSON=$(wget -qO- "$RELEASE_URL")
fi

# Extract version tag
VERSION=$(echo "$RELEASE_JSON" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: Could not find version tag"
    echo "Please check:"
    echo "  1. Repository exists: https://github.com/${REPO_OWNER}/${REPO_NAME}"
    echo "  2. Releases are available"
    exit 1
fi

# Remove 'v' prefix from version
VERSION_NUMBER="${VERSION#v}"

# GoReleaser naming convention: svp-OS-ARCH (without version)
BINARY_FILE="${BINARY_NAME}-${OS}-${ARCH}"

# Extract download URL
DOWNLOAD_URL=$(echo "$RELEASE_JSON" | grep "browser_download_url.*${BINARY_FILE}\"" | sed -E 's/.*"browser_download_url": *"([^"]+)".*/\1/')

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find binary in release"
    echo "Looking for: ${BINARY_FILE}"
    echo "Available assets:"
    echo "$RELEASE_JSON" | grep "browser_download_url" | sed -E 's/.*"browser_download_url": *"([^"]+)".*/  - \1/' | sed 's|.*/||'
    echo ""
    echo "Please check:"
    echo "  1. Repository exists: https://github.com/${REPO_OWNER}/${REPO_NAME}"
    echo "  2. Binary ${BINARY_FILE} exists in latest release"
    exit 1
fi

echo "Latest version: $VERSION"
echo "Binary: $BINARY_FILE"
echo "Download URL: $DOWNLOAD_URL"
echo ""

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

cd "$TMP_DIR"

# Download binary
echo "Downloading ${BINARY_NAME} ${VERSION}..."
download_file "$DOWNLOAD_URL" "$BINARY_FILE"

if [ ! -f "$BINARY_FILE" ]; then
    echo "Error: Download failed"
    exit 1
fi

# Download checksums (goreleaser uses checksums.txt)
CHECKSUM_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/checksums.txt"
echo "Downloading checksums..."
download_file "$CHECKSUM_URL" "checksums.txt" 2>/dev/null || true

# Verify checksum if available
if [ -f "checksums.txt" ] && command -v sha256sum &> /dev/null; then
    echo "Verifying checksum..."
    if grep "$BINARY_FILE" checksums.txt > /dev/null 2>&1; then
        if sha256sum --check --ignore-missing checksums.txt > /dev/null 2>&1; then
            echo "✓ Checksum verified"
        else
            echo "Warning: Checksum verification failed"
            echo "Continue anyway? [y/N]"
            read -r response
            if [ "$response" != "y" ] && [ "$response" != "Y" ]; then
                echo "Installation cancelled"
                exit 1
            fi
        fi
    else
        echo "Warning: Checksum not found for ${BINARY_FILE}"
    fi
else
    echo "Skipping checksum verification (checksums.txt not available or sha256sum not installed)"
fi

# Make binary executable
chmod +x "$BINARY_FILE"

# Test the binary
echo ""
echo "Testing binary..."
if ./"$BINARY_FILE" -version > /dev/null 2>&1; then
    VERSION_OUTPUT=$(./"$BINARY_FILE" -version)
    echo "✓ Binary is working"
    echo "  $VERSION_OUTPUT"
else
    echo "Error: Binary test failed"
    exit 1
fi

# Backup existing installation if present
if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    echo ""
    echo "Backing up existing installation..."
    cp "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}.backup"
    echo "✓ Backup saved to ${INSTALL_DIR}/${BINARY_NAME}.backup"
fi

# Install binary
echo ""
echo "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
mv "$BINARY_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Verify installation
if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    echo "✓ Installation successful"
else
    echo "Error: Installation failed"
    exit 1
fi

# Display final version
FINAL_VERSION=$(${INSTALL_DIR}/${BINARY_NAME} -version 2>&1)

echo ""
echo "==================================="
echo "Installation Complete!"
echo "==================================="
echo ""
echo "$FINAL_VERSION"
echo ""
echo "Usage examples:"
echo "  # Check version:"
echo "  svp -version"
echo ""
echo "  # Install Drupal:"
echo "  sudo svp -mode setup -cms drupal -domain example.com"
echo ""
echo "  # Install WordPress:"
echo "  sudo svp -mode setup -cms wordpress -domain example.com"
echo ""
echo "  # Verify configuration:"
echo "  sudo svp -mode verify"
echo ""
echo "  # Update to latest version:"
echo "  sudo svp -mode update"
echo ""
echo "For more information, see: https://github.com/${REPO_OWNER}/${REPO_NAME}"
echo ""
