# Version Management

This document explains how versioning works in Simple VPS Provisioner (svp), including how versions are set, displayed, and managed through the release process.

## Table of Contents

- [Semantic Versioning](#semantic-versioning)
- [Version Injection](#version-injection)
- [Checking Your Version](#checking-your-version)
- [Version in Different Builds](#version-in-different-builds)
- [Release Process](#release-process)
- [Updating SVP](#updating-svp)
- [For Developers](#for-developers)

## Semantic Versioning

SVP follows [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

- **MAJOR**: Incompatible API/behavior changes
- **MINOR**: New functionality in a backwards-compatible manner
- **PATCH**: Backwards-compatible bug fixes

### Examples

- `1.0.0` - First stable release
- `1.0.1` - Bug fix release
- `1.1.0` - New feature (e.g., `-keep-existing-db` flag)
- `2.0.0` - Breaking change (e.g., changed command structure)

## Version Injection

SVP uses **build-time version injection** to ensure consistency between the installed binary and GitHub releases.

### How It Works

The version is injected during compilation using Go's `-ldflags` flag:

```bash
go build -ldflags="-X main.version=1.0.24" -o svp
```

This sets the `version` variable in `main.go`:

```go
// main.go
var version = "dev"  // Default if not set during build
```

### Why Build-Time Injection?

1. **Single source of truth**: Git tags determine the version
2. **No manual updates**: No need to edit code for each release
3. **Consistency**: Binary always knows its exact version
4. **Traceability**: Version matches the Git tag

## Checking Your Version

### Display Current Version

```bash
svp -version
```

Output examples:

```bash
# Production build
Simple VPS Provisioner (svp) version 1.0.24

# Development build (no version injected)
Simple VPS Provisioner (svp) version dev
This is a development build.
Production builds should use: go build -ldflags="-X main.version=VERSION"
```

### Version in Help

The version also appears in the help text:

```bash
svp --help
# Output: Simple VPS Provisioner (svp) v1.0.24
```

## Version in Different Builds

### Production Builds (GitHub Releases)

Built by GitHub Actions with proper version injection:

```bash
# Automatically built for each tag
git tag v1.0.24
git push origin v1.0.24

# GitHub Actions runs:
go build -ldflags="-s -w -X main.version=1.0.24" -o svp
```

Binaries are named: `svp_1.0.24_linux_amd64`

### Development Builds

#### Using install.sh

Automatically detects version from git:

```bash
sudo bash install.sh

# Runs internally:
VERSION=$(git describe --tags --always --dirty)
VERSION=${VERSION#v}  # Remove 'v' prefix
go build -ldflags="-X main.version=${VERSION}" -o svp
```

Output examples:
- `1.0.24` - Built from tag v1.0.24
- `1.0.24-1-g5f3c2a1` - 1 commit after v1.0.24
- `1.0.24-1-g5f3c2a1-dirty` - Uncommitted changes present

#### Manual Build

Without version injection:

```bash
go build -o svp
# version = "dev"
```

With version injection:

```bash
VERSION="1.0.24"
go build -ldflags="-X main.version=${VERSION}" -o svp
```

#### Using build-release.sh

For local multi-platform builds:

```bash
VERSION=1.0.24 ./build-release.sh
```

This creates:
- `dist/svp-linux-amd64`
- `dist/svp-linux-arm64`
- `dist/svp-darwin-amd64`
- `dist/svp-darwin-arm64`
- `dist/checksums.txt`

## Release Process

### 1. Prepare Release

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Update CHANGELOG.md
# Move items from [Unreleased] to [1.0.24]
vim CHANGELOG.md

# Commit changelog
git add CHANGELOG.md
git commit -m "Prepare release v1.0.24"
git push origin main
```

### 2. Tag Release

```bash
# Create annotated tag
git tag -a v1.0.24 -m "Release v1.0.24"

# Push tag to GitHub
git push origin v1.0.24
```

### 3. Automated Build

GitHub Actions automatically:
1. Detects the new tag
2. Builds binaries for all platforms with version injection
3. Generates checksums
4. Creates GitHub release
5. Uploads all artifacts

### 4. Verify Release

```bash
# Check the release on GitHub
# https://github.com/willjackson/simple-vps-provisioner/releases/tag/v1.0.24

# Test installation
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash

# Check installed version
svp -version
# Should show: Simple VPS Provisioner (svp) version 1.0.24
```

## Updating SVP

### Automatic Update

SVP can update itself:

```bash
sudo svp -mode update
```

This will:
1. Check GitHub for the latest release
2. Compare with current version
3. Download new binary if available
4. Verify checksum
5. Replace the current binary
6. Create backup of old version

### Manual Update

```bash
# Using quick install script
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash

# Or manually download specific version
VERSION="1.0.24"
wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp_${VERSION}_linux_amd64
chmod +x svp_${VERSION}_linux_amd64
sudo mv svp_${VERSION}_linux_amd64 /usr/local/bin/svp
```

## For Developers

### Version Workflow

1. **During Development**
   - Version shows as `dev` or git describe output
   - No version injection needed for testing

2. **Before Release**
   - Update CHANGELOG.md
   - Commit changes
   - Tag release
   - Push tag

3. **After Tag Push**
   - GitHub Actions handles everything automatically
   - Binaries are built with proper version
   - Release is created with all artifacts

### Testing Version Display

```bash
# Build without version
go build -o svp
./svp -version
# Output: Simple VPS Provisioner (svp) version dev

# Build with version
go build -ldflags="-X main.version=1.0.24" -o svp
./svp -version
# Output: Simple VPS Provisioner (svp) version 1.0.24

# Build with git describe
VERSION=$(git describe --tags --always)
VERSION=${VERSION#v}
go build -ldflags="-X main.version=${VERSION}" -o svp
./svp -version
# Output: Simple VPS Provisioner (svp) version 1.0.24-1-g5f3c2a1
```

### Version Variable Location

```go
// main.go
package main

import (
    "flag"
    "fmt"
    "os"
    "svp/cmd"
    "svp/pkg/utils"
    "svp/types"
)

// version is set at build time via -ldflags="-X main.version=VERSION"
// Default to dev if not set during build
var version = "dev"

func main() {
    // ... rest of code
    
    // Version flag handler
    var showVersion bool
    flag.BoolVar(&showVersion, "version", false, "Show version information")
    
    if showVersion {
        fmt.Printf("Simple VPS Provisioner (svp) version %s\n", version)
        if version == "dev" {
            fmt.Println("This is a development build.")
            fmt.Println("Production builds should use: go build -ldflags=\"-X main.version=VERSION\"")
        }
        os.Exit(0)
    }
}
```

### Update Mechanism

The updater compares versions and downloads from GitHub:

```go
// pkg/updater/updater.go

const (
    GitHubRepo = "willjackson/simple-vps-provisioner"
    APIBaseURL = "https://api.github.com/repos"
)

func CheckForUpdates(currentVersion string) (string, bool, error) {
    // Fetch latest release from GitHub API
    apiURL := fmt.Sprintf("%s/%s/releases/latest", APIBaseURL, GitHubRepo)
    
    // Compare versions
    latestVersion := strings.TrimPrefix(release.TagName, "v")
    currentVersion = strings.TrimPrefix(currentVersion, "v")
    
    if latestVersion == currentVersion {
        // Already up to date
        return latestVersion, false, nil
    }
    
    return latestVersion, true, nil
}
```

### Version Naming Conventions

#### Git Tags
- Format: `vMAJOR.MINOR.PATCH`
- Examples: `v1.0.0`, `v1.0.24`, `v2.0.0`
- Always start with lowercase 'v'

#### Binary Names (GitHub Releases)
- Format: `svp_VERSION_OS_ARCH`
- Examples:
  - `svp_1.0.24_linux_amd64`
  - `svp_1.0.24_linux_arm64`
  - `svp_1.0.24_darwin_amd64`
  - `svp_1.0.24_darwin_arm64`

#### In Code
- Variable: `version` (no 'v' prefix)
- Display: "version 1.0.24" (no 'v' prefix)
- Tag comparison: Strip 'v' before comparing

### Build Scripts Reference

#### install.sh
Builds from source with automatic version detection:
```bash
# Auto-detect version from git tags
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
VERSION=${VERSION#v}

# Build with version injection
go build -ldflags="-X main.version=${VERSION}" -o svp

# Install
sudo mv svp /usr/local/bin/
```

#### build-release.sh
Builds for all platforms:
```bash
VERSION=${VERSION:-"1.0.0"}

# Build for each platform
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o dist/svp-linux-amd64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=${VERSION}" -o dist/svp-linux-arm64
# etc...

# Generate checksums
sha256sum dist/svp-* > dist/checksums.txt
```

#### .github/workflows/release.yml
Automated release builds:
```yaml
- name: Build binaries
  run: |
    VERSION=${GITHUB_REF#refs/tags/v}
    
    # Build all platforms
    GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o svp_${VERSION}_linux_amd64
    # etc...
```

## Version History

The complete version history is maintained in [CHANGELOG.md](CHANGELOG.md).

### Key Releases

- **v1.0.0** - First stable release
- **v1.0.24** - Added `-keep-existing-db` flag for reprovisioning

## Troubleshooting

### Version Shows "dev"

**Cause**: Binary was built without version injection.

**Fix**: Build with version:
```bash
go build -ldflags="-X main.version=1.0.24" -o svp
```

Or use install.sh which auto-detects version:
```bash
sudo bash install.sh
```

### Version Shows Wrong Value

**Cause**: Old binary in PATH or wrong build.

**Check**:
```bash
which svp
# Should be: /usr/local/bin/svp

svp -version
```

**Fix**: Reinstall or rebuild:
```bash
sudo bash install-from-github.sh
```

### Update Fails

**Causes**:
- Network issue
- GitHub API rate limit
- No newer version available
- Permission denied

**Debug**:
```bash
# Check current version
svp -version

# Check GitHub for latest release
curl -s https://api.github.com/repos/willjackson/simple-vps-provisioner/releases/latest | grep tag_name

# Manual update
sudo bash install-from-github.sh
```

## Best Practices

1. **Always use version injection** for production builds
2. **Tag releases** using semantic versioning
3. **Update CHANGELOG.md** before tagging
4. **Test locally** before pushing tags
5. **Use annotated tags** with meaningful messages
6. **Never edit version in code** - use build-time injection
7. **Check version** after installation to verify

## Summary

- SVP uses **semantic versioning** (MAJOR.MINOR.PATCH)
- Version is **injected at build time** using `-ldflags`
- Production binaries from GitHub have **proper versions**
- Development builds can show **"dev"** or git describe output
- Check version with **`svp -version`**
- Update with **`sudo svp -mode update`**
- Releases are **automated via GitHub Actions**

For more details on releases and deployment, see [DEPLOY.md](DEPLOY.md).
