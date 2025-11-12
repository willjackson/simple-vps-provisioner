package ssl

import (
	"fmt"
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

// ObtainCertificate obtains SSL certificate for a domain
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

	utils.Log("Obtaining SSL certificate for %s", domain)

	// Use certbot with nginx plugin
	cmd := fmt.Sprintf("certbot --nginx -d %s --non-interactive --agree-tos --email %s --redirect --no-eff-email", domain, email)
	_, err := utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %v", err)
	}

	utils.Ok("SSL certificate obtained for %s", domain)
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
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers off;
    
    # SSL session configuration
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    
    # OCSP stapling
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
