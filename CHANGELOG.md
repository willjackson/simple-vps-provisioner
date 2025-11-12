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

### Features
- **CMS Support**: Drupal and WordPress
- **Web Stack**: Nginx + PHP-FPM 8.3 (configurable)
- **Database**: MariaDB with automatic database creation
- **Security**: Firewall, PHP hardening, secure credentials
- **Isolation**: Per-domain PHP-FPM pools for better security and resource management
- **Flexibility**: Support for Git deployments and custom configurations

## [1.0.0] - YYYY-MM-DD

### Added
- First stable release
- Complete LAMP stack provisioning
- Multi-CMS support (Drupal, WordPress)
- Automated deployment from Git repositories
- Security hardening and best practices
- Comprehensive documentation

[Unreleased]: https://github.com/YOURORG/YOURREPO/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/YOURORG/YOURREPO/releases/tag/v1.0.0
