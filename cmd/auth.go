package cmd

import (
	"fmt"
	"strings"
	"svp/pkg/utils"
	"svp/pkg/web"
	"svp/types"
)

// Auth handles basic authentication operations: enable, disable, check
func Auth(cfg *types.Config) error {
	domain := cfg.PrimaryDomain
	action := cfg.AuthAction

	utils.Section(fmt.Sprintf("Basic Authentication Management for %s", domain))

	// Ensure apache2-utils is installed (provides htpasswd)
	if err := installAuthPackages(); err != nil {
		return err
	}

	switch action {
	case "enable":
		return enableAuth(domain, cfg.AuthUsername, cfg.AuthPassword)
	case "disable":
		return disableAuth(domain)
	case "check":
		return checkAuth(domain)
	default:
		return fmt.Errorf("invalid action: %s (must be enable, disable, or check)", action)
	}
}

// installAuthPackages ensures apache2-utils is installed
func installAuthPackages() error {
	if utils.CheckPackageInstalled("apache2-utils") {
		utils.Verify("apache2-utils already installed")
		return nil
	}

	utils.Log("Installing apache2-utils (provides htpasswd)...")
	_, err := utils.RunCommand("apt-get", "install", "-y", "apache2-utils")
	if err != nil {
		return fmt.Errorf("failed to install apache2-utils: %v", err)
	}

	utils.Ok("apache2-utils installed")
	return nil
}

func enableAuth(domain, username, password string) error {
	utils.Section("Enabling Basic Authentication")

	// Verify domain directory exists
	siteDir := fmt.Sprintf("/var/www/%s", domain)
	if !utils.CheckFileExists(siteDir) {
		return fmt.Errorf("site directory not found: %s", siteDir)
	}

	// Prompt for username if not provided
	if username == "" {
		fmt.Print("Enter username for basic authentication: ")
		fmt.Scanln(&username)
		if username == "" {
			return fmt.Errorf("username is required")
		}
	}

	// Prompt for password if not provided
	if password == "" {
		fmt.Print("Enter password for basic authentication: ")
		fmt.Scanln(&password)
		if password == "" {
			return fmt.Errorf("password is required")
		}
	}

	// Create .htpasswd file
	htpasswdPath := fmt.Sprintf("%s/.htpasswd", siteDir)

	utils.Log("Creating/updating .htpasswd file...")

	// Use htpasswd to create/update the password file
	// -c creates the file (we use it always to ensure single user)
	// -B uses bcrypt encryption (more secure)
	cmd := fmt.Sprintf("htpasswd -cbB %s %s %s", htpasswdPath, username, password)
	_, err := utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("failed to create .htpasswd file: %v", err)
	}

	utils.Ok(".htpasswd file created/updated")

	// Update nginx configuration
	if err := updateNginxAuthConfig(domain, htpasswdPath, true); err != nil {
		return err
	}

	// Test and reload nginx
	if err := web.ReloadNginx(); err != nil {
		return err
	}

	utils.Ok("Basic authentication enabled for %s", domain)
	fmt.Println()
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password file: %s\n", htpasswdPath)
	fmt.Println()
	fmt.Printf("Your site now requires authentication: https://%s\n", domain)

	return nil
}

func disableAuth(domain string) error {
	utils.Section("Disabling Basic Authentication")

	siteDir := fmt.Sprintf("/var/www/%s", domain)
	htpasswdPath := fmt.Sprintf("%s/.htpasswd", siteDir)

	// Check if .htpasswd exists
	if !utils.CheckFileExists(htpasswdPath) {
		utils.Warn("Basic authentication not enabled for %s", domain)
		return nil
	}

	// Remove .htpasswd file
	utils.Log("Removing .htpasswd file...")
	_, err := utils.RunCommand("rm", "-f", htpasswdPath)
	if err != nil {
		utils.Warn("Failed to remove .htpasswd file: %v", err)
	} else {
		utils.Ok(".htpasswd file removed")
	}

	// Update nginx configuration
	if err := updateNginxAuthConfig(domain, "", false); err != nil {
		return err
	}

	// Test and reload nginx
	if err := web.ReloadNginx(); err != nil {
		return err
	}

	utils.Ok("Basic authentication disabled for %s", domain)
	fmt.Println()
	fmt.Printf("Your site is now accessible without authentication: https://%s\n", domain)

	return nil
}

func checkAuth(domain string) error {
	utils.Section("Basic Authentication Status")

	siteDir := fmt.Sprintf("/var/www/%s", domain)
	htpasswdPath := fmt.Sprintf("%s/.htpasswd", siteDir)
	vhostPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)

	// Check if .htpasswd exists
	if !utils.CheckFileExists(htpasswdPath) {
		utils.Warn("Basic authentication not enabled for %s", domain)
		fmt.Println()
		fmt.Printf("To enable authentication, run: svp auth %s enable\n", domain)
		return nil
	}

	utils.Ok("Basic authentication is enabled for %s", domain)
	fmt.Println()
	fmt.Printf("Password file: %s\n", htpasswdPath)

	// Get username from .htpasswd file
	content, err := utils.RunShell(fmt.Sprintf("cat %s", htpasswdPath))
	if err == nil {
		lines := strings.Split(strings.TrimSpace(content), "\n")
		if len(lines) > 0 {
			parts := strings.Split(lines[0], ":")
			if len(parts) > 0 {
				fmt.Printf("Username: %s\n", parts[0])
			}
		}
	}

	// Check nginx configuration
	if utils.CheckFileExists(vhostPath) {
		vhostContent, err := utils.RunShell(fmt.Sprintf("cat %s", vhostPath))
		if err == nil {
			if strings.Contains(vhostContent, "auth_basic") {
				fmt.Printf("Nginx configuration: Configured\n")
			} else {
				utils.Warn("Nginx configuration missing auth_basic directives")
			}
		}
	}

	fmt.Println()
	fmt.Printf("To update credentials, run: svp auth %s enable --username USER --password PASS\n", domain)
	fmt.Printf("To disable authentication, run: svp auth %s disable\n", domain)

	return nil
}

// updateNginxAuthConfig updates nginx configuration to add or remove basic auth
func updateNginxAuthConfig(domain, htpasswdPath string, enable bool) error {
	vhostPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)

	if !utils.CheckFileExists(vhostPath) {
		return fmt.Errorf("nginx vhost not found: %s", vhostPath)
	}

	utils.Log("Updating nginx configuration...")

	// Read current config
	content, err := utils.RunShell(fmt.Sprintf("cat %s", vhostPath))
	if err != nil {
		return fmt.Errorf("failed to read vhost config: %v", err)
	}

	lines := strings.Split(content, "\n")
	var result []string
	inServerBlock := false
	serverBlockDepth := 0
	authAlreadyExists := false

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Track server blocks
		if strings.Contains(trimmedLine, "server {") {
			inServerBlock = true
			serverBlockDepth = 0
		}

		if inServerBlock {
			serverBlockDepth += strings.Count(line, "{")
			serverBlockDepth -= strings.Count(line, "}")
		}

		// Remove existing auth_basic directives
		if strings.Contains(trimmedLine, "auth_basic") || strings.Contains(trimmedLine, "auth_basic_user_file") {
			authAlreadyExists = true
			continue // Skip this line
		}

		// Add auth directives after server_name in main server block (not in location blocks)
		if enable && inServerBlock && serverBlockDepth == 1 && strings.Contains(trimmedLine, "server_name") && strings.Contains(trimmedLine, ";") {
			result = append(result, line)
			// Get the indentation of the current line
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " ")))
			result = append(result, "")
			result = append(result, fmt.Sprintf("%s# Basic Authentication", indent))
			result = append(result, fmt.Sprintf("%sauth_basic \"Restricted Access\";", indent))
			result = append(result, fmt.Sprintf("%sauth_basic_user_file %s;", indent, htpasswdPath))
			continue
		}

		result = append(result, line)

		// Check if server block ended
		if inServerBlock && serverBlockDepth == 0 && i > 0 {
			inServerBlock = false
		}
	}

	// Write updated config
	newContent := strings.Join(result, "\n")
	_, err = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", vhostPath, newContent))
	if err != nil {
		return fmt.Errorf("failed to write updated config: %v", err)
	}

	if enable {
		if authAlreadyExists {
			utils.Ok("Nginx configuration updated (auth directives replaced)")
		} else {
			utils.Ok("Nginx configuration updated (auth directives added)")
		}
	} else {
		utils.Ok("Nginx configuration updated (auth directives removed)")
	}

	return nil
}
