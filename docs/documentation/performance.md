---
layout: default
title: Performance Tuning
---

# Performance Tuning

Optimize your svp-provisioned sites for maximum performance.

## Overview

Performance tuning areas:

- ✅ PHP-FPM optimization
- ✅ Nginx caching
- ✅ Database tuning
- ✅ OPcache configuration
- ✅ Resource allocation
- ✅ Content delivery

---

## PHP-FPM Optimization

### Process Manager Tuning

Edit PHP-FPM pool config:
```bash
sudo nano /etc/php/8.3/fpm/pool.d/example.com.conf
```

**Low traffic (default):**
```ini
pm = dynamic
pm.max_children = 5
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
pm.max_requests = 500
```

**Medium traffic:**
```ini
pm = dynamic
pm.max_children = 20
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 10
pm.max_requests = 1000
```

**High traffic:**
```ini
pm = static
pm.max_children = 50
pm.max_requests = 500
```

**Calculate max_children:**
```
Available RAM = Total RAM - System RAM
max_children = Available RAM / Process Memory

Example:
(4096 MB - 1024 MB) / 50 MB = 61 max_children
```

Restart PHP-FPM:
```bash
sudo systemctl restart php8.3-fpm
```

### PHP Memory Settings

Increase PHP memory:
```bash
sudo nano /etc/php/8.3/fpm/pool.d/example.com.conf
```

Add:
```ini
php_admin_value[memory_limit] = 256M
php_admin_value[max_execution_time] = 300
php_admin_value[max_input_time] = 300
```

### OPcache Optimization

Configure OPcache:
```bash
sudo nano /etc/php/8.3/fpm/conf.d/10-opcache.ini
```

Recommended settings:
```ini
opcache.enable=1
opcache.memory_consumption=128
opcache.interned_strings_buffer=16
opcache.max_accelerated_files=10000
opcache.revalidate_freq=60
opcache.save_comments=1
opcache.fast_shutdown=1
opcache.enable_file_override=1
```

Restart:
```bash
sudo systemctl restart php8.3-fpm
```

---

## Nginx Optimization

### FastCGI Cache

Enable FastCGI caching:
```bash
sudo nano /etc/nginx/nginx.conf
```

Add in http block:
```nginx
fastcgi_cache_path /var/cache/nginx levels=1:2 
    keys_zone=DRUPAL:100m inactive=60m;
fastcgi_cache_key "$scheme$request_method$host$request_uri";
```

Create cache directory:
```bash
sudo mkdir -p /var/cache/nginx
sudo chown www-data:www-data /var/cache/nginx
```

In site config:
```bash
sudo nano /etc/nginx/sites-available/example.com.conf
```

Add in location ~ \.php$ block:
```nginx
location ~ \.php$ {
    fastcgi_pass unix:/run/php/php8.3-fpm-example.com.sock;
    include fastcgi_params;
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    
    # FastCGI cache
    fastcgi_cache DRUPAL;
    fastcgi_cache_valid 200 60m;
    fastcgi_cache_valid 404 10m;
    fastcgi_cache_bypass $http_pragma $http_authorization;
    fastcgi_no_cache $http_pragma $http_authorization;
    add_header X-Cache-Status $upstream_cache_status;
}
```

Test and reload:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

### Browser Caching

Add caching headers:
```nginx
location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

### Gzip Compression

Verify gzip enabled (usually default):
```bash
sudo nano /etc/nginx/nginx.conf
```

Ensure:
```nginx
gzip on;
gzip_vary on;
gzip_proxied any;
gzip_comp_level 6;
gzip_types text/plain text/css text/xml text/javascript
           application/json application/javascript application/xml+rss;
```

### Connection Tuning

Optimize connections:
```nginx
worker_processes auto;
worker_connections 4096;
keepalive_timeout 65;
keepalive_requests 100;
```

---

## Database Optimization

### MariaDB Configuration

Edit MariaDB config:
```bash
sudo nano /etc/mysql/mariadb.conf.d/50-server.cnf
```

**For 2GB RAM server:**
```ini
[mysqld]
innodb_buffer_pool_size = 512M
innodb_log_file_size = 128M
max_connections = 100
query_cache_size = 64M
query_cache_limit = 2M
thread_cache_size = 8
table_open_cache = 2000
```

**For 4GB RAM server:**
```ini
[mysqld]
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
max_connections = 200
query_cache_size = 128M
query_cache_limit = 4M
thread_cache_size = 16
table_open_cache = 4000
```

Restart MariaDB:
```bash
sudo systemctl restart mariadb
```

### Optimize Tables

Regular table optimization:
```bash
# Drupal
drush-example.com sql-query "OPTIMIZE TABLE cache_bootstrap, cache_config, cache_data, cache_default, cache_discovery, cache_dynamic_page_cache, cache_entity, cache_menu, cache_page, cache_render;"

# WordPress
cd /var/www/myblog.com
sudo -u admin wp db optimize
```

### Query Optimization

Enable slow query log:
```bash
sudo nano /etc/mysql/mariadb.conf.d/50-server.cnf
```

Add:
```ini
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
```

View slow queries:
```bash
sudo tail -f /var/log/mysql/slow.log
```

---

## Drupal Performance

### Page Cache

Enable page cache:
```bash
drush-example.com config-set system.performance cache.page.max_age 3600 -y
```

### CSS/JS Aggregation

```bash
drush-example.com config-set system.performance css.preprocess 1 -y
drush-example.com config-set system.performance js.preprocess 1 -y
```

### Twig Cache

Disable Twig debug (production):
```bash
# Edit services.yml
sudo nano /var/www/example.com/web/sites/default/services.yml
```

Set:
```yaml
parameters:
  twig.config:
    debug: false
    auto_reload: false
    cache: true
```

Clear cache:
```bash
drush-example.com cr
```

### Redis Cache

Install Redis:
```bash
sudo apt install redis-server php8.3-redis
sudo systemctl enable redis-server
```

Configure Drupal:
```bash
cd /var/www/example.com
sudo -u admin composer require drupal/redis
drush-example.com en redis -y
```

Add to settings.php:
```php
$settings['redis.connection']['interface'] = 'PhpRedis';
$settings['redis.connection']['host'] = '127.0.0.1';
$settings['cache']['default'] = 'cache.backend.redis';
```

### Database Cleanup

Clean old data:
```bash
# Clear cache tables
drush-example.com sql-query "TRUNCATE cache_bootstrap;"
drush-example.com sql-query "TRUNCATE cache_render;"
drush-example.com sql-query "TRUNCATE cache_data;"

# Clean watchdog
drush-example.com watchdog:delete all -y
```

---

## WordPress Performance

### Object Cache

Install Redis object cache:
```bash
sudo apt install redis-server php8.3-redis
cd /var/www/myblog.com
sudo -u admin wp plugin install redis-cache --activate
sudo -u admin wp redis enable
```

### Page Cache Plugin

```bash
sudo -u admin wp plugin install wp-super-cache --activate
# Or
sudo -u admin wp plugin install w3-total-cache --activate
```

### Database Cleanup

```bash
cd /var/www/myblog.com

# Optimize database
sudo -u admin wp db optimize

# Clean revisions
sudo -u admin wp post delete $(wp post list --post_type='revision' --format=ids)

# Clean trash
sudo -u admin wp post delete $(wp post list --post_status=trash --format=ids)

# Clean spam comments
sudo -u admin wp comment delete $(wp comment list --status=spam --format=ids)
```

---

## Resource Monitoring

### Check Server Load

```bash
# Current load
uptime

# Detailed
top
htop  # If installed

# Memory usage
free -h

# Disk usage
df -h
```

### Monitor PHP-FPM

```bash
# Process count
ps aux | grep php-fpm | wc -l

# Memory per process
ps aux | grep php-fpm | awk '{sum+=$6} END {print sum/NR/1024 " MB"}'

# View status
sudo systemctl status php8.3-fpm
```

### Monitor Nginx

```bash
# Connection count
netstat -an | grep :80 | wc -l

# Status
sudo systemctl status nginx

# Test config
sudo nginx -t
```

### Monitor MariaDB

```bash
# Connections
mysql -e "SHOW STATUS LIKE 'Threads_connected';"

# Slow queries
mysql -e "SHOW STATUS LIKE 'Slow_queries';"

# Buffer pool usage
mysql -e "SHOW STATUS LIKE 'Innodb_buffer_pool%';"
```

---

## Content Delivery

### CDN Integration

Use CDN for static assets:

**Drupal:**
```bash
cd /var/www/example.com
sudo -u admin composer require drupal/cdn
drush-example.com en cdn -y
```

Configure in admin interface.

**WordPress:**
Use plugin like:
- WP Super Cache
- W3 Total Cache
- Cloudflare

### Image Optimization

**Drupal:**
```bash
sudo apt install jpegoptim optipng
cd /var/www/example.com
sudo -u admin composer require drupal/imageapi_optimize
drush-example.com en imageapi_optimize -y
```

**WordPress:**
```bash
cd /var/www/myblog.com
sudo -u admin wp plugin install ewww-image-optimizer --activate
```

---

## Benchmarking

### Apache Bench

Test site performance:
```bash
# Install
sudo apt install apache2-utils

# Test
ab -n 1000 -c 10 https://example.com/
```

### Load Testing

```bash
# Install siege
sudo apt install siege

# Test
siege -c 25 -t 30s https://example.com/
```

---

## Performance Checklist

### Server Level

- [ ] OPcache enabled and tuned
- [ ] PHP-FPM process manager optimized
- [ ] Nginx gzip enabled
- [ ] FastCGI cache configured
- [ ] MariaDB buffer pool sized correctly

### Application Level

- [ ] Page caching enabled
- [ ] CSS/JS aggregation enabled
- [ ] Image optimization configured
- [ ] Redis/Memcached installed
- [ ] Database regularly optimized

### Content Level

- [ ] Images optimized
- [ ] CSS/JS minified
- [ ] CDN configured
- [ ] Browser caching headers set
- [ ] Unnecessary modules/plugins removed

---

## Troubleshooting Performance

### Site Loading Slowly

**Check server load:**
```bash
top
free -h
```

**Check PHP-FPM:**
```bash
sudo systemctl status php8.3-fpm
sudo tail -f /var/log/php8.3-fpm-example.com-error.log
```

**Check database:**
```bash
sudo tail -f /var/log/mysql/slow.log
```

### High Memory Usage

**Identify processes:**
```bash
ps aux --sort=-%mem | head -20
```

**Reduce PHP-FPM workers:**
```bash
sudo nano /etc/php/8.3/fpm/pool.d/example.com.conf
# Reduce pm.max_children
sudo systemctl restart php8.3-fpm
```

### Database Slow

**Check connections:**
```bash
mysql -e "SHOW PROCESSLIST;"
```

**Kill slow queries:**
```bash
mysql -e "KILL [process_id];"
```

---

[← Security Best Practices](security) | [Backup & Recovery →](backup-recovery)
