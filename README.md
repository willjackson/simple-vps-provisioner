# Simple VPS Provisioner (svp)

A Go-based command-line tool for provisioning Debian/Ubuntu VPS with LAMP stack for Drupal or WordPress.

## Features

- **Automated VPS provisioning** - Complete setup with one command
- **Multi-CMS support** - Choose between Drupal or WordPress
- **LAMP stack** - Nginx, PHP-FPM, MariaDB, Composer
- **Multi-domain support** - Provision multiple domains in one run
- **Per-domain PHP-FPM pools** - Isolated PHP processes for each site
- **Git deployment** - Deploy from existing repositories
- **SSL/HTTPS support** - Automatic Let's Encrypt certificates with DNS verification
- **DNS verification** - Automatic check before SSL certificate issuance
- **Clean reprovisioning** - Fresh database and credentials by default when reprovisioning
- **Idempotent** - Safe to run multiple times
- **Security hardening** - Firewall, PHP settings, database security, HSTS, OCSP stapling

## Requirements

**For running svp (after installation):**
- **Debian**: Version 12 (Bookworm) or 13 (Trixie)
- **Ubuntu**: 20.04 LTS, 22.04 LTS, or 24.04 LTS (recommended)
- Root access

**Note:** The tool automatically detects your OS and version, then configures appropriate package repositories. If PHP packages aren't available for your specific version yet, it will automatically use packages from the nearest stable release which are fully compatible.

**For building from source (development only):**
- Go 1.21+ (will be installed automatically by install.sh if not present)
- Git

## Installation

### Quick Install (Recommended)

**Install the latest pre-built binary from GitHub releases:**

```bash
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash
```

**Or download and review before running:**
```bash
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh -o install-svp.sh && less install-svp.sh && sudo bash install-svp.sh
```

This method:
- ✅ Downloads the latest stable release
- ✅ Verifies checksums automatically
- ✅ Works without Go installed
- ✅ Fastest installation method
- ✅ Recommended for production use

### Manual Installation from GitHub Releases

**If you prefer to manually download and install:**

```bash
# Download and install (one command) - get latest version from releases page
VERSION="1.0.30" && wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp-linux-amd64 && wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/checksums.txt && sha256sum --check --ignore-missing checksums.txt && chmod +x svp-linux-amd64 && sudo mv svp-linux-amd64 /usr/local/bin/svp && svp -version
```

<details>
<summary>Step-by-step instructions</summary>

1. **Download the binary for your system**:
   ```bash
   # Get the latest version number from https://github.com/willjackson/simple-vps-provisioner/releases
   VERSION="1.0.30"  # Replace with latest version
   
   # For Linux AMD64 (most common)
   wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp-linux-amd64

   # For Linux ARM64
   wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp-linux-arm64
   ```

2. **Verify checksum** (recommended):
   ```bash
   # Download checksums file
   wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/checksums.txt
   
   # Verify the download
   sha256sum --check --ignore-missing checksums.txt
   ```

3. **Install**:
   ```bash
   # Make executable and move to system path
   chmod +x svp-linux-amd64
   sudo mv svp-linux-amd64 /usr/local/bin/svp
   
   # Verify installation
   svp -version
   ```

</details>

### Building from Source (Development Only)

> **Note**: Building from source is for development purposes. Production use should install pre-built binaries (see Quick Install above).

**Prerequisites:**
- Go 1.21 or higher
- Git
- Root access

**Option 1: Using install.sh (Automatic)**

```bash
# Clone, checkout, and install (one command)
git clone https://github.com/willjackson/simple-vps-provisioner.git && cd simple-vps-provisioner && git checkout v1.0.24 && sudo bash install.sh
```

**Or current development version:**
```bash
git clone https://github.com/willjackson/simple-vps-provisioner.git && cd simple-vps-provisioner && sudo bash install.sh
```

<details>
<summary>What install.sh does</summary>

The `install.sh` script will:
- Install Go if not present
- Detect version from git tags automatically  
- Build with version injection
- Install to `/usr/local/bin/svp`
</details>

**Option 2: Manual Build**

```bash
# Clone, build with version, and install (one command)
git clone https://github.com/willjackson/simple-vps-provisioner.git && cd simple-vps-provisioner && VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev") && VERSION=${VERSION#v} && go build -ldflags="-X main.version=${VERSION}" -o svp && sudo mv svp /usr/local/bin/ && sudo chmod +x /usr/local/bin/svp && svp -version
```

For detailed deployment and release instructions, see [DEPLOY.md](DEPLOY.md).

## Usage

### Basic Commands

#### Install Drupal

```bash
# Fresh Drupal installation
sudo svp -mode setup -cms drupal -domain example.com

# Deploy from Git repository
sudo svp -mode setup -cms drupal \
  -domain example.com \
  -git-repo https://github.com/yourorg/yoursite.git \
  -git-branch main
```

#### Install WordPress

```bash
# Fresh WordPress installation
sudo svp -mode setup -cms wordpress -domain example.com

# Deploy from Git repository
sudo svp -mode setup -cms wordpress \
  -domain example.com \
  -git-repo https://github.com/yourorg/wpsite.git \
  -git-branch main
```

#### Multiple Domains

```bash
# Install Drupal for multiple domains
sudo svp -mode setup -cms drupal \
  -domain example.com \
  -extra-domains "staging.example.com,dev.example.com"
```

#### Verify Configuration

```bash
# Check system configuration without making changes
sudo svp -mode verify
```

#### Check Version

```bash
# Display the installed version
svp -version
```

#### Update SVP

```bash
# Check for updates and install the latest version
sudo svp -mode update
```

#### Update PHP Version

```bash
# Update PHP version for a specific domain
sudo svp -mode php-update -domain example.com -php-version 8.4
```

#### Reprovisioning Sites

When running setup for an existing domain, svp will prompt you and then reprovision the site. By default, the database and user are completely dropped and recreated with fresh credentials for security.

```bash
# Reprovision with fresh database (default)
sudo svp -mode setup -cms drupal -domain example.com
```

This will:
- Drop the existing database and database user
- Create a new database with **new credentials**
- Run a fresh Drupal site-install
- All files and code are replaced

**Keep existing database (optional):**

```bash
# Reprovision but keep same database credentials
sudo svp -mode setup -cms drupal -domain example.com -keep-existing-db
```

This will:
- Keep the existing database and user
- Drop all tables (clean slate)
- **Reuse the same credentials**
- Run a fresh Drupal site-install
- Useful for testing without changing credentials

**Note:** When importing a database with `-db /path/to/database.sql.gz`, the flag has no effect as the database is always preserved for the import.

### Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-mode` | Operation mode: `setup`, `verify`, `update`, `php-update` | `setup` |
| `-cms` | CMS to install: `drupal` or `wordpress` | `drupal` |
| `-domain` | Primary domain name (required for setup) | - |
| `-extra-domains` | Extra domains (comma-separated) | - |
| `-php-version` | PHP version to install | `8.3` |
| `-webroot` | Parent directory for sites | `/var/www` |
| `-git-repo` | Git repository URL | - |
| `-git-branch` | Git branch to checkout | `main` |
| `-drupal-root` | Drupal root path (relative to repo) | - |
| `-docroot` | Custom docroot path | - |
| `-db-engine` | Database engine: `mariadb` or `none` | `mariadb` |
| `-db` | Path to database file for import (instead of site-install) | - |
| `-keep-existing-db` | Keep existing database and drop tables (default: recreate database) | `false` |
| `-create-swap` | Create swap: `yes`, `no`, or `auto` | `auto` |
| `-firewall` | Enable UFW firewall | `true` |
| `-ssl` | Enable SSL/HTTPS with Let's Encrypt | `true` |
| `-le-email` | Let's Encrypt email (required for SSL) | - |
| `-debug` | Enable debug mode | `false` |

## What Gets Installed

### System Components

1. **Base packages**: curl, wget, git, unzip, etc.
2. **Nginx**: Web server
3. **PHP-FPM**: PHP 8.3 (or specified version) with extensions
4. **MariaDB**: Database server
5. **Composer**: PHP dependency manager
6. **Certbot**: Let's Encrypt SSL certificates (if `-le-email` provided)
7. **WP-CLI**: WordPress CLI (for WordPress installations)
8. **UFW Firewall**: Configured for HTTP, HTTPS, SSH

### Per-Domain Configuration

For each domain, the tool creates:

1. **Project directory**: `/var/www/[domain]`
2. **PHP-FPM pool**: Isolated process pool
3. **Nginx vhost**: Server configuration
4. **Database**: Dedicated database and user
5. **Drush wrapper**: `drush-[domain]` command (Drupal only)

### Configuration Files

- Main config: `/etc/simple-host-manager/`
- Site configs: `/etc/simple-host-manager/sites/`
- Database credentials: `/etc/simple-host-manager/sites/[domain].db.txt`
- Nginx vhosts: `/etc/nginx/sites-available/`
- PHP-FPM pools: `/etc/php/[version]/fpm/pool.d/`

## Examples

### Example 1: Fresh Drupal Site

```bash
sudo svp -mode setup -cms drupal -domain mysite.com
```

This will:
- Install all LAMP components
- Create a new Drupal project using Composer
- Set up database and credentials
- Configure Nginx and PHP-FPM
- Enable firewall

After completion, visit `http://mysite.com/install.php` to complete Drupal installation.

### Example 2: Deploy Existing WordPress Site

```bash
sudo svp -mode setup -cms wordpress \
  -domain myblog.com \
  -git-repo https://github.com/myorg/myblog.git \
  -git-branch production
```

This will:
- Clone the WordPress repository
- Install WP-CLI
- Create database
- Generate wp-config.php
- Configure Nginx and PHP-FPM

### Example 3: Multiple Drupal Environments

```bash
sudo svp -mode setup -cms drupal \
  -domain myapp.com \
  -extra-domains "staging.myapp.com,dev.myapp.com" \
  -git-repo https://github.com/myorg/myapp.git
```

This creates three separate environments, each with:
- Isolated project directory
- Dedicated database
- Separate PHP-FPM pool
- Individual Nginx vhost

### Example 4: Custom PHP Version

```bash
sudo svp -mode setup -cms drupal \
  -domain example.com \
  -php-version 8.4
```

### Example 5: Install with SSL/HTTPS

```bash
# Drupal with automatic SSL certificate
sudo svp -mode setup -cms drupal \
  -domain mysite.com \
  -le-email admin@mysite.com
```

This will:
- **Verify DNS points to server** (interactive check)
- Install and configure certbot
- Obtain Let's Encrypt SSL certificate
- Configure nginx with HTTPS (port 443)
- Redirect HTTP to HTTPS automatically
- Enable HSTS, OCSP stapling, and TLS 1.2/1.3
- Setup automatic certificate renewal

**DNS Verification:** Before obtaining SSL certificates, the tool automatically verifies that your domain's DNS points to the server. If DNS is not configured or points elsewhere, you'll be prompted to:
1. Check DNS again (after updating records)
2. Continue without HTTPS (HTTP only)
3. Abort setup

After completion, visit `https://mysite.com` (note the **https**)

### Example 6: Multiple Domains with SSL

```bash
sudo svp -mode setup -cms wordpress \
  -domain example.com \
  -extra-domains "www.example.com,blog.example.com" \
  -le-email webmaster@example.com
```

Each domain gets its own SSL certificate.

### Example 7: Install without SSL

```bash
# Disable SSL (HTTP only)
sudo svp -mode setup -cms drupal \
  -domain mysite.com \
  -ssl=false
```

### Example 8: Update PHP Version

```bash
# Update a domain from PHP 8.3 to PHP 8.4
sudo svp -mode php-update -domain example.com -php-version 8.4
```

This will:
- Verify the domain configuration exists
- Install PHP 8.4 if not already installed
- Create a new PHP-FPM pool with PHP 8.4
- Update Nginx configuration to use PHP 8.4
- Update the site configuration file
- Reload Nginx with the new configuration
- Remove the old PHP-FPM pool

**Note:** The tool will prompt for confirmation before making changes and show the current and new PHP versions.

### Example 9: Reprovision with Fresh Database

```bash
# Reprovision example.com with completely fresh database
sudo svp -mode setup -cms drupal -domain example.com
```

This will:
- Prompt to delete and reprovision (y/N)
- Remove the existing site directory
- **Drop the old database and user completely**
- **Create new database with fresh credentials**
- Fresh Drupal installation
- New settings.php with new database credentials

**Use case:** Complete clean slate, new credentials for security

### Example 10: Reprovision Keeping Database Credentials

```bash
# Reprovision but keep the same database credentials
sudo svp -mode setup -cms drupal -domain example.com -keep-existing-db
```

This will:
- Prompt to delete and reprovision (y/N)
- Remove the existing site directory
- **Keep the existing database and user**
- **Drop all tables** (clean database)
- **Reuse the same credentials**
- Fresh Drupal installation
- Same database credentials in settings.php

**Use case:** Testing/development where you need to maintain the same database credentials across reprovisioning

### Example 11: Ubuntu Server Setup

```bash
# Works on Ubuntu 22.04 LTS, 24.04 LTS, etc.
sudo svp -mode setup -cms drupal \
  -domain myubuntu-site.com \
  -le-email admin@myubuntu-site.com
```

The tool automatically:
- Detects Ubuntu OS and version
- Configures appropriate Ubuntu PHP repositories
- Falls back to nearest LTS if needed
- Everything else works identically to Debian

**Output example:**
```
[CREATE] Adding Sury PHP repository...
[CREATE] Detected Ubuntu jammy
[✓] Sury PHP repository added
[CREATE] Installing PHP 8.3 packages...
[✓] PHP 8.3 packages installed
```

**On Ubuntu 24.04 (Noble):**
```
[CREATE] Detected Ubuntu noble
[!] Ubuntu noble not yet supported by Sury, using jammy repository
[✓] Sury PHP repository added
```

## CMS-Specific Information

### Drupal

**Post-installation:**
- Visit `http://[domain]/install.php` to complete setup
- Or use Drush: `drush-[domain] site:install`
- Database credentials are in `/etc/simple-host-manager/sites/[domain].db.txt`

**Directory structure:**
```
/var/www/[domain]/
├── composer.json
├── vendor/
├── web/
│   ├── index.php
│   ├── sites/default/settings.php
│   └── ...
└── ...
```

**Drush usage:**
```bash
# Each domain gets its own Drush wrapper
drush-example.com status
drush-example.com cache:rebuild
drush-example.com user:login
```

### WordPress

**Post-installation:**
- Visit `http://[domain]/wp-admin/install.php` to complete setup
- Database credentials are in `/etc/simple-host-manager/sites/[domain].db.txt`

**Directory structure:**
```
/var/www/[domain]/
├── wp-config.php
├── wp-content/
├── wp-admin/
├── wp-includes/
└── ...
```

**WP-CLI usage:**
```bash
# Navigate to site directory
cd /var/www/example.com
sudo -u admin wp plugin list
sudo -u admin wp theme activate twentytwentyfour
```

## Troubleshooting

### Check Nginx status
```bash
sudo systemctl status nginx
sudo nginx -t  # Test configuration
```

### Check PHP-FPM status
```bash
sudo systemctl status php8.3-fpm
```

### Check MariaDB status
```bash
sudo systemctl status mariadb
```

### View logs
```bash
# Nginx logs
tail -f /var/log/nginx/[domain]-error.log

# PHP-FPM logs
tail -f /var/log/php8.3-fpm-[domain]-error.log

# Certbot logs
tail -f /var/log/letsencrypt/letsencrypt.log
```

### Verify configuration
```bash
sudo svp -mode verify
```

### SSL Certificate Issues

#### Check certificate status
```bash
# List all certificates
sudo certbot certificates

# Check specific domain
sudo certbot certificates -d example.com
```

#### Test certificate renewal
```bash
# Dry run (test without actually renewing)
sudo certbot renew --dry-run

# Force renewal
sudo certbot renew --force-renewal
```

#### DNS Verification

Before obtaining SSL certificates, svp automatically verifies that your domain's DNS points to the server's public IP.

**What it checks:**
1. Gets the server's public IP address
2. Resolves the domain's DNS records
3. Compares the two IP addresses

**If DNS doesn't match:**
You'll be prompted with three options:
- **Check DNS again** - Wait for DNS propagation (5-30 minutes) and recheck
- **Continue without HTTPS** - Skip SSL and use HTTP only
- **Abort setup** - Stop the installation

**Manual DNS check:**
```bash
# Get your server's public IP
wget -qO- https://ipinfo.io/ip

# Check what IP the domain resolves to
dig +short example.com

# Both should match for SSL to work
```

#### Certificate not obtained
**Common causes:**
- Domain doesn't point to server IP yet (DNS propagation)
- Firewall blocking port 80 or 443
- No email provided (`-le-email` flag)
- Rate limit reached (5 certificates per week for same domain)

**Check nginx configuration:**
```bash
sudo nginx -t
sudo systemctl status nginx
```

#### Renew expired certificate
```bash
sudo certbot renew
sudo systemctl reload nginx
```

#### Remove certificate
```bash
sudo certbot delete --cert-name example.com
```

## DNS Requirements for SSL

When using Let's Encrypt SSL certificates (`-le-email` flag), your domain **must** point to the server before certificate issuance.

**Setup Steps:**
1. **Point your domain to the server** - Update your DNS A record:
   ```
   A    @           123.45.67.89  (server IP)
   A    www         123.45.67.89  (if using www)
   ```

2. **Wait for DNS propagation** (typically 5-30 minutes)

3. **Run svp** - The tool will verify DNS automatically:
   ```bash
   sudo svp -mode setup -cms drupal \
     -domain example.com \
     -le-email admin@example.com
   ```

4. **DNS Verification** - If DNS isn't configured:
   - The tool displays your server IP
   - Shows what IP the domain currently points to
   - Prompts you to update DNS and check again

**Note:** Let's Encrypt validates domain ownership via HTTP challenge. If DNS doesn't point to your server, certificate issuance will fail.

## Security Notes

1. **DNS Verification** - Automatic check before SSL certificate issuance
2. **Database credentials** are automatically generated and saved securely
3. **Fresh credentials on reprovision** - New database credentials created by default when reprovisioning (use `-keep-existing-db` to preserve)
4. **Firewall** is enabled by default (SSH, HTTP, HTTPS)
5. **SSL/HTTPS** - Free Let's Encrypt certificates with automatic renewal
6. **HSTS** - HTTP Strict Transport Security enabled
7. **TLS 1.2/1.3** - Modern encryption protocols only
8. **OCSP Stapling** - Improved SSL performance and privacy
9. **PHP settings** are hardened (disabled dangerous functions, etc.)
10. **File permissions** are set appropriately (admin:www-data)
11. **Per-site isolation** via separate PHP-FPM pools

## Supported Operating Systems

This tool works on both Debian and Ubuntu:

### Debian Support
- **Debian 12 (Bookworm)** - Fully supported ✅
- **Debian 13 (Trixie)** - Fully supported ✅  
- **Debian 11 (Bullseye)** - Supported ✅
- **Debian Testing/Unstable** - Automatically falls back to Debian 12 packages

### Ubuntu Support
- **Ubuntu 22.04 LTS (Jammy)** - Fully supported ✅ (Most stable)
- **Ubuntu 20.04 LTS (Focal)** - Fully supported ✅
- **Ubuntu 24.04 LTS (Noble)** - Supported ✅ (uses 22.04 packages)
- **Ubuntu 18.04 LTS (Bionic)** - Supported ✅
- **Ubuntu interim releases** - Automatically falls back to 22.04 LTS

**Note:** Ubuntu 24.04 (Noble) is very new and Sury doesn't have dedicated packages yet, so it automatically uses Ubuntu 22.04 (Jammy) packages which are fully compatible.

### Automatic Version Detection

The tool automatically:
1. Detects your OS (Debian or Ubuntu)
2. Identifies your version codename
3. Maps unsupported versions to the nearest stable release
4. Configures PHP repositories accordingly

For example:
- Debian 13 (Trixie) → Uses Debian 12 (Bookworm) PHP packages
- Ubuntu 24.04 (Noble) → Uses Ubuntu 22.04 LTS (Jammy) PHP packages
- Ubuntu 23.10 (Mantic) → Uses Ubuntu 22.04 LTS (Jammy) PHP packages

This ensures compatibility even on cutting-edge or unsupported releases.

## Adaptability to Other Distributions

While this tool is optimized for Debian and Ubuntu, it can be adapted for other distributions by modifying:

1. **Package names** - `pkg/system/packages.go`
2. **Service names** - `pkg/system/services.go`
3. **File paths** - Various package files
4. **Package manager commands** - `apt-get` → `yum`, `dnf`, etc.

## Version Management

The tool uses build-time version injection to ensure consistency between installed binaries and GitHub releases.

**Check your version:**
```bash
svp -version
```

**Update to latest:**
```bash
sudo svp -mode update
```

For detailed information about versioning, building, and maintaining version consistency, see [VERSIONING.md](VERSIONING.md).

## Contributing

This tool is based on the original Bash script `setup.sh`. Contributions are welcome!

## License

MIT License

## Support

For issues and questions, please refer to the documentation or create an issue in the project repository.
