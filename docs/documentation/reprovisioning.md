---
layout: default
title: Reprovisioning Sites
---

# Reprovisioning Sites

Guide to updating, rebuilding, and reprovisioning existing sites with svp.

## Overview

Reprovisioning allows you to:

- ✅ Rebuild a site from scratch
- ✅ Update configuration
- ✅ Switch CMS versions
- ✅ Change PHP versions
- ✅ Reset to clean state

---

## When to Reprovision

### Common Scenarios

1. **Testing a clean install** - Reset site to fresh state
2. **Major updates** - Upgrade Drupal/WordPress major version
3. **Configuration changes** - Update PHP version, domain, or settings
4. **Troubleshooting** - Fix corrupted installations
5. **Development workflow** - Quickly reset test environments

---

## Reprovision with Fresh Database

### Default Behavior

Running svp on an existing domain creates a **completely fresh** installation:

```bash
sudo svp setup example.com --cms drupal
```

**Prompt appears:**
```
Directory /var/www/example.com is not empty
Delete and reprovision? [y/N]:
```

**If you answer `y`:**
1. ✅ Entire site directory deleted
2. ✅ Database dropped completely
3. ✅ Database user deleted
4. ✅ **NEW** database credentials generated
5. ✅ Fresh CMS installation
6. ✅ New SSL certificate (if needed)

**Result:**
- Clean slate
- New secure credentials
- Maximum security

---

## Reprovision Keeping Database

### Keep Existing Credentials

Use `--keep-existing-db` flag to reuse database credentials:

```bash
sudo svp setup example.com --cms drupal --keep-existing-db
```

**What happens:**
1. ✅ Site directory deleted and recreated
2. ✅ Database tables **dropped**
3. ✅ **Same** database and user kept
4. ✅ **Same** credentials reused
5. ✅ Fresh CMS installation
6. ✅ Configuration uses existing credentials

**Use when:**
- You want stable credentials for scripts
- External services depend on fixed credentials
- Testing reinstalls frequently
- Keeping same database connection strings

---

## Reprovision Workflows

### Fresh Start (Complete Reset)

```bash
# Complete clean slate
sudo svp setup example.com --cms drupal

# When prompted
Delete and reprovision? [y/N]: y
```

**Perfect for:**
- Production site initial setup
- Maximum security (new credentials)
- Moving away from compromised credentials

### Test Environment (Keep Credentials)

```bash
# Keep credentials for testing
sudo svp setup dev.example.com --cms drupal --keep-existing-db
```

**Perfect for:**
- Development environments
- Rapid testing cycles
- CI/CD pipelines
- Keeping scripts working

### Partial Update

Update only specific components:

**Change PHP version:**
```bash
sudo svp php-update example.com --php-version 8.4
```

**Update SSL certificate:**
```bash
# Using update-ssl command (recommended)
sudo svp update-ssl example.com --le-email admin@example.com --force-renewal

# Or using certbot directly
sudo certbot --nginx -d example.com --force-renewal
```

---

## Backup Before Reprovisioning

### Always Backup First!

Before reprovisioning, backup:

1. **Database:**
   ```bash
   drush-example.com sql-dump --gzip > backup-$(date +%Y%m%d).sql.gz
   ```

2. **Files:**
   ```bash
   tar -czf files-backup-$(date +%Y%m%d).tar.gz \
     /var/www/example.com/web/sites/default/files/
   ```

3. **Configuration:**
   ```bash
   cd /var/www/example.com
   drush-example.com cex
   tar -czf config-backup-$(date +%Y%m%d).tar.gz config/sync/
   ```

4. **Custom code:**
   ```bash
   tar -czf custom-backup-$(date +%Y%m%d).tar.gz \
     /var/www/example.com/web/modules/custom/ \
     /var/www/example.com/web/themes/custom/
   ```

---

## Reprovision with Data Import

### Import Database During Reprovision

```bash
sudo svp setup example.com \
  --cms drupal \
  --db /path/to/backup.sql.gz
```

**Process:**
1. Creates fresh site structure
2. Imports your database
3. Skips site-install
4. Ready to use

### Import Database After Reprovision

```bash
# 1. Reprovision fresh
sudo svp setup --cms drupal --domain example.com

# 2. Import database
drush-example.com sql-drop -y
zcat /path/to/backup.sql.gz | drush-example.com sql-cli

# 3. Restore files
rsync -avz /backup/files/ /var/www/example.com/web/sites/default/files/

# 4. Clear cache
drush-example.com cr
```

---

## Update Existing Installation

### Without Reprovisioning

Update components without full reprovision:

**Update Drupal core:**
```bash
cd /var/www/example.com
sudo -u admin composer update drupal/core-recommended --with-dependencies
drush-example.com updb
drush-example.com cr
```

**Update WordPress:**
```bash
cd /var/www/myblog.com
sudo -u admin wp core update
sudo -u admin wp core update-db
```

**Update PHP version:**
```bash
sudo svp php-update --domain example.com --php-version 8.4
```

---

## Reprovision Checklist

### Before Reprovisioning

- [ ] Backup database
- [ ] Backup files directory
- [ ] Export configuration (Drupal)
- [ ] Backup custom code
- [ ] Document current settings
- [ ] Note current credentials
- [ ] Inform users of downtime

### During Reprovisioning

- [ ] Run svp setup command
- [ ] Monitor for errors
- [ ] Verify services start
- [ ] Check SSL certificate

### After Reprovisioning

- [ ] Import database (if needed)
- [ ] Restore files
- [ ] Import configuration (Drupal)
- [ ] Test site functionality
- [ ] Check permissions
- [ ] Verify credentials work
- [ ] Update external services if credentials changed

---

## Common Reprovision Scenarios

### Scenario 1: Upgrade Drupal Major Version

```bash
# 1. Backup everything
drush-example.com sql-dump --gzip > d9-backup.sql.gz
tar -czf d9-files.tar.gz /var/www/example.com/web/sites/default/files/

# 2. Reprovision with new Drupal (will be latest version)
sudo svp setup --cms drupal --domain example.com

# 3. Don't import old database - it's D9, new site is D10
# Instead: migrate or manual upgrade
```

### Scenario 2: Switch from HTTP to HTTPS

```bash
# Reprovision from HTTP to HTTPS
sudo svp setup example.com \
  --cms drupal \
  --le-email admin@example.com \
  --db /path/to/backup.sql.gz
```

Or add SSL after reprovisioning:
```bash
# First reprovision without SSL
sudo svp setup example.com \
  --cms drupal \
  --db /path/to/backup.sql.gz

# Then add SSL
sudo svp update-ssl example.com --le-email admin@example.com
```

### Scenario 3: Fix Corrupted Installation

```bash
# Fresh start
sudo svp setup example.com --cms drupal

# Import known good database
zcat /path/to/backup.sql.gz | drush-example.com sql-cli

# Restore files
rsync -avz /backup/files/ /var/www/example.com/web/sites/default/files/
```

### Scenario 4: Reset Development Environment

```bash
# Quick reset keeping credentials
sudo svp setup dev.example.com --cms drupal --keep-existing-db

# Answer yes to prompt
Delete and reprovision? [y/N]: y
```

---

## Automation

### Automated Reprovision Script

Create `/usr/local/bin/reprovision-dev.sh`:

```bash
#!/bin/bash
set -e

DOMAIN="dev.example.com"
PROD_DOMAIN="example.com"
BACKUP_DIR="/backups"

echo "=== Reprovisioning $DOMAIN from production ==="

# 1. Export production database
echo "Exporting production database..."
drush-$PROD_DOMAIN sql-dump --gzip > $BACKUP_DIR/prod-export.sql.gz

# 2. Reprovision dev (keeps credentials)
echo "Reprovisioning dev environment..."
sudo svp setup $DOMAIN --cms drupal --keep-existing-db <<< "y"

# 3. Import production database
echo "Importing production database..."
zcat $BACKUP_DIR/prod-export.sql.gz | drush-$DOMAIN sql-cli

# 4. Sync files
echo "Syncing files..."
rsync -avz --delete \
  /var/www/$PROD_DOMAIN/web/sites/default/files/ \
  /var/www/$DOMAIN/web/sites/default/files/

# 5. Clear cache
echo "Clearing cache..."
drush-$DOMAIN cr

echo "=== Reprovision complete! ==="
```

Make executable:
```bash
sudo chmod +x /usr/local/bin/reprovision-dev.sh
```

Run:
```bash
sudo /usr/local/bin/reprovision-dev.sh
```

---

## Troubleshooting Reprovision

### Reprovision Stuck

If reprovision hangs:

```bash
# Check processes
ps aux | grep svp

# Check logs
sudo journalctl -f

# If stuck, kill and retry
sudo pkill -9 svp
```

### Database Credentials Changed

If you reprovisioned without `--keep-existing-db`:

1. **New credentials in:**
   ```bash
   sudo cat /etc/svp/sites/example.com.db.txt
   ```

2. **Update external services** with new credentials

3. **Update settings files** if manually modified

### Files Restored But Not Showing

Check permissions:
```bash
cd /var/www/example.com/web/sites/default/files
sudo chown -R admin:www-data .
sudo find . -type d -exec chmod 775 {} \;
sudo find . -type f -exec chmod 664 {} \;
```

---

## Best Practices

### 1. Always Backup First

Never reprovision without backing up:
- Database
- Files
- Custom code
- Configuration

### 2. Test on Staging

Test reprovision process on staging before production:
```bash
# Staging test
sudo svp setup staging.example.com --cms drupal

# If successful, proceed to production
```

### 3. Use Keep-Existing-DB for Development

For dev environments that reset frequently:
```bash
sudo svp setup dev.example.com --cms drupal --keep-existing-db
```

### 4. Document Changes

Keep a log of:
- When reprovisioned
- Why reprovisioned
- What changed (credentials, PHP version, etc.)
- Any issues encountered

### 5. Verify After Reprovision

Check list:
- [ ] Site loads
- [ ] Database connects
- [ ] Files display correctly
- [ ] Drush/WP-CLI works
- [ ] SSL certificate valid
- [ ] Cron jobs work
- [ ] External services connect

---

[← Configuration Files](configuration-files) | [Security Best Practices →](security)
