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
// For Debian testing/unstable (trixie, sid), uses native Debian packages instead
func AddPHPRepoIfNeeded(verifyOnly bool) error {
	// Check if Sury repo is already configured
	output, _ := utils.RunShell("grep -q 'packages.sury.org' /etc/apt/sources.list.d/* 2>/dev/null && echo 'found'")

	if strings.TrimSpace(output) == "found" {
		// Check if the configured repo has a valid/supported codename
		repoContent, _ := utils.RunShell("cat /etc/apt/sources.list.d/sury-php.list 2>/dev/null")
		
		// Extract codename from the repo line
		if strings.Contains(repoContent, "packages.sury.org/php/") {
			parts := strings.Fields(repoContent)
			if len(parts) >= 4 {
				currentCodename := parts[3]
				
				// Check if this is an unsupported/invalid codename
				unsupportedCodenames := []string{"questing", "trixie", "sid", "forky", "oracular", "plucky"}
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
					// Remove the old config and reconfigure
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

	if verifyOnly {
		utils.Fail("Sury PHP repository not configured")
		return fmt.Errorf("sury PHP repository not configured")
	}

	utils.Log("Adding Sury PHP repository...")

	// Install dependencies
	_, err := utils.RunCommand("apt-get", "install", "-y", "--no-install-recommends",
		"ca-certificates", "curl", "gnupg")
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	// Download and install GPG key
	_, err = utils.RunShell("curl -fsSL https://packages.sury.org/php/apt.gpg | gpg --dearmor -o /usr/share/keyrings/sury-keyring.gpg")
	if err != nil {
		return fmt.Errorf("failed to download GPG key: %v", err)
	}

	// Detect OS distribution (Debian or Ubuntu)
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
	// This helps catch misconfigurations where lsb_release reports wrong info
	if osReleaseID, err := utils.RunShell("grep '^ID=' /etc/os-release 2>/dev/null | cut -d= -f2 | tr -d '\"'"); err == nil {
		osReleaseID = strings.TrimSpace(strings.ToLower(osReleaseID))
		lsbID := strings.ToLower(distroID)
		
		// If there's a mismatch and os-release shows debian/ubuntu, trust it
		if osReleaseID != lsbID && (osReleaseID == "debian" || osReleaseID == "ubuntu") {
			utils.Warn("OS mismatch detected: lsb_release reports '%s' but os-release shows '%s'", distroID, osReleaseID)
			utils.Log("Using os-release ID: %s", osReleaseID)
			// Correct the distro ID
			if osReleaseID == "debian" {
				distroID = "Debian"
			} else if osReleaseID == "ubuntu" {
				distroID = "Ubuntu"
			}
		}
	}

	utils.Log("Detected %s %s", distroID, codename)

	// For Debian testing/unstable (trixie, sid, forky), use native Debian packages
	// These versions have recent PHP in their main repos and Sury packages
	// from stable releases cause dependency conflicts
	distroLower := strings.ToLower(distroID)
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

	// Sury repository may not support newer/testing Debian/Ubuntu versions yet
	// Map to supported versions or use fallback
	suryCodename := getSupportedSuryCodename(distroID, codename)
	if suryCodename != codename {
		utils.Warn("%s %s not yet supported by Sury, using %s repository", distroID, codename, suryCodename)
	}

	// Add repository
	repoLine := fmt.Sprintf("deb [signed-by=/usr/share/keyrings/sury-keyring.gpg] https://packages.sury.org/php/ %s main", suryCodename)
	_, err = utils.RunShell(fmt.Sprintf("echo '%s' > /etc/apt/sources.list.d/sury-php.list", repoLine))
	if err != nil {
		return fmt.Errorf("failed to add repository: %v", err)
	}

	// Update package lists
	_, err = utils.RunCommand("apt-get", "update", "-y")
	if err != nil {
		return fmt.Errorf("failed to update package lists: %v", err)
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
		if _, err := utils.RunCommand("apt-get", "update", "-y"); err != nil {
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

// getSupportedSuryCodename maps Debian/Ubuntu codenames to Sury-supported versions
// Sury may not support the latest testing/unstable releases immediately
func getSupportedSuryCodename(distroID, codename string) string {
	// Normalize distribution ID
	distroID = strings.ToLower(distroID)
	
	if distroID == "ubuntu" {
		// Ubuntu codename mappings
		supportedUbuntuCodenames := map[string]string{
			// Supported LTS versions
			"jammy":    "jammy",    // Ubuntu 22.04 LTS (most stable for Sury)
			"focal":    "focal",    // Ubuntu 20.04 LTS
			"bionic":   "bionic",   // Ubuntu 18.04 LTS
			
			// Noble (24.04) - may not be supported yet by Sury
			"noble":    "jammy",    // Ubuntu 24.04 LTS -> fallback to 22.04 for now
			
			// Interim releases (supported for 9 months)
			"mantic":   "jammy",    // Ubuntu 23.10 -> fallback to 22.04 LTS
			"lunar":    "jammy",    // Ubuntu 23.04 -> fallback to 22.04 LTS
			"kinetic":  "jammy",    // Ubuntu 22.10 -> fallback to 22.04 LTS
			
			// Development/Future releases
			"oracular": "jammy",    // Ubuntu 24.10 (dev) -> use 22.04 LTS
			"plucky":   "jammy",    // Ubuntu 25.04 (future) -> use 22.04 LTS
			
			// Invalid/corrupted codenames
			"questing": "jammy",    // Corrupted detection -> use 22.04 LTS
		}
		
		if mapped, ok := supportedUbuntuCodenames[codename]; ok {
			return mapped
		}
		
		// Unknown Ubuntu version - default to jammy (22.04 LTS) which is well-supported
		return "jammy"
	}
	
	// Debian codename mappings (default)
	supportedDebianCodenames := map[string]string{
		// Stable versions - fully supported
		"bookworm":  "bookworm",  // Debian 12
		"bullseye":  "bullseye",  // Debian 11
		"buster":    "buster",    // Debian 10
		
		// Testing/Unstable - may not be supported yet, fallback to latest stable
		"trixie":    "bookworm",  // Debian 13 (testing) -> use Debian 12 repo
		"forky":     "bookworm",  // Debian 14 (unstable) -> use Debian 12 repo
		"sid":       "bookworm",  // Unstable -> use Debian 12 repo
		
		// Unknown/typos - common misspellings or corrupted data
		"questing":  "bookworm",  // Likely corrupted "trixie" detection
	}
	
	// If we have a mapping, use it
	if mapped, ok := supportedDebianCodenames[codename]; ok {
		return mapped
	}
	
	// Unknown Debian codename - default to bookworm (Debian 12) as safe fallback
	// Bookworm is the current stable release and should work for most cases
	return "bookworm"
}
