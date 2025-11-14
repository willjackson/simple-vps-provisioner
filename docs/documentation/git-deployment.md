---
layout: default
title: Git Deployment
---

# Git Deployment

Deploy and manage sites from Git repositories with svp.

## Overview

svp can clone and deploy sites directly from Git repositories, perfect for:

- ✅ Version-controlled deployments
- ✅ Team collaboration
- ✅ CI/CD pipelines
- ✅ Easy rollbacks
- ✅ Branch-based workflows

---

## Basic Git Deployment

### Deploy from Repository

```bash
sudo svp setup \
  -cms drupal \
  -domain example.com \
  -git-repo https://github.com/myorg/mysite.git \
  -le-email admin@example.com
```

**What happens:**
1. Repository is cloned to `/var/www/example.com`
2. Dependencies installed via Composer
3. Database created and configured
4. Site installed (if no database imported)

### Specify Branch

```bash
sudo svp setup \
  -cms drupal \
  -domain example.com \
  -git-repo https://github.com/myorg/mysite.git \
  -git-branch production \
  -le-email admin@example.com
```

If `-git-branch` is not specified, the repository's default branch is used.

---

## Repository Structure

### Drupal Repository

**Required structure:**
```
mysite/
├── composer.json         # Required
├── composer.lock
├── web/                  # Document root
│   ├── index.php
│   ├── sites/default/
│   │   └── settings.php
│   ├── core/
│   ├── modules/
│   └── themes/
├── config/sync/          # Configuration
├── drush/
└── vendor/               # Created by composer install
```

**Important:**
- `composer.json` must be in repository root
- Drupal must be in `web/` directory
- Don't commit `vendor/` or `web/core/` - these are installed via Composer

### WordPress Repository

**Typical structure:**
```
wp-site/
├── index.php
├── wp-config.php         # Optional
├── wp-content/
│   ├── themes/
│   │   └── my-theme/
│   ├── plugins/
│   │   └── my-plugin/
│   └── uploads/
└── wp-*.php
```

**Notes:**
- You can include custom themes and plugins
- Don't commit core WordPress files if using Composer
- `wp-config.php` will be created if not present

---

## Authentication

### HTTPS (Public Repositories)

For public repos, use HTTPS:

```bash
-git-repo https://github.com/myorg/mysite.git
```

No authentication needed.

### SSH (Private Repositories)

For private repos, use SSH:

```bash
-git-repo git@github.com:myorg/mysite.git
```

**Setup SSH key:**

```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "server@example.com"

# View public key
cat ~/.ssh/id_ed25519.pub

# Add to GitHub/GitLab
# Settings → SSH Keys → Add New
```

**Test connection:**
```bash
ssh -T git@github.com
# Should show: Hi username! You've successfully authenticated...
```

### Personal Access Token

Alternative for HTTPS with private repos:

```bash
-git-repo https://TOKEN@github.com/myorg/mysite.git
```

**GitHub token:**
1. Settings → Developer settings → Personal access tokens
2. Generate new token
3. Select `repo` scope
4. Use in URL as shown above

---

## Branch Management

### Switch Branches

Change branch after initial setup:

```bash
cd /var/www/example.com
sudo -u admin git fetch
sudo -u admin git checkout develop
sudo -u admin composer install
drush-example.com updb
drush-example.com cr
```

### Multiple Environments

Different branches per environment:

```bash
# Production - main branch
sudo svp setup \
  -cms drupal \
  -domain example.com \
  -git-repo git@github.com:myorg/mysite.git \
  -git-branch main

# Staging - develop branch
sudo svp setup \
  -cms drupal \
  -domain staging.example.com \
  -git-repo git@github.com:myorg/mysite.git \
  -git-branch develop

# Development - develop branch (for feature testing)
sudo svp setup \
  -cms drupal \
  -domain dev.example.com \
  -git-repo git@github.com:myorg/mysite.git \
  -git-branch develop
```

---

## Updating Deployed Sites

### Pull Latest Changes

```bash
cd /var/www/example.com
sudo -u admin git pull
sudo -u admin composer install
drush-example.com updb
drush-example.com cim
drush-example.com cr
```

### Update Script

Create `/usr/local/bin/deploy-example.sh`:

```bash
#!/bin/bash
set -e

SITE="/var/www/example.com"
cd $SITE

echo "Pulling latest changes..."
sudo -u admin git pull

echo "Installing dependencies..."
sudo -u admin composer install --no-dev --optimize-autoloader

echo "Running database updates..."
drush-example.com updb -y

echo "Importing configuration..."
drush-example.com cim -y

echo "Clearing cache..."
drush-example.com cr

echo "Deployment complete!"
```

Make executable:
```bash
sudo chmod +x /usr/local/bin/deploy-example.sh
```

Run:
```bash
sudo /usr/local/bin/deploy-example.sh
```

---

## Monorepo Structure

### Drupal in Subdirectory

If Drupal is not in root:

```
monorepo/
├── backend/              # Drupal here
│   ├── composer.json
│   └── web/
├── frontend/             # React app
└── docs/
```

Use `-drupal-root` flag:

```bash
sudo svp setup \
  -cms drupal \
  -domain example.com \
  -git-repo https://github.com/myorg/monorepo.git \
  -drupal-root "backend" \
  -le-email admin@example.com
```

### Custom Document Root

Non-standard docroot path:

```bash
sudo svp setup \
  -cms drupal \
  -domain example.com \
  -git-repo https://github.com/myorg/mysite.git \
  -docroot "public" \
  -le-email admin@example.com
```

---

## Git Workflows

### Feature Branch Workflow

1. **Create feature branch:**
   ```bash
   git checkout -b feature/new-feature
   ```

2. **Develop locally**

3. **Push to origin:**
   ```bash
   git push origin feature/new-feature
   ```

4. **Deploy to test environment:**
   ```bash
   cd /var/www/dev.example.com
   sudo -u admin git fetch
   sudo -u admin git checkout feature/new-feature
   sudo -u admin composer install
   drush-dev.example.com cr
   ```

5. **Merge to develop after testing:**
   ```bash
   git checkout develop
   git merge feature/new-feature
   git push origin develop
   ```

6. **Deploy to staging:**
   ```bash
   cd /var/www/staging.example.com
   sudo -u admin git pull
   ```

### Gitflow Workflow

**Branches:**
- `main` - Production
- `develop` - Development
- `release/*` - Release candidates
- `feature/*` - Features
- `hotfix/*` - Emergency fixes

**Environment mapping:**
- Production → `main` branch
- Staging → `release/*` or `develop`
- Development → `develop` or `feature/*`

---

## Deployment with Database

### Deploy with Database Import

Clone repo and import existing database:

```bash
sudo svp setup \
  -cms drupal \
  -domain example.com \
  -git-repo https://github.com/myorg/mysite.git \
  -db /path/to/backup.sql.gz \
  -le-email admin@example.com
```

**Process:**
1. Clone repository
2. Install Composer dependencies
3. Create database
4. Import database dump
5. Configure settings
6. Don't run site-install

---

## .gitignore Best Practices

### Drupal .gitignore

```gitignore
# Ignore configuration file with database credentials
/web/sites/*/settings.svp.php

# Ignore paths that contain generated content
/web/sites/*/files/
/web/sites/*/private/

# Ignore Composer
/vendor/

# Ignore paths that contain user-generated content
/web/sites/default/files/

# Ignore drupal core
/web/core/
/web/modules/contrib/
/web/themes/contrib/
/web/profiles/contrib/

# Ignore temporary files
*.swp
*.swo
*~
.DS_Store
```

### WordPress .gitignore

```gitignore
# WordPress core (if using Composer)
/wp-admin/
/wp-includes/
/wp-*.php
/index.php
/license.txt
/readme.html

# Configuration
wp-config.php

# Content
/wp-content/uploads/
/wp-content/upgrade/
/wp-content/cache/

# Plugins/themes (if using Composer)
/wp-content/plugins/
/wp-content/themes/twentytwenty*/

# Keep custom plugins/themes
!/wp-content/plugins/my-plugin/
!/wp-content/themes/my-theme/

# Logs
*.log
```

---

## Rollback Procedures

### Rollback to Previous Commit

```bash
cd /var/www/example.com

# View commit history
sudo -u admin git log --oneline

# Rollback to specific commit
sudo -u admin git reset --hard abc123

# Update dependencies
sudo -u admin composer install

# Update database if needed
drush-example.com updb

# Clear cache
drush-example.com cr
```

### Rollback to Previous Tag

```bash
cd /var/www/example.com

# View tags
sudo -u admin git tag

# Checkout tag
sudo -u admin git checkout v1.2.0

# Update dependencies
sudo -u admin composer install
```

---

## Automated Deployments

### Webhook Deployment

Set up automatic deployments via webhooks:

**1. Create deploy script** (`/var/www/example.com/deploy.sh`):

```bash
#!/bin/bash
cd /var/www/example.com
sudo -u admin git pull
sudo -u admin composer install --no-dev
drush-example.com updb -y
drush-example.com cim -y
drush-example.com cr
```

**2. Setup webhook endpoint** (outside svp scope)

**3. Configure GitHub webhook:**
- Repository → Settings → Webhooks
- Payload URL: `https://example.com/deploy.php`
- Content type: `application/json`
- Events: `push`

### CI/CD Integration

**GitHub Actions example:**

```yaml
name: Deploy to Production

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy via SSH
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.HOST }}
          username: ${{ secrets.USERNAME }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd /var/www/example.com
            sudo -u admin git pull
            sudo -u admin composer install --no-dev
            drush-example.com updb -y
            drush-example.com cim -y
            drush-example.com cr
```

---

## Troubleshooting

### Git Pull Fails

**Error:** "Your local changes would be overwritten"

**Solution:**
```bash
cd /var/www/example.com
sudo -u admin git stash
sudo -u admin git pull
sudo -u admin git stash pop
```

### Permission Denied (SSH)

**Error:** "Permission denied (publickey)"

**Check:**
```bash
# Verify SSH key
ssh -T git@github.com

# Check key permissions
ls -la ~/.ssh/

# Key should be 600
chmod 600 ~/.ssh/id_ed25519
```

### Composer Install Fails

**Error:** "Your requirements could not be resolved"

**Solutions:**
```bash
# Update Composer
sudo composer self-update

# Clear cache
sudo -u admin composer clear-cache

# Try install
sudo -u admin composer install
```

---

[← Multi-Domain Setup](multi-domain) | [Database Management →](database-management)
