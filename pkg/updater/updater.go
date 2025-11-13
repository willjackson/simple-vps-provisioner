package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"svp/pkg/utils"
)

const (
	GitHubRepo = "willjackson/simple-vps-provisioner"
	APIBaseURL = "https://api.github.com/repos"
)

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckForUpdates checks if a new version is available
func CheckForUpdates(currentVersion string) (string, bool, error) {
	utils.Log("Checking for updates...")

	// Fetch latest release from GitHub API
	apiURL := fmt.Sprintf("%s/%s/releases/latest", APIBaseURL, GitHubRepo)
	output, err := utils.RunShell(fmt.Sprintf("curl -s %s", apiURL))
	if err != nil {
		return "", false, fmt.Errorf("failed to fetch release info: %v", err)
	}

	var release Release
	if err := json.Unmarshal([]byte(output), &release); err != nil {
		return "", false, fmt.Errorf("failed to parse release info: %v", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion = strings.TrimPrefix(currentVersion, "v")

	if latestVersion == currentVersion {
		utils.Ok("Already running the latest version: v%s", currentVersion)
		return latestVersion, false, nil
	}

	utils.Log("Current version: v%s", currentVersion)
	utils.Log("Latest version:  v%s", latestVersion)
	return latestVersion, true, nil
}

// Update performs the update to the latest version
func Update(currentVersion string) error {
	latestVersion, hasUpdate, err := CheckForUpdates(currentVersion)
	if err != nil {
		return err
	}

	if !hasUpdate {
		return nil
	}

	// Prompt user
	fmt.Printf("\nNew version available: v%s\n", latestVersion)
	fmt.Print("Update now? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" {
		utils.Skip("Update cancelled")
		return nil
	}

	// Determine binary name for current platform
	binaryName := getBinaryName()
	if binaryName == "" {
		return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	utils.Log("Downloading svp v%s for %s/%s...", latestVersion, runtime.GOOS, runtime.GOARCH)

	// Fetch release info
	apiURL := fmt.Sprintf("%s/%s/releases/latest", APIBaseURL, GitHubRepo)
	output, err := utils.RunShell(fmt.Sprintf("curl -s %s", apiURL))
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %v", err)
	}

	var release Release
	if err := json.Unmarshal([]byte(output), &release); err != nil {
		return fmt.Errorf("failed to parse release info: %v", err)
	}

	// Find binary and checksum URLs
	var binaryURL, checksumURL string
	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			binaryURL = asset.BrowserDownloadURL
		} else if asset.Name == "checksums.txt" {
			checksumURL = asset.BrowserDownloadURL
		}
	}

	if binaryURL == "" {
		return fmt.Errorf("binary not found in release: %s", binaryName)
	}

	// Download binary to temp location
	tmpBinary := fmt.Sprintf("/tmp/svp-update-%s", binaryName)
	_, err = utils.RunShell(fmt.Sprintf("curl -L -o %s %s", tmpBinary, binaryURL))
	if err != nil {
		return fmt.Errorf("failed to download binary: %v", err)
	}
	defer os.Remove(tmpBinary)

	// Download and verify checksum if available
	if checksumURL != "" {
		utils.Log("Verifying checksum...")
		tmpChecksum := "/tmp/svp-checksums.txt"
		_, err = utils.RunShell(fmt.Sprintf("curl -L -o %s %s", tmpChecksum, checksumURL))
		if err != nil {
			utils.Warn("Failed to download checksums, skipping verification")
		} else {
			defer os.Remove(tmpChecksum)
			// Verify checksum
			cmd := fmt.Sprintf("cd /tmp && sha256sum --check --ignore-missing %s 2>&1 | grep %s", tmpChecksum, binaryName)
			_, err = utils.RunShell(cmd)
			if err != nil {
				return fmt.Errorf("checksum verification failed")
			}
			utils.Ok("Checksum verified")
		}
	}

	// Get current binary path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %v", err)
	}

	// Backup current binary
	backupPath := exePath + ".backup"
	utils.Log("Backing up current binary to %s", backupPath)
	_, err = utils.RunCommand("cp", exePath, backupPath)
	if err != nil {
		return fmt.Errorf("failed to backup binary: %v", err)
	}

	// Replace binary
	utils.Log("Installing new version...")
	_, err = utils.RunCommand("mv", tmpBinary, exePath)
	if err != nil {
		// Restore backup
		utils.RunCommand("mv", backupPath, exePath)
		return fmt.Errorf("failed to install update: %v", err)
	}

	// Make executable
	_, err = utils.RunCommand("chmod", "+x", exePath)
	if err != nil {
		utils.Warn("Failed to set executable permission: %v", err)
	}

	// Remove backup
	os.Remove(backupPath)

	utils.Ok("Successfully updated to v%s!", latestVersion)
	fmt.Println("\nPlease restart svp to use the new version.")
	return nil
}

// getBinaryName returns the binary name for the current platform
func getBinaryName() string {
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	switch platform {
	case "linux-amd64":
		return "svp-linux-amd64"
	case "linux-arm64":
		return "svp-linux-arm64"
	case "darwin-amd64":
		return "svp-darwin-amd64"
	case "darwin-arm64":
		return "svp-darwin-arm64"
	default:
		return ""
	}
}
