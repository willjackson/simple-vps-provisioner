# Simple VPS Provisioner (svp)

A Go-based command-line tool for provisioning Debian 13 (Trixie) VPS with LAMP stack for Drupal or WordPress.

## Features

- **Automated VPS provisioning** - Complete setup with one command
- **Multi-CMS support** - Choose between Drupal or WordPress
- **LAMP stack** - Nginx, PHP-FPM, MariaDB, Composer
- **Multi-domain support** - Provision multiple domains in one run
- **Per-domain PHP-FPM pools** - Isolated PHP processes for each site
- **Git deployment** - Deploy from existing repositories
- **SSL/HTTPS support** - Automatic Let's Encrypt certificates with certbot
- **Idempotent** - Safe to run multiple times
- **Security hardening** - Firewall, PHP settings, database security, HSTS, OCSP stapling

## Requirements

- Debian 13 (Trixie) VPS
- Root access
- Go 1.21+ (for building)

## Installation

### Quick Install (Recommended)

Download and install the latest release from GitHub:

```bash
# Download the quick installer
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh -o install-svp.sh

# Run the installer
sudo bash install-svp.sh

# Verify installation
svp --help
```

### Manual Installation from GitHub Releases

1. **Download the binary for your system**:
   ```bash
   # For Linux AMD64 (most common)
   VERSION=1.0.0
   wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp-linux-amd64

   # For Linux ARM64
   wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/svp-linux-arm64
   ```

2. **Verify checksum** (recommended):
   ```bash
   wget https://github.com/willjackson/simple-vps-provisioner/releases/download/v${VERSION}/checksums.txt
   sha256sum --check --ignore-missing checksums.txt
   ```

3. **Install**:
   ```bash
   sudo mv svp-linux-amd64 /usr/local/bin/svp
   sudo chmod +x /usr/local/bin/svp
   ```

### Building from Source (Development Only)

> **Note**: Building from source is intended for development purposes. Production deployments should use pre-built binaries from GitHub Releases.

```bash
# Install Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone repository
git clone https://github.com/willjackson/simple-vps-provisioner.git
cd simple-vps-provisioner

# Build
go build -o svp

# Install
sudo mv svp /usr/local/bin/
sudo chmod +x /usr/local/bin/svp
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

### Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-mode` | Operation mode: `setup`, `verify` | `setup` |
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
- Install and configure certbot
- Obtain Let's Encrypt SSL certificate
- Configure nginx with HTTPS (port 443)
- Redirect HTTP to HTTPS automatically
- Enable HSTS, OCSP stapling, and TLS 1.2/1.3
- Setup automatic certificate renewal

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

#### Certificate not obtained
**Common causes:**
- Domain doesn't point to server IP yet (DNS propagation)
- Firewall blocking port 80 or 443
- No email provided (`-le-email` flag)
- Rate limit reached (5 certificates per week for same domain)

**Check DNS:**
```bash
dig +short example.com
# Should return your server IP
```

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

## Security Notes

1. **Database credentials** are automatically generated and saved securely
2. **Firewall** is enabled by default (SSH, HTTP, HTTPS)
3. **SSL/HTTPS** - Free Let's Encrypt certificates with automatic renewal
4. **HSTS** - HTTP Strict Transport Security enabled
5. **TLS 1.2/1.3** - Modern encryption protocols only
6. **OCSP Stapling** - Improved SSL performance and privacy
7. **PHP settings** are hardened (disabled dangerous functions, etc.)
8. **File permissions** are set appropriately (admin:www-data)
9. **Per-site isolation** via separate PHP-FPM pools

## Adaptability to Other Distributions

While this tool is designed for Debian 13, it can be adapted for other distributions by modifying:

1. **Package names** - `pkg/system/packages.go`
2. **Service names** - `pkg/system/services.go`
3. **File paths** - Various package files
4. **Package manager commands** - `apt-get` → `yum`, `dnf`, etc.

## Contributing

This tool is based on the original Bash script `setup.sh`. Contributions are welcome!

## License

MIT License

## Support

For issues and questions, please refer to the documentation or create an issue in the project repository.
