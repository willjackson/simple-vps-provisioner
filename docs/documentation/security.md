---
layout: default
title: Security Best Practices
---

# Security Best Practices

Security hardening and best practices for svp-provisioned sites.

## Overview

svp implements security by default, but additional hardening can improve protection:

- ✅ Firewall configured
- ✅ SSL/HTTPS enforced
- ✅ PHP hardening
- ✅ File permissions
- ✅ Database security
- ✅ Regular updates

---

## Built-in Security Features

### What svp Provides

**1. UFW Firewall**
- SSH (22), HTTP (80), HTTPS (443) only
- All other ports blocked
- IPv4 and IPv6 configured

**2. SSL/HTTPS**
- Let's Encrypt certificates
- Modern TLS protocols only
- Strong cipher suites
- HSTS enabled
- HTTP → HTTPS redirect

**3. PHP Security**
- `expose_php = Off`
- `display_errors = Off`
- `allow_url_fopen = Off`
- `allow_url_include = Off`
- Per-domain process isolation

**4. Database Security**
- Unique passwords per site
- Localhost-only access
- Minimal privileges granted
- No remote access

**5. File Permissions**
- Proper ownership (admin:www-data)
- Restricted settings files (400/444)
- Protected credentials (600)

---

## Additional Security Hardening

### 1. SSH Security

#### Disable Root Login

```bash
sudo nano /etc/ssh/sshd_config
```

Set:
```
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
```

Restart SSH:
```bash
sudo systemctl restart sshd
```

#### SSH Key-Only Access

```bash
# On your local machine
ssh-keygen -t ed25519 -C "your-email@example.com"

# Copy to server
ssh-copy-id admin@your-server

# Test login
ssh admin@your-server

# Disable password auth (shown above)
```

#### Change SSH Port (Optional)

```bash
sudo nano /etc/ssh/sshd_config
```

Change:
```
Port 2222
```

Update firewall:
```bash
sudo ufw allow 2222/tcp
sudo ufw delete allow 22/tcp
sudo systemctl restart sshd
```

### 2. Fail2Ban

Install fail2ban to block brute force attempts:

```bash
sudo apt install fail2ban
```

Configure:
```bash
sudo nano /etc/fail2ban/jail.local
```

Add:
```ini
[sshd]
enabled = true
port = 22
maxretry = 3
bantime = 3600

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/*error.log

[nginx-noscript]
enabled = true
port = http,https
filter = nginx-noscript
logpath = /var/log/nginx/*access.log
```

Restart:
```bash
sudo systemctl restart fail2ban
```

Check bans:
```bash
sudo fail2ban-client status sshd
```

### 3. Automatic Security Updates

Enable unattended upgrades:

```bash
sudo apt install unattended-upgrades
sudo dpkg-reconfigure unattended-upgrades
```

Configure:
```bash
sudo nano /etc/apt/apt.conf.d/50unattended-upgrades
```

Ensure enabled:
```
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}-security";
};
```

---

## Web Application Security

### 1. Security Headers

Add to Nginx configuration:

```bash
sudo nano /etc/nginx/sites-available/example.com.conf
```

Add in server block:
```nginx
# Security Headers
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;
add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;
```

Test and reload:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

### 2. Rate Limiting

Prevent brute force attacks:

```bash
sudo nano /etc/nginx/nginx.conf
```

Add in http block:
```nginx
limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
```

In site config:
```nginx
location /user/login {
    limit_req zone=login burst=5 nodelay;
}
```

### 3. Hide Server Version

```bash
sudo nano /etc/nginx/nginx.conf
```

Add in http block:
```nginx
server_tokens off;
```

Hide PHP version (already done by svp):
```ini
expose_php = Off
```

---

## Database Security

### 1. Secure MariaDB

Run security script:
```bash
sudo mysql_secure_installation
```

Answer:
- Set root password: Yes
- Remove anonymous users: Yes
- Disallow root login remotely: Yes
- Remove test database: Yes
- Reload privilege tables: Yes

### 2. Database User Privileges

Verify minimal privileges:
```bash
mysql -e "SHOW GRANTS FOR 'drupal_example_com'@'localhost';"
```

Should only show:
```
GRANT USAGE ON *.* TO 'drupal_example_com'@'localhost'
GRANT ALL PRIVILEGES ON drupal_example_com.* TO 'drupal_example_com'@'localhost'
```

### 3. Backup Credentials Securely

Credentials in `/etc/svp/sites/` must be protected:
```bash
# Verify permissions
ls -la /etc/svp/sites/*.db.txt

# Should show -rw------- (600)
# Fix if needed
sudo chmod 600 /etc/svp/sites/*.db.txt
```

---

## Drupal Security

### 1. Security Modules

Install security modules:
```bash
cd /var/www/example.com
sudo -u admin composer require drupal/security_review
sudo -u admin composer require drupal/paranoia
drush-example.com en security_review paranoia -y
```

### 2. Security Review

Run security review:
```bash
drush-example.com security-review
```

Fix any issues found.

### 3. Restrict Admin Access

Restrict /admin to specific IPs:

```bash
sudo nano /etc/nginx/sites-available/example.com.conf
```

Add:
```nginx
location ~ ^/admin {
    allow 1.2.3.4;      # Your IP
    allow 5.6.7.8;      # Office IP
    deny all;
    try_files $uri /index.php?$query_string;
}
```

### 4. Update Regularly

```bash
# Check for updates
drush-example.com pm:security

# Update modules
cd /var/www/example.com
sudo -u admin composer update --with-dependencies
drush-example.com updb -y
drush-example.com cr
```

---

## WordPress Security

### 1. Security Plugins

Install security plugins:
```bash
cd /var/www/myblog.com
sudo -u admin wp plugin install wordfence --activate
sudo -u admin wp plugin install all-in-one-wp-security-and-firewall --activate
```

### 2. Disable File Editing

Add to `wp-config.php`:
```php
define('DISALLOW_FILE_EDIT', true);
```

### 3. Limit Login Attempts

```bash
sudo -u admin wp plugin install limit-login-attempts-reloaded --activate
```

### 4. Hide wp-login.php

Use a security plugin to change login URL.

### 5. Update Regularly

```bash
cd /var/www/myblog.com
sudo -u admin wp core update
sudo -u admin wp plugin update --all
sudo -u admin wp theme update --all
```

---

## File Security

### 1. Proper Permissions

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

### 2. Protect Sensitive Files

**Drupal - in Nginx config:**
```nginx
location ~ /\.ht {
    deny all;
}

location ~ /\. {
    deny all;
}

location ~ ^/sites/[^/]+/files/.*\.php$ {
    deny all;
}
```

**WordPress - in Nginx config:**
```nginx
location ~ /\.ht {
    deny all;
}

location ~* /(?:uploads|files)/.*\.php$ {
    deny all;
}
```

---

## Monitoring & Logging

### 1. Monitor Logs

**Nginx errors:**
```bash
sudo tail -f /var/log/nginx/error.log
sudo tail -f /var/log/nginx/example.com.error.log
```

**PHP errors:**
```bash
sudo tail -f /var/log/php8.3-fpm-example.com-error.log
```

**Auth attempts:**
```bash
sudo tail -f /var/log/auth.log
```

### 2. Log Rotation

Verify logrotate is configured:
```bash
cat /etc/logrotate.d/nginx
cat /etc/logrotate.d/php8.3-fpm
```

### 3. Intrusion Detection

Install AIDE:
```bash
sudo apt install aide
sudo aideinit
sudo mv /var/lib/aide/aide.db.new /var/lib/aide/aide.db
```

Check for changes:
```bash
sudo aide --check
```

---

## Backup Security

### 1. Encrypt Backups

```bash
# Backup and encrypt
tar -czf - /var/www/example.com | gpg --encrypt -r admin@example.com > backup.tar.gz.gpg

# Decrypt and restore
gpg --decrypt backup.tar.gz.gpg | tar -xzf -
```

### 2. Secure Backup Storage

- Store backups off-server
- Use encrypted storage
- Limit access to backups
- Test restore procedures

---

## Security Checklist

### Initial Setup

- [ ] SSH key-only authentication
- [ ] Disable root login
- [ ] UFW firewall enabled
- [ ] SSL/HTTPS configured
- [ ] Strong passwords for all users
- [ ] Database secured

### Ongoing Maintenance

- [ ] Regular updates (weekly)
- [ ] Monitor logs (daily)
- [ ] Security audits (monthly)
- [ ] Backup verification (monthly)
- [ ] Certificate renewal checks
- [ ] Review user access

### After Security Incident

- [ ] Review all logs
- [ ] Check file integrity
- [ ] Review user accounts
- [ ] Change all passwords
- [ ] Update all software
- [ ] Review firewall rules
- [ ] Consider reprovisioning

---

## Security Tools

### Scan for Vulnerabilities

**Drupal:**
```bash
drush-example.com pm:security
```

**WordPress:**
```bash
cd /var/www/myblog.com
sudo -u admin wp plugin list --field=update
```

### Test SSL Configuration

```bash
# Command line
echo | openssl s_client -connect example.com:443 -servername example.com

# Online test
# https://www.ssllabs.com/ssltest/
```

### Port Scanning

Check what ports are exposed:
```bash
sudo nmap -sS -p 1-65535 localhost
```

Should only show: 22 (SSH), 80 (HTTP), 443 (HTTPS)

---

## Incident Response

### If Site is Compromised

1. **Isolate site:**
   ```bash
   sudo systemctl stop nginx
   sudo systemctl stop php8.3-fpm
   ```

2. **Review logs:**
   ```bash
   sudo grep -r "suspicious" /var/log/nginx/
   sudo grep -r "error" /var/log/php*/
   ```

3. **Check files:**
   ```bash
   sudo find /var/www -name "*.php" -mtime -7
   ```

4. **Restore from backup:**
   ```bash
   # Restore known good backup
   ```

5. **Reprovision if needed:**
   ```bash
   sudo svp setup example.com --cms drupal
   ```

6. **Change all passwords:**
   - Database
   - Admin users
   - SSH keys
   - API keys

---

[← Reprovisioning](reprovisioning) | [Performance Tuning →](performance)
