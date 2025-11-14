# Simple VPS Provisioner (svp)

**Automate your VPS setup in minutes with a single command.**

Simple VPS Provisioner is a command-line tool that transforms a fresh Debian or Ubuntu VPS into a production-ready LAMP environment for Drupal or WordPress.

## ⚠️ Intended Use

This tool is designed for **fresh VPS installations** running the latest versions of:
- **Debian** (11, 12, or 13 Trixie)
- **Ubuntu LTS** (20.04, 22.04, or 24.04)

**Important:** Run `svp` on a newly provisioned VPS *before* installing other software. Running on a system with existing web server, database, or PHP configurations may cause conflicts. For best results, start with a clean OS installation.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

**📚 [Full Documentation](https://willjackson.github.io/simple-vps-provisioner/)**

---

## Features

- ✅ **Complete LAMP Stack** - Nginx, PHP-FPM, MariaDB, optional SSL
- ✅ **One Command Setup** - From fresh VPS to running site in minutes
- ✅ **Multi-CMS** - Drupal and WordPress support
- ✅ **Multi-Domain** - Multiple sites on one server
- ✅ **Git Deployment** - Clone and deploy from repositories
- ✅ **Optional SSL** - Let's Encrypt with auto-renewal (opt-in via --le-email)
- ✅ **Basic Authentication** - Password-protect sites with HTTP basic auth
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
sudo svp setup example.com --cms drupal --le-email admin@example.com
```

**WordPress:**
```bash
sudo svp setup example.com --cms wordpress --le-email admin@example.com
```

That's it! Visit `https://example.com` (HTTPS enabled with --le-email) 🎉

**HTTP-only setup (no SSL):**
```bash
sudo svp setup example.com --cms drupal
```
Visit `http://example.com`

---

## System Requirements

- **OS**: Debian 11-13 or Ubuntu 20.04-24.04 LTS
- **Access**: Root/sudo privileges

---

## Common Use Cases

### Deploy from Git Repository

```bash
sudo svp setup example.com \
  --cms drupal \
  --git-repo https://github.com/yourorg/yoursite.git \
  --le-email admin@example.com
```

### Multiple Environments

```bash
sudo svp setup example.com \
  --cms drupal \
  --extra-domains "staging.example.com,dev.example.com" \
  --le-email admin@example.com
```

### Import Existing Database

```bash
sudo svp setup example.com \
  --cms drupal \
  --db /path/to/backup.sql.gz \
  --le-email admin@example.com
```

### Update PHP Version

```bash
sudo svp php-update example.com --php-version 8.4
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

# Update SSL certificates (check renewal or enable SSL)
sudo svp update-ssl example.com check

# Check basic authentication status
sudo svp auth example.com check

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
| **SSL/HTTPS** | Let's Encrypt certificates (optional, enabled via --le-email) |
| **Firewall** | UFW configured (SSH, HTTP, HTTPS) |
| **Composer** | PHP dependency manager |
| **Drush/WP-CLI** | CMS-specific tools |

**Note:** SSL is disabled by default. To enable HTTPS, provide `--le-email` during setup or use `svp update-ssl` after initial setup.

---

## Key Configuration Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--cms` | `drupal` or `wordpress` | `drupal` |
| `--domain` | Primary domain (required) | - |
| `--php-version` | PHP version | `8.4` |
| `--le-email` | Let's Encrypt email (enables SSL when provided) | - |
| `--git-repo` | Git repository URL | - |
| `--git-branch` | Git branch (uses default if not specified) | - |
| `--db` | Database file to import | - |

**SSL Behavior:** SSL is disabled by default. When you provide `--le-email`, SSL is automatically enabled with Let's Encrypt certificates.

[View all flags →](https://willjackson.github.io/simple-vps-provisioner/documentation/command-line)

---

## SSL Management

### Enable SSL After Initial Setup

If you initially set up a site without SSL, you can enable it later:

```bash
# Enable SSL for an existing site
sudo svp update-ssl example.com --le-email admin@example.com
```

### Check SSL Certificate Status

```bash
# Check certificate renewal status
sudo svp update-ssl example.com check
```

### SSL During Initial Setup

```bash
# Enable SSL during setup (recommended)
sudo svp setup example.com --cms drupal --le-email admin@example.com
```

Certificates are automatically renewed via a cron job.

---

## Basic Authentication

### Password-Protect Your Site

Add an extra layer of security by requiring username/password authentication to access your site. This is useful for staging environments, development sites, or restricting access to specific content.

### Enable Authentication

```bash
# Enable with interactive prompts
sudo svp auth example.com enable

# Or provide credentials via flags
sudo svp auth example.com enable --username admin --password securepass123
```

The `auth` command automatically installs `apache2-utils` (htpasswd) if not already present. Changes take effect immediately.

### Check Authentication Status

```bash
# View current authentication status
sudo svp auth example.com check
```

This shows whether authentication is enabled and which username is configured.

### Update Credentials

To change the username or password, simply enable authentication again with new credentials:

```bash
# Replace existing credentials
sudo svp auth example.com enable --username newuser --password newpass456
```

This replaces the existing authentication configuration. Only one username/password pair is supported per domain.

### Disable Authentication

```bash
# Remove authentication requirement
sudo svp auth example.com disable
```

This removes the password protection and makes the site publicly accessible again.

### Important Notes

- Only one username/password combination per domain
- Credentials can be provided via `--username` and `--password` flags or entered interactively
- Changes take effect immediately (no server restart required)
- Re-enabling authentication replaces existing credentials
- Works with both HTTP and HTTPS sites

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
sudo svp setup mysite.com --cms drupal --le-email admin@mysite.com
```

### WordPress from Git
```bash
sudo svp setup blog.com --cms wordpress --git-repo https://github.com/me/blog.git --le-email admin@blog.com
```

### Staging + Production
```bash
sudo svp setup myapp.com --cms drupal --extra-domains "staging.myapp.com" --le-email dev@myapp.com
```

[View more examples →](https://willjackson.github.io/simple-vps-provisioner/examples)

---

<div align="center">

**[Get Started](https://willjackson.github.io/simple-vps-provisioner/getting-started) • [Documentation](https://willjackson.github.io/simple-vps-provisioner/documentation) • [Examples](https://willjackson.github.io/simple-vps-provisioner/examples)**

Made with ❤️ for the Drupal and WordPress communities

</div>
