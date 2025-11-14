---
layout: default
title: Database Management
---

# Database Management

Complete guide to database operations with svp.

## Overview

svp automatically creates and configures MariaDB databases for each site with:

- ✅ Secure random passwords
- ✅ Per-site database and user
- ✅ Proper permissions
- ✅ Automatic configuration in CMS

---

## Database Creation

### Automatic Setup

When provisioning a site, svp automatically:

1. Creates database
2. Creates database user
3. Generates secure password
4. Grants proper permissions
5. Configures CMS settings

### Database Naming

Databases are named based on domain:

```
example.com           → drupal_example_com
staging.example.com   → drupal_staging_example_com
myblog.com            → wordpress_myblog_com
```

**Pattern:** `{cms}_{domain_with_underscores}`

---

## Database Credentials

### View Credentials

Credentials are stored in `/etc/svp/sites/`:

```bash
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

### Access Database

Using stored credentials:

```bash
# Read credentials
source /etc/svp/sites/example.com.db.txt

# Connect to database
mysql -u "$Username" -p"$Password" "$Database"
```

Or directly:
```bash
# Drupal
drush-example.com sql-cli

# WordPress
cd /var/www/myblog.com
sudo -u admin wp db cli
```

---

## Database Import

### During Initial Setup

Import database during provisioning:

```bash
sudo svp setup example.com \
  --cms drupal \
  --git-repo https://github.com/myorg/mysite.git \
  --db /path/to/backup.sql.gz \
  --le-email admin@example.com
```

**Supported formats:**
- `.sql` - Plain SQL dump
- `.sql.gz` - Gzip compressed dump

### After Setup

Import database after site is provisioned:

```bash
# Drupal
cd /var/www/example.com
drush-example.com sql-drop -y
zcat /path/to/backup.sql.gz | drush-example.com sql-cli

# WordPress
cd /var/www/myblog.com
sudo -u admin wp db reset --yes
zcat /path/to/backup.sql.gz | sudo -u admin wp db cli
```

### Manual Import

Using mysql directly:

```bash
# Uncompressed
mysql -u drupal_example_com -p drupal_example_com < backup.sql

# Compressed
zcat backup.sql.gz | mysql -u drupal_example_com -p drupal_example_com
```

---

## Database Export

### Export Database

**Drupal:**
```bash
# Plain SQL
drush-example.com sql-dump > backup.sql

# Compressed
drush-example.com sql-dump --gzip > backup-$(date +%Y%m%d).sql.gz

# With structure only (no data)
drush-example.com sql-dump --structure > structure.sql

# Specific tables
drush-example.com sql-dump --tables=node,users > backup.sql
```

**WordPress:**
```bash
cd /var/www/myblog.com

# Plain SQL
sudo -u admin wp db export backup.sql

# Compressed
sudo -u admin wp db export backup.sql
gzip backup.sql

# Add timestamp
sudo -u admin wp db export backup-$(date +%Y%m%d).sql
```

### Automated Backups

Create `/usr/local/bin/backup-db.sh`:

```bash
#!/bin/bash
DOMAIN="example.com"
BACKUP_DIR="/backups/databases"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR

# Backup database
drush-$DOMAIN sql-dump --gzip > $BACKUP_DIR/$DOMAIN-$DATE.sql.gz

# Keep only last 30 days
find $BACKUP_DIR -name "$DOMAIN-*.sql.gz" -mtime +30 -delete

echo "Backup complete: $BACKUP_DIR/$DOMAIN-$DATE.sql.gz"
```

Make executable:
```bash
sudo chmod +x /usr/local/bin/backup-db.sh
```

Schedule with cron:
```bash
sudo crontab -e

# Daily backup at 2 AM
0 2 * * * /usr/local/bin/backup-db.sh
```

---

## Database Operations

### List Databases

```bash
mysql -e "SHOW DATABASES;"
```

### Database Size

```bash
mysql -e "SELECT table_schema AS 'Database', 
  ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'Size (MB)' 
  FROM information_schema.TABLES 
  GROUP BY table_schema;"
```

### Optimize Database

**Drupal:**
```bash
# Via Drush
drush-example.com sql-query "OPTIMIZE TABLE cache_bootstrap, cache_config, cache_data, cache_default, cache_discovery, cache_dynamic_page_cache, cache_entity, cache_menu, cache_page, cache_render;"

# Or specific tables
drush-example.com sql-query "OPTIMIZE TABLE cache_render;"
```

**WordPress:**
```bash
cd /var/www/myblog.com
sudo -u admin wp db optimize
```

**Manual:**
```bash
mysql -u drupal_example_com -p drupal_example_com -e "OPTIMIZE TABLE table_name;"
```

### Repair Database

```bash
# WordPress
cd /var/www/myblog.com
sudo -u admin wp db repair

# Manual
mysql -u drupal_example_com -p drupal_example_com -e "REPAIR TABLE table_name;"
```

---

## Database Users

### View Users

```bash
mysql -e "SELECT User, Host FROM mysql.user WHERE User LIKE 'drupal%' OR User LIKE 'wordpress%';"
```

### Change Password

```bash
# Generate new password
NEW_PASS=$(openssl rand -base64 32)

# Update in MySQL
mysql -e "ALTER USER 'drupal_example_com'@'localhost' IDENTIFIED BY '$NEW_PASS';"

# Update credentials file
sudo sed -i "s/Password: .*/Password: $NEW_PASS/" /etc/svp/sites/example.com.db.txt

# Update Drupal settings
sudo sed -i "s/'password' => '.*'/'password' => '$NEW_PASS'/" /var/www/example.com/web/sites/default/settings.svp.php
```

### Grant Permissions

```bash
mysql -e "GRANT ALL PRIVILEGES ON drupal_example_com.* TO 'drupal_example_com'@'localhost';"
mysql -e "FLUSH PRIVILEGES;"
```

---

## Database Migrations

### Migrate to New Database

**1. Export from old database:**
```bash
mysqldump -u old_user -p old_database | gzip > migration.sql.gz
```

**2. Create new database:**
```bash
sudo svp setup --cms drupal --domain example.com
```

**3. Import to new database:**
```bash
zcat migration.sql.gz | drush-example.com sql-cli
```

### Migrate Between Servers

**Old server:**
```bash
# Export database
drush-example.com sql-dump --gzip > export.sql.gz

# Transfer to new server
scp export.sql.gz user@new-server:/tmp/
```

**New server:**
```bash
# Import database
zcat /tmp/export.sql.gz | drush-example.com sql-cli
```

---

## Database Configuration

### MariaDB Configuration

Main config: `/etc/mysql/mariadb.conf.d/50-server.cnf`

**Important settings:**
```ini
[mysqld]
max_connections = 100
innodb_buffer_pool_size = 512M
query_cache_size = 64M
```

Restart after changes:
```bash
sudo systemctl restart mariadb
```

### Per-Site Settings

**Drupal** (`/var/www/example.com/web/sites/default/settings.svp.php`):
```php
$databases['default']['default'] = [
  'database' => 'drupal_example_com',
  'username' => 'drupal_example_com',
  'password' => '[password]',
  'host' => 'localhost',
  'port' => '3306',
  'driver' => 'mysql',
  'prefix' => '',
  'collation' => 'utf8mb4_general_ci',
];
```

**WordPress** (`/var/www/myblog.com/wp-config.php`):
```php
define('DB_NAME', 'wordpress_myblog_com');
define('DB_USER', 'wordpress_myblog_com');
define('DB_PASSWORD', '[password]');
define('DB_HOST', 'localhost');
define('DB_CHARSET', 'utf8mb4');
define('DB_COLLATE', '');
```

---

## Database Reprovisioning

### Fresh Database

Drop and recreate database:

```bash
# Reprovision with fresh database
sudo svp setup example.com --cms drupal

# When prompted about existing directory:
Delete and reprovision? [y/N]: y
```

Creates:
- New database
- New credentials
- Fresh installation

### Keep Existing Database

Reprovision but reuse database:

```bash
sudo svp setup example.com --cms drupal --keep-existing-db
```

This:
- Keeps same database name
- Keeps same credentials
- Drops all tables
- Reinstalls fresh CMS

**Use when:**
- You want to test reinstall
- You need to keep credentials stable
- Automated scripts depend on fixed credentials

---

## Troubleshooting

### Can't Connect to Database

**Check credentials:**
```bash
sudo cat /etc/svp/sites/example.com.db.txt
```

**Test connection:**
```bash
mysql -u drupal_example_com -p
# Enter password from credentials file
```

**Check MariaDB is running:**
```bash
sudo systemctl status mariadb
```

### Database Import Fails

**Error:** "Access denied"

Check user permissions:
```bash
mysql -e "SHOW GRANTS FOR 'drupal_example_com'@'localhost';"
```

**Error:** "Table already exists"

Drop tables first:
```bash
drush-example.com sql-drop -y
```

### Database Too Large

**Check size:**
```bash
drush-example.com sql-query "SELECT table_name, 
  ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size (MB)' 
  FROM information_schema.TABLES 
  WHERE table_schema = 'drupal_example_com' 
  ORDER BY (data_length + index_length) DESC 
  LIMIT 10;"
```

**Clear cache tables:**
```bash
drush-example.com sql-query "TRUNCATE cache_bootstrap;"
drush-example.com sql-query "TRUNCATE cache_render;"
# Repeat for other cache tables
```

### Corrupt Tables

**Check tables:**
```bash
drush-example.com sql-query "CHECK TABLE table_name;"
```

**Repair:**
```bash
drush-example.com sql-query "REPAIR TABLE table_name;"
```

---

## Performance Tuning

### Optimize Queries

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

Restart:
```bash
sudo systemctl restart mariadb
```

View slow queries:
```bash
sudo tail -f /var/log/mysql/slow.log
```

### Add Indexes

Find missing indexes:
```bash
drush-example.com sql-query "
  SELECT * FROM information_schema.TABLES 
  WHERE table_schema = 'drupal_example_com' 
  AND index_length = 0;"
```

---

[← Git Deployment](git-deployment) | [PHP Version Management →](php-versions)
