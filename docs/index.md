---
layout: default
title: Home
---

# Simple VPS Provisioner

**Automate your VPS setup in minutes, not hours.**

Simple VPS Provisioner (svp) is a command-line tool that transforms a fresh Debian or Ubuntu VPS into a production-ready LAMP environment for Drupal or WordPress.

## What It Does

With a single command, svp provisions your entire server stack:

- ✅ **Web Server**: Nginx with optimized configuration
- ✅ **PHP**: PHP-FPM with per-domain isolation
- ✅ **Database**: MariaDB with automatic setup
- ✅ **SSL/HTTPS**: Let's Encrypt certificates with auto-renewal
- ✅ **Security**: UFW firewall, hardened PHP settings, secure credentials
- ✅ **CMS**: Complete Drupal or WordPress installation
- ✅ **Git Deploy**: Clone and deploy from repositories
- ✅ **Multi-Domain**: Support for multiple sites on one server

## Quick Start

```bash
# Install svp
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash

# Provision a Drupal site
sudo svp setup -cms drupal -domain example.com -le-email admin@example.com

# Provision a WordPress site
sudo svp setup -cms wordpress -domain myblog.com -le-email admin@myblog.com
```

That's it! Your site is ready at `https://example.com`

## Why Simple VPS Provisioner?

### 🚀 **Fast Setup**
Go from a fresh VPS to a running website in minutes. No manual configuration required.

### 🔒 **Secure by Default**
- Automatic firewall configuration
- SSL/HTTPS with Let's Encrypt
- Hardened PHP settings
- Secure database credentials
- Per-domain process isolation

### 🎯 **Purpose-Built**
Optimized specifically for Drupal and WordPress on Debian/Ubuntu VPS. No bloat, no unnecessary features.

### 🔄 **Idempotent**
Safe to run multiple times. Won't break existing installations.

### 🌐 **Multi-Site Ready**
Run multiple domains on a single VPS with isolated environments.

### 🔧 **Flexible**
- Choose your PHP version
- Deploy from Git repositories
- Import existing databases
- Custom configuration options

## Supported Platforms

- **Debian**: 11 (Bullseye), 12 (Bookworm), 13 (Trixie)
- **Ubuntu**: 20.04 LTS, 22.04 LTS, 24.04 LTS

The tool automatically detects your OS and configures packages accordingly.

## Use Cases

### Fresh Drupal Installation
Perfect for starting new Drupal projects with best practices built-in.

```bash
sudo svp setup -cms drupal -domain mysite.com -le-email admin@mysite.com
```

### Deploy Existing Site
Clone and configure an existing Drupal or WordPress site from Git.

```bash
sudo svp setup -cms drupal \
  -domain mysite.com \
  -git-repo https://github.com/myorg/mysite.git \
  -git-branch production \
  -le-email admin@mysite.com
```

### Multiple Staging Environments
Set up production, staging, and development environments on one server.

```bash
sudo svp setup -cms drupal \
  -domain mysite.com \
  -extra-domains "staging.mysite.com,dev.mysite.com" \
  -le-email admin@mysite.com
```

### WordPress with Existing Database
Import an existing WordPress database during setup.

```bash
sudo svp setup -cms wordpress \
  -domain myblog.com \
  -db /path/to/backup.sql.gz \
  -le-email admin@myblog.com
```

## What's Next?

<div class="button-row">
  <a href="{{ site.baseurl }}/getting-started" class="button">Get Started →</a>
  <a href="{{ site.baseurl }}/documentation/" class="button secondary">Read Documentation</a>
  <a href="{{ site.baseurl }}/examples/" class="button secondary">View Examples</a>
</div>

## Features at a Glance

| Feature | Description |
|---------|-------------|
| **Automated Setup** | Complete LAMP stack in one command |
| **CMS Support** | Drupal and WordPress ready |
| **SSL/HTTPS** | Automatic Let's Encrypt certificates |
| **Security** | Firewall, hardening, secure defaults |
| **Multi-Domain** | Multiple sites per server |
| **Git Deploy** | Clone from repositories |
| **PHP Versions** | Choose 8.1, 8.2, 8.3, or 8.4 |
| **Database Import** | Restore from existing backups |
| **Per-Domain Pools** | Isolated PHP-FPM processes |
| **Auto-Updates** | Self-update capability built-in |

## Community & Support

- **GitHub**: [willjackson/simple-vps-provisioner](https://github.com/willjackson/simple-vps-provisioner)
- **Issues**: [Report bugs or request features](https://github.com/willjackson/simple-vps-provisioner/issues)
- **Changelog**: [View version history](https://github.com/willjackson/simple-vps-provisioner/blob/main/CHANGELOG.md)

## License

Simple VPS Provisioner is open source software licensed under the MIT License.

---

<style>
.button-row {
  margin: 2em 0;
  display: flex;
  gap: 1em;
  flex-wrap: wrap;
}
.button {
  display: inline-block;
  padding: 0.75em 1.5em;
  background: #159957;
  color: white;
  text-decoration: none;
  border-radius: 4px;
  font-weight: bold;
}
.button:hover {
  background: #0e7a46;
  color: white;
}
.button.secondary {
  background: #606c71;
}
.button.secondary:hover {
  background: #505860;
}
</style>
