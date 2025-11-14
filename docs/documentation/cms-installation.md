---
layout: default
title: CMS Installation
---

# CMS Installation

Guide to installing and configuring Drupal and WordPress with svp.

## Drupal Installation

### Fresh Drupal Site

Install the latest Drupal with default settings:

```bash
sudo svp setup \
  --cms drupal \
  --domain example.com \
  --le-email admin@example.com
```

**What gets installed:**
- Drupal 10 (latest stable)
- Composer-based structure
- Standard installation profile
- Recommended modules

**Directory structure:**
```
/var/www/example.com/
├── composer.json
├── vendor/
├── web/
│   ├── index.php
│   ├── sites/default/
│   │   ├── settings.php
│   │   ├── settings.svp.php (svp-managed, gitignored)
│   │   └── files/
│   ├── core/
│   ├── modules/
│   └── themes/
└── config/sync/
```

### First Login

After installation completes, you'll receive a one-time login link:

```
Login to admin: https://example.com/user/reset/1/...
```

Use this link to set your admin password.

### Using Drush

Each domain gets a custom drush alias:

```bash
# Clear cache
drush-example.com cr

# Get login link
drush-example.com uli

# Check status
drush-example.com status

# Import configuration
drush-example.com cim

# Export configuration
drush-example.com cex
```

### Deploy Existing Drupal Site

Deploy from a Git repository:

```bash
sudo svp setup \
  --cms drupal \
  --domain example.com \
  --git-repo https://github.com/myorg/mysite.git \
  --git-branch main \
  --le-email admin@example.com
```

**Requirements for your repository:**
- Must have `composer.json` in root
- Drupal must be in `web/` directory (or specify with `-docroot`)
- Configuration should be in `config/sync/`

### Import Database

Restore from existing database dump:

```bash
sudo svp setup \
  --cms drupal \
  --domain example.com \
  --git-repo https://github.com/myorg/mysite.git \
  --db /path/to/backup.sql.gz \
  --le-email admin@example.com
```

**Supported formats:**
- `.sql` - Plain SQL
- `.sql.gz` - Gzip compressed

### Post-Installation Steps

#### 1. Sync Files

If you have existing files, sync them:

```bash
rsync -avz /backup/files/ /var/www/example.com/web/sites/default/files/
sudo chown -R admin:www-data /var/www/example.com/web/sites/default/files/
sudo chmod -R 775 /var/www/example.com/web/sites/default/files/
```

#### 2. Import Configuration

If you have exported configuration:

```bash
cd /var/www/example.com
drush-example.com cim
```

#### 3. Update Database

Run database updates:

```bash
drush-example.com updb
```

---

## WordPress Installation

### Fresh WordPress Site

Install latest WordPress:

```bash
sudo svp setup \
  --cms wordpress \
  --domain myblog.com \
  --le-email admin@myblog.com
```

**What gets installed:**
- Latest WordPress version
- WP-CLI tool
- Standard WordPress structure
- Ready for installation wizard

### Complete Installation

After provisioning, visit your site to complete the installation:

```
https://myblog.com/wp-admin/install.php
```

Fill in:
- Site Title
- Admin Username
- Admin Password
- Admin Email

### Using WP-CLI

WordPress sites have WP-CLI available:

```bash
cd /var/www/myblog.com

# Install plugin
sudo -u admin wp plugin install jetpack --activate

# Update WordPress
sudo -u admin wp core update

# List themes
sudo -u admin wp theme list

# Create user
sudo -u admin wp user create bob bob@example.com --role=editor
```

### Deploy Existing WordPress Site

Deploy from Git repository:

```bash
sudo svp setup \
  --cms wordpress \
  --domain myblog.com \
  --git-repo https://github.com/myorg/wp-site.git \
  --le-email admin@myblog.com
```

**Repository structure:**
```
wp-site/
├── wp-config.php (optional, will be created)
├── wp-content/
│   ├── themes/
│   ├── plugins/
│   └── uploads/
├── index.php
└── wp-*.php
```

### Import Database

Restore existing WordPress database:

```bash
sudo svp setup \
  --cms wordpress \
  --domain myblog.com \
  --db /path/to/wp-backup.sql.gz \
  --le-email admin@myblog.com
```

**After import, update URLs:**

```bash
cd /var/www/myblog.com
sudo -u admin wp search-replace 'http://oldsite.com' 'https://myblog.com'
```

---

## Database Credentials

Database credentials are stored securely:

```bash
# View credentials
sudo cat /etc/svp/sites/example.com.db.txt
```

**Output:**
```
Database: drupal_example_com
Username: drupal_example_com
Password: [secure-random-password]
Host: localhost
Port: 3306
```

These credentials are automatically configured in:
- **Drupal**: `web/sites/default/settings.svp.php`
- **WordPress**: `wp-config.php`

---

## File Permissions

Proper file permissions are set automatically:

**Drupal:**
```bash
# Files directory
/var/www/example.com/web/sites/default/files/
Owner: admin:www-data
Permissions: 775 (directories), 664 (files)

# Settings files
settings.php: 444 (read-only)
settings.svp.php: 400 (read-only, owner only)
```

**WordPress:**
```bash
# Uploads directory
/var/www/myblog.com/wp-content/uploads/
Owner: admin:www-data
Permissions: 775 (directories), 664 (files)

# Config file
wp-config.php: 400 (read-only, owner only)
```

---

## Common Tasks

### Update CMS

**Drupal:**
```bash
cd /var/www/example.com
sudo -u admin composer update drupal/core-recommended --with-dependencies
drush-example.com updb
drush-example.com cr
```

**WordPress:**
```bash
cd /var/www/myblog.com
sudo -u admin wp core update
sudo -u admin wp core update-db
```

### Install Modules/Plugins

**Drupal:**
```bash
cd /var/www/example.com
sudo -u admin composer require drupal/admin_toolbar
drush-example.com en admin_toolbar -y
drush-example.com cr
```

**WordPress:**
```bash
cd /var/www/myblog.com
sudo -u admin wp plugin install contact-form-7 --activate
```

### Backup Site

**Drupal:**
```bash
# Database
drush-example.com sql-dump --gzip > backup-$(date +%Y%m%d).sql.gz

# Files
tar -czf files-backup-$(date +%Y%m%d).tar.gz \
  /var/www/example.com/web/sites/default/files/
```

**WordPress:**
```bash
# Database
cd /var/www/myblog.com
sudo -u admin wp db export backup-$(date +%Y%m%d).sql
gzip backup-$(date +%Y%m%d).sql

# Files
tar -czf wp-content-backup-$(date +%Y%m%d).tar.gz \
  /var/www/myblog.com/wp-content/
```

---

## Troubleshooting

### Drupal Installation Fails

Check permissions:
```bash
cd /var/www/example.com
ls -la web/sites/default/
```

Should show `settings.php` and `settings.svp.php` with correct ownership.

### WordPress Can't Connect to Database

Verify credentials in wp-config.php:
```bash
cd /var/www/myblog.com
grep DB_ wp-config.php
```

Compare with credentials file:
```bash
sudo cat /etc/svp/sites/myblog.com.db.txt
```

### File Upload Issues

Check files directory permissions:
```bash
# Drupal
ls -la /var/www/example.com/web/sites/default/files/

# WordPress
ls -la /var/www/myblog.com/wp-content/uploads/
```

Fix if needed:
```bash
sudo chown -R admin:www-data /path/to/files/
sudo chmod -R 775 /path/to/files/
```

---

[← Back to Documentation]({{ site.baseurl }}/documentation) | [SSL Configuration →](ssl-configuration)
