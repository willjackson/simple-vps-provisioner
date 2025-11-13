package cmd

import (
	"fmt"
	"strings"
	"svp/pkg/config"
	"svp/pkg/ssl"
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

	// Socket path will be used for verification
	socketPath := fmt.Sprintf("/run/php/php%s-fpm-%s.sock", newPHPVersion, domain)

	// Verify new PHP-FPM service is running
	utils.Section("Verifying PHP-FPM Service")
	newServiceName := fmt.Sprintf("php%s-fpm", newPHPVersion)
	if err := system.EnsureServiceRunning(newServiceName, false); err != nil {
		return fmt.Errorf("new PHP-FPM service is not running: %v", err)
	}

	// Verify the socket file exists
	if !utils.CheckFileExists(socketPath) {
		return fmt.Errorf("PHP-FPM socket not found: %s", socketPath)
	}
	utils.Ok("PHP-FPM socket verified: %s", socketPath)

	// Update Nginx snippets for new PHP version if needed
	utils.Section("Updating Nginx Configuration")
	if err := web.EnsureSnippets(newPHPVersion); err != nil {
		utils.Warn("Failed to update Nginx snippets: %v", err)
	}

	// Update Nginx vhost to use new PHP version
	if err := web.CreateNginxVhost(domain, siteConfig.Webroot, newPHPVersion); err != nil {
		return fmt.Errorf("failed to update Nginx vhost: %v", err)
	}

	// Reconfigure SSL if certificate exists (vhost recreation removes SSL config)
	utils.Section("Restoring SSL Configuration")
	if err := ssl.ReconfigureSSL(domain); err != nil {
		utils.Warn("Failed to reconfigure SSL: %v", err)
		utils.Warn("You may need to run manually: certbot install --nginx -d %s --cert-name %s --redirect", domain, domain)
	}

	// Update site configuration
	utils.Section("Updating Site Configuration")
	if err := config.WriteSiteConfig(domain, newPHPVersion, siteConfig.Webroot); err != nil {
		return fmt.Errorf("failed to update site config: %v", err)
	}
	utils.Ok("Site configuration updated")

	// Check and fix database host configuration before reloading
	utils.Section("Checking Database Configuration")
	if err := fixDatabaseHost(domain, siteConfig.Webroot); err != nil {
		utils.Warn("Could not check database configuration: %v", err)
	}

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
	
	// Quick site accessibility check
	utils.Section("Verifying Site Access")
	utils.Log("Testing site accessibility...")
	
	// Check if site responds
	curlCmd := fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code}' -m 5 http://localhost -H 'Host: %s'", domain)
	httpCode, err := utils.RunShell(curlCmd)
	if err != nil {
		utils.Warn("Could not test site: %v", err)
	} else {
		httpCode = strings.TrimSpace(httpCode)
		if httpCode == "200" {
			utils.Ok("Site is accessible (HTTP %s)", httpCode)
		} else if httpCode == "500" || httpCode == "502" || httpCode == "503" {
			utils.Warn("Site returned HTTP %s - there may be an issue", httpCode)
			fmt.Println()
			fmt.Println("Troubleshooting steps:")
			fmt.Printf("  1. Check PHP-FPM logs: tail -f /var/log/php%s-fpm-%s-error.log\n", newPHPVersion, domain)
			fmt.Printf("  2. Check Nginx logs: tail -f /var/log/nginx/%s-error.log\n", domain)
			fmt.Printf("  3. Verify socket: ls -la %s\n", socketPath)
			fmt.Printf("  4. Test PHP-FPM: sudo -u www-data php%s -v\n", newPHPVersion)
			fmt.Printf("  5. Clear Drupal cache: drush-%s cr\n", domain)
			fmt.Println()
		} else {
			utils.Verify("Site returned HTTP %s", httpCode)
		}
	}
	
	fmt.Println()
	fmt.Printf("Visit: http://%s or https://%s to verify\n", domain, domain)
	fmt.Println()

	return nil
}

// fixDatabaseHost checks and fixes database host configuration in settings files
func fixDatabaseHost(domain, webroot string) error {
	settingsDir := fmt.Sprintf("%s/web/sites/default", webroot)
	settingsFiles := []string{
		fmt.Sprintf("%s/settings.svp.php", settingsDir),
		fmt.Sprintf("%s/settings.php", settingsDir),
	}

	for _, settingsFile := range settingsFiles {
		if !utils.CheckFileExists(settingsFile) {
			continue
		}

		// Check if file contains database host = 'db'
		checkCmd := fmt.Sprintf("grep -E \"'host'.*=>.*'db'\" %s", settingsFile)
		_, err := utils.RunShell(checkCmd)
		if err != nil {
			// Not found or error - continue to next file
			continue
		}

		// Found 'db' as hostname - needs fixing
		utils.Warn("Found database host 'db' in %s", settingsFile)
		fmt.Println("This needs to be changed to 'localhost' for non-Docker environments.")
		fmt.Print("Automatically fix this? [Y/n]: ")
		
		var response string
		fmt.Scanln(&response)
		if response == "n" || response == "N" {
			utils.Skip("Skipping database host fix")
			return nil
		}

		utils.Log("Fixing database host in %s...", settingsFile)

		// Make writable
		_, _ = utils.RunCommand("chmod", "u+w", settingsDir)
		_, _ = utils.RunCommand("chmod", "u+w", settingsFile)

		// Fix the host
		fixCmd := fmt.Sprintf("sed -i \"s/'host' => 'db'/'host' => 'localhost'/g\" %s", settingsFile)
		_, err = utils.RunShell(fixCmd)
		if err != nil {
			utils.Err("Failed to fix database host: %v", err)
			return err
		}

		// Make read-only again
		_, _ = utils.RunCommand("chmod", "444", settingsFile)
		_, _ = utils.RunCommand("chmod", "555", settingsDir)

		utils.Ok("Database host fixed in %s", settingsFile)

		// Clear Drupal cache if drush is available
		drushCmd := fmt.Sprintf("drush-%s", domain)
		if utils.CommandExists(drushCmd) {
			utils.Log("Clearing Drupal cache...")
			_, _ = utils.RunCommand(drushCmd, "cr")
			utils.Ok("Cache cleared")
		}

		return nil
	}

	utils.Verify("Database host configuration looks correct")
	return nil
}
