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
curl -fsSL https://raw.githubusercontent.com/YOURORG/YOURREPO/main/install-from-github.sh -o install-svp.sh

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
   wget https://github.com/YOURORG/YOURREPO/releases/download/v${VERSION}/svp-linux-amd64

   # For Linux ARM64
   wget https://github.com/YOURORG/YOURREPO/releases/download/v${VERSION}/svp-linux-arm64
   ```

2. **Verify checksum** (recommended):
   ```bash
   wget https://github.com/YOURORG/YOURREPO/releases/download/v${VERSION}/checksums.txt
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
git clone https://github.com/YOURORG/YOURREPO.git
cd YOURREPO

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
| `-debug` | Enable debug mode | `false` |

## What Gets Installed

### System Components

1. **Base packages**: curl, wget, git, unzip, etc.
2. **Nginx**: Web server
3. **PHP-FPM**: PHP 8.3 (or specified version) with extensions
4. **MariaDB**: Database server
5. **Composer**: PHP dependency manager
6. **WP-CLI**: WordPress CLI (for WordPress installations)
7. **UFW Firewall**: Configured for HTTP, HTTPS, SSH

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
```

### Verify configuration
```bash
sudo svp -mode verify
```

## Security Notes

1. **Database credentials** are automatically generated and saved securely
2. **Firewall** is enabled by default (SSH, HTTP, HTTPS)
3. **PHP settings** are hardened (disabled dangerous functions, etc.)
4. **File permissions** are set appropriately (admin:www-data)
5. **Per-site isolation** via separate PHP-FPM pools

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
