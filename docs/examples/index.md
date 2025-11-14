---
layout: default
title: Examples & Use Cases
---

# Examples & Use Cases

Real-world examples and common use cases for Simple VPS Provisioner.

## Table of Contents

- [Quick Start Examples](#quick-start-examples)
- [Production Deployments](#production-deployments)
- [Development Workflows](#development-workflows)
- [Migration Scenarios](#migration-scenarios)
- [Advanced Configurations](#advanced-configurations)

---

## Quick Start Examples

### Fresh Drupal Site

The simplest way to get started with a new Drupal site:

```bash
sudo svp setup \
  --cms drupal \
  --domain mysite.com \
  --le-email admin@mysite.com
```

**What you get:**
- ✅ Fresh Drupal 10 installation
- ✅ Composer-based structure
- ✅ HTTPS with Let's Encrypt
- ✅ MariaDB database with secure credentials
- ✅ PHP 8.3 with optimized settings
- ✅ UFW firewall configured

**Time:** 5-10 minutes

**Access:**
- Site: `https://mysite.com`
- Admin: One-time login link provided

---

### Fresh WordPress Site

```bash
sudo svp setup \
  --cms wordpress \
  --domain myblog.com \
  --le-email admin@myblog.com
```

**What you get:**
- ✅ Latest WordPress
- ✅ WP-CLI installed
- ✅ HTTPS enabled
- ✅ Secure database setup
- ✅ Proper file permissions

**Next steps:**
```bash
# Complete installation
https://myblog.com/wp-admin/install.php
```

---

## Production Deployments

### Deploy from Git Repository

Deploy an existing Drupal site from GitHub:

```bash
sudo svp setup production.mysite.com \
  --cms drupal \
  --git-repo git@github.com:mycompany/mysite.git \
  --git-branch main \
  --le-email devops@mycompany.com
```

**Perfect for:**
- Deploying existing projects
- CI/CD pipelines
- Team collaborations

**Repository structure supported:**
```
mysite/
├── composer.json
├── web/
│   ├── index.php
│   └── sites/default/
├── config/sync/
└── drush/drush.yml
```

---

### Import Existing Database

Migrate a site with an existing database:

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --git-repo git@github.com:mycompany/mysite.git \
  --db /home/admin/mysite-backup-2024-01-15.sql.gz \
  --le-email admin@mysite.com
```

**Steps:**
1. Upload database backup to server
2. Run svp with `-db` flag
3. Database is imported automatically
4. Site is ready to use

**Supported formats:**
- `.sql` - Plain SQL dump
- `.sql.gz` - Gzip compressed

**Important:** Don't forget to sync files:
```bash
rsync -avz user@old-server:/var/www/mysite/web/sites/default/files/ \
  /var/www/mysite.com/web/sites/default/files/
```

---

### Multi-Environment Setup

Production, staging, and development on one server:

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --extra-domains "staging.mysite.com,dev.mysite.com" \
  --git-repo git@github.com:mycompany/mysite.git \
  --le-email devops@mycompany.com
```

**Creates:**
- `mysite.com` - Production
- `staging.mysite.com` - Staging
- `dev.mysite.com` - Development

**Each environment has:**
- Separate database
- Isolated PHP-FPM pool
- Individual SSL certificate
- Independent file storage

**Switching branches:**
```bash
# Staging → develop branch
cd /var/www/staging.mysite.com
sudo -u admin git checkout develop

# Dev → feature branch
cd /var/www/dev.mysite.com
sudo -u admin git checkout feature/new-feature
```

---

## Development Workflows

### Local to Production

#### Step 1: Local Development

Develop locally with DDEV, Lando, or Docker:

```bash
# Your local Drupal site
~/Sites/mysite/
```

#### Step 2: Push to Git

```bash
cd ~/Sites/mysite
git add .
git commit -m "Ready for production"
git push origin main
```

#### Step 3: Deploy to VPS

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --git-repo https://github.com/mycompany/mysite.git \
  --le-email admin@mysite.com
```

#### Step 4: Import Database

```bash
# Export from local
drush sql-dump --gzip > mysite-db.sql.gz

# Upload to server
scp mysite-db.sql.gz admin@mysite.com:/home/admin/

# Import on server
cd /var/www/mysite.com
drush-mysite.com sql-drop -y
zcat /home/admin/mysite-db.sql.gz | drush-mysite.com sql-cli

# Sync files
rsync -avz ~/Sites/mysite/web/sites/default/files/ \
  admin@mysite.com:/var/www/mysite.com/web/sites/default/files/
```

---

### Feature Branch Testing

Test a feature branch on a dedicated environment:

```bash
sudo svp setup feature-xyz.mysite.com \
  --cms drupal \
  --git-repo git@github.com:mycompany/mysite.git \
  --git-branch feature/xyz \
  --le-email admin@mysite.com
```

After testing, tear down:
```bash
# Remove site
sudo rm -rf /var/www/feature-xyz.mysite.com

# Remove database
sudo mysql -e "DROP DATABASE drupal_feature_xyz_mysite_com;"
sudo mysql -e "DROP USER 'drupal_feature_xyz_mysite_com'@'localhost';"

# Remove SSL certificate
sudo certbot delete --cert-name feature-xyz.mysite.com
```

---

## Migration Scenarios

### Migrate from Shared Hosting

#### Step 1: Backup Current Site

On shared host:
```bash
# Files
tar -czf mysite-files.tar.gz public_html/

# Database
mysqldump -u username -p database_name | gzip > mysite-db.sql.gz
```

#### Step 2: Download Backups

```bash
scp user@shared-host:mysite-files.tar.gz ./
scp user@shared-host:mysite-db.sql.gz ./
```

#### Step 3: Upload to VPS

```bash
scp mysite-*.tar.gz admin@vps-ip:/home/admin/
scp mysite-db.sql.gz admin@vps-ip:/home/admin/
```

#### Step 4: Provision Site

```bash
# On VPS
sudo svp setup mysite.com \
  --cms drupal \
  --db /home/admin/mysite-db.sql.gz \
  --le-email admin@mysite.com
```

#### Step 5: Extract Files

```bash
cd /var/www/mysite.com/web/sites/default
sudo tar -xzf /home/admin/mysite-files.tar.gz
sudo chown -R admin:www-data files/
sudo chmod -R 775 files/
```

#### Step 6: Update DNS

Point DNS to new VPS:
```
A    mysite.com    NEW_VPS_IP
```

---

### Migrate from Another VPS

Using rsync for minimal downtime:

#### Step 1: Provision New Server

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --ssl=false
```

#### Step 2: Sync Files (Incremental)

```bash
rsync -avz --delete \
  user@old-vps:/var/www/mysite/web/sites/default/files/ \
  /var/www/mysite.com/web/sites/default/files/
```

#### Step 3: Export/Import Database

On old server:
```bash
drush sql-dump --gzip > mysite-latest.sql.gz
```

On new server:
```bash
scp user@old-vps:mysite-latest.sql.gz /tmp/
cd /var/www/mysite.com
zcat /tmp/mysite-latest.sql.gz | drush-mysite.com sql-cli
```

#### Step 4: Final Sync (Downtime)

Put old site in maintenance mode, final sync:
```bash
drush @old pm-enable maintenance_mode
rsync -avz --delete \
  user@old-vps:/var/www/mysite/web/sites/default/files/ \
  /var/www/mysite.com/web/sites/default/files/
```

#### Step 5: Enable SSL

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --le-email admin@mysite.com
```

#### Step 6: Switch DNS

Update DNS to point to new server.

---

## Advanced Configurations

### Custom PHP Version

Use PHP 8.4 for latest features:

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --php-version 8.4 \
  --le-email admin@mysite.com
```

### Upgrade PHP Version

Upgrade existing site:

```bash
sudo svp php-update mysite.com \
  --php-version 8.4
```

---

### Monorepo Structure

Repository with Drupal in subdirectory:

```
monorepo/
├── backend/          # Drupal here
│   ├── composer.json
│   └── web/
├── frontend/         # React app
└── docs/
```

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --git-repo https://github.com/mycompany/monorepo.git \
  --drupal-root "backend" \
  --le-email admin@mysite.com
```

---

### HTTP Only (Internal Network)

For internal tools without SSL:

```bash
sudo svp setup internal.mycompany.local \
  --cms drupal \
  --ssl=false \
  --firewall=false
```

---

### Custom Document Root

Non-standard docroot:

```bash
sudo svp setup mysite.com \
  --cms drupal \
  --docroot "public_html" \
  --le-email admin@mysite.com
```

---

## WordPress Examples

### Basic WordPress

```bash
sudo svp setup myblog.com \
  --cms wordpress \
  --le-email admin@myblog.com
```

### WordPress from Git

```bash
sudo svp setup myblog.com \
  --cms wordpress \
  --git-repo https://github.com/mycompany/wp-site.git \
  --le-email admin@myblog.com
```

### WordPress with Database Import

```bash
sudo svp setup myblog.com \
  --cms wordpress \
  --db /home/admin/wp-backup.sql.gz \
  --le-email admin@myblog.com
```

**After import:**
```bash
cd /var/www/myblog.com
sudo -u admin wp search-replace 'http://old-domain.com' 'https://myblog.com'
```

---

## Troubleshooting Examples

### Test DNS Before SSL

```bash
# Check if DNS is configured
dig +short mysite.com

# Should return your server IP
123.45.67.89

# Test without SSL first
sudo svp setup mysite.com --cms drupal --ssl=false

# Add SSL later
sudo certbot --nginx -d mysite.com --non-interactive --agree-tos --email admin@mysite.com
```

---

### Reprovision with Fresh Database

Clean slate with new credentials:

```bash
sudo svp setup mysite.com --cms drupal
```

When prompted:
```
Directory /var/www/mysite.com is not empty
Delete and reprovision? [y/N]: y
```

Result:
- New database with NEW credentials
- Fresh Drupal installation
- All old data removed

---

### Reprovision Keeping Database

Keep credentials, clear tables:

```bash
sudo svp setup mysite.com --cms drupal --keep-existing-db
```

Result:
- Same database credentials
- All tables dropped
- Fresh Drupal installation
- Useful for testing without changing credentials

---

## Best Practices

### 1. Always Test Without SSL First

```bash
# Step 1: Test HTTP
sudo svp setup --cms drupal --domain example.com --ssl=false

# Step 2: Verify site works
curl -I http://example.com

# Step 3: Add SSL
sudo certbot --nginx -d example.com
```

### 2. Use Git for All Production Sites

```bash
# Benefits:
# - Version control
# - Easy rollbacks
# - Team collaboration
# - CI/CD integration

sudo svp setup example.com \
  --cms drupal \
  --git-repo git@github.com:company/site.git \
  --le-email admin@example.com
```

### 3. Separate Environments

```bash
# Never mix production and development
sudo svp setup example.com \
  --cms drupal \
  --extra-domains "staging.example.com"
```

### 4. Regular Backups

```bash
# Automated backup script
#!/bin/bash
DOMAIN="example.com"
BACKUP_DIR="/backups/$(date +%Y%m%d)"

mkdir -p $BACKUP_DIR

# Database
drush-$DOMAIN sql-dump --gzip > $BACKUP_DIR/$DOMAIN-db.sql.gz

# Files
tar -czf $BACKUP_DIR/$DOMAIN-files.tar.gz \
  /var/www/$DOMAIN/web/sites/default/files

# Code (if not in Git)
tar -czf $BACKUP_DIR/$DOMAIN-code.tar.gz \
  --exclude='sites/default/files' \
  /var/www/$DOMAIN
```

---

## More Help

- [Full Documentation]({{ site.baseurl }}/documentation/)
- [Command-Line Reference]({{ site.baseurl }}/documentation/command-line)
- [Troubleshooting Guide]({{ site.baseurl }}/documentation/troubleshooting)
- [GitHub Issues](https://github.com/willjackson/simple-vps-provisioner/issues)

---

[← Back to Home]({{ site.baseurl }}/) | [Documentation →]({{ site.baseurl }}/documentation/)
