package ssl

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"svp/pkg/system"
	"svp/pkg/utils"
)

// InstallCertbot installs certbot for Let's Encrypt
func InstallCertbot(verifyOnly bool) error {
	if utils.CheckPackageInstalled("certbot") && utils.CheckPackageInstalled("python3-certbot-nginx") {
		utils.Verify("Certbot already installed")
		return nil
	}

	if verifyOnly {
		utils.Fail("Certbot not installed")
		return fmt.Errorf("certbot not installed")
	}

	utils.Log("Installing Certbot...")
	_, err := utils.RunCommand("apt-get", "install", "-y", "certbot", "python3-certbot-nginx")
	if err != nil {
		return fmt.Errorf("failed to install certbot: %v", err)
	}

	utils.Ok("Certbot installed")
	return nil
}

// GetPublicIP retrieves the server's public IP address
func GetPublicIP() (string, error) {
	utils.Log("Getting server's public IP address...")
	
	// Try multiple services in case one is down
	services := []string{
		"wget -qO- https://ipinfo.io/ip",
		"wget -qO- https://api.ipify.org",
		"wget -qO- https://icanhazip.com",
	}
	
	for _, cmd := range services {
		ip, err := utils.RunShell(cmd)
		if err == nil {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				utils.Log("Server public IP: %s", ip)
				return ip, nil
			}
		}
	}
	
	return "", fmt.Errorf("failed to retrieve public IP from any service")
}

// GetDomainIP retrieves the IP address that a domain resolves to
func GetDomainIP(domain string) (string, error) {
	utils.Log("Looking up DNS for %s...", domain)
	
	// Try dig first (more reliable)
	if utils.CommandExists("dig") {
		cmd := fmt.Sprintf("dig +short %s | grep -E '^[0-9.]+$' | head -n1", domain)
		ip, err := utils.RunShell(cmd)
		if err == nil {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip, nil
			}
		}
	}
	
	// Fallback to nslookup
	if utils.CommandExists("nslookup") {
		cmd := fmt.Sprintf("nslookup %s | grep 'Address:' | tail -n1 | awk '{print $2}'", domain)
		ip, err := utils.RunShell(cmd)
		if err == nil {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip, nil
			}
		}
	}
	
	// Fallback to host command
	if utils.CommandExists("host") {
		cmd := fmt.Sprintf("host %s | grep 'has address' | awk '{print $4}' | head -n1", domain)
		ip, err := utils.RunShell(cmd)
		if err == nil {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip, nil
			}
		}
	}
	
	return "", fmt.Errorf("failed to resolve DNS for %s", domain)
}

// VerifyDNSAndPrompt checks if domain DNS points to the server's IP
// If not, prompts user to update DNS or continue without HTTPS
func VerifyDNSAndPrompt(domain string) error {
	utils.Section(fmt.Sprintf("DNS Verification for %s", domain))
	
	// Get server's public IP
	serverIP, err := GetPublicIP()
	if err != nil {
		utils.Warn("Could not determine server's public IP: %v", err)
		utils.Warn("Skipping DNS verification")
		return nil
	}
	
	// Get domain's DNS IP
	domainIP, err := GetDomainIP(domain)
	if err != nil {
		utils.Warn("Could not resolve DNS for %s: %v", domain, err)
		utils.Warn("Domain may not be configured yet")
		
		fmt.Println()
		fmt.Println("DNS cannot be resolved for this domain.")
		fmt.Printf("Server IP: %s\n", serverIP)
		fmt.Println()
		fmt.Println("Please ensure your DNS records are configured:")
		fmt.Printf("  A record: %s -> %s\n", domain, serverIP)
		fmt.Println()
		
		return promptUserAction(domain)
	}
	
	// Compare IPs
	if serverIP == domainIP {
		utils.Ok("DNS correctly points to server (IP: %s)", serverIP)
		return nil
	}
	
	// DNS mismatch
	utils.Warn("DNS mismatch detected!")
	fmt.Println()
	fmt.Printf("  Server IP:  %s\n", serverIP)
	fmt.Printf("  Domain IP:  %s\n", domainIP)
	fmt.Println()
	fmt.Println("Let's Encrypt requires the domain to point to this server.")
	fmt.Println("Please update your DNS records:")
	fmt.Printf("  A record: %s -> %s\n", domain, serverIP)
	fmt.Println()
	fmt.Println("DNS propagation can take 5-30 minutes.")
	fmt.Println()
	
	return promptUserAction(domain)
}

// promptUserAction prompts the user for action when DNS verification fails
func promptUserAction(domain string) error {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Println("What would you like to do?")
		fmt.Println("  1) Check DNS again (after updating records)")
		fmt.Println("  2) Continue without HTTPS (HTTP only)")
		fmt.Println("  3) Abort setup")
		fmt.Print("\nChoice [1/2/3]: ")
		
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %v", err)
		}
		
		choice := strings.TrimSpace(input)
		
		switch choice {
		case "1":
			// Check again
			fmt.Println()
			utils.Log("Checking DNS again...")
			
			serverIP, err := GetPublicIP()
			if err != nil {
				utils.Warn("Could not determine server's public IP: %v", err)
				continue
			}
			
			domainIP, err := GetDomainIP(domain)
			if err != nil {
				utils.Warn("Could not resolve DNS for %s: %v", domain, err)
				fmt.Println()
				continue
			}
			
			if serverIP == domainIP {
				utils.Ok("DNS now correctly points to server (IP: %s)", serverIP)
				return nil
			}
			
			utils.Warn("DNS still points to %s (expected %s)", domainIP, serverIP)
			fmt.Println()
			continue
			
		case "2":
			// Continue without HTTPS
			utils.Warn("Continuing without HTTPS - site will be HTTP only")
			return fmt.Errorf("skipping SSL: DNS not configured")
			
		case "3":
			// Abort
			return fmt.Errorf("setup aborted by user")
			
		default:
			fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
			fmt.Println()
		}
	}
}

// ObtainCertificate obtains SSL certificate for a domain using a two-phase approach:
// Phase 1: Obtain certificate without modifying nginx (safer, won't break site if it fails)
// Phase 2: Configure nginx (done by caller after verifying certificate exists)
func ObtainCertificate(domain, email string) error {
	// Check if certificate already exists
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	if utils.CheckFileExists(certPath) {
		utils.Verify("SSL certificate already exists for %s", domain)
		return nil
	}

	if email == "" {
		utils.Skip("No Let's Encrypt email provided, skipping SSL")
		return nil
	}
	
	// Verify DNS before attempting certificate
	if err := VerifyDNSAndPrompt(domain); err != nil {
		return err
	}

	utils.Log("Obtaining SSL certificate for %s", domain)

	// PHASE 1: Obtain certificate ONLY - do not modify nginx
	// This ensures that if certificate obtainment fails (e.g., rate limit),
	// nginx configuration remains unchanged and the site continues to work over HTTP
	cmd := fmt.Sprintf("certbot certonly --nginx -d %s --non-interactive --agree-tos --email %s --no-eff-email", domain, email)
	_, err := utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %v", err)
	}

	// Verify certificate was actually created
	if !utils.CheckFileExists(certPath) {
		return fmt.Errorf("certificate file not found after obtainment: %s", certPath)
	}

	utils.Ok("SSL certificate obtained for %s", domain)
	return nil
}

// ConfigureNginxSSL configures nginx to use an existing SSL certificate
// This is PHASE 2 of SSL setup - only call after ObtainCertificate succeeds
func ConfigureNginxSSL(domain string) error {
	// Verify certificate exists first
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	if !utils.CheckFileExists(certPath) {
		return fmt.Errorf("certificate not found: %s", certPath)
	}

	utils.Log("Configuring nginx with SSL for %s...", domain)

	// Use certbot install to configure nginx with existing certificate
	// The --redirect flag adds HTTP to HTTPS redirect
	cmd := fmt.Sprintf("certbot install --nginx -d %s --cert-name %s --non-interactive --redirect", domain, domain)
	_, err := utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("failed to configure nginx SSL: %v", err)
	}

	utils.Ok("Nginx configured with SSL for %s", domain)
	return nil
}

// SetupAutoRenewal configures automatic certificate renewal
func SetupAutoRenewal(verifyOnly bool) error {
	// Check if systemd timer exists
	if utils.CheckFileExists("/lib/systemd/system/certbot.timer") {
		if err := system.EnsureServiceRunning("certbot.timer", verifyOnly); err != nil {
			return err
		}
		utils.Verify("Certbot auto-renewal configured")
		return nil
	}

	if verifyOnly {
		utils.Fail("Certbot auto-renewal not configured")
		return fmt.Errorf("certbot auto-renewal not configured")
	}

	utils.Log("Setting up auto-renewal...")
	_, _ = utils.RunCommand("systemctl", "enable", "certbot.timer")
	_, _ = utils.RunCommand("systemctl", "start", "certbot.timer")

	utils.Ok("Certbot auto-renewal configured")
	return nil
}

// FixSSLDocroot ensures the SSL server block uses the correct document root
func FixSSLDocroot(domain, webroot string) error {
	vhostPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)
	
	if !utils.CheckFileExists(vhostPath) {
		return fmt.Errorf("nginx vhost not found: %s", vhostPath)
	}

	utils.Log("Fixing SSL docroot for %s...", domain)

	// Read current config
	content, err := utils.RunShell(fmt.Sprintf("cat %s", vhostPath))
	if err != nil {
		return fmt.Errorf("failed to read vhost config: %v", err)
	}

	// Check if SSL is configured (look for 443 listener or ssl directives)
	if !strings.Contains(content, "listen 443") && !strings.Contains(content, "listen [::]:443") {
		utils.Skip("SSL not configured yet (no 443 listener), skipping docroot fix")
		return nil
	}

	// Replace any incorrect root paths in the SSL server block (port 443)
	// Strategy: Find the ssl server block and ensure it has the correct root directive
	lines := strings.Split(content, "\n")
	var result []string
	inSSLBlock := false
	rootFixed := false
	
	for i, line := range lines {
		// Detect start of SSL server block
		if strings.Contains(line, "listen 443") || strings.Contains(line, "listen [::]:443") {
			inSSLBlock = true
		}
		
		// Fix root directive in SSL block
		if inSSLBlock && strings.Contains(line, "root ") && !strings.Contains(line, webroot) {
			// Replace with correct root
			indent := strings.Repeat(" ", len(line)-len(strings.TrimLeft(line, " ")))
			result = append(result, fmt.Sprintf("%sroot %s;", indent, webroot))
			rootFixed = true
			utils.Log("Fixed root directive in SSL block")
		} else {
			result = append(result, line)
		}
		
		// Detect end of server block
		if inSSLBlock && strings.TrimSpace(line) == "}" {
			// Check if this is the closing brace of the server block
			// Count opening and closing braces up to this point
			openBraces := 0
			for j := 0; j <= i; j++ {
				openBraces += strings.Count(lines[j], "{")
				openBraces -= strings.Count(lines[j], "}")
			}
			if openBraces == 0 {
				inSSLBlock = false
			}
		}
	}
	
	if !rootFixed {
		utils.Verify("SSL docroot already correct")
		return nil
	}

	// Write fixed config
	fixedContent := strings.Join(result, "\n")
	_, err = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", vhostPath, fixedContent))
	if err != nil {
		return fmt.Errorf("failed to write fixed config: %v", err)
	}

	utils.Ok("SSL docroot fixed for %s", domain)
	return nil
}

// ReconfigureSSL reconfigures SSL for a domain if certificate exists
// This is useful after updating vhost configuration that removed SSL blocks
func ReconfigureSSL(domain string) error {
	// Check if certificate exists
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	if !utils.CheckFileExists(certPath) {
		utils.Skip("No SSL certificate found for %s", domain)
		return nil
	}

	utils.Log("Reconfiguring SSL for %s...", domain)

	// Use certbot install to reconfigure nginx
	cmd := fmt.Sprintf("certbot install --nginx -d %s --cert-name %s --non-interactive --redirect", domain, domain)
	_, err := utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("failed to reconfigure SSL: %v", err)
	}

	utils.Ok("SSL reconfigured for %s", domain)
	return nil
}

// EnhanceSSLConfig adds advanced security settings to nginx SSL configuration
func EnhanceSSLConfig(domain string) error {
	vhostPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)
	
	if !utils.CheckFileExists(vhostPath) {
		return fmt.Errorf("nginx vhost not found: %s", vhostPath)
	}

	utils.Log("Enhancing SSL configuration for %s...", domain)

	// Read current config
	content, err := utils.RunShell(fmt.Sprintf("cat %s", vhostPath))
	if err != nil {
		return fmt.Errorf("failed to read vhost config: %v", err)
	}

	// Check if already enhanced
	if strings.Contains(content, "Strict-Transport-Security") {
		utils.Verify("SSL configuration already enhanced")
		return nil
	}

	// Check if SSL is configured (look for 443 listener)
	if !strings.Contains(content, "listen 443") && !strings.Contains(content, "listen [::]:443") {
		utils.Warn("SSL not configured yet (no 443 listener), skipping enhancement")
		return nil
	}

	// Add enhanced SSL configuration after the ssl_certificate_key line
	enhancedConfig := `
    # Enhanced SSL Security Settings
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;
    
    # HSTS (HTTP Strict Transport Security)
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
`

	// Find the ssl_certificate_key line and insert enhanced config after it
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, line)
		// After ssl_certificate_key line, add enhanced config
		if strings.Contains(line, "ssl_certificate_key") && strings.Contains(line, ";") {
			result = append(result, enhancedConfig)
		}
	}
	enhancedContent := strings.Join(result, "\n")

	// Write enhanced config
	_, err = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", vhostPath, enhancedContent))
	if err != nil {
		return fmt.Errorf("failed to write enhanced config: %v", err)
	}

	utils.Ok("SSL configuration enhanced for %s", domain)
	return nil
}
