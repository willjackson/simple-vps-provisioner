package cmd

import (
	"fmt"
	"svp/pkg/cms"
	"svp/pkg/config"
	"svp/pkg/database"
	"svp/pkg/system"
	"svp/pkg/utils"
	"svp/pkg/web"
	"path/filepath"
	"strings"
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

	utils.Section(fmt.Sprintf("%s Installations", strings.Title(cfg.CMS)))
	fmt.Printf("Installing %s for domains: %s\n", cfg.CMS, strings.Join(domains, ", "))

	for _, domain := range domains {
		if cfg.CMS == "drupal" {
			err := cms.InstallDrupal(domain, cfg.Webroot, cfg.GitRepo, cfg.GitBranch,
				cfg.DrupalRoot, cfg.Docroot, config.SitesDir)
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
			if cfg.DrupalRoot != "" {
				siteWebroot = filepath.Join(siteWebroot, cfg.DrupalRoot)
			}
			if cfg.Docroot != "" {
				siteWebroot = filepath.Join(siteWebroot, cfg.Docroot)
			}
			// Drupal typically serves from web/ directory
			siteWebroot = filepath.Join(siteWebroot, "web")
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

		// Create Drush wrapper for Drupal
		if cfg.CMS == "drupal" {
			drushDir := domainDir
			if cfg.DrupalRoot != "" {
				drushDir = filepath.Join(drushDir, cfg.DrupalRoot)
			}
			if err := cms.CreateDrushWrapper(domain, drushDir); err != nil {
				utils.Warn("Failed to create Drush wrapper: %v", err)
			}
		}
	}

	// Reload Nginx
	utils.Section("Nginx")
	if err := web.ReloadNginx(); err != nil {
		return err
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
