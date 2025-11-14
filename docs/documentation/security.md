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
- Let's Encrypt certificates (when --le-email is provided)
- Modern TLS protocols only
- Strong cipher suites
- HSTS enabled
- HTTP → HTTPS redirect
- SSL disabled by default for flexibility

**3. Basic Authentication**
- Password-protect entire sites with HTTP Basic Auth
- Perfect for staging/development environments
- Bcrypt-hashed password storage
- Simple enable/disable management
- One-command setup

**4. PHP Security**
- `expose_php = Off`
- `display_errors = Off`
- `allow_url_fopen = Off`
- `allow_url_include = Off`
- Per-domain process isolation

**5. Database Security**
- Unique passwords per site
- Localhost-only access
- Minimal privileges granted
- No remote access

**6. File Permissions**
- Proper ownership (admin:www-data)
- Restricted settings files (400/444)
- Protected credentials (600)

---

## Basic Authentication for Sites

### Overview

HTTP Basic Authentication provides a simple, effective way to password-protect entire sites. This is particularly useful for:

- **Staging environments** - Prevent public access while developing
- **Development sites** - Protect works-in-progress
- **Client previews** - Share private previews securely
- **Internal tools** - Restrict access to team members

### Using the auth Command

svp makes it simple to enable, disable, or check authentication status for any domain.

#### Enable Authentication (Interactive)

The easiest way is to enable interactively, which will prompt for credentials:

```bash
sudo svp auth staging.example.com enable
```

You'll be prompted to enter:
- Username
- Password (hidden while typing)

#### Enable Authentication (Non-Interactive)

For automation or scripting, provide credentials via flags:

```bash
sudo svp auth staging.example.com enable \
  --username admin \
  --password mySecurePassword123
```

#### Check Authentication Status

Verify if authentication is enabled for a domain:

```bash
sudo svp auth staging.example.com check
```

Output shows:
- Whether authentication is enabled
- Location of .htpasswd file
- Current configuration status

#### Disable Authentication

Remove authentication and allow public access:

```bash
sudo svp auth staging.example.com disable
```

This will:
- Remove authentication from Nginx configuration
- Delete the .htpasswd file
- Reload Nginx to apply changes immediately

### How It Works

**Behind the scenes:**
1. Installs `apache2-utils` if not present (for htpasswd command)
2. Creates a `.htpasswd` file in the site directory
3. Stores password using bcrypt hashing (secure)
4. Updates Nginx configuration with `auth_basic` directives
5. Reloads Nginx configuration (no restart needed)

**Security notes:**
- Passwords are hashed with bcrypt (strong protection)
- Only one username/password pair per domain
- Enabling again replaces existing credentials
- .htpasswd file has restricted permissions (600)

### Common Use Cases

#### Protect Staging Environment

```bash
# Set up staging with authentication from the start
sudo svp setup staging.mysite.com \
  --cms drupal \
  --le-email admin@mysite.com

# Enable authentication
sudo svp auth staging.mysite.com enable \
  --username client \
  --password preview2024
```

#### Secure Development Sites

```bash
# Enable auth on dev environment
sudo svp auth dev.mysite.com enable

# Share credentials with team
# Username: team
# Password: [as entered]
```

#### Temporary Protection

```bash
# Enable during development
sudo svp auth mysite.com enable

# Disable when ready for launch
sudo svp auth mysite.com disable
```

### Best Practices

**1. Use Strong Passwords**

Generate secure passwords:
```bash
# Generate random password
openssl rand -base64 16
```

**2. Combine with SSL/HTTPS**

Always use HTTPS when using Basic Auth to encrypt credentials in transit:
```bash
sudo svp setup staging.mysite.com \
  --cms drupal \
  --le-email admin@mysite.com

sudo svp auth staging.mysite.com enable
```

**3. Different Credentials per Environment**

Use unique passwords for each environment:
```bash
sudo svp auth staging.mysite.com enable --username staging --password staging123
sudo svp auth dev.mysite.com enable --username dev --password dev456
```

**4. Document Credentials Securely**

Store credentials in a password manager, not in plain text files or emails.

**5. Remove When Not Needed**

Disable authentication when moving to production:
```bash
sudo svp auth mysite.com disable
```

### Limitations

- Only one username/password combination per domain
- Authentication applies to the entire site (not per-page)
- Basic Auth is not suitable for production user authentication
- Use CMS authentication (Drupal/WordPress login) for actual users

### Troubleshooting

**Authentication Not Working**

1. Check if enabled:
   ```bash
   sudo svp auth example.com check
   ```

2. Verify Nginx configuration:
   ```bash
   sudo nginx -t
   sudo cat /etc/nginx/sites-available/example.com.conf | grep auth_basic
   ```

3. Check .htpasswd file exists:
   ```bash
   ls -la /var/www/example.com/.htpasswd
   ```

**Forgot Password**

Simply re-enable authentication with new credentials:
```bash
sudo svp auth example.com enable --username admin --password newPassword123
```

**Remove Old Credentials**

Disable and re-enable:
```bash
sudo svp auth example.com disable
sudo svp auth example.com enable
```

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
- [ ] SSL/HTTPS configured (use --le-email or update-ssl command)
- [ ] Strong passwords for all users
- [ ] Database secured
- [ ] Basic Auth enabled for staging/dev sites (use auth command)

### Ongoing Maintenance

- [ ] Regular updates (weekly)
- [ ] Monitor logs (daily)
- [ ] Security audits (monthly)
- [ ] Backup verification (monthly)
- [ ] Certificate renewal checks
- [ ] Review user access
- [ ] Review and update Basic Auth credentials

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

### Enable or Update SSL

For sites without SSL or to renew certificates:

```bash
# Enable SSL for a site
sudo svp update-ssl example.com enable --le-email admin@example.com

# Check SSL status
sudo svp update-ssl example.com check

# Renew certificate
sudo svp update-ssl example.com renew
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
