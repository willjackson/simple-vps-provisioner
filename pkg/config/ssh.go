package config

import (
	"fmt"
	"strings"
	"svp/pkg/utils"
)

// EnsureAdminSSHKey ensures admin user has SSH key and GitHub is in known_hosts
func EnsureAdminSSHKey(adminUser string) error {
	sshDir := fmt.Sprintf("/home/%s/.ssh", adminUser)
	keyPath := fmt.Sprintf("%s/id_rsa", sshDir)
	pubKeyPath := fmt.Sprintf("%s.pub", keyPath)
	knownHostsPath := fmt.Sprintf("%s/known_hosts", sshDir)

	// Ensure .ssh directory exists
	if !utils.CheckDirExists(sshDir) {
		utils.Log("Creating .ssh directory for %s", adminUser)
		if err := utils.EnsureDir(sshDir); err != nil {
			return fmt.Errorf("failed to create .ssh directory: %v", err)
		}
		_, _ = utils.RunCommand("chown", fmt.Sprintf("%s:%s", adminUser, adminUser), sshDir)
		_, _ = utils.RunCommand("chmod", "700", sshDir)
	}

	// Check if SSH key exists
	if !utils.CheckFileExists(keyPath) {
		utils.Log("Generating SSH key for %s...", adminUser)
		
		// Generate RSA 4096 key
		cmd := fmt.Sprintf("sudo -u %s ssh-keygen -t rsa -b 4096 -f %s -N '' -C '%s@vps'", adminUser, keyPath, adminUser)
		_, err := utils.RunShell(cmd)
		if err != nil {
			return fmt.Errorf("failed to generate SSH key: %v", err)
		}
		
		utils.Ok("SSH key generated")
		
		// Display public key
		pubKey, err := utils.RunShell(fmt.Sprintf("cat %s", pubKeyPath))
		if err != nil {
			return fmt.Errorf("failed to read public key: %v", err)
		}
		
		fmt.Println()
		fmt.Println("==========================================================")
		fmt.Println("SSH PUBLIC KEY - Add this to your Git repository:")
		fmt.Println("==========================================================")
		fmt.Println(strings.TrimSpace(pubKey))
		fmt.Println("==========================================================")
		fmt.Println()
		fmt.Println("For GitHub:")
		fmt.Println("  1. Go to https://github.com/settings/keys")
		fmt.Println("  2. Click 'New SSH key'")
		fmt.Println("  3. Paste the key above")
		fmt.Println()
		fmt.Println("For GitLab:")
		fmt.Println("  1. Go to https://gitlab.com/-/profile/keys")
		fmt.Println("  2. Paste the key above")
		fmt.Println()
		fmt.Print("Press Enter after adding the key to continue...")
		fmt.Scanln()
	} else {
		utils.Verify("SSH key already exists for %s", adminUser)
	}

	// Add GitHub to known_hosts
	if !utils.CheckFileExists(knownHostsPath) {
		utils.RunShell(fmt.Sprintf("touch %s", knownHostsPath))
		_, _ = utils.RunCommand("chown", fmt.Sprintf("%s:%s", adminUser, adminUser), knownHostsPath)
		_, _ = utils.RunCommand("chmod", "644", knownHostsPath)
	}

	// Check if GitHub already in known_hosts
	knownHosts, _ := utils.RunShell(fmt.Sprintf("cat %s", knownHostsPath))
	if !strings.Contains(knownHosts, "github.com") {
		utils.Log("Adding GitHub to known_hosts...")
		cmd := fmt.Sprintf("ssh-keyscan -H github.com >> %s 2>/dev/null", knownHostsPath)
		_, _ = utils.RunShell(cmd)
	}
	
	// Add GitLab to known_hosts
	if !strings.Contains(knownHosts, "gitlab.com") {
		utils.Log("Adding GitLab to known_hosts...")
		cmd := fmt.Sprintf("ssh-keyscan -H gitlab.com >> %s 2>/dev/null", knownHostsPath)
		_, _ = utils.RunShell(cmd)
	}

	return nil
}
