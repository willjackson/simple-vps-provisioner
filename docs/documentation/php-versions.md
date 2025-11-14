---
layout: default
title: PHP Version Management
---

# PHP Version Management

Manage PHP versions for your sites with svp.

## Overview

svp supports multiple PHP versions with:

- ✅ PHP 8.4, 8.3, 8.2, 8.1
- ✅ Per-domain PHP versions
- ✅ Easy version switching
- ✅ Isolated PHP-FPM pools
- ✅ Automatic configuration

---

## Available PHP Versions

| Version | Status | Recommended For |
|---------|--------|-----------------|
| PHP 8.4 | Latest | New projects, testing |
| PHP 8.3 | **Recommended** | Production sites |
| PHP 8.2 | Stable | Legacy compatibility |
| PHP 8.1 | Stable | Older projects |

**Default:** PHP 8.3 is used if no version specified.

---

## Specify PHP Version

### During Initial Setup

```bash
# Use default PHP 8.3
sudo svp setup --cms drupal --domain example.com

# Specify PHP 8.4
sudo svp setup --cms drupal --domain example.com --php-version 8.4

# Specify PHP 8.2
sudo svp setup --cms drupal --domain example.com --php-version 8.2
```

### Different Versions Per Site

Each site can use a different PHP version:

```bash
# Site 1 with PHP 8.3
sudo svp setup --cms drupal --domain site1.com --php-version 8.3

# Site 2 with PHP 8.4
sudo svp setup --cms drupal --domain site2.com --php-version 8.4

# Site 3 with PHP 8.2
sudo svp setup --cms wordpress --domain site3.com --php-version 8.2
```

---

## Update PHP Version

### Using php-update Command

Update PHP version for existing site:

```bash
sudo svp php-update --domain example.com --php-version 8.4
```

**This automatically:**
1. Installs PHP 8.4 if not present
2. Creates new PHP-FPM pool for the domain
3. Updates Nginx configuration
4. Removes old PHP-FPM pool
5. Restarts services

### Manual Update Process

If you prefer manual control:

**1. Install new PHP version:**
```bash
sudo apt update
sudo apt install php8.4-fpm php8.4-cli php8.4-common \
  php8.4-mysql php8.4-gd php8.4-curl php8.4-mbstring \
  php8.4-xml php8.4-zip php8.4-opcache
```

**2. Create PHP-FPM pool:**
```bash
sudo cp /etc/php/8.3/fpm/pool.d/example.com.conf \
       /etc/php/8.4/fpm/pool.d/example.com.conf

# Edit socket path
sudo sed -i 's/php8.3-fpm/php8.4-fpm/g' \
       /etc/php/8.4/fpm/pool.d/example.com.conf
```

**3. Update Nginx:**
```bash
sudo sed -i 's/php8.3-fpm/php8.4-fpm/g' \
       /etc/nginx/sites-available/example.com.conf
```

**4. Restart services:**
```bash
sudo systemctl restart php8.4-fpm
sudo systemctl reload nginx
```

**5. Remove old pool:**
```bash
sudo rm /etc/php/8.3/fpm/pool.d/example.com.conf
sudo systemctl restart php8.3-fpm
```

---

## PHP Configuration

### PHP-FPM Pool Settings

Each domain has its own PHP-FPM pool:

```bash
/etc/php/8.3/fpm/pool.d/example.com.conf
```

**Default settings:**
```ini
[example.com]
user = admin
group = www-data
listen = /run/php/php8.3-fpm-example.com.sock
listen.owner = www-data
listen.group = www-data
pm = dynamic
pm.max_children = 5
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
```

### Adjust PHP-FPM Workers

For high-traffic sites:

```bash
sudo nano /etc/php/8.3/fpm/pool.d/example.com.conf
```

Increase workers:
```ini
pm.max_children = 20
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 10
```

Restart PHP-FPM:
```bash
sudo systemctl restart php8.3-fpm
```

### PHP Memory Limit

Increase memory per domain:

```bash
sudo nano /etc/php/8.3/fpm/pool.d/example.com.conf
```

Add:
```ini
php_admin_value[memory_limit] = 256M
php_admin_value[upload_max_filesize] = 64M
php_admin_value[post_max_size] = 64M
```

Restart:
```bash
sudo systemctl restart php8.3-fpm
```

### PHP INI Settings

Global PHP settings:
```bash
/etc/php/8.3/fpm/php.ini
```

Per-domain PHP settings can be added to pool config:
```ini
php_admin_value[max_execution_time] = 300
php_admin_value[max_input_time] = 300
php_admin_flag[display_errors] = off
```

---

## PHP Extensions

### Installed Extensions

View installed extensions:

```bash
php8.3 -m
```

Common extensions installed by svp:
- mysql
- gd
- curl
- mbstring
- xml
- zip
- opcache
- intl
- soap

### Install Additional Extensions

```bash
# Example: Install Redis
sudo apt install php8.3-redis

# Example: Install Memcached
sudo apt install php8.3-memcached

# Example: Install ImageMagick
sudo apt install php8.3-imagick
```

Restart PHP-FPM after installing:
```bash
sudo systemctl restart php8.3-fpm
```

---

## Multiple PHP Versions

### Running Multiple Versions

You can have multiple PHP versions installed simultaneously:

```bash
# Check installed versions
dpkg -l | grep php | grep fpm

# Should show multiple versions:
php8.2-fpm
php8.3-fpm
php8.4-fpm
```

Each site uses its specified version independently.

### Check Site's PHP Version

**Via web:**
Create `info.php` in document root:
```php
<?php
phpinfo();
```

Visit: `https://example.com/info.php`

**Via CLI:**
```bash
# Check pool config
grep listen /etc/php/*/fpm/pool.d/example.com.conf

# Shows which PHP version is used
/etc/php/8.3/fpm/pool.d/example.com.conf:listen = /run/php/php8.3-fpm-example.com.sock
```

---

## PHP-FPM Management

### Service Control

```bash
# Status
sudo systemctl status php8.3-fpm

# Start
sudo systemctl start php8.3-fpm

# Stop
sudo systemctl stop php8.3-fpm

# Restart
sudo systemctl restart php8.3-fpm

# Reload (graceful)
sudo systemctl reload php8.3-fpm
```

### View Running Pools

```bash
# All PHP-FPM processes
ps aux | grep php-fpm

# Specific version
ps aux | grep php8.3-fpm
```

### Pool Status

Check pool status page (if enabled):

```bash
sudo curl http://localhost/php-fpm-status
```

---

## Performance Tuning

### OPcache Configuration

OPcache is enabled by default. View settings:

```bash
php8.3 -i | grep opcache
```

Adjust OPcache:
```bash
sudo nano /etc/php/8.3/fpm/conf.d/10-opcache.ini
```

Recommended settings:
```ini
opcache.enable=1
opcache.memory_consumption=128
opcache.interned_strings_buffer=8
opcache.max_accelerated_files=10000
opcache.revalidate_freq=60
opcache.fast_shutdown=1
```

### Process Manager Tuning

For better performance, adjust PM settings:

**Low traffic (default):**
```ini
pm = dynamic
pm.max_children = 5
```

**Medium traffic:**
```ini
pm = dynamic
pm.max_children = 20
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 10
```

**High traffic:**
```ini
pm = static
pm.max_children = 50
```

**Calculate max_children:**
```
max_children = (Total RAM - System RAM) / PHP process size

Example:
(2048 MB - 512 MB) / 50 MB = ~30 max_children
```

---

## Troubleshooting

### 502 Bad Gateway

**Cause:** PHP-FPM not running or socket issue

**Check PHP-FPM:**
```bash
sudo systemctl status php8.3-fpm
```

**Check socket:**
```bash
ls -la /run/php/php8.3-fpm-example.com.sock
```

**Restart PHP-FPM:**
```bash
sudo systemctl restart php8.3-fpm
```

### PHP Version Not Working

**Check Nginx config:**
```bash
grep fastcgi_pass /etc/nginx/sites-available/example.com.conf
```

Should show correct socket:
```
fastcgi_pass unix:/run/php/php8.3-fpm-example.com.sock;
```

**Test Nginx config:**
```bash
sudo nginx -t
```

### Extension Not Loading

**Check if installed:**
```bash
dpkg -l | grep php8.3-[extension]
```

**Install if missing:**
```bash
sudo apt install php8.3-[extension]
sudo systemctl restart php8.3-fpm
```

**Verify loaded:**
```bash
php8.3 -m | grep [extension]
```

---

## Version-Specific Features

### PHP 8.4

New features:
- Property hooks
- Asymmetric visibility
- New array functions
- Performance improvements

### PHP 8.3

Features:
- Typed class constants
- Dynamic class constant fetch
- Readonly amendments
- Better performance

### PHP 8.2

Features:
- Readonly classes
- Disjunctive Normal Form types
- Null, false, and true standalone types

### PHP 8.1

Features:
- Enums
- Fibers
- Readonly properties
- First-class callable syntax

---

## Migration Guide

### Upgrading from PHP 8.2 to 8.3

**1. Test on staging:**
```bash
sudo svp php-update --domain staging.example.com --php-version 8.3
```

**2. Test application:**
- Check error logs: `sudo tail -f /var/log/php8.3-fpm-staging.example.com-error.log`
- Test all functionality
- Check deprecated warnings

**3. Update production:**
```bash
sudo svp php-update --domain example.com --php-version 8.3
```

**4. Monitor:**
```bash
sudo tail -f /var/log/php8.3-fpm-example.com-error.log
```

### Rollback if Needed

```bash
sudo svp php-update --domain example.com --php-version 8.2
```

---

## Logs

### PHP-FPM Logs

Per-domain error log:
```bash
sudo tail -f /var/log/php8.3-fpm-example.com-error.log
```

Main PHP-FPM log:
```bash
sudo tail -f /var/log/php8.3-fpm.log
```

### Slow Log

Enable slow log in pool config:
```ini
slowlog = /var/log/php8.3-fpm-example.com-slow.log
request_slowlog_timeout = 5s
```

View slow requests:
```bash
sudo tail -f /var/log/php8.3-fpm-example.com-slow.log
```

---

[← Database Management](database-management) | [Configuration Files →](configuration-files)
