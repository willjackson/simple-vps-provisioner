package config

import (
	"fmt"
	"svp/pkg/utils"
	"os"
	"strings"
	"svp/types"
)

const (
	ConfigDir   = "/etc/simple-host-manager"
	SitesDir    = "/etc/simple-host-manager/sites"
	PHPConfFile = "/etc/simple-host-manager/php.conf"
)

// EnsureConfigDirs creates the configuration directories if they don't exist
func EnsureConfigDirs() error {
	if err := utils.EnsureDir(ConfigDir); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	if err := utils.EnsureDir(SitesDir); err != nil {
		return fmt.Errorf("failed to create sites directory: %v", err)
	}
	return nil
}

// WriteSiteConfig writes configuration for a site
func WriteSiteConfig(domain, phpVersion, webroot string) error {
	configPath := fmt.Sprintf("%s/%s.conf", SitesDir, domain)

	dateStr, _ := utils.RunShell("date")
	configContent := fmt.Sprintf(`# Site configuration for %s
DOMAIN='%s'
PHP_VERSION='%s'
WEBROOT='%s'
CREATED='%s'
`, domain, domain, phpVersion, webroot, dateStr)

	_, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", configPath, configContent))
	if err != nil {
		return fmt.Errorf("failed to write site config: %v", err)
	}

	_, _ = utils.RunCommand("chmod", "0644", configPath)
	return nil
}

// ReadPHPVersions reads current and previous PHP versions from config
func ReadPHPVersions() (types.PHPVersions, error) {
	var versions types.PHPVersions

	if !utils.CheckFileExists(PHPConfFile) {
		return versions, nil
	}

	content, err := os.ReadFile(PHPConfFile)
	if err != nil {
		return versions, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "CURRENT_VERSION=") {
			versions.Current = strings.Trim(strings.TrimPrefix(line, "CURRENT_VERSION="), "'\"")
		} else if strings.HasPrefix(line, "PREVIOUS_VERSION=") {
			versions.Previous = strings.Trim(strings.TrimPrefix(line, "PREVIOUS_VERSION="), "'\"")
		}
	}

	return versions, nil
}

// WritePHPVersions writes current and previous PHP versions to config
func WritePHPVersions(current, previous string) error {
	configContent := fmt.Sprintf(`# PHP version tracking
CURRENT_VERSION='%s'
PREVIOUS_VERSION='%s'
`, current, previous)

	_, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", PHPConfFile, configContent))
	if err != nil {
		return fmt.Errorf("failed to write PHP versions: %v", err)
	}

	_, _ = utils.RunCommand("chmod", "0644", PHPConfFile)
	return nil
}

// SetCurrentPHPIfEmpty sets the current PHP version if not already set
func SetCurrentPHPIfEmpty(defaultVersion string) error {
	versions, err := ReadPHPVersions()
	if err != nil {
		return err
	}

	if versions.Current == "" {
		return WritePHPVersions(defaultVersion, "")
	}

	return nil
}

// EnsureAdminUser creates an admin user if it doesn't exist
func EnsureAdminUser(verifyOnly bool) error {
	// Check if admin user exists
	_, err := utils.RunCommand("id", "admin")
	if err == nil {
		utils.Verify("Admin user already exists")
		return nil
	}

	if verifyOnly {
		utils.Fail("Admin user does not exist")
		return fmt.Errorf("admin user does not exist")
	}

	utils.Log("Creating admin user...")

	// Create admin user
	_, err = utils.RunCommand("useradd", "-m", "-s", "/bin/bash", "admin")
	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	// Add to www-data group
	_, err = utils.RunCommand("usermod", "-a", "-G", "www-data", "admin")
	if err != nil {
		return fmt.Errorf("failed to add admin to www-data group: %v", err)
	}

	// Add to sudo group
	_, err = utils.RunCommand("usermod", "-a", "-G", "sudo", "admin")
	if err != nil {
		return fmt.Errorf("failed to add admin to sudo group: %v", err)
	}

	// Set password (prompt user)
	utils.Log("Set password for admin user:")
	_, err = utils.RunShell("passwd admin")
	if err != nil {
		utils.Warn("Failed to set admin password: %v", err)
	}

	utils.Ok("Admin user created")
	return nil
}
