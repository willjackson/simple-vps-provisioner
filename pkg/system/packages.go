package system

import (
	"fmt"
	"svp/pkg/utils"
	"strings"
)

// BasePackages returns the list of base packages to install
var BasePackages = []string{
	"ca-certificates", "gnupg", "lsb-release", "curl", "wget",
	"unzip", "git", "ufw", "apt-transport-https", "acl",
	"nano", "jq", "htop", "bind9-dnsutils",
}

// AddPHPRepoIfNeeded adds the Sury PHP repository if not already configured
// For Ubuntu: Uses PPA method (add-apt-repository ppa:ondrej/php)
// For Debian: Uses direct deb line with GPG key
// For Debian testing/unstable (trixie, sid): Uses native Debian packages instead
func AddPHPRepoIfNeeded(verifyOnly bool) error {
	// Detect OS distribution first to determine method
	distroID, err := utils.RunShell("lsb_release -si")
	if err != nil {
		return fmt.Errorf("failed to get distribution ID: %v", err)
	}
	distroID = strings.TrimSpace(distroID)

	// Get distribution codename
	codename, err := utils.RunShell("lsb_release -sc")
	if err != nil {
		return fmt.Errorf("failed to get distribution codename: %v", err)
	}
	codename = strings.TrimSpace(codename)

	// Verify OS detection - cross-check with /etc/os-release if available
	if osReleaseID, err := utils.RunShell("grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'"); err == nil {
		osReleaseID = strings.TrimSpace(strings.ToLower(osReleaseID))
		lsbID := strings.ToLower(distroID)
		
		// If there's a mismatch and os-release shows debian/ubuntu, trust it
		if osReleaseID != lsbID && (osReleaseID == "debian" || osReleaseID == "ubuntu") {
			utils.Warn("OS mismatch detected: lsb_release reports '%s' but os-release shows '%s'", distroID, osReleaseID)
			utils.Log("Using os-release ID: %s", osReleaseID)
			if osReleaseID == "debian" {
				distroID = "Debian"
			} else if osReleaseID == "ubuntu" {
				distroID = "Ubuntu"
			}
		}
	}

	distroLower := strings.ToLower(distroID)
	
	// Check if already configured
	if distroLower == "ubuntu" {
		// For Ubuntu, check ALL possible PPA configurations and validate codenames
		// We need to check multiple locations because old runs might have created different files
		var needsReconfiguration bool
		var hasValidConfig bool
		
		// Check all sources.list.d files for ondrej/php
		files, _ := utils.RunShell("grep -l 'ondrej/php' /etc/apt/sources.list.d/*.list 2>/dev/null || true")
		if strings.TrimSpace(files) != "" {
			fileList := strings.Split(strings.TrimSpace(files), "\n")
			
			for _, file := range fileList {
				if file == "" {
					continue
				}
				
				repoContent, _ := utils.RunShell(fmt.Sprintf("cat %s 2>/dev/null", file))
				lines := strings.Split(repoContent, "\n")
				
				for _, line := range lines {
					if strings.HasPrefix(line, "deb") && strings.Contains(line, "ondrej/php") {
						parts := strings.Fields(line)
						if len(parts) >= 4 {
							currentCodename := parts[3]
							supportedCodename := getSupportedPPACodename(currentCodename)
							
							if supportedCodename != currentCodename {
								utils.Warn("Found PPA in %s with unsupported codename: %s", file, currentCodename)
								needsReconfiguration = true
								// Remove this file
								_, _ = utils.RunCommand("rm", "-f", file)
							} else {
								hasValidConfig = true
							}
						}
					}
				}
			}
		}
		
		// If we found a valid configuration and no invalid ones, we're good
		if hasValidConfig && !needsReconfiguration {
			utils.Verify("Sury PHP PPA already configured")
			return nil
		}
		
		// If we need reconfiguration, log it
		if needsReconfiguration {
			utils.Log("Reconfiguring PPA with correct codename...")
		}
	} else {
		// For Debian, check if repo is already configured
		output, _ := utils.RunShell("grep -q 'packages.sury.org' /etc/apt/sources.list.d/* 2>/dev/null && echo 'found'")
		if strings.TrimSpace(output) == "found" {
			// Check if the configured repo has a valid/supported codename
			repoContent, _ := utils.RunShell("cat /etc/apt/sources.list.d/sury-php.list 2>/dev/null")
			
			if strings.Contains(repoContent, "packages.sury.org/php/") {
				parts := strings.Fields(repoContent)
				if len(parts) >= 4 {
					currentCodename := parts[3]
					
					// Check if this is an unsupported/invalid codename
					unsupportedCodenames := []string{"trixie", "sid", "forky"}
					isUnsupported := false
					for _, unsupported := range unsupportedCodenames {
						if currentCodename == unsupported {
							isUnsupported = true
							break
						}
					}
					
					if isUnsupported {
						utils.Warn("Found invalid/unsupported Sury repository codename: %s", currentCodename)
						utils.Log("Reconfiguring with supported codename...")
						_, _ = utils.RunCommand("rm", "-f", "/etc/apt/sources.list.d/sury-php.list")
					} else {
						utils.Verify("Sury PHP repo already configured")
						return nil
					}
				} else {
					utils.Verify("Sury PHP repo already configured")
					return nil
				}
			} else {
				utils.Verify("Sury PHP repo already configured")
				return nil
			}
		}
	}

	if verifyOnly {
		utils.Fail("Sury PHP repository not configured")
		return fmt.Errorf("sury PHP repository not configured")
	}

	// Get version ID for better logging
	versionID := ""
	if vid, err := utils.RunShell("grep '^VERSION_ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'"); err == nil {
		versionID = strings.TrimSpace(vid)
	}

	// Display detection results
	if versionID != "" {
		utils.Log("Detected %s %s (version %s)", distroID, codename, versionID)
	} else {
		utils.Log("Detected %s %s", distroID, codename)
	}

	// For Debian testing/unstable (trixie, sid, forky), use native Debian packages
	if distroLower == "debian" {
		testingVersions := []string{"trixie", "sid", "forky"}
		for _, testVer := range testingVersions {
			if codename == testVer {
				utils.Log("Debian %s detected - using native Debian PHP packages", codename)
				utils.Verify("Native Debian repositories will be used (Sury not needed)")
				return nil
			}
		}
	}

	utils.Log("Adding Sury PHP repository...")

	// Install base dependencies
	_, err = utils.RunCommand("apt-get", "install", "-y", "--no-install-recommends",
		"ca-certificates", "curl", "gnupg", "lsb-release")
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	// Ubuntu: Use PPA method
	if distroLower == "ubuntu" {
		return addPHPRepoUbuntu()
	}

	// Debian: Use direct deb line method
	return addPHPRepoDebian(codename)
}

// addPHPRepoUbuntu adds the Sury PHP repository for Ubuntu using PPA method
func addPHPRepoUbuntu() error {
	utils.Log("Adding Sury PHP PPA for Ubuntu...")

	// Get current codename to check if it's supported
	codename, err := utils.RunShell("lsb_release -sc")
	if err != nil {
		return fmt.Errorf("failed to get codename: %v", err)
	}
	codename = strings.TrimSpace(codename)

	// Map to supported Ubuntu codename for PPA
	ppaCodename := getSupportedPPACodename(codename)
	if ppaCodename != codename {
		utils.Warn("Ubuntu %s not yet supported by PPA, using %s packages", codename, ppaCodename)
	}

	// Ensure software-properties-common is installed (provides add-apt-repository)
	utils.Log("Installing software-properties-common...")
	_, err = utils.RunCommand("apt-get", "install", "-y", "software-properties-common")
	if err != nil {
		return fmt.Errorf("failed to install software-properties-common: %v", err)
	}

	// Clean up any old PPA files that might exist from previous runs
	// This ensures we don't have conflicting configurations
	utils.Log("Cleaning up any old PPA configurations...")
	// Remove any files with ondrej and php in the name
	_, _ = utils.RunShell("rm -f /etc/apt/sources.list.d/*ondrej*php*.list 2>/dev/null || true")
	// Also check for generic ppa files that might contain ondrej/php
	_, _ = utils.RunShell("grep -l 'ondrej/php' /etc/apt/sources.list.d/*.list 2>/dev/null | xargs rm -f 2>/dev/null || true")
	// Remove from main sources.list if present (shouldn't be, but just in case)
	_, _ = utils.RunShell("sed -i '/ondrej\/php/d' /etc/apt/sources.list 2>/dev/null || true")

	// Import PPA GPG key using modern method
	utils.Log("Importing PPA GPG key...")
	// Download key to proper location for modern apt
	keyringPath := "/etc/apt/keyrings/ondrej-php-ppa.gpg"
	_, err = utils.RunShell("mkdir -p /etc/apt/keyrings")
	if err != nil {
		return fmt.Errorf("failed to create keyrings directory: %v", err)
	}
	
	// Download and convert key in one step
	_, err = utils.RunShell(fmt.Sprintf("curl -fsSL 'https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x4F4EA0AAE5267A6C' | gpg --dearmor -o %s 2>/dev/null", keyringPath))
	if err != nil {
		utils.Warn("Failed to download key from keyserver, trying alternative method...")
		// Try using apt-key as fallback (though deprecated)
		_, err = utils.RunShell("apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 4F4EA0AAE5267A6C 2>/dev/null || true")
		if err != nil {
			utils.Warn("Alternative key import also failed, continuing anyway...")
		}
	}
	
	// Verify key file was created
	if utils.CheckFileExists(keyringPath) {
		utils.Ok("PPA GPG key imported successfully")
	} else {
		utils.Warn("Key file not created, but continuing (may already be in apt-key)")
	}

	// Manually create the sources.list.d file with the mapped codename
	// This is more reliable than add-apt-repository for unsupported versions
	utils.Log("Adding PPA sources for Ubuntu %s...", ppaCodename)
	
	// Use modern signed-by syntax if we have the keyring file
	var sourcesLine string
	if utils.CheckFileExists(keyringPath) {
		sourcesLine = fmt.Sprintf("deb [signed-by=%s] https://ppa.launchpadcontent.net/ondrej/php/ubuntu %s main", keyringPath, ppaCodename)
	} else {
		// Fallback to old method without signed-by (relies on apt-key)
		sourcesLine = fmt.Sprintf("deb https://ppa.launchpadcontent.net/ondrej/php/ubuntu %s main", ppaCodename)
	}
	
	sourcesFile := "/etc/apt/sources.list.d/ondrej-ubuntu-php.list"
	
	_, err = utils.RunShell(fmt.Sprintf("echo '%s' > %s", sourcesLine, sourcesFile))
	if err != nil {
		return fmt.Errorf("failed to create PPA sources file: %v", err)
	}

	// Also add the source repository (deb-src)
	sourcesLineSrc := fmt.Sprintf("# deb-src https://ppa.launchpadcontent.net/ondrej/php/ubuntu %s main", ppaCodename)
	_, err = utils.RunShell(fmt.Sprintf("echo '%s' >> %s", sourcesLineSrc, sourcesFile))
	if err != nil {
		utils.Warn("Failed to add deb-src line: %v", err)
	}

	// Update package lists
	utils.Log("Updating package lists...")
	_, err = utils.RunCommand("apt-get", "update")
	if err != nil {
		return fmt.Errorf("failed to update package lists: %v", err)
	}

	utils.Ok("Sury PHP PPA added successfully")
	return nil
}

// addPHPRepoDebian adds the Sury PHP repository for Debian using direct deb line method
func addPHPRepoDebian(codename string) error {
	utils.Log("Adding Sury PHP repository for Debian...")

	// Download and install GPG key
	utils.Log("Downloading Sury GPG key...")
	_, err := utils.RunShell("curl -fsSL https://packages.sury.org/php/apt.gpg | gpg --dearmor --yes -o /usr/share/keyrings/sury-keyring.gpg")
	if err != nil {
		return fmt.Errorf("failed to download GPG key: %v", err)
	}

	// Verify GPG key was installed
	if !utils.CheckFileExists("/usr/share/keyrings/sury-keyring.gpg") {
		return fmt.Errorf("GPG key file not created at /usr/share/keyrings/sury-keyring.gpg")
	}
	utils.Ok("Sury GPG key installed")

	// Map to supported codename
	suryCodename := getSupportedSuryCodenameDebian(codename)
	if suryCodename != codename {
		utils.Warn("Debian %s not yet supported by Sury, using %s repository", codename, suryCodename)
	}

	// Add repository
	utils.Log("Adding Sury repository for Debian %s...", suryCodename)
	repoLine := fmt.Sprintf("deb [signed-by=/usr/share/keyrings/sury-keyring.gpg] https://packages.sury.org/php/ %s main", suryCodename)
	_, err = utils.RunShell(fmt.Sprintf("echo '%s' > /etc/apt/sources.list.d/sury-php.list", repoLine))
	if err != nil {
		return fmt.Errorf("failed to add repository: %v", err)
	}

	// Update package lists with retry logic
	utils.Log("Updating package lists...")
	maxRetries := 3
	var updateErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			utils.Warn("Retry %d/%d: Updating package lists...", i, maxRetries-1)
		}
		
		output, err := utils.RunCommand("apt-get", "update")
		if err == nil {
			updateErr = nil
			break
		}
		
		updateErr = err
		
		// Check for specific error patterns
		if strings.Contains(output, "418") || strings.Contains(output, "I'm a teapot") {
			utils.Warn("Received HTTP 418 error from Sury repository")
			utils.Warn("This typically indicates rate limiting or temporary blocking")
			utils.Warn("Waiting 10 seconds before retry...")
			_, _ = utils.RunShell("sleep 10")
			continue
		}
		
		if strings.Contains(output, "not signed") || strings.Contains(output, "NO_PUBKEY") {
			utils.Warn("Repository signing key issue detected")
			utils.Warn("Attempting to reinstall GPG key...")
			_, _ = utils.RunShell("curl -fsSL https://packages.sury.org/php/apt.gpg | gpg --dearmor --yes -o /usr/share/keyrings/sury-keyring.gpg")
			continue
		}
		
		// For other errors, wait briefly before retry
		if i < maxRetries-1 {
			_, _ = utils.RunShell("sleep 5")
		}
	}
	
	if updateErr != nil {
		return fmt.Errorf("failed to update package lists after %d attempts: %v\n\nTroubleshooting steps:\n  1. Check network connectivity: ping packages.sury.org\n  2. Try again in a few minutes (may be rate limited)\n  3. Check if IP is blocked: curl -I https://packages.sury.org/php/\n  4. Consider using native distribution packages if available", maxRetries, updateErr)
	}

	utils.Ok("Sury PHP repository added")
	return nil
}

// EnsureBasePackages ensures all base packages are installed
func EnsureBasePackages(verifyOnly bool) error {
	utils.Verify("Checking base packages...")

	var missing []string
	for _, pkg := range BasePackages {
		if !utils.CheckPackageInstalled(pkg) {
			missing = append(missing, pkg)
		}
	}

	if len(missing) > 0 {
		if verifyOnly {
			utils.Fail("Missing packages: %s", strings.Join(missing, ", "))
			return fmt.Errorf("missing packages")
		}

		utils.Log("Installing missing packages: %s", strings.Join(missing, ", "))

		// Update package lists
		if _, err := utils.RunCommand("apt-get", "update"); err != nil {
			return fmt.Errorf("failed to update package lists: %v", err)
		}

		// Upgrade existing packages
		if _, err := utils.RunCommand("apt-get", "upgrade", "-y"); err != nil {
			return fmt.Errorf("failed to upgrade packages: %v", err)
		}

		// Install missing packages
		args := append([]string{"install", "-y", "--no-install-recommends"}, missing...)
		if _, err := utils.RunCommand("apt-get", args...); err != nil {
			return fmt.Errorf("failed to install packages: %v", err)
		}

		utils.Ok("Base packages installed")
	} else {
		utils.Ok("All base packages installed")
	}

	return nil
}

// getSupportedSuryCodenameDebian maps Debian codenames to Sury-supported versions
func getSupportedSuryCodenameDebian(codename string) string {
	supportedDebianCodenames := map[string]string{
		// Stable versions - fully supported
		"bookworm": "bookworm", // Debian 12
		"bullseye": "bullseye", // Debian 11
		"buster":   "buster",   // Debian 10
		
		// Testing/Unstable - use latest stable
		"trixie": "bookworm", // Debian 13 (testing) -> use Debian 12
		"forky":  "bookworm", // Debian 14 (unstable) -> use Debian 12
		"sid":    "bookworm", // Unstable -> use Debian 12
	}
	
	if mapped, ok := supportedDebianCodenames[codename]; ok {
		return mapped
	}
	
	// Unknown - default to bookworm
	return "bookworm"
}

// getSupportedPPACodename maps Ubuntu codenames to PPA-supported versions
func getSupportedPPACodename(codename string) string {
	supportedUbuntuCodenames := map[string]string{
		// Supported LTS versions (PPA has these)
		"noble":  "noble",  // Ubuntu 24.04 LTS
		"jammy":  "jammy",  // Ubuntu 22.04 LTS
		"focal":  "focal",  // Ubuntu 20.04 LTS
		"bionic": "bionic", // Ubuntu 18.04 LTS
		
		// Interim/development releases - map to latest LTS
		"oracular": "noble",    // Ubuntu 24.10 -> use 24.04 LTS
		"plucky":   "noble",    // Ubuntu 25.04 -> use 24.04 LTS
		"questing": "noble",    // Ubuntu 25.10 -> use 24.04 LTS
		"mantic":   "jammy",    // Ubuntu 23.10 -> use 22.04 LTS
		"lunar":    "jammy",    // Ubuntu 23.04 -> use 22.04 LTS
		"kinetic":  "jammy",    // Ubuntu 22.10 -> use 22.04 LTS
	}
	
	if mapped, ok := supportedUbuntuCodenames[codename]; ok {
		return mapped
	}
	
	// Unknown Ubuntu version - default to noble (24.04 LTS) as it's the latest LTS
	return "noble"
}
