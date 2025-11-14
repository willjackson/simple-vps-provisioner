---
layout: default
title: Documentation
---

# Documentation

Complete reference for Simple VPS Provisioner.

## Table of Contents

### Installation & Setup
- [Getting Started]({{ site.baseurl }}/getting-started)
- [System Requirements](#system-requirements)
- [Installation Methods](#installation-methods)
- [First Time Setup](#first-time-setup)

### Core Features
- [CMS Installation](cms-installation)
- [SSL/HTTPS Configuration](ssl-configuration)
- [Basic Authentication](security#basic-authentication-for-sites)
- [Multi-Domain Setup](multi-domain)
- [Git Deployment](git-deployment)
- [Database Management](database-management)
- [PHP Version Management](php-versions)

### Command Reference
- [Command-Line Options](command-line)
- [Configuration Files](configuration-files)
- [Environment Variables](#environment-variables)

### Advanced Topics
- [Reprovisioning Sites](reprovisioning)
- [Security Best Practices](security)
- [Performance Tuning](performance)
- [Backup & Recovery](backup-recovery)
- [Troubleshooting](troubleshooting)

---

## System Requirements

### Supported Operating Systems

**Debian:**
- Debian 13 (Trixie) - Fully supported ✅
- Debian 12 (Bookworm) - Fully supported ✅
- Debian 11 (Bullseye) - Supported ✅

**Ubuntu:**
- Ubuntu 24.04 LTS (Noble) - Supported ✅
- Ubuntu 22.04 LTS (Jammy) - Fully supported ✅ (Most stable)
- Ubuntu 20.04 LTS (Focal) - Fully supported ✅

### Required Access

- **Root access** (sudo privileges)
- **SSH access** to the VPS
- **Port access**: 22 (SSH), 80 (HTTP), 443 (HTTPS)

---

## Installation Methods

### Method 1: Quick Install (Recommended)

For most users, the quick install script is the easiest option:

```bash
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash
```

**Pros:**
- ✅ Fastest installation
- ✅ Automatic checksum verification
- ✅ No Go compiler needed
- ✅ Gets latest stable release

**Use when:**
- First-time installation
- Production servers
- You want the stable release

### Method 2: Manual Download

Download and verify manually:

```bash
# Set version (get latest from releases page)
VERSION="1.0.30"

# Download binary
wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp-linux-amd64

# Download checksums
wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/checksums.txt

# Verify
sha256sum --check --ignore-missing checksums.txt

# Install
chmod +x svp-linux-amd64
sudo mv svp-linux-amd64 /usr/local/bin/svp

# Verify
svp --version
```

**Use when:**
- You want to review the installer
- Air-gapped installations
- Corporate environments requiring manual verification

### Method 3: Build from Source

For development or customization:

```bash
# Clone repository
git clone https://github.com/willjackson/simple-vps-provisioner.git
cd simple-vps-provisioner

# Build and install
sudo bash install.sh
```

**Use when:**
- Contributing to development
- Customizing the tool
- Testing unreleased features

---

## First Time Setup

After installation, your first provisioning looks like this:

### 1. Prepare Your Domain

Configure DNS A record:
```
A    example.com    YOUR_SERVER_IP
```

Wait 5-30 minutes for DNS propagation.

### 2. Run Setup Command

```bash
sudo svp setup example.com \
  --cms drupal \
  --le-email admin@example.com
```

Note: SSL is enabled when you provide `--le-email`. Omit it to set up without SSL initially.

### 3. Wait for Completion

svp will automatically:
1. ✅ Update system packages
2. ✅ Install Nginx
3. ✅ Install PHP-FPM
4. ✅ Install MariaDB
5. ✅ Create database and user
6. ✅ Install CMS
7. ✅ Obtain SSL certificate
8. ✅ Configure firewall

### 4. Access Your Site

Visit `https://example.com` to see your site!

For Drupal, you'll get a one-time login link:
```
Login to admin: https://example.com/user/reset/...
```

---

## Environment Variables

svp respects these environment variables:

### DEBUG

Enable debug output:
```bash
DEBUG=1 svp setup example.com --cms drupal
# or use the --debug flag:
svp setup example.com --cms drupal --debug
```

Shows detailed command execution and troubleshooting information.

---

## Configuration Files

### System-Wide Configuration

**Location:** `/etc/svp/`

```
/etc/svp/
├── php.conf              # Current PHP version
└── sites/                # Per-site configurations
    ├── example.com.conf  # Site config
    └── example.com.db.txt # Database credentials
```

### Per-Site Configuration

**Site Config** (`/etc/svp/sites/example.com.conf`):
```bash
# Site configuration for example.com
DOMAIN='example.com'
PHP_VERSION='8.3'
WEBROOT='/var/www/example.com/web'
CREATED='Mon Jan 15 10:30:00 UTC 2024'
SSL_EMAIL='admin@example.com'  # If SSL enabled
```

**Database Credentials** (`/etc/svp/sites/example.com.db.txt`):
```
Database: drupal_example_com
Username: drupal_example_com
Password: [auto-generated secure password]
Host: localhost
Port: 3306
```

⚠️ **Important:** Keep database credential files secure with 600 permissions.

---

## Next Topics

Explore detailed documentation for specific features:

- [CMS Installation Guide](cms-installation) - Drupal and WordPress specifics
- [SSL/HTTPS Configuration](ssl-configuration) - Let's Encrypt setup
- [Multi-Domain Setup](multi-domain) - Multiple sites on one server
- [Git Deployment](git-deployment) - Deploy from repositories
- [Database Management](database-management) - Import, export, backup
- [PHP Version Management](php-versions) - Update PHP versions
- [Reprovisioning](reprovisioning) - Update existing sites
- [Security Best Practices](security) - Hardening and protection
- [Troubleshooting](troubleshooting) - Common issues and solutions

---

[← Back to Home]({{ site.baseurl }}/) | [Examples →]({{ site.baseurl }}/examples/)
