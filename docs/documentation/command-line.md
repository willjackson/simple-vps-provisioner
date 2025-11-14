---
layout: default
title: Command-Line Reference
---

# Command-Line Reference

Complete reference for all svp command-line options and commands.

## Table of Contents

- [Commands](#commands)
- [Global Flags](#global-flags)
- [CMS Options](#cms-options)
- [Domain Configuration](#domain-configuration)
- [PHP Configuration](#php-configuration)
- [SSL/Security Options](#sslsecurity-options)
- [Git Deployment](#git-deployment)
- [Database Options](#database-options)
- [System Options](#system-options)

---

## Commands

svp uses positional commands instead of flags.

### Setup Command

Perform full VPS provisioning.

```bash
svp setup DOMAIN [options]
```

**What it does:**
- Installs system packages
- Configures web server
- Installs PHP-FPM
- Sets up database
- Installs CMS
- Obtains SSL certificates (if enabled)
- Configures firewall

**Example:**
```bash
sudo svp setup example.com --cms drupal
```

### Verify Command

Check configuration without making changes.

```bash
svp verify
```

**What it checks:**
- Base packages installed
- Nginx running
- PHP-FPM running
- Database running
- Composer installed
- Firewall configured
- SSL certificates (if configured)

**Example output:**
```
==========================================================
  Configuration Verification
==========================================================

=== Base Packages ===
[✓] All base packages installed

=== Nginx ===
[✓] Nginx already installed
[✓] nginx running

=== PHP ===
[✓] PHP 8.3 packages already installed
[✓] php8.3-fpm running

==========================================================
[✓] All checks passed!
==========================================================
```

### Update Command

Update svp to the latest version.

```bash
svp update
```

**What it does:**
- Checks GitHub for latest release
- Downloads new binary
- Verifies checksum
- Replaces current installation
- Creates backup of old version

**Example:**
```bash
$ sudo svp update

[CREATE] Checking for updates...
Current version: v1.0.24
Latest version:  v1.0.30

New version available: v1.0.30
Update now? [y/N]: y

[CREATE] Downloading svp v1.0.30...
[CREATE] Verifying checksum...
[✓] Checksum verified
[CREATE] Installing new version...
[✓] Successfully updated to v1.0.30!
```

### PHP Update Command

Update PHP version for a domain.

```bash
svp php-update DOMAIN --php-version VERSION
```

**What it does:**
- Installs new PHP version
- Creates new PHP-FPM pool
- Updates Nginx configuration
- Reconfigures SSL (if needed)
- Updates site configuration
- Removes old PHP pool

**Example:**
```bash
sudo svp php-update example.com --php-version 8.4
```

### SSL Management Command

Manage SSL certificates for a domain.

```bash
svp update-ssl DOMAIN ACTION [options]
```

**Actions:**
- `enable` - Enable SSL and obtain Let's Encrypt certificate
- `disable` - Disable SSL and switch to HTTP only
- `renew` - Force renewal of existing SSL certificate
- `check` - Check SSL certificate status and expiration

**What it does:**
- Manages SSL certificates for existing domains
- Enables/disables HTTPS configuration
- Renews certificates before expiration
- Validates certificate status

**Examples:**

Enable SSL for an existing domain:
```bash
sudo svp update-ssl example.com enable --le-email admin@example.com
```

Disable SSL and switch to HTTP:
```bash
sudo svp update-ssl example.com disable
```

Force certificate renewal:
```bash
sudo svp update-ssl example.com renew
```

Check certificate status:
```bash
sudo svp update-ssl example.com check
```

**Note:** The `enable` action requires `--le-email` to obtain a certificate.

---

## Global Flags

### --version

Display version information.

```bash
svp --version
```

**Output:**
```
Simple VPS Provisioner (svp) version 1.0.30
```

### --debug

Enable debug mode for troubleshooting.

```bash
svp --debug setup example.com --cms drupal
```

Enables verbose output showing all command execution.

---

## CMS Options

### --cms

Choose CMS to install.

```bash
--cms drupal     # Install Drupal (default)
--cms wordpress  # Install WordPress
```

**Examples:**

```bash
# Drupal
sudo svp setup example.com --cms drupal

# WordPress
sudo svp setup myblog.com --cms wordpress
```

---

## Domain Configuration

### DOMAIN (positional argument)

Primary domain name (required for setup and php-update commands).

**Usage:**
```bash
svp setup DOMAIN [options]
svp php-update DOMAIN [options]
```

**Example:**
```bash
sudo svp setup example.com --cms drupal
```

### --extra-domains

Additional domains (comma-separated).

```bash
--extra-domains "staging.example.com,dev.example.com"
```

Each domain gets:
- Separate directory
- Dedicated database
- Isolated PHP-FPM pool
- Individual Nginx vhost

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --extra-domains "staging.example.com,dev.example.com"
```

Creates three separate environments:
- `https://example.com`
- `https://staging.example.com`
- `https://dev.example.com`

---

## PHP Configuration

### --php-version

PHP version to install.

```bash
--php-version 8.4  # Default
--php-version 8.3
--php-version 8.2
--php-version 8.1
```

**Example:**
```bash
sudo svp setup example.com --cms drupal --php-version 8.4
```

**Available versions:**
- PHP 8.4 (latest, default, recommended)
- PHP 8.3
- PHP 8.2
- PHP 8.1

---

## SSL/Security Options

### --ssl

Enable or disable SSL/HTTPS.

```bash
--ssl=true   # Enable SSL explicitly
--ssl=false  # Disable SSL (HTTP only, default)
```

**Default behavior:**
- SSL is disabled by default (false)
- Automatically enabled when `--le-email` is provided
- Can be explicitly set with `--ssl=true` or `--ssl=false`

**Note:** When SSL is enabled, it requires `--le-email` to obtain certificates.

**Example (with SSL via --le-email):**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --le-email admin@example.com
```

**Example (with SSL explicitly):**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --ssl=true \
  --le-email admin@example.com
```

**Example (HTTP only - omit --le-email):**
```bash
sudo svp setup example.com --cms drupal
```

### --le-email

Let's Encrypt email for SSL certificates.

```bash
--le-email admin@example.com
```

**Behavior:**
- Automatically enables SSL when provided
- No need to specify `--ssl=true` separately

**Required when:**
- Enabling SSL during setup
- Obtaining SSL certificates

**Used for:**
- Certificate expiration notices
- Important updates from Let's Encrypt

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --le-email admin@example.com
```

### --firewall

Enable or disable UFW firewall.

```bash
--firewall=true   # Enable (default)
--firewall=false  # Disable
```

**Example:**
```bash
sudo svp setup example.com --cms drupal --firewall=true
```

---

## Git Deployment

### --git-repo

Git repository URL to clone.

```bash
--git-repo https://github.com/myorg/mysite.git
--git-repo git@github.com:myorg/mysite.git
```

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --git-repo https://github.com/myorg/mysite.git
```

### --git-branch

Git branch to checkout (optional).

```bash
--git-branch production
--git-branch develop
--git-branch main
```

**Note:** If not specified, uses the repository's default branch.

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --git-repo https://github.com/myorg/mysite.git \
  --git-branch production
```

### --drupal-root

Drupal root directory (relative to repo).

```bash
--drupal-root ""          # Auto-detect (default)
--drupal-root "drupal"
--drupal-root "backend"
```

**Use when:**
- Drupal is in a subdirectory of your Git repo
- Using a monorepo structure

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --git-repo https://github.com/myorg/monorepo.git \
  --drupal-root "backend"
```

### --docroot

Custom document root path.

```bash
--docroot "web"           # Default for Drupal
--docroot "public"
--docroot "htdocs"
```

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --git-repo https://github.com/myorg/mysite.git \
  --docroot "public_html"
```

---

## Database Options

### --db-engine

Database engine to use.

```bash
--db-engine mariadb  # Default
--db-engine none     # Skip database installation
```

**Example (with database):**
```bash
sudo svp setup example.com --cms drupal --db-engine mariadb
```

**Example (without database):**
```bash
sudo svp setup example.com --cms drupal --db-engine none
```

### --db

Path to database file for import.

```bash
--db /path/to/backup.sql
--db /path/to/backup.sql.gz  # Compressed files supported
```

**Supported formats:**
- `.sql` - Plain SQL
- `.sql.gz` - Gzip compressed

**When specified:**
- Database is imported instead of site-install
- Drupal config import is skipped
- Existing database is cleared first

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --db /home/admin/backup.sql.gz
```

### --keep-existing-db

Keep existing database when reprovisioning.

```bash
--keep-existing-db=false  # Drop database completely (default)
--keep-existing-db=true   # Keep database, reuse credentials
```

**Default behavior (false):**
- Drops entire database and user
- Creates fresh database with NEW credentials
- Maximum security

**With flag (true):**
- Keeps existing database and user
- Drops all tables only
- Reuses SAME credentials

**Example (fresh database):**
```bash
sudo svp setup example.com --cms drupal
```

**Example (keep credentials):**
```bash
sudo svp setup example.com --cms drupal --keep-existing-db
```

---

## System Options

### --webroot

Parent directory for sites.

```bash
--webroot /var/www        # Default
--webroot /home/sites
--webroot /srv/www
```

Sites are created as:
- `/var/www/example.com/`
- `/var/www/staging.example.com/`

**Example:**
```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --webroot /srv/www
```

### --create-swap

Create swap space.

```bash
--create-swap auto  # Create if RAM < 2GB (default)
--create-swap yes   # Always create
--create-swap no    # Never create
```

**auto mode:**
- Creates 2GB swap if total RAM < 2GB
- Skips if RAM >= 2GB

**Example:**
```bash
sudo svp setup example.com --cms drupal --create-swap yes
```

---

## Complete Examples

### Minimal Drupal (HTTP only, no SSL)

```bash
sudo svp setup example.com --cms drupal
```

Note: SSL is disabled by default when --le-email is not provided.

### Production Drupal with SSL

```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --le-email admin@example.com
```

### Drupal from Git with Custom PHP

```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --php-version 8.4 \
  --git-repo https://github.com/myorg/mysite.git \
  --git-branch production \
  --le-email admin@example.com
```

### WordPress with Database Import

```bash
sudo svp setup \
  myblog.com \
  --cms wordpress \
  --db /home/admin/wp-backup.sql.gz \
  --le-email admin@myblog.com
```

### Multi-Environment Setup

```bash
sudo svp setup \
  mysite.com \
  --cms drupal \
  --extra-domains "staging.mysite.com,dev.mysite.com" \
  --git-repo https://github.com/myorg/mysite.git \
  --le-email admin@mysite.com
```

### Reprovision with Fresh Database

```bash
sudo svp setup example.com --cms drupal
```

### Reprovision Keeping Credentials

```bash
sudo svp setup example.com --cms drupal --keep-existing-db
```

### Enable SSL on Existing Site

```bash
sudo svp update-ssl example.com enable --le-email admin@example.com
```

### Disable SSL on Existing Site

```bash
sudo svp update-ssl example.com disable
```

### Renew SSL Certificate

```bash
sudo svp update-ssl example.com renew
```

---

## Tips

### SSL Configuration

**When to use --le-email vs update-ssl:**

Use `--le-email` during initial setup:
```bash
sudo svp setup example.com --cms drupal --le-email admin@example.com
```

Use `update-ssl` to manage SSL on existing sites:
```bash
sudo svp update-ssl example.com enable --le-email admin@example.com
sudo svp update-ssl example.com disable
sudo svp update-ssl example.com renew
sudo svp update-ssl example.com check
```

**SSL Behavior and Defaults:**
- SSL is disabled by default (HTTP only)
- Providing `--le-email` automatically enables SSL
- You can explicitly disable SSL with `--ssl=false` even if `--le-email` is provided
- Use `update-ssl` command to add/remove SSL after initial setup
- Certificates auto-renew via certbot, but you can force renewal with `update-ssl renew`

### Using Quotes

Use quotes for values with special characters:

```bash
--extra-domains "site1.com,site2.com"
--git-repo "https://github.com/org/repo.git"
```

### Combining Flags

All flags can be combined:

```bash
sudo svp setup \
  example.com \
  --cms drupal \
  --php-version 8.4 \
  --webroot /srv/www \
  --git-repo https://github.com/myorg/mysite.git \
  --git-branch production \
  --db /path/to/backup.sql.gz \
  --le-email admin@example.com \
  --firewall=true \
  --create-swap yes
```

### Getting Help

View all options:

```bash
svp --help
```

Or just run svp with no arguments:

```bash
svp
```

---

[← Documentation]({{ site.baseurl }}/documentation/) | [SSL Configuration →](ssl-configuration)
