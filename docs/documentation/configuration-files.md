---
layout: default
title: Configuration Files
---

# Configuration Files

Reference guide for all svp configuration files and their locations.

## Overview

svp manages configuration files in several locations:

- `/etc/svp/` - svp-specific configuration
- `/etc/nginx/` - Web server configuration
- `/etc/php/*/fpm/` - PHP-FPM configuration
- Site-specific settings in each site directory

---

## svp Configuration Files

### Site Configuration Directory

**Location:** `/etc/svp/`

```
/etc/svp/
├── php.conf              # Current PHP version
└── sites/                # Per-site configurations
    ├── example.com.conf  # Site config
    └── example.com.db.txt # Database credentials
```

### Site Config File

**Location:** `/etc/svp/sites/example.com.conf`

**Contents:**
```bash
# Site configuration for example.com
DOMAIN='example.com'
PHP_VERSION='8.3'
WEBROOT='/var/www/example.com/web'
CREATED='Mon Jan 15 10:30:00 UTC 2024'
```

### Database Credentials

**Location:** `/etc/svp/sites/example.com.db.txt`

**Contents:**
```
Database: drupal_example_com
Username: drupal_example_com
Password: [auto-generated-secure-password]
Host: localhost
Port: 3306
```

**Permissions:** `600` (read/write owner only)

```bash
# View credentials
sudo cat /etc/svp/sites/example.com.db.txt

# Proper permissions
sudo chmod 600 /etc/svp/sites/example.com.db.txt
```

---

## Nginx Configuration

### Main Nginx Config

**Location:** `/etc/nginx/nginx.conf`

Key settings svp doesn't modify (uses defaults):
```nginx
user www-data;
worker_processes auto;
pid /run/nginx.pid;
```

### Site Virtual Hosts

**Location:** `/etc/nginx/sites-available/example.com.conf`

**Symlinked to:** `/etc/nginx/sites-enabled/example.com.conf`

**Example structure:**
```nginx
# HTTP to HTTPS redirect
server {
    listen 80;
    listen [::]:80;
    server_name example.com;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name example.com;

    root /var/www/example.com/web;
    index index.php index.html;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;

    # PHP-FPM
    location ~ \.php$ {
        fastcgi_pass unix:/run/php/php8.3-fpm-example.com.sock;
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }

    # Additional Drupal/WordPress specific configs...
}
```

### Edit Nginx Config

```bash
# Edit site config
sudo nano /etc/nginx/sites-available/example.com.conf

# Test configuration
sudo nginx -t

# Reload if test passes
sudo systemctl reload nginx
```

---

## PHP-FPM Configuration

### PHP-FPM Pool Config

**Location:** `/etc/php/8.3/fpm/pool.d/example.com.conf`

**Contents:**
```ini
[example.com]
user = admin
group = www-data

listen = /run/php/php8.3-fpm-example.com.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = dynamic
pm.max_children = 5
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3

pm.max_requests = 500

chdir = /

php_admin_value[error_log] = /var/log/php8.3-fpm-example.com-error.log
php_admin_flag[log_errors] = on
```

### PHP INI Settings

**Main PHP config:** `/etc/php/8.3/fpm/php.ini`

**Per-pool overrides** in pool config:
```ini
php_admin_value[memory_limit] = 256M
php_admin_value[upload_max_filesize] = 64M
php_admin_value[post_max_size] = 64M
php_admin_value[max_execution_time] = 300
php_admin_flag[display_errors] = off
```

---

## CMS Configuration Files

### Drupal Settings

**Main settings:** `/var/www/example.com/web/sites/default/settings.php`

**svp-managed settings:** `/var/www/example.com/web/sites/default/settings.svp.php`

**Contents of settings.svp.php:**
```php
<?php
// Managed by Simple VPS Provisioner
// Do not edit manually - changes will be overwritten

$databases['default']['default'] = [
  'database' => 'drupal_example_com',
  'username' => 'drupal_example_com',
  'password' => '[password-from-db-txt]',
  'host' => 'localhost',
  'port' => '3306',
  'driver' => 'mysql',
  'prefix' => '',
  'collation' => 'utf8mb4_general_ci',
];

$settings['file_private_path'] = '/var/www/example.com/private';
$settings['config_sync_directory'] = '../config/sync';
```

**Main settings.php includes:**
```php
// Include svp-managed settings
if (file_exists($app_root . '/' . $site_path . '/settings.svp.php')) {
  include $app_root . '/' . $site_path . '/settings.svp.php';
}
```

### WordPress Configuration

**Location:** `/var/www/myblog.com/wp-config.php`

**Database settings:**
```php
define('DB_NAME', 'wordpress_myblog_com');
define('DB_USER', 'wordpress_myblog_com');
define('DB_PASSWORD', '[password-from-db-txt]');
define('DB_HOST', 'localhost');
define('DB_CHARSET', 'utf8mb4');
define('DB_COLLATE', '');
```

---

## SSL Configuration

### Let's Encrypt Certificates

**Location:** `/etc/letsencrypt/live/example.com/`

```
/etc/letsencrypt/live/example.com/
├── fullchain.pem   # Certificate + intermediate
├── privkey.pem     # Private key
├── cert.pem        # Certificate only
└── chain.pem       # Intermediate certificate
```

### Certificate Renewal Config

**Location:** `/etc/letsencrypt/renewal/example.com.conf`

**Contents:**
```ini
version = 1.1.0
archive_dir = /etc/letsencrypt/archive/example.com
cert = /etc/letsencrypt/live/example.com/cert.pem
privkey = /etc/letsencrypt/live/example.com/privkey.pem
chain = /etc/letsencrypt/live/example.com/chain.pem
fullchain = /etc/letsencrypt/live/example.com/fullchain.pem

[renewalparams]
authenticator = nginx
installer = nginx
account = [account-hash]
```

---

## Firewall Configuration

### UFW Rules

**View rules:**
```bash
sudo ufw status verbose
```

**Rules file:** `/etc/ufw/user.rules`

**Default rules set by svp:**
```
# Allow SSH
-A ufw-user-input -p tcp --dport 22 -j ACCEPT

# Allow HTTP
-A ufw-user-input -p tcp --dport 80 -j ACCEPT

# Allow HTTPS
-A ufw-user-input -p tcp --dport 443 -j ACCEPT
```

---

## Log Files

### Nginx Logs

**Main logs:**
```
/var/log/nginx/access.log
/var/log/nginx/error.log
```

**Per-site logs:**
```
/var/log/nginx/example.com.access.log
/var/log/nginx/example.com.error.log
```

### PHP-FPM Logs

**Per-domain logs:**
```
/var/log/php8.3-fpm-example.com-error.log
```

**Main PHP-FPM log:**
```
/var/log/php8.3-fpm.log
```

### MariaDB Logs

**Error log:**
```
/var/log/mysql/error.log
```

**Slow query log (if enabled):**
```
/var/log/mysql/slow.log
```

---

## Configuration Templates

### Nginx Drupal Template

Stored in svp source - used during provisioning.

**Key sections:**
- Document root configuration
- Clean URLs support
- PHP-FPM integration
- Security headers
- Asset caching
- Access restrictions

### Nginx WordPress Template

Similar structure but WordPress-specific:
- Permalink support
- wp-admin access
- PHP-FPM integration
- Asset caching
- Security restrictions

---

## File Permissions

### Recommended Permissions

**Site directories:**
```bash
/var/www/example.com/
Owner: admin:www-data
Directories: 755
Files: 644
```

**Drupal files directory:**
```bash
/var/www/example.com/web/sites/default/files/
Owner: admin:www-data
Directories: 775
Files: 664
```

**WordPress uploads:**
```bash
/var/www/myblog.com/wp-content/uploads/
Owner: admin:www-data
Directories: 775
Files: 664
```

**Configuration files:**
```bash
settings.php: 444 (read-only)
settings.svp.php: 400 (owner read-only)
wp-config.php: 400 (owner read-only)
```

**Database credentials:**
```bash
/etc/svp/sites/example.com.db.txt: 600 (owner only)
```

### Fix Permissions

**Drupal:**
```bash
cd /var/www/example.com
sudo chown -R admin:www-data .
sudo find . -type d -exec chmod 755 {} \;
sudo find . -type f -exec chmod 644 {} \;
sudo chmod 775 web/sites/default/files/
sudo chmod 444 web/sites/default/settings.php
sudo chmod 400 web/sites/default/settings.svp.php
```

**WordPress:**
```bash
cd /var/www/myblog.com
sudo chown -R admin:www-data .
sudo find . -type d -exec chmod 755 {} \;
sudo find . -type f -exec chmod 644 {} \;
sudo chmod 775 wp-content/uploads/
sudo chmod 400 wp-config.php
```

---

## Editing Configuration

### Safe Edit Procedure

1. **Backup original:**
   ```bash
   sudo cp /etc/nginx/sites-available/example.com.conf \
          /etc/nginx/sites-available/example.com.conf.backup
   ```

2. **Edit file:**
   ```bash
   sudo nano /etc/nginx/sites-available/example.com.conf
   ```

3. **Test configuration:**
   ```bash
   sudo nginx -t
   ```

4. **Apply if test passes:**
   ```bash
   sudo systemctl reload nginx
   ```

5. **Rollback if needed:**
   ```bash
   sudo cp /etc/nginx/sites-available/example.com.conf.backup \
          /etc/nginx/sites-available/example.com.conf
   sudo systemctl reload nginx
   ```

---

## Configuration Backup

### Backup All Configs

```bash
#!/bin/bash
BACKUP_DIR="/backups/config-$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# svp configs
sudo cp -r /etc/svp/ $BACKUP_DIR/

# Nginx configs
sudo cp -r /etc/nginx/sites-available/ $BACKUP_DIR/nginx/

# PHP-FPM pools
sudo cp -r /etc/php/8.3/fpm/pool.d/ $BACKUP_DIR/php-fpm/

# SSL certificates
sudo cp -r /etc/letsencrypt/ $BACKUP_DIR/letsencrypt/

echo "Backup complete: $BACKUP_DIR"
```

---

[← PHP Version Management](php-versions) | [Reprovisioning →](reprovisioning)
