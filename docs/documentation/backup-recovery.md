---
layout: default
title: Backup & Recovery
---

# Backup & Recovery

Complete guide to backing up and restoring svp-provisioned sites.

## Overview

Critical data to backup:

- ✅ Database
- ✅ Files directory
- ✅ Custom code
- ✅ Configuration
- ✅ SSL certificates
- ✅ Server configuration

---

## What to Backup

### Priority 1: Database

**Contains:**
- All content
- User accounts
- Configuration (Drupal)
- Comments and data

**Frequency:** Daily

### Priority 2: Files

**Contains:**
- Uploaded images
- Documents
- User files
- Media

**Frequency:** Daily

### Priority 3: Custom Code

**Contains:**
- Custom modules/plugins
- Custom themes
- Custom scripts

**Frequency:** After changes

### Priority 4: Configuration

**Contains:**
- Nginx configs
- PHP-FPM pools
- SSL certificates
- Site settings

**Frequency:** After changes

---

## Database Backup

### Automated Database Backup

Create `/usr/local/bin/backup-database.sh`:

```bash
#!/bin/bash
DOMAIN="example.com"
BACKUP_DIR="/backups/databases"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p $BACKUP_DIR

# Backup database
echo "Backing up database for $DOMAIN..."
drush-$DOMAIN sql-dump --gzip > $BACKUP_DIR/$DOMAIN-$DATE.sql.gz

# Keep only last N days
find $BACKUP_DIR -name "$DOMAIN-*.sql.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup complete: $BACKUP_DIR/$DOMAIN-$DATE.sql.gz"
```

Make executable:
```bash
sudo chmod +x /usr/local/bin/backup-database.sh
```

Schedule daily:
```bash
sudo crontab -e

# Daily at 2 AM
0 2 * * * /usr/local/bin/backup-database.sh
```

### Manual Database Backup

**Drupal:**
```bash
# Compressed
drush-example.com sql-dump --gzip > backup-$(date +%Y%m%d).sql.gz

# Uncompressed
drush-example.com sql-dump > backup.sql
```

**WordPress:**
```bash
cd /var/www/myblog.com

# Export
sudo -u admin wp db export backup-$(date +%Y%m%d).sql

# Compress
gzip backup-$(date +%Y%m%d).sql
```

**Direct MySQL:**
```bash
source /etc/svp/sites/example.com.db.txt
mysqldump -u "$Username" -p"$Password" "$Database" | gzip > backup.sql.gz
```

---

## Files Backup

### Files Directory Only

**Drupal:**
```bash
tar -czf files-backup-$(date +%Y%m%d).tar.gz \
  /var/www/example.com/web/sites/default/files/
```

**WordPress:**
```bash
tar -czf uploads-backup-$(date +%Y%m%d).tar.gz \
  /var/www/myblog.com/wp-content/uploads/
```

### Entire Site

```bash
tar -czf site-backup-$(date +%Y%m%d).tar.gz \
  --exclude='*/files/*' \
  --exclude='*/uploads/*' \
  /var/www/example.com/
```

### Automated Files Backup

Create `/usr/local/bin/backup-files.sh`:

```bash
#!/bin/bash
DOMAIN="example.com"
BACKUP_DIR="/backups/files"
RETENTION_DAYS=7
DATE=$(date +%Y%m%d)

mkdir -p $BACKUP_DIR

# Backup files
echo "Backing up files for $DOMAIN..."
tar -czf $BACKUP_DIR/$DOMAIN-files-$DATE.tar.gz \
  /var/www/$DOMAIN/web/sites/default/files/

# Keep only last N days
find $BACKUP_DIR -name "$DOMAIN-files-*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup complete: $BACKUP_DIR/$DOMAIN-files-$DATE.tar.gz"
```

---

## Complete Site Backup

### All-in-One Backup Script

Create `/usr/local/bin/backup-site.sh`:

```bash
#!/bin/bash
set -e

DOMAIN="example.com"
BACKUP_DIR="/backups/$DOMAIN"
DATE=$(date +%Y%m%d-%H%M%S)
SITE_DIR="/var/www/$DOMAIN"

mkdir -p $BACKUP_DIR/$DATE

echo "=== Complete Backup for $DOMAIN ==="

# 1. Database
echo "Backing up database..."
drush-$DOMAIN sql-dump --gzip > $BACKUP_DIR/$DATE/database.sql.gz

# 2. Files
echo "Backing up files..."
tar -czf $BACKUP_DIR/$DATE/files.tar.gz \
  $SITE_DIR/web/sites/default/files/

# 3. Code (excluding vendor and files)
echo "Backing up code..."
tar -czf $BACKUP_DIR/$DATE/code.tar.gz \
  --exclude='vendor' \
  --exclude='web/sites/default/files' \
  $SITE_DIR/

# 4. Configuration
echo "Backing up configuration..."
if [ -d "$SITE_DIR/config/sync" ]; then
  tar -czf $BACKUP_DIR/$DATE/config.tar.gz $SITE_DIR/config/sync/
fi

# 5. Site configs
echo "Backing up site configs..."
tar -czf $BACKUP_DIR/$DATE/site-configs.tar.gz \
  /etc/nginx/sites-available/$DOMAIN.conf \
  /etc/php/8.3/fpm/pool.d/$DOMAIN.conf \
  /etc/svp/sites/$DOMAIN.*

# 6. SSL certificates
echo "Backing up SSL certificates..."
if [ -d "/etc/letsencrypt/live/$DOMAIN" ]; then
  tar -czf $BACKUP_DIR/$DATE/ssl.tar.gz /etc/letsencrypt/live/$DOMAIN/
fi

# Create manifest
cat > $BACKUP_DIR/$DATE/manifest.txt <<EOF
Backup Date: $DATE
Domain: $DOMAIN
Database: database.sql.gz
Files: files.tar.gz
Code: code.tar.gz
Config: config.tar.gz
Site Configs: site-configs.tar.gz
SSL: ssl.tar.gz
EOF

# Cleanup old backups (keep last 7 days)
find $BACKUP_DIR -maxdepth 1 -type d -mtime +7 -exec rm -rf {} \;

echo "=== Backup Complete ==="
echo "Location: $BACKUP_DIR/$DATE"
du -sh $BACKUP_DIR/$DATE
```

Make executable:
```bash
sudo chmod +x /usr/local/bin/backup-site.sh
```

Run:
```bash
sudo /usr/local/bin/backup-site.sh
```

---

## Remote Backup

### Backup to Remote Server

Using rsync:
```bash
#!/bin/bash
BACKUP_DIR="/backups"
REMOTE_USER="backup"
REMOTE_HOST="backup.example.com"
REMOTE_DIR="/backups/vps1"

# Sync backups to remote server
rsync -avz --delete \
  $BACKUP_DIR/ \
  $REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR/

echo "Remote backup complete"
```

### Cloud Storage

**AWS S3:**
```bash
#!/bin/bash
# Install: sudo apt install awscli

BUCKET="my-backups"
BACKUP_DIR="/backups"

aws s3 sync $BACKUP_DIR/ s3://$BUCKET/$(hostname)/
```

**Google Cloud Storage:**
```bash
#!/bin/bash
# Install: sudo apt install google-cloud-sdk

BUCKET="my-backups"
BACKUP_DIR="/backups"

gsutil -m rsync -r $BACKUP_DIR/ gs://$BUCKET/$(hostname)/
```

---

## Database Restore

### Restore Database

**Drupal:**
```bash
# Drop current database
drush-example.com sql-drop -y

# Import backup
zcat backup.sql.gz | drush-example.com sql-cli

# Update database
drush-example.com updb -y

# Clear cache
drush-example.com cr
```

**WordPress:**
```bash
cd /var/www/myblog.com

# Drop current database
sudo -u admin wp db reset --yes

# Import backup
zcat backup.sql.gz | sudo -u admin wp db cli

# Update URLs if needed
sudo -u admin wp search-replace 'oldsite.com' 'newsite.com'
```

**Direct MySQL:**
```bash
source /etc/svp/sites/example.com.db.txt
mysql -u "$Username" -p"$Password" "$Database" < backup.sql
```

---

## Files Restore

### Restore Files Directory

**Drupal:**
```bash
# Extract to temporary location
tar -xzf files-backup.tar.gz -C /tmp/

# Sync to site
rsync -avz --delete /tmp/var/www/example.com/web/sites/default/files/ \
  /var/www/example.com/web/sites/default/files/

# Fix permissions
cd /var/www/example.com/web/sites/default/files
sudo chown -R admin:www-data .
sudo find . -type d -exec chmod 775 {} \;
sudo find . -type f -exec chmod 664 {} \;
```

**WordPress:**
```bash
# Extract
tar -xzf uploads-backup.tar.gz -C /tmp/

# Sync
rsync -avz --delete /tmp/var/www/myblog.com/wp-content/uploads/ \
  /var/www/myblog.com/wp-content/uploads/

# Fix permissions
sudo chown -R admin:www-data /var/www/myblog.com/wp-content/uploads/
```

---

## Complete Site Restore

### Restore from Full Backup

```bash
#!/bin/bash
set -e

DOMAIN="example.com"
BACKUP_DATE="20240115-140000"
BACKUP_DIR="/backups/$DOMAIN/$BACKUP_DATE"
SITE_DIR="/var/www/$DOMAIN"

echo "=== Restoring $DOMAIN from $BACKUP_DATE ==="

# 1. Stop services
echo "Stopping services..."
sudo systemctl stop php8.3-fpm
sudo systemctl stop nginx

# 2. Restore database
echo "Restoring database..."
drush-$DOMAIN sql-drop -y
zcat $BACKUP_DIR/database.sql.gz | drush-$DOMAIN sql-cli

# 3. Restore files
echo "Restoring files..."
tar -xzf $BACKUP_DIR/files.tar.gz -C /

# 4. Restore code
echo "Restoring code..."
rm -rf $SITE_DIR
tar -xzf $BACKUP_DIR/code.tar.gz -C /

# 5. Restore configuration
echo "Restoring configuration..."
tar -xzf $BACKUP_DIR/config.tar.gz -C $SITE_DIR/

# 6. Restore site configs
echo "Restoring site configs..."
tar -xzf $BACKUP_DIR/site-configs.tar.gz -C /

# 7. Fix permissions
echo "Fixing permissions..."
cd $SITE_DIR
sudo chown -R admin:www-data .
sudo find . -type d -exec chmod 755 {} \;
sudo find . -type f -exec chmod 644 {} \;
sudo chmod 775 web/sites/default/files/

# 8. Restart services
echo "Restarting services..."
sudo systemctl start php8.3-fpm
sudo systemctl start nginx

# 9. Clear cache
echo "Clearing cache..."
drush-$DOMAIN cr

echo "=== Restore Complete ==="
```

---

## Disaster Recovery

### Server Failure Recovery

**1. Provision new server:**
```bash
# Get latest release
curl -fsSL https://raw.githubusercontent.com/willjackson/simple-vps-provisioner/main/install-from-github.sh | sudo bash

# Provision site (no database import yet)
sudo svp setup example.com --cms drupal --ssl=false
```

**2. Restore from backup:**
```bash
# Copy backup to new server
scp -r /backups/example.com/latest/ admin@new-server:/tmp/

# On new server
sudo /usr/local/bin/restore-site.sh
```

**3. Update DNS:**
```
A    example.com    NEW_SERVER_IP
```

**4. Enable SSL:**
```bash
sudo certbot --nginx -d example.com
```

---

## Backup Verification

### Test Restore Process

Regularly test your backups:

```bash
# Create test site
sudo svp setup test.example.com --cms drupal --ssl=false

# Restore backup to test site
drush-test.example.com sql-drop -y
zcat /backups/example.com/latest/database.sql.gz | drush-test.example.com sql-cli

# Verify test site works
curl http://test.example.com/

# Clean up
sudo rm -rf /var/www/test.example.com
```

### Backup Checklist

- [ ] Database backup completed
- [ ] Files backup completed
- [ ] Configuration backup completed
- [ ] Backups stored off-server
- [ ] Backups tested within last month
- [ ] Restoration procedure documented
- [ ] Recovery time acceptable

---

## Backup Automation

### Complete Backup System

**1. Create backup script:** (shown above)

**2. Schedule with cron:**
```bash
sudo crontab -e

# Daily database backup at 2 AM
0 2 * * * /usr/local/bin/backup-database.sh

# Weekly full backup on Sunday at 3 AM
0 3 * * 0 /usr/local/bin/backup-site.sh

# Monthly sync to remote storage
0 4 1 * * /usr/local/bin/sync-to-cloud.sh
```

**3. Monitor backups:**
Create `/usr/local/bin/check-backups.sh`:
```bash
#!/bin/bash
BACKUP_DIR="/backups"
MAX_AGE_HOURS=48

# Check if recent backup exists
RECENT=$(find $BACKUP_DIR -type f -mtime -2 | wc -l)

if [ $RECENT -eq 0 ]; then
  echo "WARNING: No recent backups found!"
  # Send alert email
  mail -s "Backup Alert" admin@example.com <<< "No backups in last 48 hours"
  exit 1
fi

echo "Backup check passed: $RECENT files found"
```

---

## Backup Best Practices

### 1. 3-2-1 Rule

- **3** copies of data
- **2** different storage types
- **1** copy off-site

### 2. Regular Testing

Test restore procedure monthly

### 3. Retention Policy

Example:
- Daily: Keep 7 days
- Weekly: Keep 4 weeks
- Monthly: Keep 12 months

### 4. Automated Monitoring

Alert if backups fail

### 5. Encrypted Backups

Encrypt sensitive data:
```bash
tar -czf - /var/www/example.com | \
  gpg --encrypt -r admin@example.com > backup.tar.gz.gpg
```

---

## Backup Storage

### Local Storage

- Fast restore
- Lower cost
- Risk: Server failure loses backups

### Remote Server

- Off-site protection
- Moderate cost
- Requires network bandwidth

### Cloud Storage

- Highly durable
- Scalable
- Pay per use
- Examples: AWS S3, Google Cloud Storage, Backblaze B2

---

## Troubleshooting

### Backup Fails

**Check disk space:**
```bash
df -h
```

**Check permissions:**
```bash
ls -la /backups/
```

**Check logs:**
```bash
sudo tail -f /var/log/syslog
```

### Restore Fails

**Verify backup integrity:**
```bash
tar -tzf backup.tar.gz
```

**Check database import:**
```bash
zcat backup.sql.gz | head -100
```

---

[← Performance Tuning](performance) | [Back to Documentation]({{ site.baseurl }}/documentation)
