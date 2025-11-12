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
func AddPHPRepoIfNeeded(verifyOnly bool) error {
	// Check if Sury repo is already configured
	output, _ := utils.RunShell("grep -q 'packages.sury.org' /etc/apt/sources.list.d/* 2>/dev/null && echo 'found'")

	if strings.TrimSpace(output) == "found" {
		utils.Verify("Sury PHP repo already configured")
		return nil
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

	// Get distribution codename
	codename, err := utils.RunShell("lsb_release -sc")
	if err != nil {
		return fmt.Errorf("failed to get distribution codename: %v", err)
	}
	codename = strings.TrimSpace(codename)

	// Add repository
	repoLine := fmt.Sprintf("deb [signed-by=/usr/share/keyrings/sury-keyring.gpg] https://packages.sury.org/php/ %s main", codename)
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
