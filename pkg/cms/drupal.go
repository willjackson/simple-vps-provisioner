package cms

import (
	"fmt"
	"path/filepath"
	"strings"
	"svp/pkg/database"
	"svp/pkg/utils"
)

// InstallDrupal installs a Drupal site for a domain
// Returns settingsSVPAdded flag indicating if settings.svp.php was created
func InstallDrupal(domain, webroot, gitRepo, gitBranch, drupalRoot, docroot string, sitesDir string, dbImport string, keepExistingDB bool) (settingsSVPAdded bool, err error) {
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
				return false, fmt.Errorf("aborted: directory not empty")
			}
			utils.Log("Removing existing directory...")
			if _, err := utils.RunCommand("rm", "-rf", domainDir); err != nil {
				return false, fmt.Errorf("failed to remove directory: %v", err)
			}
		}
	}

	// Ensure webroot exists
	if err := utils.EnsureDir(domainDir); err != nil {
		return false, fmt.Errorf("failed to create domain directory: %v", err)
	}

	// Set ownership to admin:www-data
	_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), domainDir)
	_, _ = utils.RunCommand("chmod", "-R", "775", domainDir)

	var projectDir string

	// Check if Git repo is specified
	if gitRepo != "" {
		utils.Log("Cloning Git repository: %s (branch: %s)", gitRepo, gitBranch)

		// Clone repository as admin user
		// CD to webroot first to avoid getcwd issues after directory deletion
		if !utils.CheckDirExists(filepath.Join(domainDir, ".git")) {
			_, err := utils.RunShell(fmt.Sprintf("cd %s && sudo -u %s git clone -b %s %s %s", webroot, adminUser, gitBranch, gitRepo, domainDir))
			if err != nil {
				return false, fmt.Errorf("failed to clone repository: %v", err)
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
				return false, fmt.Errorf("failed to create Drupal project: %v", err)
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

	// Handle database setup based on keepExistingDB flag
	dbName, dbUser, dbPass := "", "", ""
	existingDB := false
	
	if keepExistingDB {
		// Check if database credentials already exist
		if existingDBName, existingDBUser, existingDBPass, exists := database.ReadDatabaseCredentials(domain, sitesDir); exists {
			dbName = existingDBName
			dbUser = existingDBUser
			dbPass = existingDBPass
			existingDB = true
			utils.Ok("Using existing database credentials")
			
			// If we have existing database and drush is available, drop all tables
			if utils.CheckFileExists(filepath.Join(composerDir, "vendor/bin/drush")) {
				utils.Log("Clearing existing database tables...")
				if err := DropDatabaseTables(composerDir, adminUser, domain); err != nil {
					utils.Warn("Failed to drop tables with drush: %v", err)
					utils.Log("Falling back to SQL method...")
					// Fallback to SQL method if drush fails
					dropTablesCmd := fmt.Sprintf("mysql -u%s -p%s %s -e \"SET FOREIGN_KEY_CHECKS = 0; SET GROUP_CONCAT_MAX_LEN=32768; SET @tables = NULL; SELECT GROUP_CONCAT('\\`', table_name, '\\`') INTO @tables FROM information_schema.tables WHERE table_schema = '%s'; SELECT IFNULL(@tables,'dummy') INTO @tables; SET @tables = CONCAT('DROP TABLE IF EXISTS ', @tables); PREPARE stmt FROM @tables; EXECUTE stmt; DEALLOCATE PREPARE stmt; SET FOREIGN_KEY_CHECKS = 1;\"", dbUser, dbPass, dbName, dbName)
					_, err = utils.RunShell(dropTablesCmd)
					if err != nil {
						utils.Warn("Failed to drop tables (may be empty): %v", err)
					} else {
						utils.Ok("Tables dropped using SQL")
					}
				}
			} else {
				utils.Warn("Drush not available yet, will clear database after composer install")
			}
		} else {
			// No existing credentials, create new database
			var err error
			dbName, dbUser, dbPass, err = database.CreateDatabase(domain, sitesDir)
			if err != nil {
				return false, fmt.Errorf("failed to create database: %v", err)
			}
		}
	} else {
		// Default behavior: Drop existing database completely and create new one
		if err := database.DropDatabase(domain, sitesDir); err != nil {
			utils.Warn("Failed to drop existing database: %v", err)
		}
		
		// Create fresh database with new credentials
		var err error
		dbName, dbUser, dbPass, err = database.CreateDatabase(domain, sitesDir)
		if err != nil {
			return false, fmt.Errorf("failed to create database: %v", err)
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
		
		// If we have existing database but couldn't drop tables earlier (drush wasn't available),
		// do it now after drush is installed
		if existingDB && utils.CheckFileExists(filepath.Join(composerDir, "vendor/bin/drush")) {
			// Check if tables still exist (they would if we couldn't drop earlier)
			checkTables := fmt.Sprintf("mysql -u%s -p%s %s -e 'SHOW TABLES;' | wc -l", dbUser, dbPass, dbName)
			tableCount, err := utils.RunShell(checkTables)
			if err == nil && strings.TrimSpace(tableCount) != "0" {
				utils.Log("Database has existing tables, clearing now...")
				if err := DropDatabaseTables(composerDir, adminUser, domain); err != nil {
					utils.Warn("Failed to drop tables with drush: %v", err)
				}
			}
		}
	}

	// Import database if provided (database already cleared above)
	if dbImport != "" {
		if !utils.CheckFileExists(dbImport) {
			return false, fmt.Errorf("database file not found: %s", dbImport)
		}
		
		utils.Log("Importing database from %s...", dbImport)
		
		// Import database (handle .gz files)
		var importCmd string
		if strings.HasSuffix(strings.ToLower(dbImport), ".gz") {
			// Use zcat for compressed files
			importCmd = fmt.Sprintf("zcat < %s | mysql -u%s -p%s %s", dbImport, dbUser, dbPass, dbName)
		} else {
			// Direct import for .sql files
			importCmd = fmt.Sprintf("mysql -u%s -p%s %s < %s", dbUser, dbPass, dbName, dbImport)
		}
		
		_, err := utils.RunShell(importCmd)
		if err != nil {
			return false, fmt.Errorf("database import failed: %v", err)
		}
		
		utils.Ok("Database imported successfully")
	}

	// Determine sites/default directory
	sitesDefaultDir := composerDir
	if docroot != "" {
		sitesDefaultDir = filepath.Join(sitesDefaultDir, docroot)
	}
	sitesDefaultDir = filepath.Join(sitesDefaultDir, "web", "sites", "default")

	// Handle settings.php creation or update
	settingsFile := filepath.Join(sitesDefaultDir, "settings.php")
	settingsSVPFile := filepath.Join(sitesDefaultDir, "settings.svp.php")
	settingsSVPAdded = false
	
	// Check if directory exists
	if !utils.CheckDirExists(sitesDefaultDir) {
		return false, fmt.Errorf("sites/default directory not found: %s", sitesDefaultDir)
	}

	// Calculate config directory path relative to settings.php
	configDir := filepath.Join(composerDir, "config", "sync")
	if err := utils.EnsureDir(configDir); err != nil {
		utils.Warn("Failed to create config directory: %v", err)
	} else {
		_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), configDir)
		utils.Ok("Config directory created: %s", configDir)
	}
	configSyncPath := "../config/sync"

	// Create drush directory and drush.yml
	drushDir := filepath.Join(composerDir, "drush")
	if err := utils.EnsureDir(drushDir); err != nil {
		utils.Warn("Failed to create drush directory: %v", err)
	} else {
		drushYmlPath := filepath.Join(drushDir, "drush.yml")
		if !utils.CheckFileExists(drushYmlPath) {
			drushYml := fmt.Sprintf("options:\n  uri: 'http://%s'\n", domain)
			_, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", drushYmlPath, drushYml))
			if err != nil {
				utils.Warn("Failed to create drush.yml: %v", err)
			} else {
				_, _ = utils.RunCommand("chown", fmt.Sprintf("%s:www-data", adminUser), drushYmlPath)
				utils.Ok("Created drush/drush.yml with base URL")
			}
		} else {
			utils.Verify("drush/drush.yml already exists")
		}
	}

	// Prepare database configuration
	dbConfig := fmt.Sprintf(`<?php
/**
 * SVP-managed database and configuration settings
 * Generated by Simple VPS Provisioner
 * 
 * This file contains production database credentials and is automatically
 * added to .gitignore to prevent committing sensitive information.
 */

// Database configuration
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

	if !utils.CheckFileExists(settingsFile) {
		// No existing settings.php - create a new one normally
		utils.Log("Creating settings.php...")

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

		// Append configuration directly to settings.php
		_, err = utils.RunShell(fmt.Sprintf("cat >> %s <<'EOF'\n\n%s\nEOF", settingsFile, dbConfig))
		if err != nil {
			return false, fmt.Errorf("failed to write settings.php: %v", err)
		}

		// Set proper permissions
		_, _ = utils.RunCommand("chmod", "444", settingsFile)
		_, _ = utils.RunCommand("chmod", "555", sitesDefaultDir)

		utils.Ok("settings.php created")
	} else {
		// Existing settings.php found - use settings.svp.php pattern
		utils.Log("settings.php already exists - creating settings.svp.php")
		
		// Make directories writable
		_, _ = utils.RunCommand("chmod", "u+w", sitesDefaultDir)
		_, _ = utils.RunCommand("chmod", "u+w", settingsFile)
		
		// Create settings.svp.php with our configuration
		_, err = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", settingsSVPFile, dbConfig))
		if err != nil {
			return false, fmt.Errorf("failed to create settings.svp.php: %v", err)
		}
		utils.Ok("Created settings.svp.php")
		
		// Check if include statement already exists in settings.php
		content, err := utils.RunShell(fmt.Sprintf("cat %s", settingsFile))
		if err == nil && !strings.Contains(content, "settings.svp.php") {
			// Append include statement to settings.php
			includeStatement := `

/**
 * Include SVP-managed settings
 * 
 * This file is managed by Simple VPS Provisioner and contains database
 * credentials. It is automatically added to .gitignore.
 */
if (file_exists($app_root . '/' . $site_path . '/settings.svp.php')) {
  include $app_root . '/' . $site_path . '/settings.svp.php';
}
`
			_, err = utils.RunShell(fmt.Sprintf("cat >> %s <<'EOF'\n%s\nEOF", settingsFile, includeStatement))
			if err != nil {
				return false, fmt.Errorf("failed to add include to settings.php: %v", err)
			}
			utils.Ok("Added settings.svp.php include to settings.php")
		} else if err == nil {
			utils.Verify("settings.svp.php already included in settings.php")
		}
		
		// Set proper permissions
		_, _ = utils.RunCommand("chmod", "444", settingsFile)
		_, _ = utils.RunCommand("chmod", "444", settingsSVPFile)
		_, _ = utils.RunCommand("chmod", "555", sitesDefaultDir)
		
		// Add settings.svp.php to .gitignore
		gitignorePath := filepath.Join(composerDir, ".gitignore")
		if utils.CheckFileExists(gitignorePath) {
			// Check if already in .gitignore
			gitignoreContent, err := utils.RunShell(fmt.Sprintf("cat %s", gitignorePath))
			if err == nil && !strings.Contains(gitignoreContent, "settings.svp.php") {
				utils.Log("Adding settings.svp.php to .gitignore...")
				_, _ = utils.RunShell(fmt.Sprintf("echo '\n# SVP-managed database credentials\nweb/sites/*/settings.svp.php' >> %s", gitignorePath))
				utils.Ok("Added settings.svp.php to .gitignore")
			} else if err == nil {
				utils.Verify("settings.svp.php already in .gitignore")
			}
		} else {
			// Create .gitignore if it doesn't exist
			utils.Log("Creating .gitignore with settings.svp.php...")
			gitignoreContent := `# SVP-managed database credentials
web/sites/*/settings.svp.php
`
			_, _ = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", gitignorePath, gitignoreContent))
			_, _ = utils.RunCommand("chown", fmt.Sprintf("%s:www-data", adminUser), gitignorePath)
			utils.Ok("Created .gitignore with settings.svp.php")
		}
		
		settingsSVPAdded = true
	}

	// Create public files directory
	filesDir := filepath.Join(sitesDefaultDir, "files")
	if err := utils.EnsureDir(filesDir); err != nil {
		return settingsSVPAdded, fmt.Errorf("failed to create files directory: %v", err)
	}
	_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), filesDir)
	_, _ = utils.RunCommand("chmod", "775", filesDir)
	utils.Ok("Public files directory created")

	// Set proper ownership
	_, _ = utils.RunCommand("chown", "-R", fmt.Sprintf("%s:www-data", adminUser), domainDir)

	utils.Ok("Drupal installation complete for %s", domain)
	return settingsSVPAdded, nil
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

// DropDatabaseTables drops all tables from the database using drush sql-drop
func DropDatabaseTables(projectDir, adminUser, domain string) error {
	drushPath := filepath.Join(projectDir, "vendor/bin/drush")
	
	if !utils.CheckFileExists(drushPath) {
		utils.Warn("Drush not found at %s, cannot drop tables", drushPath)
		return fmt.Errorf("drush not found")
	}
	
	utils.Log("Dropping all tables from database for %s using drush...", domain)
	
	// Use drush sql-drop to drop all tables
	cmd := fmt.Sprintf("cd %s && sudo -u %s %s sql-drop -y", projectDir, adminUser, drushPath)
	_, err := utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("failed to drop database tables: %v", err)
	}
	
	utils.Ok("All tables dropped from database")
	return nil
}

// generateHashSalt generates a random hash salt for Drupal
func generateHashSalt() string {
	salt, _ := database.GeneratePassword(64)
	return salt
}

// InstallDrupalSite runs drush site-install if needed (when no database import provided)
func InstallDrupalSite(domain, projectDir, adminUser, dbImport, sitesDir string) error {
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

	// If database was imported, skip site-install
	if dbImport != "" {
		utils.Verify("Database imported, skipping site-install")
		return nil
	}

	// Read database credentials
	dbName, dbUser, dbPass, exists := database.ReadDatabaseCredentials(domain, sitesDir)
	if !exists {
		return fmt.Errorf("database credentials not found for %s", domain)
	}

	// Fresh install with site-install
	utils.Log("Installing Drupal via drush site-install...")

	// Construct database URL: mysql://user:pass@host/dbname
	dbURL := fmt.Sprintf("mysql://%s:%s@localhost/%s", dbUser, dbPass, dbName)

	cmd := fmt.Sprintf("cd %s && sudo -u %s %s site-install standard -y --account-name=admin --account-pass=admin --site-name='%s' --db-url='%s'", projectDir, adminUser, drushPath, domain, dbURL)
	_, err = utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("drush site-install failed: %v", err)
	}

	utils.Ok("Drupal site installed")
	return nil
}

// ImportDrupalConfig imports configuration via drush cim
func ImportDrupalConfig(domain, projectDir, adminUser string, hasDBImport bool) error {
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

	// If we did site-install (not database import), get UUID from config and set it in DB
	if !hasDBImport {
		utils.Log("Getting UUID from config...")
		
		// Check if system.site.yml exists in config
		systemSiteConfig := filepath.Join(configDir, "system.site.yml")
		if utils.CheckFileExists(systemSiteConfig) {
			// Extract UUID from system.site.yml
			uuidOutput, err := utils.RunShell(fmt.Sprintf("grep '^uuid:' %s | awk '{print $2}'", systemSiteConfig))
			if err == nil && strings.TrimSpace(uuidOutput) != "" {
				configUUID := strings.TrimSpace(uuidOutput)
				utils.Log("Setting site UUID to match config: %s", configUUID)
				
				// Set the UUID in the database
				cmd := fmt.Sprintf("cd %s && sudo -u %s %s config:set system.site uuid %s -y", projectDir, adminUser, drushPath, configUUID)
				_, err = utils.RunShell(cmd)
				if err != nil {
					utils.Warn("Failed to set UUID: %v", err)
				} else {
					utils.Ok("UUID set successfully")
				}
			}
		}
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

// GetDrupalLoginLink generates and returns a one-time login link using drush uli
func GetDrupalLoginLink(projectDir, adminUser string) (string, error) {
	drushPath := filepath.Join(projectDir, "vendor/bin/drush")
	
	if !utils.CheckFileExists(drushPath) {
		return "", fmt.Errorf("drush not found")
	}
	
	// Run drush uli to get login link
	cmd := fmt.Sprintf("cd %s && sudo -u %s %s uli", projectDir, adminUser, drushPath)
	output, err := utils.RunShell(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to generate login link: %v", err)
	}
	
	return strings.TrimSpace(output), nil
}
