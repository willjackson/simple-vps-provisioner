---
layout: default
title: Troubleshooting
---

# Troubleshooting

Common issues and their solutions when using Simple VPS Provisioner.

## Installation Issues

### "Permission denied" errors

**Problem:** Installation fails with permission denied errors.

**Solution:** Run with sudo:
```bash
sudo bash install.sh
```

### "Command not found: svp"

**Problem:** After installation, `svp` command is not found.

**Solution:**
1. Verify installation:
   ```bash
   which svp
   ls -l /usr/local/bin/svp
   ```

2. If not found, manually copy:
   ```bash
   sudo cp svp /usr/local/bin/
   sudo chmod +x /usr/local/bin/svp
   ```

## SSL/HTTPS Issues

### Let's Encrypt certificate failure

**Problem:** SSL certificate generation fails.

**Causes & Solutions:**

1. **DNS not propagated**
   - Wait 15-30 minutes after DNS changes
   - Verify DNS: `dig +short example.com`
   - Should return your server IP

2. **Port 80 blocked**
   - Check firewall: `sudo ufw status`
   - Temporarily allow: `sudo ufw allow 80`

3. **Rate limiting**
   - Let's Encrypt has rate limits (5 certs/week per domain)
   - Wait 7 days or use certbot staging for testing:
   ```bash
   sudo certbot --nginx -d example.com --staging
   ```

4. **Invalid email**
   - Ensure `--le-email` is provided with a valid email address
   - Example: `admin@example.com`

### "Failed to obtain certificate" with existing Nginx

**Problem:** Certificate fails when Nginx already running.

**Solution 1: Use update-ssl command (recommended)**
```bash
# Retry SSL setup for existing site
sudo svp update-ssl example.com enable --le-email admin@example.com
```

**Solution 2: Manual approach**
```bash
# Stop Nginx temporarily
sudo systemctl stop nginx

# Run svp
sudo svp setup example.com --cms drupal --le-email admin@example.com

# Nginx will be restarted automatically
```

### Adding SSL to existing HTTP-only site

**Problem:** Site was set up without SSL, need to add it later.

**Solution:**
```bash
# Use update-ssl command
sudo svp update-ssl example.com enable --le-email admin@example.com

# This will:
# 1. Obtain Let's Encrypt certificate
# 2. Update Nginx configuration for HTTPS
# 3. Set up HTTP to HTTPS redirect
```

### SSL certificate needs renewal

**Problem:** Certificate expired or about to expire.

**Solution:**
```bash
# Renew using update-ssl
sudo svp update-ssl example.com renew

# Or check auto-renewal status
sudo systemctl status certbot.timer
```

## Database Issues

### "Access denied for user"

**Problem:** Database connection fails.

**Solution:**
Check credentials in `/etc/svp/sites/example.com.db.txt`:
```bash
sudo cat /etc/svp/sites/example.com.db.txt
```

Use these credentials in your CMS settings.php or wp-config.php.

### Database import fails

**Problem:** `-db` flag doesn't import database.

**Solution:**
1. Verify file exists and has correct permissions
2. Supported formats: `.sql`, `.sql.gz`
3. Manual import:
   ```bash
   # Get credentials
   source /etc/svp/sites/example.com.db.txt

   # Import
   gunzip < backup.sql.gz | mysql -u $USERNAME -p$PASSWORD $DATABASE
   ```

## PHP Issues

### Wrong PHP version active

**Problem:** Site uses old PHP version after update.

**Solution:**
```bash
# Update PHP version for domain
sudo svp php-update example.com --php-version 8.4

# Restart PHP-FPM
sudo systemctl restart php8.4-fpm

# Verify
php -v
```

### PHP extensions missing

**Problem:** Module required by CMS not loaded.

**Solution:**
Common extensions are installed automatically. For additional ones:
```bash
# Example: GD library
sudo apt install php8.3-gd

# Restart PHP-FPM
sudo systemctl restart php8.3-fpm
```

## Git Deployment Issues

### "Permission denied (publickey)"

**Problem:** Git clone fails with SSH key error.

**Solution:**
1. Use HTTPS instead:
   ```bash
   --git-repo https://github.com/user/repo.git
   ```

2. Or configure SSH key:
   ```bash
   # Generate key
   ssh-keygen -t ed25519

   # Add to GitHub
   cat ~/.ssh/id_ed25519.pub
   ```

### Git clone hangs or times out

**Problem:** Clone appears to hang.

**Solution:**
- Check repository is accessible
- Verify network connectivity
- Try smaller repository first
- Use `DEBUG=1` to see progress:
  ```bash
  DEBUG=1 sudo svp setup --cms drupal --domain example.com --git-repo https://...
  ```

## Firewall Issues

### Can't access site after provisioning

**Problem:** Site unreachable via browser.

**Solution:**
Check firewall rules:
```bash
# Check status
sudo ufw status verbose

# Should show:
# 22/tcp  ALLOW   (SSH)
# 80/tcp  ALLOW   (HTTP)
# 443/tcp ALLOW   (HTTPS)

# If missing, add:
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

### SSH locked out after setup

**Problem:** Can't SSH after svp runs.

**Solution:**
UFW is configured to allow SSH by default. If locked out:
- Contact your VPS provider to access via console
- Disable UFW: `sudo ufw disable`
- Re-enable with SSH: `sudo ufw allow 22 && sudo ufw enable`

---

## Basic Authentication Issues

### Site prompts for password unexpectedly

**Problem:** Browser prompts for username/password when accessing site.

**Solution:**
Basic Authentication is enabled. Check status:
```bash
sudo svp auth example.com check
```

If you don't want authentication, disable it:
```bash
sudo svp auth example.com disable
```

### Can't log in with credentials

**Problem:** Username/password doesn't work.

**Causes & Solutions:**

1. **Wrong credentials**
   - Double-check username and password
   - Passwords are case-sensitive

2. **Need to reset credentials**
   ```bash
   # Re-enable with new credentials
   sudo svp auth example.com enable \
     --username newuser \
     --password newpassword
   ```

3. **Check .htpasswd file exists**
   ```bash
   ls -la /var/www/example.com/.htpasswd
   sudo cat /var/www/example.com/.htpasswd
   ```

### Authentication not working after enabling

**Problem:** Enabled auth but site still accessible without password.

**Solution:**

1. Verify authentication is enabled:
   ```bash
   sudo svp auth example.com check
   ```

2. Check Nginx configuration:
   ```bash
   sudo cat /etc/nginx/sites-available/example.com.conf | grep auth_basic
   ```

3. Reload Nginx:
   ```bash
   sudo nginx -t
   sudo systemctl reload nginx
   ```

4. Clear browser cache or try incognito mode

### Remove authentication completely

**Problem:** Need to remove all authentication.

**Solution:**
```bash
# Disable authentication
sudo svp auth example.com disable

# Verify it's removed
sudo svp auth example.com check

# Should show: Authentication is not enabled
```

## CMS-Specific Issues

### Drupal: "Settings file is not writable"

**Problem:** Drupal installation can't write settings.php.

**Solution:**
```bash
# Fix permissions
cd /var/www/example.com
sudo chmod 666 web/sites/default/settings.php
sudo chmod 777 web/sites/default

# After install, secure:
sudo chmod 444 web/sites/default/settings.php
sudo chmod 555 web/sites/default
```

### WordPress: "Error establishing database connection"

**Problem:** WordPress can't connect to database.

**Solution:**
1. Get correct credentials:
   ```bash
   sudo cat /etc/svp/sites/example.com.db.txt
   ```

2. Update `wp-config.php`:
   ```bash
   cd /var/www/example.com
   nano wp-config.php

   # Update these lines:
   define('DB_NAME', 'wordpress_example_com');
   define('DB_USER', 'wordpress_example_com');
   define('DB_PASSWORD', 'your_password_from_db_txt');
   define('DB_HOST', 'localhost');
   ```

## Performance Issues

### Site loads slowly

**Problem:** Website response time is slow.

**Solutions:**

1. **Enable OPcache** (should be default):
   ```bash
   # Check if enabled
   php -i | grep opcache.enable

   # Should show: opcache.enable => On
   ```

2. **Increase PHP-FPM workers**:
   ```bash
   # Edit pool config
   sudo nano /etc/php/8.3/fpm/pool.d/example.com.conf

   # Increase these values:
   pm.max_children = 20
   pm.start_servers = 5
   pm.min_spare_servers = 5
   pm.max_spare_servers = 10

   # Restart
   sudo systemctl restart php8.3-fpm
   ```

3. **Check resources**:
   ```bash
   # Memory usage
   free -h

   # CPU usage
   top

   # Disk I/O
   iostat
   ```

### High memory usage

**Problem:** Server running out of RAM.

**Solution:**
```bash
# Check what's using memory
sudo ps aux --sort=-%mem | head

# Optimize PHP-FPM if needed
# Reduce pm.max_children based on available RAM
# Formula: max_children = (RAM - 500MB) / 50MB
```

## Verification & Debugging

### Check service status

```bash
# Nginx
sudo systemctl status nginx

# PHP-FPM
sudo systemctl status php8.3-fpm

# MariaDB
sudo systemctl status mariadb

# UFW
sudo ufw status
```

### View logs

```bash
# Nginx access log
sudo tail -f /var/log/nginx/access.log

# Nginx error log
sudo tail -f /var/log/nginx/error.log

# PHP-FPM error log
sudo tail -f /var/log/php8.3-fpm.log

# Site-specific error log
sudo tail -f /var/log/nginx/example.com.error.log
```

### Debug mode

Run svp with debug output:
```bash
DEBUG=1 sudo svp setup example.com --cms drupal --le-email admin@example.com
```

This shows detailed command execution and can help identify where issues occur.

## Getting Help

If you're still experiencing issues:

1. **Check logs** using commands above
2. **Run in debug mode** with `DEBUG=1`
3. **Verify system requirements** match supported OS versions
4. **Search existing issues**: [GitHub Issues](https://github.com/willjackson/simple-vps-provisioner/issues)
5. **Create new issue**: Include:
   - OS version: `cat /etc/os-release`
   - svp version: `svp --version`
   - Command used
   - Error messages
   - Relevant log output

---

[← Back to Documentation]({{ site.baseurl }}/documentation/) | [Examples →]({{ site.baseurl }}/examples/)
