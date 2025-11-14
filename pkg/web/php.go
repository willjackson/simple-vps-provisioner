package web

import (
	"fmt"
	"svp/pkg/system"
	"svp/pkg/utils"
	"strings"
)

// PHPPackages returns the list of PHP packages for a given version
func PHPPackages(version string) []string {
	return []string{
		fmt.Sprintf("php%s-fpm", version),
		fmt.Sprintf("php%s-cli", version),
		fmt.Sprintf("php%s-common", version),
		fmt.Sprintf("php%s-mbstring", version),
		fmt.Sprintf("php%s-xml", version),
		fmt.Sprintf("php%s-gd", version),
		fmt.Sprintf("php%s-curl", version),
		fmt.Sprintf("php%s-zip", version),
		fmt.Sprintf("php%s-intl", version),
		fmt.Sprintf("php%s-sqlite3", version),
		fmt.Sprintf("php%s-readline", version),
		fmt.Sprintf("php%s-mysql", version),
		fmt.Sprintf("php%s-opcache", version),
	}
}

// InstallPHP installs a specific PHP version with required extensions
func InstallPHP(version string, verifyOnly bool) error {
	// Ensure Sury repo is added or native packages are available
	if err := system.AddPHPRepoIfNeeded(version, verifyOnly); err != nil {
		return err
	}

	packages := PHPPackages(version)
	var missing []string

	for _, pkg := range packages {
		if !utils.CheckPackageInstalled(pkg) {
			missing = append(missing, pkg)
		}
	}

	// Only update package cache if we have missing packages to install
	// This prevents unnecessary update errors when everything is already installed
	if len(missing) > 0 && !verifyOnly {
		utils.Log("Updating package cache...")
		if _, err := utils.RunCommand("apt-get", "update", "-y"); err != nil {
			utils.Warn("Failed to update package cache: %v", err)
			// Continue anyway - packages might still be installable
		}
	}

	if len(missing) > 0 {
		if verifyOnly {
			utils.Fail("Missing PHP %s packages: %s", version, strings.Join(missing, ", "))
			return fmt.Errorf("missing PHP packages")
		}

		// Check if the first package is available before attempting full install
		firstPkg := missing[0]
		checkCmd := fmt.Sprintf("apt-cache policy %s | grep -q 'Candidate:' && apt-cache policy %s | grep -v 'Candidate: (none)'", firstPkg, firstPkg)
		_, err := utils.RunShell(checkCmd)
		if err != nil {
			// Package not available - provide helpful error
			utils.Err("PHP %s packages are not available in your distribution's repositories", version)
			
			// Detect OS to provide specific guidance
			codename, _ := utils.RunShell("lsb_release -sc")
			codename = strings.TrimSpace(codename)
			
			if codename == "trixie" || codename == "sid" {
				utils.Err("Debian %s ships with PHP 8.4+ by default", codename)
				utils.Err("Please use PHP 8.4 instead: svp -php-version 8.4")
				return fmt.Errorf("PHP %s not available on Debian %s (use PHP 8.4+)", version, codename)
			}
			
			utils.Err("Try running: apt-cache search php%s to see available versions", version)
			return fmt.Errorf("PHP %s packages not found in repositories", version)
		}

		utils.Log("Installing PHP %s packages: %s", version, strings.Join(missing, ", "))
		args := append([]string{"install", "-y", "--no-install-recommends"}, missing...)
		_, err = utils.RunCommand("apt-get", args...)
		if err != nil {
			return fmt.Errorf("failed to install PHP packages: %v", err)
		}

		utils.Ok("PHP %s packages installed", version)
	} else {
		utils.Verify("PHP %s packages already installed", version)
	}

	// Ensure PHP-FPM service is running
	serviceName := fmt.Sprintf("php%s-fpm", version)
	if err := system.EnsureServiceRunning(serviceName, verifyOnly); err != nil {
		return err
	}

	return nil
}

// HardenPHPIni applies security hardening to PHP configuration
func HardenPHPIni(version string, verifyOnly bool) error {
	fpmIni := fmt.Sprintf("/etc/php/%s/fpm/php.ini", version)

	if !utils.CheckFileExists(fpmIni) {
		utils.Warn("PHP FPM ini file not found: %s", fpmIni)
		return nil
	}

	utils.Log("Hardening PHP %s configuration...", version)

	settings := map[string]string{
		"expose_php":              "Off",
		"display_errors":          "Off",
		"log_errors":              "On",
		"max_execution_time":      "300",
		"max_input_time":          "300",
		"memory_limit":            "512M",
		"post_max_size":           "64M",
		"upload_max_filesize":     "64M",
		"max_file_uploads":        "20",
		"allow_url_fopen":         "On",
		"allow_url_include":       "Off",
		"session.cookie_httponly": "1",
		"session.cookie_secure":   "1",
		"session.use_strict_mode": "1",
	}

	if verifyOnly {
		// In verify mode, just check if settings exist
		utils.Verify("Would harden PHP %s configuration", version)
		return nil
	}

	// Apply settings using sed
	for key, value := range settings {
		// Escape special characters for sed
		escapedKey := strings.ReplaceAll(key, ".", "\\.")
		sedCmd := fmt.Sprintf("sed -i 's/^;*%s =.*/%s = %s/' %s", escapedKey, key, value, fpmIni)
		_, _ = utils.RunShell(sedCmd)
	}

	utils.Ok("PHP %s configuration hardened", version)

	// Restart PHP-FPM to apply changes
	serviceName := fmt.Sprintf("php%s-fpm", version)
	if err := system.RestartService(serviceName); err != nil {
		utils.Warn("Failed to restart %s: %v", serviceName, err)
	}

	return nil
}

// CreatePHPPool creates a PHP-FPM pool for a specific site
func CreatePHPPool(domain, version, webroot string) error {
	poolFile := fmt.Sprintf("/etc/php/%s/fpm/pool.d/%s.conf", version, domain)
	socketPath := fmt.Sprintf("/run/php/php%s-fpm-%s.sock", version, domain)

	// For open_basedir, use the project root (parent of webroot if it ends with /web)
	// This allows access to vendor/ directory
	projectRoot := webroot
	if strings.HasSuffix(webroot, "/web") {
		// Get parent directory (removes /web)
		projectRoot = webroot[:len(webroot)-4]
	}

	poolConfig := fmt.Sprintf(`; PHP-FPM pool for %s
[%s]
user = www-data
group = www-data
listen = %s
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = dynamic
pm.max_children = 10
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
pm.max_requests = 500

; Environment
env[HOSTNAME] = $HOSTNAME
env[PATH] = /usr/local/bin:/usr/bin:/bin
env[TMP] = /tmp
env[TMPDIR] = /tmp
env[TEMP] = /tmp

; PHP admin values
php_admin_value[error_log] = /var/log/php%s-fpm-%s-error.log
php_admin_flag[log_errors] = on
php_admin_value[memory_limit] = 512M

; Security
php_admin_value[open_basedir] = %s:/tmp:/usr/share/php
php_admin_value[upload_tmp_dir] = /tmp
php_admin_value[session.save_path] = /tmp
`, domain, domain, socketPath, version, domain, projectRoot)

	if utils.CheckFileExists(poolFile) {
		utils.Log("Updating PHP %s pool for %s", version, domain)
	} else {
		utils.Log("Creating PHP %s pool for %s", version, domain)
	}

	// Always write the pool config to ensure it's up to date
	_, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", poolFile, poolConfig))
	if err != nil {
		return fmt.Errorf("failed to create PHP pool: %v", err)
	}

	// Always restart PHP-FPM to load/reload the pool and create the socket
	serviceName := fmt.Sprintf("php%s-fpm", version)
	utils.Log("Restarting %s to load pool...", serviceName)
	if err := system.RestartService(serviceName); err != nil {
		return fmt.Errorf("failed to restart PHP-FPM: %v", err)
	}

	// Verify the socket was created
	if !utils.CheckFileExists(socketPath) {
		return fmt.Errorf("PHP-FPM socket was not created: %s (check PHP-FPM logs)", socketPath)
	}

	utils.Ok("PHP pool configured for %s", domain)

	return nil
}

// InstallComposer installs Composer globally
func InstallComposer(verifyOnly bool) error {
	if utils.CommandExists("composer") {
		// Try to get version
		version, err := utils.RunShell("timeout 5 composer --version 2>&1 | head -n1")
		if err == nil {
			utils.Verify("Composer already installed (%s)", strings.TrimSpace(version))
			return nil
		}

		utils.Warn("Composer command exists but not responding, reinstalling...")
		_, _ = utils.RunCommand("rm", "-f", "/usr/local/bin/composer")
	}

	if verifyOnly {
		utils.Fail("Composer not installed")
		return fmt.Errorf("composer not installed")
	}

	utils.Log("Installing Composer...")

	// Download signature
	utils.Log("Downloading Composer signature...")
	expectedSig, err := utils.RunShell("timeout 30 wget -q -O - https://composer.github.io/installer.sig 2>/dev/null")
	if err != nil {
		return fmt.Errorf("failed to download Composer signature: %v", err)
	}
	expectedSig = strings.TrimSpace(expectedSig)

	// Download installer
	utils.Log("Downloading Composer installer...")
	_, err = utils.RunShell("timeout 30 php -r \"copy('https://getcomposer.org/installer', 'composer-setup.php');\" 2>/dev/null")
	if err != nil {
		return fmt.Errorf("failed to download Composer installer: %v", err)
	}

	// Verify signature
	utils.Log("Verifying Composer signature...")
	actualSig, err := utils.RunShell("php -r \"echo hash_file('SHA384', 'composer-setup.php');\"")
	if err != nil {
		return fmt.Errorf("failed to verify signature: %v", err)
	}
	actualSig = strings.TrimSpace(actualSig)

	if expectedSig != actualSig {
		_, _ = utils.RunCommand("rm", "-f", "composer-setup.php")
		return fmt.Errorf("composer signature verification failed")
	}

	// Install Composer
	utils.Log("Installing Composer...")
	_, err = utils.RunShell("php composer-setup.php --quiet --install-dir=/usr/local/bin --filename=composer")
	if err != nil {
		return fmt.Errorf("failed to install Composer: %v", err)
	}

	// Cleanup
	_, _ = utils.RunCommand("rm", "-f", "composer-setup.php")

	utils.Ok("Composer installed successfully")
	return nil
}
