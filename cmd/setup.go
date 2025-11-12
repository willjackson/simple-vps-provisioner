package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"svp/pkg/cms"
	"svp/pkg/config"
	"svp/pkg/database"
	"svp/pkg/ssl"
	"svp/pkg/system"
	"svp/pkg/utils"
	"svp/pkg/web"
	"svp/types"
)

// FullSetup performs a complete VPS provisioning
func FullSetup(cfg *types.Config) error {
	utils.Section("Starting Full Setup")
	fmt.Printf("Domain: %s\n", cfg.PrimaryDomain)
	fmt.Printf("CMS: %s\n", cfg.CMS)
	fmt.Println()

	// Ensure configuration directories
	if err := config.EnsureConfigDirs(); err != nil {
		return err
	}

	// Set hostname
	utils.Section("Hostname")
	if err := system.SetProjectHostname(cfg.PrimaryDomain, cfg.VerifyOnly); err != nil {
		return err
	}

	// Install base packages
	utils.Section("Base Packages")
	if err := system.EnsureBasePackages(cfg.VerifyOnly); err != nil {
		return err
	}

	// Create swap if needed
	utils.Section("Swap")
	if err := system.CreateSwapIfNeeded(cfg.CreateSwap, cfg.VerifyOnly); err != nil {
		return err
	}

	// Install Nginx
	utils.Section("Nginx")
	if err := web.InstallNginx(cfg.VerifyOnly); err != nil {
		return err
	}

	// Install PHP
	utils.Section(fmt.Sprintf("PHP %s", cfg.PHPVersion))
	if err := web.InstallPHP(cfg.PHPVersion, cfg.VerifyOnly); err != nil {
		return err
	}
	if err := web.HardenPHPIni(cfg.PHPVersion, cfg.VerifyOnly); err != nil {
		return err
	}

	// Install database
	utils.Section("Database")
	if err := database.InstallMariaDB(cfg.DBEngine, cfg.VerifyOnly); err != nil {
		return err
	}

	// Install Composer
	utils.Section("Composer")
	if err := web.InstallComposer(cfg.VerifyOnly); err != nil {
		return err
	}

	// Install WP-CLI if WordPress
	if cfg.CMS == "wordpress" {
		utils.Section("WP-CLI")
		if err := cms.InstallWPCLI(cfg.VerifyOnly); err != nil {
			return err
		}
	}

	// Create admin user
	utils.Section("Admin User")
	if err := config.EnsureAdminUser(cfg.VerifyOnly); err != nil {
		return err
	}

	// Prepare webroot
	utils.Section("Webroot")
	if err := utils.EnsureDir(cfg.Webroot); err != nil {
		return fmt.Errorf("failed to create webroot: %v", err)
	}

	// Install CMS for each domain
	domains := []string{cfg.PrimaryDomain}
	if cfg.ExtraDomains != "" {
		extraDomains := strings.Split(cfg.ExtraDomains, ",")
		for _, d := range extraDomains {
			d = strings.TrimSpace(d)
			if d != "" {
				domains = append(domains, d)
			}
		}
	}

	// Resolve database import path to absolute path
	dbImportPath := cfg.DBImport
	if dbImportPath != "" && !filepath.IsAbs(dbImportPath) {
		var err error
		dbImportPath, err = filepath.Abs(dbImportPath)
		if err != nil {
			return fmt.Errorf("failed to resolve database path: %v", err)
		}
		utils.Log("Resolved database path: %s", dbImportPath)
	}

	utils.Section(fmt.Sprintf("%s Installations", strings.Title(cfg.CMS)))
	fmt.Printf("Installing %s for domains: %s\n", cfg.CMS, strings.Join(domains, ", "))

	for _, domain := range domains {
		if cfg.CMS == "drupal" {
			err := cms.InstallDrupal(domain, cfg.Webroot, cfg.GitRepo, cfg.GitBranch,
				cfg.DrupalRoot, cfg.Docroot, config.SitesDir, dbImportPath)
			if err != nil {
				return fmt.Errorf("failed to install Drupal for %s: %v", domain, err)
			}
		} else if cfg.CMS == "wordpress" {
			err := cms.InstallWordPress(domain, cfg.Webroot, cfg.GitRepo, cfg.GitBranch, config.SitesDir)
			if err != nil {
				return fmt.Errorf("failed to install WordPress for %s: %v", domain, err)
			}
		}
	}

	// Configure Nginx
	utils.Section("Nginx Configuration")
	if err := web.EnsureSnippets(cfg.PHPVersion); err != nil {
		return err
	}

	// Configure sites
	utils.Section("Configuring Sites")
	for _, domain := range domains {
		domainDir := filepath.Join(cfg.Webroot, domain)

		// Determine webroot for this domain
		siteWebroot := domainDir
		if cfg.CMS == "drupal" {
			// Auto-detect Drupal root by finding index.php
			if cfg.DrupalRoot != "" {
				siteWebroot = filepath.Join(siteWebroot, cfg.DrupalRoot)
			} else {
				// Look for index.php in common locations
				for _, subdir := range []string{"drupal/web", "app/web", "backend/web", "web"} {
					potentialPath := filepath.Join(domainDir, subdir)
					if utils.CheckFileExists(filepath.Join(potentialPath, "index.php")) {
						siteWebroot = potentialPath
						utils.Log("Auto-detected docroot: %s", subdir)
						break
					}
				}
			}
			
			if cfg.Docroot != "" {
				siteWebroot = filepath.Join(domainDir, cfg.Docroot)
			}
		}

		utils.Log("Configuring site: %s", domain)

		// Write site config
		if err := config.WriteSiteConfig(domain, cfg.PHPVersion, siteWebroot); err != nil {
			return err
		}

		// Create PHP-FPM pool
		if err := web.CreatePHPPool(domain, cfg.PHPVersion, siteWebroot); err != nil {
			return err
		}

		// Create Nginx vhost
		if err := web.CreateNginxVhost(domain, siteWebroot, cfg.PHPVersion); err != nil {
			return err
		}

		// Clean up old nginx config without .conf extension
		oldVhost := fmt.Sprintf("/etc/nginx/sites-available/%s", domain)
		oldLink := fmt.Sprintf("/etc/nginx/sites-enabled/%s", domain)
		if utils.CheckFileExists(oldVhost) {
			utils.Log("Removing old nginx config: %s", oldVhost)
			_, _ = utils.RunCommand("rm", "-f", oldVhost, oldLink)
		}

		// Create Drush alias for Drupal
		if cfg.CMS == "drupal" {
			drushDir := domainDir
			if cfg.DrupalRoot != "" {
				drushDir = filepath.Join(drushDir, cfg.DrupalRoot)
			} else {
				// Auto-detect composer.json location
				for _, subdir := range []string{"drupal", "app", "backend"} {
					potentialPath := filepath.Join(domainDir, subdir)
					if utils.CheckFileExists(filepath.Join(potentialPath, "composer.json")) {
						drushDir = potentialPath
						break
					}
				}
			}
			
			// Get admin user
			adminUser := "admin"
			output, err := utils.RunShell("getent group www-data | cut -d: -f4")
			if err == nil {
				members := strings.Split(strings.TrimSpace(output), ",")
				for _, member := range members {
					if member != "" && member != "www-data" {
						adminUser = member
						break
					}
				}
			}
			
			if err := cms.CreateDrushAlias(domain, drushDir, adminUser); err != nil {
				utils.Warn("Failed to create Drush alias: %v", err)
			}
			
			// Install Drupal site if not already installed (skipped if db provided)
			if err := cms.InstallDrupalSite(domain, drushDir, adminUser, dbImportPath); err != nil {
				utils.Warn("Failed to install Drupal site: %v", err)
			}
			
			// Import configuration only if database was NOT imported
			if dbImportPath == "" {
				if err := cms.ImportDrupalConfig(domain, drushDir, adminUser, false); err != nil {
					utils.Warn("Failed to import configuration: %v", err)
				}
			} else {
				utils.Skip("Skipping config import (database was imported)")
			}
		}
	}

	// Reload Nginx
	utils.Section("Nginx")
	if err := web.ReloadNginx(); err != nil {
		return err
	}

	// Install Certbot and obtain SSL certificates if enabled and email provided
	if cfg.SSLEnable && cfg.LEEmail != "" {
		utils.Section("SSL Certificates")
		if err := ssl.InstallCertbot(cfg.VerifyOnly); err != nil {
			return err
		}

		// Obtain/reconfigure certificates for all domains
		for _, domain := range domains {
			domainDir := filepath.Join(cfg.Webroot, domain)
			
			// Determine webroot for this domain (same logic as earlier)
			siteWebroot := domainDir
			if cfg.CMS == "drupal" {
				if cfg.DrupalRoot != "" {
					siteWebroot = filepath.Join(siteWebroot, cfg.DrupalRoot)
				} else {
					for _, subdir := range []string{"drupal/web", "app/web", "backend/web", "web"} {
						potentialPath := filepath.Join(domainDir, subdir)
						if utils.CheckFileExists(filepath.Join(potentialPath, "index.php")) {
							siteWebroot = potentialPath
							break
						}
					}
				}
				if cfg.Docroot != "" {
					siteWebroot = filepath.Join(domainDir, cfg.Docroot)
				}
			}
			
			// Check if certificate already exists
			certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
			if utils.CheckFileExists(certPath) {
				utils.Log("Reconfiguring nginx with existing SSL certificate for %s", domain)
				// Use certbot install to reconfigure nginx (since we recreated vhost)
				cmd := fmt.Sprintf("certbot install --nginx --cert-name %s --non-interactive --redirect", domain)
				if _, err := utils.RunShell(cmd); err != nil {
					utils.Warn("Failed to reconfigure SSL for %s: %v", domain, err)
					continue
				}
				utils.Ok("SSL reconfigured for %s", domain)
			} else {
				// Obtain new certificate
				if err := ssl.ObtainCertificate(domain, cfg.LEEmail); err != nil {
					utils.Warn("Failed to obtain SSL for %s: %v", domain, err)
					continue
				}
			}

			// Fix SSL docroot to match HTTP docroot
			if err := ssl.FixSSLDocroot(domain, siteWebroot); err != nil {
				utils.Warn("Failed to fix SSL docroot for %s: %v", domain, err)
			}

			// Enhance SSL configuration with better security settings
			if err := ssl.EnhanceSSLConfig(domain); err != nil {
				utils.Warn("Failed to enhance SSL config for %s: %v", domain, err)
			}
		}

		// Setup auto-renewal
		if err := ssl.SetupAutoRenewal(cfg.VerifyOnly); err != nil {
			return err
		}

		// Reload nginx to apply enhanced SSL settings
		if err := web.ReloadNginx(); err != nil {
			return err
		}
	}

	// Setup firewall
	utils.Section("Firewall")
	if err := system.SetupFirewall(cfg.UFWEnable, cfg.VerifyOnly); err != nil {
		return err
	}

	// Save PHP version
	if err := config.SetCurrentPHPIfEmpty(cfg.PHPVersion); err != nil {
		return err
	}

	// Print summary
	fmt.Println()
	fmt.Println("==========================================================")
	utils.Ok("Setup Complete!")
	fmt.Println("==========================================================")
	fmt.Println()
	fmt.Printf("CMS: %s\n", cfg.CMS)
	fmt.Printf("Domain(s): %s\n", strings.Join(domains, ", "))
	fmt.Printf("Webroot: %s\n", cfg.Webroot)
	fmt.Printf("PHP Version: %s\n", cfg.PHPVersion)
	fmt.Println()
	fmt.Println("Next steps:")
	if cfg.CMS == "drupal" {
		for _, domain := range domains {
			fmt.Printf("  • Complete Drupal installation: http://%s/install.php\n", domain)
			fmt.Printf("  • Or use Drush: drush-%s site:install\n", domain)
		}
	} else {
		for _, domain := range domains {
			fmt.Printf("  • Complete WordPress installation: http://%s/wp-admin/install.php\n", domain)
		}
	}
	fmt.Println()
	fmt.Println("Database credentials saved in:", config.SitesDir)
	fmt.Println()

	return nil
}
