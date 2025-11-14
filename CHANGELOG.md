# Changelog

All notable changes to Simple VPS Provisioner (svp) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **BREAKING: New command structure** - Commands are now positional arguments instead of flags. Use `svp setup` instead of `svp -mode setup`. Running `svp` with no arguments now shows help instead of attempting setup
- **Context-sensitive help** - Each command (`setup`, `verify`, `update`, `php-update`) now has its own help text with relevant options only. Use `svp setup -help` to see setup-specific options
- **Git branch behavior** - The `-git-branch` flag no longer defaults to "main". When not specified, the repository's default branch is used automatically
- **Simplified help output** - Main help now shows only 2-3 examples with reference to full documentation for complete examples

### Fixed
- **CRITICAL: SSL certificate failures no longer break sites** - Completely redesigned SSL certificate obtainment to use two-phase approach: Phase 1 obtains certificate WITHOUT modifying nginx (using `certbot certonly`), Phase 2 configures nginx only if certificate was successfully obtained. This prevents sites from becoming inaccessible when certificate obtainment fails (e.g., rate limits, network issues) because nginx configuration is never modified unless we have a valid certificate
- **PHP update now preserves SSL configuration** - `php-update` mode now automatically reconfigures SSL/HTTPS after updating Nginx vhost, preventing sites from becoming HTTP-only after PHP version changes
- **PHP-FPM pool creation now always restarts service** - Fixed issue where socket files weren't created when pool configuration already existed, causing connection refused errors
- **Socket verification after pool creation** - PHP pool creation now verifies the socket file was created successfully and fails with clear error message if not

### Changed
- **Improved PHP pool creation reliability** - Pool configuration files are now always written and PHP-FPM is always restarted to ensure sockets are created, even when updating existing pools
- **Better error messages for PHP-FPM issues** - Clearer guidance when socket creation fails, directing users to check PHP-FPM logs

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
- **Fresh database on reprovisioning (default)** - By default, drops entire database and user completely when reprovisioning, creates fresh database with new credentials for better security
- **Optional database preservation** - New `-keep-existing-db` flag to preserve existing database and credentials when reprovisioning (drops tables only)
- `DropDatabase()` function in `pkg/database/mariadb.go` to completely remove database and user
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
- **Database**: MariaDB with automatic database creation; fresh credentials by default on reprovisioning
- **Security**: Firewall, PHP hardening, secure credentials, DNS verification for SSL, fresh database credentials on reprovisioning
- **SSL/HTTPS**: Let's Encrypt certificates with automatic DNS verification
- **Isolation**: Per-domain PHP-FPM pools for better security and resource management
- **Flexibility**: Support for Git deployments and custom configurations
- **Clean Reprovisioning**: Fresh database and credentials by default; optional flag to preserve database

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
