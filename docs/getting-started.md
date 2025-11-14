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
svp --version
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
sudo svp setup example.com \
  --cms drupal \
  --le-email admin@example.com
```

For WordPress:

```bash
sudo svp setup example.com \
  --cms wordpress \
  --le-email admin@example.com
```

### Step 3: Watch the Magic Happen

svp will automatically:
1. Install and configure Nginx
2. Install PHP-FPM 8.3
3. Install and secure MariaDB
4. Create database and user
5. Install Composer
6. Install your CMS
7. Obtain SSL certificate (if --le-email provided)
8. Configure firewall

The process takes 5-15 minutes depending on your VPS speed.

### Step 4: Access Your Site

Once complete, you'll see a summary with:
- **Site URL**: `https://example.com` (if --le-email was provided) or `http://example.com` (default)
- **Database credentials**: `/etc/svp/sites/example.com.db.txt`
- **Login link**: One-time admin login (Drupal)

HTTPS is only available if you provided the `--le-email` flag during setup. By default, sites run on HTTP only. You can enable SSL later using the `update-ssl` command (see below).

## Common Options

### Choose PHP Version

```bash
sudo svp setup example.com --cms drupal --php-version 8.4
```

### Deploy from Git

```bash
sudo svp setup example.com \
  --cms drupal \
  --git-repo https://github.com/myorg/mysite.git \
  --git-branch production \
  --le-email admin@example.com
```

### Multiple Domains

```bash
sudo svp setup example.com \
  --cms drupal \
  --extra-domains "staging.example.com,dev.example.com" \
  --le-email admin@example.com

# Optional: Password-protect staging/dev environments
sudo svp auth staging.example.com enable
sudo svp auth dev.example.com enable
```

### Import Existing Database

```bash
sudo svp setup example.com \
  --cms drupal \
  --db /path/to/backup.sql.gz \
  --le-email admin@example.com
```

### HTTP Only (No SSL)

By default, sites are created without SSL. Simply omit the `--le-email` flag:

```bash
sudo svp setup example.com --cms drupal
```

SSL is only enabled when you provide the `--le-email` flag during setup.

## Managing SSL After Setup

You can enable, check, renew, or disable SSL at any time using the `update-ssl` command.

### Enable SSL After Initial Setup

If you created a site without SSL, you can enable it later:

```bash
sudo svp update-ssl example.com --le-email admin@example.com
```

This will:
- Install Certbot (if not already installed)
- Obtain a Let's Encrypt SSL certificate
- Update Nginx configuration to use HTTPS
- Set up automatic certificate renewal

### Check SSL Status

To see the current SSL status for a site:

```bash
sudo svp update-ssl example.com --check
```

This displays:
- Whether SSL is enabled
- Certificate expiration date (if enabled)
- Certificate domains covered

### Renew SSL Certificate

Certificates renew automatically, but you can manually renew:

```bash
sudo svp update-ssl example.com --renew
```

### Disable SSL

To remove SSL and revert to HTTP only:

```bash
sudo svp update-ssl example.com --disable
```

This will:
- Update Nginx to use HTTP only
- Keep the certificate in place (but not use it)
- Stop automatic renewal

## Password Protection with Basic Authentication

Basic authentication provides a simple way to password-protect your sites. This is particularly useful for:
- Staging and development environments
- Sites under construction
- Internal tools and dashboards
- Preventing search engine indexing during development

### Enable Authentication

You can enable basic auth with interactive prompts:

```bash
sudo svp auth example.com enable
```

Or provide credentials directly via flags:

```bash
sudo svp auth example.com enable --username admin --password secure123
```

This will:
- Install apache2-utils (if not already installed)
- Create a password file using htpasswd
- Update Nginx configuration to require authentication
- Reload Nginx to apply changes

### Check Authentication Status

To see if authentication is enabled for a site:

```bash
sudo svp auth example.com check
```

This displays:
- Whether basic auth is enabled or disabled
- The username configured (if enabled)
- Location of the password file

### Disable Authentication

To remove password protection:

```bash
sudo svp auth example.com disable
```

This will:
- Update Nginx configuration to remove auth requirement
- Keep the password file in place (but not use it)
- Reload Nginx to apply changes

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
- Certbot (optional, installed only when SSL is enabled via --le-email)
- UFW firewall
- WP-CLI (for WordPress)
- apache2-utils (provides htpasswd for basic authentication)

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
sudo svp verify
```

## Getting Help

- **Documentation**: [Read full documentation]({{ site.baseurl }}/documentation/)
- **Examples**: [View more examples]({{ site.baseurl }}/examples/)
- **GitHub Issues**: [Report problems](https://github.com/willjackson/simple-vps-provisioner/issues)

## What's Next?

- [Read detailed documentation]({{ site.baseurl }}/documentation/)
- [Explore examples and use cases]({{ site.baseurl }}/examples/)
- [Learn about development]({{ site.baseurl }}/development/)

---

[← Back to Home]({{ site.baseurl }}/) | [Documentation →]({{ site.baseurl }}/documentation/)
