---
layout: default
title: Multi-Domain Setup
---

# Multi-Domain Setup

Run multiple websites on a single VPS with isolated environments.

## Overview

svp supports hosting multiple domains on one server. Each domain gets:

- ✅ Separate directory
- ✅ Isolated PHP-FPM pool
- ✅ Dedicated database
- ✅ Independent SSL certificate
- ✅ Isolated file storage
- ✅ Separate Nginx vhost

---

## Basic Multi-Domain Setup

### Using Extra Domains Flag

Provision multiple domains in one command:

```bash
sudo svp setup example.com \
  --cms drupal \
  --extra-domains "staging.example.com,dev.example.com" \
  --le-email admin@example.com
```

**This creates:**
```
/var/www/
├── example.com/          # Production
├── staging.example.com/  # Staging
└── dev.example.com/      # Development
```

Each with:
- Own database (e.g., `drupal_example_com`, `drupal_staging_example_com`)
- Own PHP-FPM pool
- Own SSL certificate
- Own Nginx configuration

---

## Multiple Separate Setups

### Add Domains One at a Time

Run svp multiple times for different domains:

```bash
# First site
sudo svp setup site1.com --cms drupal --le-email admin@site1.com

# Second site
sudo svp setup site2.com --cms drupal --le-email admin@site2.com

# Third site
sudo svp setup blog.com --cms wordpress --le-email admin@blog.com
```

**Advantage:** Each site can have:
- Different CMS (Drupal vs WordPress)
- Different PHP version
- Different Git repository
- Independent provisioning

---

## Environment-Based Setup

### Production + Staging + Development

Perfect for development workflow:

```bash
sudo svp setup myapp.com \
  --cms drupal \
  --extra-domains "staging.myapp.com,dev.myapp.com" \
  --git-repo https://github.com/company/myapp.git \
  --le-email devops@company.com

# Password-protect non-production environments
sudo svp auth staging.myapp.com enable --username team --password stagingPass
sudo svp auth dev.myapp.com enable --username dev --password devPass
```

**Workflow:**

1. **Development** (`dev.myapp.com`)
   - Feature branch testing
   - Rapid iteration
   - Latest code
   - Password-protected from public

2. **Staging** (`staging.myapp.com`)
   - Pre-production testing
   - QA environment
   - Password-protected for client previews
   - Mirror of production data

3. **Production** (`myapp.com`)
   - Live site
   - Stable branch only
   - Production data

**Switch branches per environment:**

```bash
# Development - feature branch
cd /var/www/dev.myapp.com
sudo -u admin git checkout feature/new-feature

# Staging - develop branch
cd /var/www/staging.myapp.com
sudo -u admin git checkout develop

# Production - main branch
cd /var/www/myapp.com
sudo -u admin git checkout main
```

---

## Resource Allocation

### PHP-FPM Pool Isolation

Each domain has isolated PHP processes:

```bash
# View running pools
sudo systemctl status php8.3-fpm

# Check pool configs
ls -l /etc/php/8.3/fpm/pool.d/
```

**Per-domain pool configuration:**
```
/etc/php/8.3/fpm/pool.d/
├── example.com.conf
├── staging.example.com.conf
└── dev.example.com.conf
```

### Database Isolation

Each domain has its own database:

```bash
# List databases
sudo mysql -e "SHOW DATABASES LIKE '%example%';"
```

**Output:**
```
drupal_example_com
drupal_staging_example_com
drupal_dev_example_com
```

### Process Isolation Benefits

1. **Security** - One site compromise doesn't affect others
2. **Performance** - Limit resources per site
3. **Debugging** - Easy to identify which site has issues
4. **Flexibility** - Different PHP versions per site

---

## Managing Multiple Sites

### Drush Aliases

Each Drupal site gets its own drush command:

```bash
# Production
drush-example.com status

# Staging
drush-staging.example.com status

# Development
drush-dev.example.com status
```

### WP-CLI for WordPress

```bash
# Site 1
cd /var/www/site1.com
sudo -u admin wp plugin list

# Site 2
cd /var/www/site2.com
sudo -u admin wp theme list
```

### Database Credentials

Each site has separate credentials:

```bash
# Production
sudo cat /etc/svp/sites/example.com.db.txt

# Staging
sudo cat /etc/svp/sites/staging.example.com.db.txt

# Development
sudo cat /etc/svp/sites/dev.example.com.db.txt
```

---

## SSL for Multiple Domains

### Automatic SSL

Each domain gets its own certificate:

```bash
# View all certificates
sudo certbot certificates
```

**Output:**
```
Certificate Name: example.com
  Domains: example.com
  
Certificate Name: staging.example.com
  Domains: staging.example.com
  
Certificate Name: dev.example.com
  Domains: dev.example.com
```

### Renewal

All certificates auto-renew via certbot timer:

```bash
sudo systemctl status certbot.timer
```

---

## DNS Configuration

### Setup DNS for Multiple Domains

Each domain needs an A record pointing to your server:

```
A    example.com             123.45.67.89
A    staging.example.com     123.45.67.89
A    dev.example.com         123.45.67.89
A    site1.com               123.45.67.89
A    site2.com               123.45.67.89
```

**For subdomains:**
```
A    staging                 123.45.67.89
A    dev                     123.45.67.89
```

Or use wildcard (if all subdomains point to same server):
```
A    *                       123.45.67.89
```

---

## Server Resources

### Calculating Capacity

**PHP-FPM Memory:**
- Each PHP worker: ~50MB
- Default per site: 5 workers = 250MB
- 3 sites: ~750MB for PHP alone

**Total estimate for 3 sites:**
- PHP-FPM: 750MB
- Nginx: 100MB
- MariaDB: 400MB
- System: 300MB
- **Total: ~1.5GB minimum**

**Recommendation:**
- 1-3 sites: 2GB RAM VPS
- 4-8 sites: 4GB RAM VPS
- 9+ sites: 8GB+ RAM VPS

### Monitor Resources

```bash
# Overall memory
free -h

# Per-process memory
ps aux --sort=-%mem | head -20

# PHP-FPM pool processes
ps aux | grep php-fpm
```

---

## Nginx Configuration

### Virtual Hosts

Each domain gets its own Nginx config:

```bash
ls -l /etc/nginx/sites-available/
```

**Output:**
```
example.com.conf
staging.example.com.conf
dev.example.com.conf
```

### Check Configuration

```bash
# Test all configs
sudo nginx -t

# Reload after changes
sudo systemctl reload nginx
```

---

## Different PHP Versions

### Run Different PHP per Domain

```bash
# Site 1 with PHP 8.3
sudo svp setup site1.com --cms drupal --php-version 8.3

# Site 2 with PHP 8.4
sudo svp setup site2.com --cms drupal --php-version 8.4
```

Each site uses its specified PHP-FPM version independently.

---

## Data Synchronization

### Sync Staging from Production

Copy database and files from production to staging:

**1. Export production database:**
```bash
drush-example.com sql-dump --gzip > prod-db.sql.gz
```

**2. Import to staging:**
```bash
zcat prod-db.sql.gz | drush-staging.example.com sql-cli
```

**3. Sync files:**
```bash
rsync -avz --delete \
  /var/www/example.com/web/sites/default/files/ \
  /var/www/staging.example.com/web/sites/default/files/
```

**4. Clear cache:**
```bash
drush-staging.example.com cr
```

### Automated Sync Script

Create `/usr/local/bin/sync-to-staging.sh`:

```bash
#!/bin/bash
set -e

PROD="/var/www/example.com"
STAGING="/var/www/staging.example.com"

echo "Syncing production to staging..."

# Database
echo "Exporting production database..."
drush-example.com sql-dump --gzip > /tmp/prod-sync.sql.gz

echo "Importing to staging..."
zcat /tmp/prod-sync.sql.gz | drush-staging.example.com sql-cli
rm /tmp/prod-sync.sql.gz

# Files
echo "Syncing files..."
rsync -avz --delete \
  $PROD/web/sites/default/files/ \
  $STAGING/web/sites/default/files/

# Clear cache
echo "Clearing staging cache..."
drush-staging.example.com cr

echo "Sync complete!"
```

Make executable:
```bash
sudo chmod +x /usr/local/bin/sync-to-staging.sh
```

Run:
```bash
sudo /usr/local/bin/sync-to-staging.sh
```

---

## Removing Sites

### Delete a Domain

To remove a site completely:

```bash
# 1. Remove SSL certificate
sudo certbot delete --cert-name staging.example.com

# 2. Remove Nginx config
sudo rm /etc/nginx/sites-enabled/staging.example.com.conf
sudo rm /etc/nginx/sites-available/staging.example.com.conf
sudo systemctl reload nginx

# 3. Stop and remove PHP-FPM pool
sudo rm /etc/php/8.3/fpm/pool.d/staging.example.com.conf
sudo systemctl restart php8.3-fpm

# 4. Drop database
sudo mysql <<EOF
DROP DATABASE IF EXISTS drupal_staging_example_com;
DROP USER IF EXISTS 'drupal_staging_example_com'@'localhost';
EOF

# 5. Remove site directory
sudo rm -rf /var/www/staging.example.com

# 6. Remove site config
sudo rm /etc/svp/sites/staging.example.com.*
```

---

## Best Practices

### 1. Use Consistent Naming

```
example.com           # Production
staging.example.com   # Staging
dev.example.com       # Development
```

### 2. Separate Git Branches

- Production: `main` or `master`
- Staging: `develop` or `staging`
- Development: `develop` or feature branches

### 3. Different Databases

Never share databases between environments. Always use separate databases.

### 4. Monitor Resources

Keep an eye on server resources:
```bash
# Check memory
free -h

# Check disk
df -h

# Check load
uptime
```

### 5. Regular Backups

Back up each site independently:
```bash
# Production
drush-example.com sql-dump --gzip > prod-backup.sql.gz

# Staging
drush-staging.example.com sql-dump --gzip > staging-backup.sql.gz
```

---

## Troubleshooting

### Wrong Site Loading

Check Nginx configuration:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

Verify server_name in config:
```bash
grep server_name /etc/nginx/sites-available/example.com.conf
```

### PHP-FPM Issues

Check pool is running:
```bash
sudo systemctl status php8.3-fpm
```

View pool logs:
```bash
sudo tail -f /var/log/php8.3-fpm-example.com-error.log
```

### Database Connection

Verify credentials:
```bash
sudo cat /etc/svp/sites/example.com.db.txt
```

Test connection:
```bash
mysql -u drupal_example_com -p
```

---

[← SSL Configuration](ssl-configuration) | [Git Deployment →](git-deployment)
