---
layout: default
title: SSL/HTTPS Configuration
---

# SSL/HTTPS Configuration

Complete guide to SSL/HTTPS setup with Let's Encrypt and svp.

## Overview

svp automatically configures SSL/HTTPS using Let's Encrypt certificates. This provides:

- ✅ Free SSL certificates
- ✅ Automatic renewal
- ✅ A+ SSL rating (modern ciphers)
- ✅ HTTP to HTTPS redirect
- ✅ HSTS (HTTP Strict Transport Security)

---

## Basic SSL Setup

### Default Behavior

**SSL is disabled by default.** To enable SSL, provide an email address with `--le-email`:

```bash
sudo svp setup example.com \
  --cms drupal \
  --le-email admin@example.com
```

When `--le-email` is provided, this will:
1. Obtain certificate from Let's Encrypt
2. Configure Nginx for HTTPS
3. Set up automatic HTTP → HTTPS redirect
4. Enable secure SSL settings

### Prerequisites

Before obtaining SSL certificates:

1. **DNS must be configured** - Domain must point to your server
2. **Port 80 open** - Required for Let's Encrypt validation
3. **Valid email** - For expiration notices

**Verify DNS:**
```bash
dig +short example.com
# Should return your server IP
```

---

## SSL Options

### Enable SSL

To enable SSL, simply provide your email address:

```bash
sudo svp setup example.com \
  --cms drupal \
  --le-email admin@example.com
```

### Disable SSL (HTTP Only)

By default, SSL is disabled. Simply omit the `--le-email` flag:

```bash
sudo svp setup example.com \
  --cms drupal
```

**Use cases:**
- Local development
- Internal network sites
- Testing before DNS is ready

---

## Multi-Domain SSL

When using multiple domains, each gets its own certificate:

```bash
sudo svp setup example.com \
  --cms drupal \
  --extra-domains "staging.example.com,dev.example.com" \
  --le-email admin@example.com
```

**Results in:**
- `example.com` - Separate certificate
- `staging.example.com` - Separate certificate
- `dev.example.com` - Separate certificate

**Tip:** Combine SSL with Basic Authentication to secure non-production environments:

```bash
# Password-protect staging while keeping SSL
sudo svp auth staging.example.com enable
sudo svp auth dev.example.com enable
```

This gives you both HTTPS encryption and password protection for staging/dev sites.

---

## Certificate Management

### View Certificate Details

```bash
# List all certificates
sudo certbot certificates

# View specific certificate
sudo certbot certificates -d example.com
```

### Manual Certificate Renewal

Certificates auto-renew, but you can force renewal:

```bash
# Renew using update-ssl (recommended)
sudo svp update-ssl example.com renew

# Or using certbot directly
# Renew specific certificate
sudo certbot renew --cert-name example.com

# Renew all certificates
sudo certbot renew
```

### Test Renewal

Test renewal without actually renewing:

```bash
sudo certbot renew --dry-run
```

### Check Auto-Renewal

Certbot automatically renews certificates via systemd timer:

```bash
# Check timer status
sudo systemctl status certbot.timer

# View timer schedule
sudo systemctl list-timers certbot.timer
```

---

## Managing SSL Certificates

### The update-ssl Command

Update or enable SSL certificates for existing sites using the `update-ssl` command:

```bash
sudo svp update-ssl example.com enable --le-email admin@example.com
```

**What it does:**
- Obtains or renews Let's Encrypt certificate
- Updates Nginx configuration for HTTPS
- Sets up HTTP to HTTPS redirect
- Configures secure SSL settings

**Examples:**
```bash
# Enable SSL for a site
sudo svp update-ssl example.com enable --le-email admin@example.com

# Check SSL status
sudo svp update-ssl example.com check

# Renew SSL certificate
sudo svp update-ssl example.com renew

# Disable SSL
sudo svp update-ssl example.com disable
```

### Adding SSL After Initial Setup

If you initially set up without SSL, you can add it later using either method:

**Method 1: Using update-ssl command (recommended)**
```bash
sudo svp update-ssl example.com enable --le-email admin@example.com
```

**Method 2: Using certbot directly**
```bash
sudo certbot --nginx -d example.com \
  --non-interactive \
  --agree-tos \
  --email admin@example.com
```

Both methods will:
1. Obtain certificate
2. Update Nginx configuration
3. Set up HTTPS redirect

---

## SSL Configuration Files

### Certificate Location

Certificates are stored in:
```
/etc/letsencrypt/live/example.com/
├── fullchain.pem   # Certificate + intermediate
├── privkey.pem     # Private key
├── cert.pem        # Certificate only
└── chain.pem       # Intermediate certificate
```

### Nginx SSL Configuration

SSL settings in Nginx vhost:

```nginx
# /etc/nginx/sites-available/example.com.conf

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name example.com;

    # SSL Certificate
    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;

    # SSL Settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256...';
    ssl_prefer_server_ciphers on;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=31536000" always;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # ... rest of config
}

# HTTP to HTTPS redirect
server {
    listen 80;
    listen [::]:80;
    server_name example.com;
    return 301 https://$server_name$request_uri;
}
```

---

## Troubleshooting

### Certificate Fails to Generate

**Error:** "Failed to obtain certificate"

**Common causes:**

1. **DNS not pointing to server**
   ```bash
   # Check DNS
   dig +short example.com

   # Should match your server IP
   curl -4 ifconfig.me
   ```

2. **Port 80 blocked**
   ```bash
   # Check firewall
   sudo ufw status

   # Allow port 80
   sudo ufw allow 80/tcp
   ```

3. **Domain validation failed**
   - Ensure domain is accessible via HTTP first
   - Check Nginx is running: `sudo systemctl status nginx`

**Solution:** Use the update-ssl command to retry:
```bash
sudo svp update-ssl example.com enable --le-email admin@example.com
```

### Rate Limit Errors

**Error:** "Too many certificates already issued"

**Solution:**
Let's Encrypt limits to 5 certificates per domain per week.

Wait 7 days or use staging environment for testing with certbot:

```bash
# Use certbot with staging for testing
sudo certbot --nginx -d example.com --staging
```

### Certificate Expired

Check certificate status:
```bash
sudo certbot certificates
```

If expired, renew:
```bash
# Using update-ssl (recommended)
sudo svp update-ssl example.com renew

# Or using certbot directly
sudo certbot renew --force-renewal
```

### Wrong Certificate Served

Check Nginx configuration:
```bash
# Test config
sudo nginx -t

# View SSL config for domain
sudo grep -A 10 "ssl_certificate" /etc/nginx/sites-available/example.com.conf
```

Restart Nginx:
```bash
sudo systemctl restart nginx
```

---

## Security Best Practices

### SSL Test

Test your SSL configuration:
- [SSL Labs Test](https://www.ssllabs.com/ssltest/)
- Should achieve A or A+ rating

### Force HTTPS

Ensure HTTP redirects to HTTPS:
```bash
curl -I http://example.com
# Should return: HTTP/1.1 301 Moved Permanently
# Location: https://example.com/
```

### Check Certificate Expiry

Monitor certificate expiration:
```bash
echo | openssl s_client -connect example.com:443 2>/dev/null | openssl x509 -noout -dates
```

Should show:
```
notBefore=Jan 15 00:00:00 2024 GMT
notAfter=Apr 15 23:59:59 2024 GMT
```

Certbot auto-renews at 30 days before expiry.

---

## Advanced Configuration

### Custom SSL Settings

To customize SSL settings, edit Nginx configuration:

```bash
sudo nano /etc/nginx/sites-available/example.com.conf
```

Then test and reload:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

### Wildcard Certificates

svp doesn't support wildcard certificates by default. For wildcard:

```bash
sudo certbot certonly --dns-route53 -d example.com -d '*.example.com'
```

Then manually update Nginx configuration.

### Multiple Domain Certificate

Request single certificate for multiple domains:

```bash
sudo certbot certonly --nginx \
  -d example.com \
  -d www.example.com \
  -d api.example.com
```

---

## Certificate Renewal

### Auto-Renewal

Certbot automatically renews certificates via systemd:

```bash
# Check auto-renewal service
sudo systemctl list-timers certbot.timer

# View renewal configuration
sudo cat /etc/letsencrypt/renewal/example.com.conf
```

### Manual Renewal Cron

Alternative manual cron setup (not needed with systemd):

```bash
# Add to crontab
sudo crontab -e

# Add this line
0 0 * * * certbot renew --quiet --post-hook "systemctl reload nginx"
```

### Renewal Hooks

Add hooks for post-renewal actions:

```bash
sudo nano /etc/letsencrypt/renewal/example.com.conf
```

Add:
```ini
[renewalparams]
renew_hook = systemctl reload nginx
```

---

## HTTP Only Sites

### Setting Up HTTP Only

For development or testing (simply omit --le-email):

```bash
sudo svp setup dev.example.local \
  --cms drupal
```

**Use when:**
- Testing before DNS is ready
- Local development (.local domains)
- Internal network sites
- Behind reverse proxy that handles SSL

### Adding SSL Later

Convert HTTP to HTTPS:

1. Configure DNS to point to server
2. Enable SSL using update-ssl:
   ```bash
   sudo svp update-ssl example.com enable --le-email admin@example.com
   ```

   Or using certbot directly:
   ```bash
   sudo certbot --nginx -d example.com
   ```

3. Verify:
   ```bash
   curl -I https://example.com
   ```

---

## Certificate Types

### DV (Domain Validation)

Let's Encrypt provides DV certificates:
- Validates domain ownership only
- Free and automated
- Suitable for most websites
- 90-day validity (auto-renews)

### EV/OV Certificates

For EV (Extended Validation) or OV (Organization Validation):
- Not provided by Let's Encrypt
- Purchase from commercial CA
- Manually install in Nginx
- Annual renewal required

---

[← CMS Installation](cms-installation) | [Multi-Domain Setup →](multi-domain)
