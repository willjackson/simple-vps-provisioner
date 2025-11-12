package system

import (
	"fmt"
	"svp/pkg/utils"
	"strings"
)

// SetupFirewall configures UFW firewall
func SetupFirewall(enable bool, verifyOnly bool) error {
	if !enable {
		utils.Skip("Firewall configuration disabled")
		return nil
	}

	utils.Section("Firewall")

	// Check if UFW is installed
	if !utils.CheckPackageInstalled("ufw") {
		if verifyOnly {
			utils.Fail("UFW not installed")
			return fmt.Errorf("ufw not installed")
		}

		utils.Log("Installing UFW...")
		_, err := utils.RunCommand("apt-get", "install", "-y", "ufw")
		if err != nil {
			return fmt.Errorf("failed to install UFW: %v", err)
		}
	}

	// Check UFW status
	status, _ := utils.RunCommand("ufw", "status")
	isActive := strings.Contains(status, "Status: active")

	if !isActive {
		if verifyOnly {
			utils.Fail("UFW not active")
			return fmt.Errorf("ufw not active")
		}

		utils.Log("Configuring UFW firewall...")

		// Allow SSH before enabling
		if _, err := utils.RunCommand("ufw", "allow", "22/tcp"); err != nil {
			return fmt.Errorf("failed to allow SSH: %v", err)
		}

		// Allow HTTP
		if _, err := utils.RunCommand("ufw", "allow", "80/tcp"); err != nil {
			return fmt.Errorf("failed to allow HTTP: %v", err)
		}

		// Allow HTTPS
		if _, err := utils.RunCommand("ufw", "allow", "443/tcp"); err != nil {
			return fmt.Errorf("failed to allow HTTPS: %v", err)
		}

		// Enable UFW (non-interactive)
		_, err := utils.RunShell("echo 'y' | ufw enable")
		if err != nil {
			return fmt.Errorf("failed to enable UFW: %v", err)
		}

		utils.Ok("UFW firewall configured and active")
	} else {
		utils.Verify("UFW already active")
	}

	return nil
}
