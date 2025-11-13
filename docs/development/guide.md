---
layout: default
title: Complete Development Guide
---

# Complete Development Guide

Comprehensive guide for developing and contributing to Simple VPS Provisioner.

[View Quick Development Overview →]({{ site.baseurl }}/development/)

## Table of Contents

- [Development Environment](#development-environment)
- [Code Architecture](#code-architecture)
- [Development Workflows](#development-workflows)
- [Testing Guide](#testing-guide)
- [Contributing Process](#contributing-process)
- [Release Process](#release-process)

---

## Development Environment

### Prerequisites

**Required:**
- Go 1.21 or higher
- Git
- Linux/macOS or WSL on Windows

**Recommended:**
- Debian 12 or Ubuntu 22.04 VM for testing
- VS Code with Go extension
- Docker for testing in clean environments

### Initial Setup

```bash
# Clone repository
git clone https://github.com/willjackson/simple-vps-provisioner.git
cd simple-vps-provisioner

# Initialize modules
go mod tidy

# Build
go build -o svp

# Run
./svp -version
```

---

## Code Architecture

### Package Organization

```
pkg/
├── cms/          # CMS installation (Drupal, WordPress)
├── config/       # Configuration files and management
├── database/     # MariaDB setup and operations
├── ssl/          # Let's Encrypt and SSL certificates
├── system/       # System packages, firewall, services
├── updater/      # Self-update functionality
├── utils/        # Shared utilities (logging, exec, checks)
└── web/          # Nginx and PHP-FPM configuration
```

### Design Principles

1. **Separation of Concerns** - Each package handles one domain
2. **No Circular Dependencies** - Clear dependency hierarchy
3. **Idempotency** - Safe to run multiple times
4. **Verify-Only Pattern** - Check before modify

### Example Function Pattern

```go
// InstallSomething installs and configures something
func InstallSomething(verifyOnly bool) error {
    // Check if already done
    if isAlreadyDone() {
        utils.Verify("Something already configured")
        return nil
    }
    
    // If verify-only mode, report failure
    if verifyOnly {
        utils.Fail("Something not configured")
        return fmt.Errorf("something not configured")
    }
    
    // Actually do the work
    utils.Log("Installing something...")
    if err := doWork(); err != nil {
        return fmt.Errorf("failed to install: %v", err)
    }
    
    utils.Ok("Something installed successfully")
    return nil
}
```

---

## Development Workflows

### Adding a New Feature

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Identify Package**
   - System operations → `pkg/system/`
   - Web server → `pkg/web/`
   - Database → `pkg/database/`
   - CMS-specific → `pkg/cms/`

3. **Implement Feature**
   ```go
   // pkg/system/myfeature.go
   package system
   
   import "svp/pkg/utils"
   
   func ConfigureMyFeature(verifyOnly bool) error {
       // Implementation
   }
   ```

4. **Add CLI Flag (if needed)**
   ```go
   // types/types.go
   type Config struct {
       MyFeature bool
   }
   
   // main.go
   flag.BoolVar(&cfg.MyFeature, "my-feature", false, "Enable my feature")
   ```

5. **Update Command Handler**
   ```go
   // cmd/setup.go
   if cfg.MyFeature {
       if err := system.ConfigureMyFeature(cfg.VerifyOnly); err != nil {
           return err
       }
   }
   ```

6. **Test Locally**
   ```bash
   go build -o svp
   sudo ./svp -mode setup -my-feature
   ```

7. **Document**
   - Update README.md
   - Update docs/
   - Update CHANGELOG.md

8. **Commit and Push**
   ```bash
   git add .
   git commit -m "Add my feature"
   git push origin feature/my-feature
   ```

---

## Testing Guide

### Manual Testing

1. **Spin up test VM**
   ```bash
   # Debian 12 or Ubuntu 22.04
   multipass launch --name svp-test 22.04
   multipass shell svp-test
   ```

2. **Build and copy**
   ```bash
   go build -o svp
   multipass transfer svp svp-test:/home/ubuntu/
   ```

3. **Test on VM**
   ```bash
   multipass shell svp-test
   sudo ./svp -mode setup -cms drupal -domain test.local -ssl=false
   ```

4. **Verify**
   ```bash
   sudo ./svp -mode verify
   curl -I http://test.local
   ```

### Testing Checklist

- [ ] Fresh Debian 12 installation
- [ ] Fresh Ubuntu 22.04 installation
- [ ] Drupal installation
- [ ] WordPress installation
- [ ] Multi-domain setup
- [ ] Git deployment
- [ ] Database import
- [ ] SSL certificate
- [ ] PHP version update
- [ ] Verify mode
- [ ] Update mode

---

## Contributing Process

### 1. Fork and Clone

```bash
# Fork on GitHub, then:
git clone https://github.com/YOUR_USERNAME/simple-vps-provisioner.git
cd simple-vps-provisioner

# Add upstream
git remote add upstream https://github.com/willjackson/simple-vps-provisioner.git
```

### 2. Create Feature Branch

```bash
git checkout -b feature/my-feature
```

### 3. Make Changes

- Write clear, concise code
- Follow existing patterns
- Add comments for exported functions
- Handle errors properly

### 4. Test Thoroughly

- Build without errors
- Test on real Debian/Ubuntu
- Check all modes work
- Verify doesn't break existing functionality

### 5. Commit

```bash
git add .
git commit -m "Add feature: description"
```

**Good commit messages:**
- "Add support for PostgreSQL database"
- "Fix SSL certificate renewal issue"
- "Update PHP version detection logic"

**Bad commit messages:**
- "fix"
- "update"
- "changes"

### 6. Push and Create PR

```bash
git push origin feature/my-feature
```

Then create Pull Request on GitHub with:
- Clear title
- Description of changes
- Why the change is needed
- How it was tested

### 7. Address Feedback

```bash
# Make changes
git add .
git commit -m "Address review feedback"
git push origin feature/my-feature
```

---

## Release Process

### Semantic Versioning

Format: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Release Steps

#### 1. Prepare

```bash
# Ensure on main branch
git checkout main
git pull origin main

# Update CHANGELOG.md
vim CHANGELOG.md
# Move [Unreleased] items to [1.0.31]

# Commit
git add CHANGELOG.md
git commit -m "Prepare release v1.0.31"
git push origin main
```

#### 2. Tag

```bash
# Create annotated tag
git tag -a v1.0.31 -m "Release v1.0.31"

# Push tag
git push origin v1.0.31
```

#### 3. Automated Build

GitHub Actions will automatically:
1. Build for all platforms
2. Generate checksums
3. Create GitHub release
4. Upload binaries

#### 4. Verify

```bash
# Check release page
https://github.com/willjackson/simple-vps-provisioner/releases

# Test installation
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash

# Check version
svp -version
```

---

## Code Guidelines

### File Organization

```go
// Package declaration
package system

// Imports (stdlib first, then external, then local)
import (
    "fmt"
    "strings"
    
    "svp/pkg/utils"
)

// Constants
const (
    DefaultValue = "default"
)

// Type definitions
type Config struct {
    Value string
}

// Exported functions (documented)
// PublicFunction does something useful
func PublicFunction() error {
    // Implementation
}

// Unexported functions
func privateHelper() error {
    // Implementation
}
```

### Error Handling

```go
// Good ✅
if err := DoSomething(); err != nil {
    return fmt.Errorf("failed to do something: %v", err)
}

// Bad ❌
err := DoSomething()
if err != nil {
    return err  // No context
}

// Bad ❌
_ = DoSomething()  // Ignored error
```

### Logging

```go
// Use utils logger
utils.Log("Starting operation...")
utils.Verify("Checking status")
utils.Ok("Operation complete")

// Don't print directly
fmt.Println("Message")  // ❌
```

---

## Debugging

### Enable Debug Mode

```bash
# Environment variable
export DEBUG=1
./svp -mode setup

# Or flag
./svp -debug -mode setup
```

### Common Issues

**Build fails:**
```bash
export GO111MODULE=on
go mod tidy
go build
```

**Import errors:**
```bash
# Use full import paths
import "svp/pkg/utils"  // ✓
import "./pkg/utils"     // ✗
```

**Module issues:**
```bash
go clean -modcache
go mod download
go mod tidy
```

---

## Resources

### Documentation
- [README.md](https://github.com/willjackson/simple-vps-provisioner/blob/main/README.md)
- [CHANGELOG.md](https://github.com/willjackson/simple-vps-provisioner/blob/main/CHANGELOG.md)
- [VERSIONING.md](https://github.com/willjackson/simple-vps-provisioner/blob/main/VERSIONING.md)

### Go Resources
- [Official Go Docs](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### External Tools
- [Nginx Docs](https://nginx.org/en/docs/)
- [PHP-FPM](https://www.php.net/manual/en/install.fpm.php)
- [Drupal](https://www.drupal.org/docs)

---

## Getting Help

- **GitHub Issues**: [Report bugs](https://github.com/willjackson/simple-vps-provisioner/issues)
- **Discussions**: [Ask questions](https://github.com/willjackson/simple-vps-provisioner/discussions)

---

[← Back to Development Overview]({{ site.baseurl }}/development/) | [Documentation →]({{ site.baseurl }}/documentation/)
