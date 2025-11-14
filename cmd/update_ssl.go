package cmd

import (
	"fmt"
	"strings"
	"svp/pkg/ssl"
	"svp/pkg/system"
	"svp/pkg/utils"
	"svp/types"
)

// UpdateSSL handles SSL operations: enable, disable, renew, check
func UpdateSSL(cfg *types.Config) error {
	domain := cfg.PrimaryDomain
	action := cfg.SSLAction
	email := cfg.LEEmail

	utils.Section(fmt.Sprintf("SSL Management for %s", domain))

	// Ensure certbot is installed for all actions
	if err := ssl.InstallCertbot(false); err != nil {
		return err
	}

	switch action {
	case "enable":
		return enableSSL(domain, email)
	case "disable":
		return disableSSL(domain)
	case "renew":
		return renewSSL(domain)
	case "check":
		return checkSSL(domain)
	default:
		return fmt.Errorf("invalid action: %s (must be enable, disable, renew, or check)", action)
	}
}

func enableSSL(domain, email string) error {
	utils.Section("Enabling SSL")

	// Check if certificate already exists
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	if utils.CheckFileExists(certPath) {
		utils.Warn("SSL certificate already exists for %s", domain)
		utils.Log("Reconfiguring nginx with existing certificate...")

		// Reconfigure nginx to use existing certificate
		if err := ssl.ConfigureNginxSSL(domain); err != nil {
			return err
		}

		// Reload nginx
		if err := system.ReloadNginx(); err != nil {
			return err
		}

		utils.Ok("SSL enabled for %s", domain)
		return nil
	}

	// Email is required for obtaining new certificates
	if email == "" {
		fmt.Print("Please enter an email address for Let's Encrypt notifications: ")
		var userEmail string
		fmt.Scanln(&userEmail)
		if userEmail == "" {
			return fmt.Errorf("email address is required to obtain SSL certificate")
		}
		email = userEmail
	}

	// Obtain certificate
	if err := ssl.ObtainCertificate(domain, email); err != nil {
		return err
	}

	// Configure nginx
	if err := ssl.ConfigureNginxSSL(domain); err != nil {
		return err
	}

	// Fix docroot if needed
	webroot := fmt.Sprintf("/var/www/%s/web", domain)
	if err := ssl.FixSSLDocroot(domain, webroot); err != nil {
		utils.Warn("Failed to fix SSL docroot: %v", err)
	}

	// Enhance SSL config
	if err := ssl.EnhanceSSLConfig(domain); err != nil {
		utils.Warn("Failed to enhance SSL config: %v", err)
	}

	// Reload nginx
	if err := system.ReloadNginx(); err != nil {
		return err
	}

	// Setup auto-renewal
	if err := ssl.SetupAutoRenewal(false); err != nil {
		utils.Warn("Failed to setup auto-renewal: %v", err)
	}

	utils.Ok("SSL enabled for %s", domain)
	fmt.Println()
	fmt.Printf("Your site is now available at https://%s\n", domain)

	return nil
}

func disableSSL(domain string) error {
	utils.Section("Disabling SSL")

	// Check if certificate exists
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	if !utils.CheckFileExists(certPath) {
		utils.Warn("No SSL certificate found for %s", domain)
		return nil
	}

	// Remove SSL configuration from nginx
	utils.Log("Removing SSL configuration from nginx...")

	vhostPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)
	if !utils.CheckFileExists(vhostPath) {
		return fmt.Errorf("nginx vhost not found: %s", vhostPath)
	}

	// Read current config
	content, err := utils.RunShell(fmt.Sprintf("cat %s", vhostPath))
	if err != nil {
		return fmt.Errorf("failed to read vhost config: %v", err)
	}

	// Remove SSL server blocks (port 443) and keep only HTTP (port 80)
	lines := strings.Split(content, "\n")
	var result []string
	inSSLBlock := false
	braceCount := 0

	for _, line := range lines {
		// Detect start of SSL server block
		if strings.Contains(line, "listen 443") || strings.Contains(line, "listen [::]:443") {
			inSSLBlock = true
			braceCount = 0
		}

		// Count braces if in SSL block
		if inSSLBlock {
			braceCount += strings.Count(line, "{")
			braceCount -= strings.Count(line, "}")

			// Skip lines in SSL block
			if braceCount == 0 && strings.Contains(line, "}") {
				inSSLBlock = false
			}
			continue
		}

		// Keep non-SSL lines, but remove SSL-related directives from HTTP block
		if !strings.Contains(line, "ssl_certificate") &&
		   !strings.Contains(line, "ssl_") &&
		   !strings.Contains(line, "Strict-Transport-Security") {
			result = append(result, line)
		}
	}

	// Write updated config
	newContent := strings.Join(result, "\n")
	_, err = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", vhostPath, newContent))
	if err != nil {
		return fmt.Errorf("failed to write updated config: %v", err)
	}

	// Reload nginx
	if err := system.ReloadNginx(); err != nil {
		return err
	}

	utils.Ok("SSL disabled for %s", domain)
	utils.Warn("Certificate files remain in /etc/letsencrypt/live/%s", domain)
	utils.Log("To re-enable SSL, run: svp update-ssl %s enable", domain)
	fmt.Println()
	fmt.Printf("Your site is now available at http://%s\n", domain)

	return nil
}

func renewSSL(domain string) error {
	utils.Section("Renewing SSL Certificate")

	// Check if certificate exists
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	if !utils.CheckFileExists(certPath) {
		return fmt.Errorf("no SSL certificate found for %s", domain)
	}

	utils.Log("Renewing SSL certificate for %s...", domain)

	// Force renewal
	cmd := fmt.Sprintf("certbot renew --cert-name %s --force-renewal --nginx", domain)
	_, err := utils.RunShell(cmd)
	if err != nil {
		return fmt.Errorf("failed to renew certificate: %v", err)
	}

	// Reload nginx
	if err := system.ReloadNginx(); err != nil {
		return err
	}

	utils.Ok("SSL certificate renewed for %s", domain)

	// Show certificate info
	return checkSSL(domain)
}

func checkSSL(domain string) error {
	utils.Section("SSL Certificate Status")

	// Check if certificate exists
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain)
	if !utils.CheckFileExists(certPath) {
		utils.Warn("No SSL certificate found for %s", domain)
		fmt.Println()
		fmt.Printf("To enable SSL, run: svp update-ssl %s enable --le-email your@email.com\n", domain)
		return nil
	}

	utils.Ok("SSL certificate found for %s", domain)
	fmt.Println()

	// Show certificate details
	utils.Log("Certificate details:")

	// Get certificate expiry date
	cmd := fmt.Sprintf("openssl x509 -in %s -noout -enddate", certPath)
	expiry, err := utils.RunShell(cmd)
	if err == nil {
		fmt.Printf("  Expiry: %s\n", strings.TrimSpace(expiry))
	}

	// Get certificate subject
	cmd = fmt.Sprintf("openssl x509 -in %s -noout -subject", certPath)
	subject, err := utils.RunShell(cmd)
	if err == nil {
		fmt.Printf("  Subject: %s\n", strings.TrimSpace(subject))
	}

	// Get certificate issuer
	cmd = fmt.Sprintf("openssl x509 -in %s -noout -issuer", certPath)
	issuer, err := utils.RunShell(cmd)
	if err == nil {
		fmt.Printf("  Issuer: %s\n", strings.TrimSpace(issuer))
	}

	fmt.Println()

	// Check certbot status for this certificate
	utils.Log("Certbot certificate information:")
	cmd = fmt.Sprintf("certbot certificates --cert-name %s", domain)
	certInfo, err := utils.RunShell(cmd)
	if err == nil {
		fmt.Println(certInfo)
	}

	return nil
}
