package system

import (
	"fmt"
	"svp/pkg/utils"
	"strconv"
	"strings"
)

// CreateSwapIfNeeded creates swap space if needed based on configuration
func CreateSwapIfNeeded(createSwap string, verifyOnly bool) error {
	if createSwap == "no" {
		utils.Skip("Swap creation disabled by config")
		return nil
	}

	// Check if swap is already active
	output, _ := utils.RunShell("swapon --show --noheadings")
	if strings.TrimSpace(output) != "" {
		utils.Verify("Swap already active")
		return nil
	}

	// Get total memory in MB
	memInfo, err := utils.RunShell("awk '/MemTotal/ {print $2}' /proc/meminfo")
	if err != nil {
		return fmt.Errorf("failed to read memory info: %v", err)
	}

	totalKB, err := strconv.Atoi(strings.TrimSpace(memInfo))
	if err != nil {
		return fmt.Errorf("failed to parse memory info: %v", err)
	}
	totalMB := totalKB / 1024

	// Determine if swap is needed
	needSwap := false
	if createSwap == "yes" {
		needSwap = true
	} else if createSwap == "auto" && totalMB < 2000 {
		needSwap = true
	}

	if !needSwap {
		utils.Skip("Swap not needed (system has %d MB RAM)", totalMB)
		return nil
	}

	if verifyOnly {
		utils.Fail("Swap not configured (system has %d MB RAM)", totalMB)
		return fmt.Errorf("swap not configured")
	}

	utils.Log("Creating 2G swap...")

	// Create swap file if it doesn't exist
	if !utils.CheckFileExists("/swapfile") {
		// Try fallocate first, fall back to dd
		_, err := utils.RunCommand("fallocate", "-l", "2G", "/swapfile")
		if err != nil {
			utils.Warn("fallocate failed, using dd instead...")
			_, err = utils.RunCommand("dd", "if=/dev/zero", "of=/swapfile", "bs=1M", "count=2048")
			if err != nil {
				return fmt.Errorf("failed to create swap file: %v", err)
			}
		}

		// Set permissions
		if _, err := utils.RunCommand("chmod", "600", "/swapfile"); err != nil {
			return fmt.Errorf("failed to set swap file permissions: %v", err)
		}

		// Make swap
		if _, err := utils.RunCommand("mkswap", "/swapfile"); err != nil {
			return fmt.Errorf("failed to make swap: %v", err)
		}
	}

	// Enable swap
	_, _ = utils.RunCommand("swapon", "/swapfile")

	// Add to fstab if not already there
	fstabContent, _ := utils.RunShell("cat /etc/fstab")
	if !strings.Contains(fstabContent, "/swapfile") {
		_, err := utils.RunShell("echo '/swapfile none swap sw 0 0' >> /etc/fstab")
		if err != nil {
			return fmt.Errorf("failed to add swap to fstab: %v", err)
		}
	}

	// Set swappiness
	_, _ = utils.RunCommand("sysctl", "vm.swappiness=10")

	// Make swappiness persistent
	swappinessConf := "/etc/sysctl.d/99-swap.conf"
	if !utils.CheckFileExists(swappinessConf) {
		_, err := utils.RunShell("echo 'vm.swappiness=10' > " + swappinessConf)
		if err != nil {
			return fmt.Errorf("failed to make swappiness persistent: %v", err)
		}
	}

	utils.Ok("Swap configured and active")
	return nil
}
