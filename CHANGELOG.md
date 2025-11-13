# Changelog

All notable changes to Simple VPS Provisioner (svp) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of Simple VPS Provisioner (svp)
- Support for Drupal provisioning with Composer and Drush
- Support for WordPress provisioning with WP-CLI
- Debian 13 (Trixie) support
- Multi-domain provisioning capability
- Git repository deployment
- Per-domain PHP-FPM pools
- Nginx virtual host configuration
- MariaDB database setup with auto-generated credentials
- UFW firewall configuration
- Swap management (auto/yes/no)
- System verification mode
- Comprehensive CLI with flag-based configuration
- **DNS verification before SSL certificate issuance** - Automatically checks if domain DNS points to server before attempting Let's Encrypt certificate
- Interactive DNS check with options to retry, continue without HTTPS, or abort
- Multiple fallback methods for IP and DNS resolution (dig, nslookup, host)
- **Database reuse on reprovisioning** - Reuses existing database credentials and drops tables with drush sql-drop instead of creating new databases
- Automatic table dropping using drush sql-drop with SQL fallback
- Existing database credential detection and reuse
- **PHP version update mode** - Update PHP version for existing domains with `php-update` mode
- Automatic PHP-FPM pool migration when updating PHP versions
- Seamless PHP version switching with zero-downtime updates
- **Build-time version injection** - Ensures consistency between installed binary and GitHub releases
- `-version` flag to display installed version
- Automatic version detection from git tags during builds
- Comprehensive version management documentation (VERSIONING.md)
- **install-from-github.sh** - Quick installer for pre-built binaries from GitHub releases
- Automatic checksum verification during installation
- Improved install.sh with better documentation and version detection
- Comprehensive installation testing guide (INSTALLATION_TESTING.md)
- Updated README with accurate installation instructions for all methods

### Features
- **CMS Support**: Drupal and WordPress
- **Web Stack**: Nginx + PHP-FPM 8.3 (configurable)
- **Database**: MariaDB with automatic database creation and reuse on reprovisioning
- **Security**: Firewall, PHP hardening, secure credentials, DNS verification for SSL
- **SSL/HTTPS**: Let's Encrypt certificates with automatic DNS verification
- **Isolation**: Per-domain PHP-FPM pools for better security and resource management
- **Flexibility**: Support for Git deployments and custom configurations
- **Smart Reprovisioning**: Reuses databases and drops tables cleanly with drush

## [1.0.0] - YYYY-MM-DD

### Added
- First stable release
- Complete LAMP stack provisioning
- Multi-CMS support (Drupal, WordPress)
- Automated deployment from Git repositories
- Security hardening and best practices
- Comprehensive documentation

[Unreleased]: https://github.com/willjackson/simple-vps-provisioner/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/willjackson/simple-vps-provisioner/releases/tag/v1.0.0
