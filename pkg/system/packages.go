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
		// For Ubuntu, check if PPA is already added
		output, _ := utils.RunShell("grep -r 'ondrej/php' /etc/apt/sources.list.d/ 2>/dev/null")
		if strings.TrimSpace(output) != "" {
			utils.Verify("Sury PHP PPA already configured")
			return nil
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

	// Ensure software-properties-common is installed (provides add-apt-repository)
	utils.Log("Installing software-properties-common...")
	_, err := utils.RunCommand("apt-get", "install", "-y", "software-properties-common")
	if err != nil {
		return fmt.Errorf("failed to install software-properties-common: %v", err)
	}

	// Add the PPA using add-apt-repository
	utils.Log("Adding PPA: ppa:ondrej/php...")
	_, err = utils.RunCommand("add-apt-repository", "-y", "ppa:ondrej/php")
	if err != nil {
		return fmt.Errorf("failed to add PPA: %v\n\nTroubleshooting:\n  1. Check network: ping ppa.launchpad.net\n  2. Verify PPA exists: https://launchpad.net/~ondrej/+archive/ubuntu/php\n  3. Try manual: sudo add-apt-repository ppa:ondrej/php", err)
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
