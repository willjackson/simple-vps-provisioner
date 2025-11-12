package cms

import (
	"fmt"
	"path/filepath"
	"strings"
	"svp/pkg/database"
	"svp/pkg/utils"
)

// InstallDrupal installs a Drupal site for a domain
func InstallDrupal(domain, webroot, gitRepo, gitBranch, drupalRoot, docroot string, sitesDir string) error {
	utils.Section("Installing Drupal for " + domain)

	// Get admin username from www-data group
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

	domainDir := filepath.Join(webroot, domain)

	// Check if directory exists and is not empty
	if utils.CheckDirExists(domainDir) {
		entries, err := utils.RunShell(fmt.Sprintf("ls -A %s | wc -l", domainDir))
		if err == nil && strings.TrimSpace(entries) != "0" {
			utils.Warn("Directory %s is not empty", domainDir)
			fmt.Print("Delete and reprovision? [y/N]: ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				return fmt.Errorf("aborted: directory not empty")
			}
			utils.Log("Removing existing directory...")
			if _, err := utils.RunCommand("rm", "-rf", domainDir); err != nil {
				return fmt.Errorf("failed to remove directory: %v", err)
			}
		}
	}

	// Ensure webroot exists
	if err := utils.EnsureDir(domainDir); err != nil {
		return fmt.Errorf("failed to create domain directory: %v", err)
	}

	// Set ownership to admin:www-data
	_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), domainDir)
	_, _ = utils.RunCommand("chmod", "-R", "775", domainDir)

	var projectDir string

	// Check if Git repo is specified
	if gitRepo != "" {
		utils.Log("Cloning Git repository: %s (branch: %s)", gitRepo, gitBranch)

		// Clone repository as admin user
		if !utils.CheckDirExists(filepath.Join(domainDir, ".git")) {
			_, err := utils.RunShell(fmt.Sprintf("sudo -u %s git clone -b %s %s %s", adminUser, gitBranch, gitRepo, domainDir))
			if err != nil {
				return fmt.Errorf("failed to clone repository: %v", err)
			}
			utils.Ok("Repository cloned successfully")
		} else {
			utils.Verify("Repository already cloned")
		}

		projectDir = domainDir
	} else {
		// Fresh Drupal installation
		utils.Log("Creating fresh Drupal installation")
		projectDir = domainDir

		// Check if composer.json exists
		composerJSON := filepath.Join(projectDir, "composer.json")
		if !utils.CheckFileExists(composerJSON) {
			utils.Log("Creating new Drupal project via Composer...")

			// Create Drupal project as admin user
			_, err := utils.RunShell(fmt.Sprintf("cd %s && sudo -u %s composer create-project drupal/recommended-project . --no-interaction", projectDir, adminUser))
			if err != nil {
				return fmt.Errorf("failed to create Drupal project: %v", err)
			}

			utils.Ok("Drupal project created")
		} else {
			utils.Verify("Composer.json already exists")
		}
	}

	// Determine Drupal root directory
	composerDir := projectDir
	if drupalRoot != "" {
		composerDir = filepath.Join(projectDir, drupalRoot)
	} else {
		// Auto-detect composer.json location
		if !utils.CheckFileExists(filepath.Join(projectDir, "composer.json")) {
			// Look in common subdirectories
			for _, subdir := range []string{"drupal", "app", "backend"} {
				potentialPath := filepath.Join(projectDir, subdir)
				if utils.CheckFileExists(filepath.Join(potentialPath, "composer.json")) {
					composerDir = potentialPath
					utils.Log("Auto-detected Drupal root: %s", subdir)
					break
				}
			}
		}
	}

	// Run composer install
	composerJSON := filepath.Join(composerDir, "composer.json")
	if utils.CheckFileExists(composerJSON) {
		utils.Log("Running composer install...")
		_, err := utils.RunShell(fmt.Sprintf("cd %s && sudo -u %s composer install --no-interaction --prefer-dist", composerDir, adminUser))
		if err != nil {
			utils.Warn("Composer install failed: %v", err)
		} else {
			utils.Ok("Composer dependencies installed")
		}

		// Check if Drush is already in composer.json
		hasDrush, _ := utils.RunShell(fmt.Sprintf("grep -q 'drush/drush' %s && echo 'yes'", composerJSON))
		if strings.TrimSpace(hasDrush) != "yes" {
			utils.Log("Installing Drush...")
			_, err = utils.RunShell(fmt.Sprintf("cd %s && sudo -u %s composer require drush/drush --no-interaction", composerDir, adminUser))
			if err != nil {
				utils.Warn("Failed to install Drush: %v", err)
			} else {
				utils.Ok("Drush installed")
			}
		} else {
			utils.Verify("Drush already in composer.json")
		}
	}

	// Create database
	dbName, dbUser, dbPass, err := database.CreateDatabase(domain, sitesDir)
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	// Determine sites/default directory
	sitesDefaultDir := composerDir
	if docroot != "" {
		sitesDefaultDir = filepath.Join(sitesDefaultDir, docroot)
	}
	sitesDefaultDir = filepath.Join(sitesDefaultDir, "web", "sites", "default")

	// Create settings.php if it doesn't exist
	settingsFile := filepath.Join(sitesDefaultDir, "settings.php")
	if !utils.CheckFileExists(settingsFile) {
		utils.Log("Creating settings.php...")

		// Check if directory exists
		if !utils.CheckDirExists(sitesDefaultDir) {
			return fmt.Errorf("sites/default directory not found: %s", sitesDefaultDir)
		}

		// Copy default settings if exists
		defaultSettings := filepath.Join(sitesDefaultDir, "default.settings.php")
		if utils.CheckFileExists(defaultSettings) {
			_, _ = utils.RunCommand("cp", defaultSettings, settingsFile)
		} else {
			_, _ = utils.RunShell(fmt.Sprintf("touch %s", settingsFile))
		}

		// Make writable
		_, _ = utils.RunCommand("chmod", "u+w", sitesDefaultDir)
		_, _ = utils.RunCommand("chmod", "u+w", settingsFile)

		// Calculate config directory path relative to settings.php
		// settings.php is at: composerDir/web/sites/default/settings.php
		// config should be at: composerDir/config/sync
		// So from settings.php: ../../../config/sync
		configDir := filepath.Join(composerDir, "config", "sync")
		if err := utils.EnsureDir(configDir); err != nil {
			utils.Warn("Failed to create config directory: %v", err)
		} else {
			_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), configDir)
			utils.Ok("Config directory created: %s", configDir)
		}
		configSyncPath := "../config/sync"

		// Append database configuration
		dbConfig := fmt.Sprintf(`

// Database configuration added by setup script
$databases['default']['default'] = [
  'database' => '%s',
  'username' => '%s',
  'password' => '%s',
  'host' => 'localhost',
  'port' => '3306',
  'driver' => 'mysql',
  'prefix' => '',
  'collation' => 'utf8mb4_general_ci',
  'init_commands' => [
    'isolation_level' => 'SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED',
  ],
];

// Trusted host patterns
$settings['trusted_host_patterns'] = [
  '^%s$',
];

// Config sync directory
$settings['config_sync_directory'] = '%s';

// Hash salt
$settings['hash_salt'] = '%s';
`, dbName, dbUser, dbPass, strings.ReplaceAll(domain, ".", "\\."), configSyncPath, generateHashSalt())

		_, err = utils.RunShell(fmt.Sprintf("cat >> %s <<'EOF'\n%s\nEOF", settingsFile, dbConfig))
		if err != nil {
			return fmt.Errorf("failed to write settings.php: %v", err)
		}

		// Set proper permissions
		_, _ = utils.RunCommand("chmod", "444", settingsFile)
		_, _ = utils.RunCommand("chmod", "555", sitesDefaultDir)

		utils.Ok("settings.php created")

		// Create public files directory
		filesDir := filepath.Join(sitesDefaultDir, "files")
		if err := utils.EnsureDir(filesDir); err != nil {
			return fmt.Errorf("failed to create files directory: %v", err)
		}
		_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), filesDir)
		_, _ = utils.RunCommand("chmod", "775", filesDir)
		utils.Ok("Public files directory created")
	} else {
		utils.Verify("settings.php already exists")
		
		// Update config_sync_directory if needed
		_, _ = utils.RunCommand("chmod", "u+w", settingsFile)
		content, err := utils.RunShell(fmt.Sprintf("cat %s", settingsFile))
		if err == nil {
			// Check if config_sync_directory is set correctly
			if !strings.Contains(content, "$settings['config_sync_directory']") || !strings.Contains(content, "../config/sync") {
				utils.Log("Updating config_sync_directory in settings.php...")
				
				// Remove old config_sync_directory lines
				_, _ = utils.RunShell(fmt.Sprintf("sed -i '/config_sync_directory/d' %s", settingsFile))
				
				// Append correct config path
				configUpdate := "\n// Config sync directory\n$settings['config_sync_directory'] = '../config/sync';\n"
				_, _ = utils.RunShell(fmt.Sprintf("cat >> %s <<'EOF'\n%s\nEOF", settingsFile, configUpdate))
				utils.Ok("Config sync directory updated")
			}
		}
		_, _ = utils.RunCommand("chmod", "444", settingsFile)
	}

	// Set proper ownership
	_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), domainDir)

	utils.Ok("Drupal installation complete for %s", domain)
	return nil
}

// CreateDrushAlias creates a Drush site alias
func CreateDrushAlias(domain, projectDir, adminUser string) error {
	aliasDir := "/etc/drush/sites"
	if err := utils.EnsureDir(aliasDir); err != nil {
		return fmt.Errorf("failed to create alias directory: %v", err)
	}

	// Sanitize domain for alias name (replace dots with underscores)
	aliasName := strings.ReplaceAll(domain, ".", "_")
	aliasFile := filepath.Join(aliasDir, fmt.Sprintf("%s.site.yml", aliasName))

	// Drush 9+ YAML alias format
	aliasContent := fmt.Sprintf(`# Drush alias for %s
%s:
  root: %s/web
  uri: http://%s
  user: %s
`, domain, aliasName, projectDir, domain, adminUser)

	if utils.CheckFileExists(aliasFile) {
		utils.Log("Updating Drush alias for %s", domain)
	} else {
		utils.Log("Creating Drush alias for %s", domain)
	}

	_, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", aliasFile, aliasContent))
	if err != nil {
		return fmt.Errorf("failed to create alias: %v", err)
	}

	utils.Ok("Drush alias: @%s", aliasName)
	
	// Also create a shell wrapper for convenience
	return CreateDrushWrapper(domain, projectDir)
}

// CreateDrushWrapper creates a shell wrapper script for easy drush access
func CreateDrushWrapper(domain, projectDir string) error {
	wrapperPath := fmt.Sprintf("/usr/local/bin/drush-%s", domain)
	drushPath := filepath.Join(projectDir, "vendor/bin/drush")

	wrapperScript := fmt.Sprintf(`#!/bin/bash
# Drush wrapper for %s
cd %s || exit 1
exec %s "$@"
`, domain, projectDir, drushPath)

	utils.Log("Creating Drush wrapper for %s", domain)

	_, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", wrapperPath, wrapperScript))
	if err != nil {
		return fmt.Errorf("failed to create Drush wrapper: %v", err)
	}

	_, err = utils.RunCommand("chmod", "+x", wrapperPath)
	if err != nil {
		return fmt.Errorf("failed to make Drush wrapper executable: %v", err)
	}

	utils.Ok("Drush wrapper created: drush-%s", domain)
	return nil
}

// generateHashSalt generates a random hash salt for Drupal
func generateHashSalt() string {
	salt, _ := database.GeneratePassword(64)
	return salt
}

// InstallDrupalSite runs drush site-install if needed
func InstallDrupalSite(domain, projectDir, adminUser string) error {
	drushPath := filepath.Join(projectDir, "vendor/bin/drush")

	if !utils.CheckFileExists(drushPath) {
		utils.Warn("Drush not found, skipping site install")
		return nil
	}

	// Check if Drupal is already installed
	output, err := utils.RunShell(fmt.Sprintf("cd %s && %s status --field=bootstrap 2>/dev/null", projectDir, drushPath))
	if err == nil && strings.Contains(output, "Successful") {
		utils.Verify("Drupal already installed")
		return nil
	}

	utils.Log("Installing Drupal via drush site-install...")

	cmd := fmt.Sprintf("cd %s && sudo -u %s %s site-install minimal -y --account-name=admin --account-pass=admin", projectDir, adminUser, drushPath)
	_, err2 := utils.RunShell(cmd)
	if err2 != nil {
		return fmt.Errorf("drush site-install failed: %v", err2)
	}

	utils.Ok("Drupal site installed")
	return nil
}

// ImportDrupalConfig imports configuration via drush cim
func ImportDrupalConfig(domain, projectDir, adminUser string) error {
	drushPath := filepath.Join(projectDir, "vendor/bin/drush")

	if !utils.CheckFileExists(drushPath) {
		utils.Warn("Drush not found, skipping config import")
		return nil
	}

	// Check if config/sync directory exists
	configDir := filepath.Join(projectDir, "config/sync")
	if !utils.CheckDirExists(configDir) {
		utils.Skip("No config/sync directory found, skipping config import")
		return nil
	}

	// Check if config directory has actual config files (look for .yml files)
	output, err := utils.RunShell(fmt.Sprintf("ls -A %s 2>/dev/null | grep -E '\\.yml$' | wc -l", configDir))
	if err != nil {
		utils.Skip("Cannot check config directory, skipping config import")
		return nil
	}
	
	ymlCount := strings.TrimSpace(output)
	if ymlCount == "0" || ymlCount == "" {
		utils.Skip("No YAML config files found in config/sync, skipping config import")
		return nil
	}

	utils.Log("Importing Drupal configuration (%s config files)...", ymlCount)

	cmd := fmt.Sprintf("cd %s && sudo -u %s %s config-import -y", projectDir, adminUser, drushPath)
	_, err2 := utils.RunShell(cmd)
	if err2 != nil {
		utils.Warn("Config import failed: %v", err2)
		return nil
	}

	utils.Ok("Configuration imported")
	return nil
}
