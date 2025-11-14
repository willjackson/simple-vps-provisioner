---
layout: default
title: Development Guide
---

# Development Guide

Guide for contributing to and developing Simple VPS Provisioner.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Building from Source](#building-from-source)
- [Code Architecture](#code-architecture)
- [Contributing](#contributing)
- [Testing](#testing)
- [Release Process](#release-process)

---

## Development Setup

### Prerequisites

- **Go 1.21+** - [Install Go](https://go.dev/dl/)
- **Git** - For version control
- **Debian or Ubuntu VM** - For testing (recommended)
- **Text editor** - VS Code, GoLand, Vim, etc.

### Clone Repository

```bash
git clone https://github.com/willjackson/simple-vps-provisioner.git
cd simple-vps-provisioner
```

### Install Dependencies

```bash
# Initialize Go modules
go mod tidy

# Verify dependencies
go mod verify
```

### Build Development Version

```bash
# Simple build
go build -o svp

# Build with version
VERSION=$(git describe --tags --always)
go build -ldflags="-X main.version=${VERSION}" -o svp

# Test
./svp --version
```

---

## Project Structure

```
svp/
├── .github/workflows/     # CI/CD pipelines
│   ├── release.yml        # Automated releases
│   └── test.yml          # Build and test
│
├── cmd/                   # Command implementations
│   ├── setup.go          # Full provisioning logic
│   ├── verify.go         # Configuration verification
│   ├── update.go         # Self-update logic
│   └── php_update.go     # PHP version updates
│
├── pkg/                   # Core packages
│   ├── cms/              # CMS-specific logic
│   ├── config/           # Configuration management
│   ├── database/         # Database operations
│   ├── ssl/              # SSL/TLS operations
│   ├── system/           # System-level operations
│   ├── updater/          # Self-update functionality
│   ├── utils/            # Utility functions
│   └── web/              # Web server operations
│
├── types/                # Shared type definitions
├── docs/                 # GitHub Pages documentation
├── main.go               # CLI entry point
└── *.sh                  # Build and install scripts
```

See full structure details in [Project Architecture](#code-architecture).

---

## Building from Source

### Development Build

```bash
go build -o svp
```

### Build All Platforms

```bash
VERSION=1.0.30 ./build-release.sh
```

### Install Locally

```bash
sudo bash install.sh
```

---

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/my-feature`)
3. Make changes with clear commits
4. Test on Debian/Ubuntu
5. Push and create pull request

See [full contributing guidelines](#contributing-1) for details.

---

[View Complete Development Guide →]({{ site.baseurl }}/development/guide)

---

[← Back to Home]({{ site.baseurl }}/) | [Documentation →]({{ site.baseurl }}/documentation/)
