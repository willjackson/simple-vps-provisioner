---
layout: default
title: Getting Started
---

# Getting Started with Simple VPS Provisioner

This guide will help you install svp and provision your first site.

## Prerequisites

- **Debian**: Version 11, 12, or 13
- **Ubuntu**: 20.04 LTS, 22.04 LTS, or 24.04 LTS
- **Root access** to your VPS
- **DNS configured** (for SSL/HTTPS)

## Installation

### Quick Install (Recommended)

The easiest way to install svp is using our automated installer:

```bash
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash
```

This script will:
- ✅ Download the latest release
- ✅ Verify checksums
- ✅ Install to `/usr/local/bin/svp`
- ✅ Make it executable

### Manual Installation

If you prefer to review the installer first:

```bash
# Download and review
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh -o install-svp.sh
less install-svp.sh

# Run after review
sudo bash install-svp.sh
```

### Verify Installation

```bash
svp -version
# Output: Simple VPS Provisioner (svp) version 1.0.30
```

## Your First Site

### Step 1: Configure DNS

Before running svp, ensure your domain points to your server:

```
A    example.com          123.45.67.89
A    www.example.com      123.45.67.89
```

DNS propagation typically takes 5-30 minutes.

### Step 2: Run Setup

For a fresh Drupal installation:

```bash
sudo svp -mode setup \
  -cms drupal \
  -domain example.com \
  -le-email admin@example.com
```

For WordPress:

```bash
sudo svp -mode setup \
  -cms wordpress \
  -domain example.com \
  -le-email admin@example.com
```

### Step 3: Watch the Magic Happen

svp will automatically:
1. Install and configure Nginx
2. Install PHP-FPM 8.3
3. Install and secure MariaDB
4. Create database and user
5. Install Composer
6. Install your CMS
7. Obtain SSL certificate
8. Configure firewall

The process takes 5-15 minutes depending on your VPS speed.

### Step 4: Access Your Site

Once complete, you'll see a summary with:
- **Site URL**: `https://example.com`
- **Database credentials**: `/etc/svp/sites/example.com.db.txt`
- **Login link**: One-time admin login (Drupal)

## Common Options

### Choose PHP Version

```bash
sudo svp -mode setup -cms drupal -domain example.com -php-version 8.4
```

### Deploy from Git

```bash
sudo svp -mode setup \
  -cms drupal \
  -domain example.com \
  -git-repo https://github.com/myorg/mysite.git \
  -git-branch production \
  -le-email admin@example.com
```

### Multiple Domains

```bash
sudo svp -mode setup \
  -cms drupal \
  -domain example.com \
  -extra-domains "staging.example.com,dev.example.com" \
  -le-email admin@example.com
```

### Import Existing Database

```bash
sudo svp -mode setup \
  -cms drupal \
  -domain example.com \
  -db /path/to/backup.sql.gz \
  -le-email admin@example.com
```

### HTTP Only (No SSL)

```bash
sudo svp -mode setup -cms drupal -domain example.com -ssl=false
```

## Understanding the Output

svp provides clear, color-coded output:

- 🟢 **[CREATE]**: Creating or installing something
- 🔵 **[VERIFY]**: Checking if already exists
- ⚪ **[SKIP]**: Skipping (already done or disabled)
- 🟡 **[FIX]**: Repairing or fixing configuration
- ✅ **[✓]**: Success
- ❌ **[✗]**: Failure
- ⚠️ **[!]**: Warning

## What Gets Installed

### System Packages
- Base utilities (curl, wget, git, etc.)
- Nginx web server
- PHP-FPM and extensions
- MariaDB database server
- Composer
- Certbot (for SSL)
- UFW firewall
- WP-CLI (for WordPress)

### Directory Structure

```
/var/www/example.com/          # Your site directory
├── composer.json              # Dependencies
├── vendor/                    # Composer packages
├── web/                       # Document root
│   ├── index.php
│   └── sites/default/
│       ├── settings.php       # Main settings
│       ├── settings.svp.php   # SVP-managed settings (gitignored)
│       └── files/             # Uploaded files
└── config/sync/               # Drupal configuration

/etc/svp/                      # SVP configuration
├── sites/                     # Per-site configs
│   ├── example.com.conf       # Site configuration
│   └── example.com.db.txt     # Database credentials
└── php.conf                   # PHP version tracking
```

### Configuration Files

- **Nginx vhost**: `/etc/nginx/sites-available/example.com.conf`
- **PHP-FPM pool**: `/etc/php/8.3/fpm/pool.d/example.com.conf`
- **Database credentials**: `/etc/svp/sites/example.com.db.txt`

## Next Steps

### For Drupal

```bash
# Get one-time login link
drush-example.com uli

# Check site status
drush-example.com status

# Clear cache
drush-example.com cr

# Import configuration
drush-example.com cim
```

### For WordPress

```bash
# Navigate to site
cd /var/www/example.com

# List plugins
sudo -u admin wp plugin list

# Activate theme
sudo -u admin wp theme activate twentytwentyfour
```

## Troubleshooting

### DNS Not Configured

If you see this error:

```
[!] DNS mismatch detected!
```

**Solution**: Configure your DNS A record to point to your server IP and wait for propagation.

### SSL Certificate Failed

**Common causes:**
- DNS not pointing to server
- Port 80/443 blocked by firewall
- Rate limit reached (5 certs per week)

**Check DNS:**
```bash
dig +short example.com
# Should return your server IP
```

### Site Returns 502 Bad Gateway

**Check PHP-FPM:**
```bash
sudo systemctl status php8.3-fpm
sudo tail -f /var/log/php8.3-fpm-example.com-error.log
```

### Database Connection Failed

**Check credentials:**
```bash
sudo cat /etc/svp/sites/example.com.db.txt
```

### General Verification

Run the verify command to check configuration:

```bash
sudo svp -mode verify
```

## Getting Help

- **Documentation**: [Read full documentation](/documentation/)
- **Examples**: [View more examples](/examples/)
- **GitHub Issues**: [Report problems](https://github.com/willjackson/simple-vps-provisioner/issues)

## What's Next?

- [Read detailed documentation](/documentation/)
- [Explore examples and use cases](/examples/)
- [Learn about development](/development/)

---

[← Back to Home](/) | [Documentation →](/documentation/)
