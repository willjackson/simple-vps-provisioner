# Simple VPS Provisioner (svp)

**Automate your VPS setup in minutes with a single command.**

Simple VPS Provisioner is a command-line tool that transforms a fresh Debian or Ubuntu VPS into a production-ready LAMP environment for Drupal or WordPress.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

**📚 [Full Documentation](https://willjackson.github.io/simple-vps-provisioner/)**

---

## Features

- ✅ **Complete LAMP Stack** - Nginx, PHP-FPM, MariaDB, SSL
- ✅ **One Command Setup** - From fresh VPS to running site in minutes
- ✅ **Multi-CMS** - Drupal and WordPress support
- ✅ **Multi-Domain** - Multiple sites on one server
- ✅ **Git Deployment** - Clone and deploy from repositories
- ✅ **Automatic SSL** - Let's Encrypt with auto-renewal
- ✅ **Security First** - Firewall, hardening, isolation
- ✅ **Idempotent** - Safe to run multiple times

---

## Quick Start

### 1. Install svp

```bash
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash
```

### 2. Provision Your Site

**Drupal:**
```bash
sudo svp setup --cms drupal --domain example.com --le-email admin@example.com
```

**WordPress:**
```bash
sudo svp setup --cms wordpress --domain example.com --le-email admin@example.com
```

That's it! Visit `https://example.com` 🎉

---

## System Requirements

- **OS**: Debian 11-13 or Ubuntu 20.04-24.04 LTS
- **Access**: Root/sudo privileges

---

## Common Use Cases

### Deploy from Git Repository

```bash
sudo svp setup \
  --cms drupal \
  --domain example.com \
  --git-repo https://github.com/yourorg/yoursite.git \
  --le-email admin@example.com
```

### Multiple Environments

```bash
sudo svp setup \
  --cms drupal \
  --domain example.com \
  --extra-domains "staging.example.com,dev.example.com" \
  --le-email admin@example.com
```

### Import Existing Database

```bash
sudo svp setup \
  --cms drupal \
  --domain example.com \
  --db /path/to/backup.sql.gz \
  --le-email admin@example.com
```

### Update PHP Version

```bash
sudo svp php-update --domain example.com --php-version 8.4
```

---

## Essential Commands

```bash
# Check version
svp --version

# Verify configuration
sudo svp verify

# Update svp
sudo svp update

# View help
svp
```

---

## What Gets Installed

For each site, svp configures:

| Component | Description |
|-----------|-------------|
| **Nginx** | Web server with optimized config |
| **PHP-FPM** | Isolated per-domain pools |
| **MariaDB** | Database with secure credentials |
| **SSL/HTTPS** | Let's Encrypt certificates |
| **Firewall** | UFW configured (SSH, HTTP, HTTPS) |
| **Composer** | PHP dependency manager |
| **Drush/WP-CLI** | CMS-specific tools |

---

## Key Configuration Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--cms` | `drupal` or `wordpress` | `drupal` |
| `--domain` | Primary domain (required) | - |
| `--php-version` | PHP version | `8.4` |
| `--le-email` | Let's Encrypt email for SSL | - |
| `--git-repo` | Git repository URL | - |
| `--git-branch` | Git branch (uses default if not specified) | - |
| `--db` | Database file to import | - |

[View all flags →](https://willjackson.github.io/simple-vps-provisioner/documentation/command-line)

---

## Documentation

📖 **Full documentation available at: [willjackson.github.io/simple-vps-provisioner](https://willjackson.github.io/simple-vps-provisioner/)**

### Quick Links

- [Getting Started Guide](https://willjackson.github.io/simple-vps-provisioner/getting-started)
- [Complete Documentation](https://willjackson.github.io/simple-vps-provisioner/documentation)
- [Examples & Use Cases](https://willjackson.github.io/simple-vps-provisioner/examples)
- [Command-Line Reference](https://willjackson.github.io/simple-vps-provisioner/documentation/command-line)
- [Troubleshooting](https://willjackson.github.io/simple-vps-provisioner/documentation/troubleshooting)

---

## Installation Methods

### Quick Install (Recommended)

Installs the latest stable release:

```bash
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash
```

### Manual Download

```bash
VERSION="1.0.35"  # Check releases for latest
wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp-linux-amd64
chmod +x svp-linux-amd64
sudo mv svp-linux-amd64 /usr/local/bin/svp
```

### Build from Source

For development:

```bash
git clone https://github.com/willjackson/simple-vps-provisioner.git
cd simple-vps-provisioner
sudo bash install.sh
```

[More installation options →](https://willjackson.github.io/simple-vps-provisioner/getting-started)

---

## Support & Contributing

- **📖 Documentation**: [willjackson.github.io/simple-vps-provisioner](https://willjackson.github.io/simple-vps-provisioner/)
- **🐛 Bug Reports**: [GitHub Issues](https://github.com/willjackson/simple-vps-provisioner/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/willjackson/simple-vps-provisioner/discussions)

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Quick Examples

### Fresh Drupal Site
```bash
sudo svp setup --cms drupal --domain mysite.com --le-email admin@mysite.com
```

### WordPress from Git
```bash
sudo svp setup --cms wordpress --domain blog.com --git-repo https://github.com/me/blog.git --le-email admin@blog.com
```

### Staging + Production
```bash
sudo svp setup --cms drupal --domain myapp.com --extra-domains "staging.myapp.com" --le-email dev@myapp.com
```

[View more examples →](https://willjackson.github.io/simple-vps-provisioner/examples)

---

<div align="center">

**[Get Started](https://willjackson.github.io/simple-vps-provisioner/getting-started) • [Documentation](https://willjackson.github.io/simple-vps-provisioner/documentation) • [Examples](https://willjackson.github.io/simple-vps-provisioner/examples)**

Made with ❤️ for the Drupal and WordPress communities

</div>
