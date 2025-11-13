package cmd

import (
	"fmt"
	"svp/pkg/config"
	"svp/pkg/system"
	"svp/pkg/utils"
	"svp/pkg/web"
	"svp/types"
)

// PHPUpdate updates the PHP version for a specific domain
func PHPUpdate(cfg *types.Config) error {
	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println("  PHP Version Update")
	fmt.Println("==========================================================")
	fmt.Println()

	domain := cfg.PrimaryDomain
	newPHPVersion := cfg.PHPVersion

	if domain == "" {
		return fmt.Errorf("domain is required for PHP update mode")
	}

	if newPHPVersion == "" {
		return fmt.Errorf("PHP version is required for PHP update mode")
	}

	// Read site configuration
	utils.Section("Reading Site Configuration")
	siteConfig, err := config.ReadSiteConfig(domain)
	if err != nil {
		utils.Err("Failed to read site config: %v", err)
		return fmt.Errorf("site config not found - domain may not be configured: %s", domain)
	}

	utils.Ok("Found configuration for %s", domain)
	utils.Log("Current PHP version: %s", siteConfig.PHPVersion)
	utils.Log("New PHP version: %s", newPHPVersion)
	utils.Log("Webroot: %s", siteConfig.Webroot)

	// Check if already using the requested version
	if siteConfig.PHPVersion == newPHPVersion {
		utils.Skip("Domain %s is already using PHP %s", domain, newPHPVersion)
		return nil
	}

	// Confirm with user
	fmt.Println()
	fmt.Printf("This will update %s from PHP %s to PHP %s\n", domain, siteConfig.PHPVersion, newPHPVersion)
	fmt.Print("Continue? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		utils.Skip("PHP update cancelled")
		return nil
	}

	// Install new PHP version if not already installed
	utils.Section(fmt.Sprintf("Installing PHP %s", newPHPVersion))
	if err := web.InstallPHP(newPHPVersion, false); err != nil {
		return fmt.Errorf("failed to install PHP %s: %v", newPHPVersion, err)
	}

	// Harden PHP configuration
	if err := web.HardenPHPIni(newPHPVersion, false); err != nil {
		utils.Warn("Failed to harden PHP configuration: %v", err)
	}

	// Create new PHP-FPM pool for this domain with new version
	utils.Section("Creating PHP-FPM Pool")
	if err := web.CreatePHPPool(domain, newPHPVersion, siteConfig.Webroot); err != nil {
		return fmt.Errorf("failed to create PHP-FPM pool: %v", err)
	}

	// Update Nginx snippets for new PHP version if needed
	utils.Section("Updating Nginx Configuration")
	if err := web.EnsureSnippets(newPHPVersion); err != nil {
		utils.Warn("Failed to update Nginx snippets: %v", err)
	}

	// Update Nginx vhost to use new PHP version
	if err := web.CreateNginxVhost(domain, siteConfig.Webroot, newPHPVersion); err != nil {
		return fmt.Errorf("failed to update Nginx vhost: %v", err)
	}

	// Update site configuration
	utils.Section("Updating Site Configuration")
	if err := config.WriteSiteConfig(domain, newPHPVersion, siteConfig.Webroot); err != nil {
		return fmt.Errorf("failed to update site config: %v", err)
	}
	utils.Ok("Site configuration updated")

	// Test and reload Nginx
	utils.Section("Reloading Nginx")
	if err := web.ReloadNginx(); err != nil {
		return fmt.Errorf("failed to reload Nginx: %v", err)
	}

	// Stop old PHP-FPM pool (if it exists)
	utils.Section("Cleaning Up Old PHP Pool")
	oldPoolFile := fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", siteConfig.PHPVersion, domain)
	if utils.CheckFileExists(oldPoolFile) {
		utils.Log("Removing old PHP %s pool for %s", siteConfig.PHPVersion, domain)
		_, err := utils.RunCommand("rm", "-f", oldPoolFile)
		if err != nil {
			utils.Warn("Failed to remove old pool file: %v", err)
		}

		// Restart old PHP-FPM to unload the pool
		oldServiceName := fmt.Sprintf("php%s-fpm", siteConfig.PHPVersion)
		if err := system.RestartService(oldServiceName); err != nil {
			utils.Warn("Failed to restart old PHP-FPM service: %v", err)
		} else {
			utils.Ok("Old PHP pool removed")
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("==========================================================")
	utils.Ok("PHP Update Complete!")
	fmt.Println("==========================================================")
	fmt.Println()
	fmt.Printf("Domain: %s\n", domain)
	fmt.Printf("Old PHP Version: %s\n", siteConfig.PHPVersion)
	fmt.Printf("New PHP Version: %s\n", newPHPVersion)
	fmt.Println()
	fmt.Printf("Visit: http://%s or https://%s to verify\n", domain, domain)
	fmt.Println()

	return nil
}
