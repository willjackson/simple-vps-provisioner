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
	}

	// Install Drush locally
	utils.Log("Installing Drush...")
	_, err = utils.RunShell(fmt.Sprintf("cd %s && sudo -u %s composer require drush/drush --no-interaction", composerDir, adminUser))
	if err != nil {
		utils.Warn("Failed to install Drush: %v", err)
	} else {
		utils.Ok("Drush installed")
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
];

// Trusted host patterns
$settings['trusted_host_patterns'] = [
  '^%s$',
];

// Hash salt
$settings['hash_salt'] = '%s';
`, dbName, dbUser, dbPass, strings.ReplaceAll(domain, ".", "\\."), generateHashSalt())

		_, err = utils.RunShell(fmt.Sprintf("cat >> %s <<'EOF'\n%s\nEOF", settingsFile, dbConfig))
		if err != nil {
			return fmt.Errorf("failed to write settings.php: %v", err)
		}

		// Set proper permissions
		_, _ = utils.RunCommand("chmod", "444", settingsFile)
		_, _ = utils.RunCommand("chmod", "555", sitesDefaultDir)

		utils.Ok("settings.php created")
	} else {
		utils.Verify("settings.php already exists")
	}

	// Set proper ownership
	_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), domainDir)

	utils.Ok("Drupal installation complete for %s", domain)
	return nil
}

// CreateDrushWrapper creates a Drush wrapper script for a domain
func CreateDrushWrapper(domain, projectDir string) error {
	wrapperPath := fmt.Sprintf("/usr/local/bin/drush-%s", domain)

	wrapperScript := fmt.Sprintf(`#!/bin/bash
# Drush wrapper for %s
cd %s || exit 1
exec ./vendor/bin/drush "$@"
`, domain, projectDir)

	if utils.CheckFileExists(wrapperPath) {
		utils.Verify("Drush wrapper already exists for %s", domain)
		return nil
	}

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
