package web

import (
	"fmt"
	"svp/pkg/system"
	"svp/pkg/utils"
)

// InstallNginx installs and configures Nginx
func InstallNginx(verifyOnly bool) error {
	if utils.CheckPackageInstalled("nginx") {
		utils.Verify("Nginx already installed")

		// Ensure service is enabled and running
		if err := system.EnsureServiceRunning("nginx", verifyOnly); err != nil {
			return err
		}
	} else {
		if verifyOnly {
			utils.Fail("Nginx not installed")
			return fmt.Errorf("nginx not installed")
		}

		utils.Log("Installing Nginx...")
		_, err := utils.RunCommand("apt-get", "install", "-y", "--no-install-recommends", "nginx")
		if err != nil {
			return fmt.Errorf("failed to install Nginx: %v", err)
		}

		if err := system.EnableService("nginx"); err != nil {
			return err
		}
		if err := system.StartService("nginx"); err != nil {
			return err
		}

		utils.Ok("Nginx installed and running")
	}

	return nil
}

// ReloadNginx reloads the Nginx configuration
func ReloadNginx() error {
	utils.Log("Testing Nginx configuration...")

	// Test config first
	output, err := utils.RunCommand("nginx", "-t")
	if err != nil {
		utils.Err("Nginx configuration test failed:")
		utils.Err(output)
		return fmt.Errorf("nginx config test failed: %v", err)
	}

	utils.Ok("Nginx configuration is valid")

	// Reload
	utils.Log("Reloading Nginx...")
	if err := system.ReloadService("nginx"); err != nil {
		return err
	}

	utils.Ok("Nginx reloaded successfully")
	return nil
}

// EnsureSnippets creates common Nginx snippet files
func EnsureSnippets(phpVersion string) error {
	snippetsDir := "/etc/nginx/snippets"

	if err := utils.EnsureDir(snippetsDir); err != nil {
		return fmt.Errorf("failed to create snippets directory: %v", err)
	}

	// PHP-FPM snippet
	phpSnippet := fmt.Sprintf(`# PHP-FPM configuration for version %s
location ~ \.php$ {
    include snippets/fastcgi-php.conf;
    fastcgi_pass unix:/run/php/php%s-fpm-$pool.sock;
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    include fastcgi_params;
}

# Deny access to .htaccess files
location ~ /\.ht {
    deny all;
}
`, phpVersion, phpVersion)

	phpSnippetPath := fmt.Sprintf("%s/php%s-fpm.conf", snippetsDir, phpVersion)
	if !utils.CheckFileExists(phpSnippetPath) {
		utils.Log("Creating PHP-FPM snippet: %s", phpSnippetPath)
		if _, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", phpSnippetPath, phpSnippet)); err != nil {
			return fmt.Errorf("failed to create PHP snippet: %v", err)
		}
	} else {
		utils.Verify("PHP-FPM snippet already exists")
	}

	// Security headers snippet
	securitySnippet := `# Security headers
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
`

	securitySnippetPath := fmt.Sprintf("%s/security-headers.conf", snippetsDir)
	if !utils.CheckFileExists(securitySnippetPath) {
		utils.Log("Creating security headers snippet")
		if _, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", securitySnippetPath, securitySnippet)); err != nil {
			return fmt.Errorf("failed to create security headers snippet: %v", err)
		}
	} else {
		utils.Verify("Security headers snippet already exists")
	}

	return nil
}

// CreateNginxVhost creates an Nginx virtual host configuration
func CreateNginxVhost(domain, webroot, phpVersion string) error {
	vhostPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", domain)
	vhostLink := fmt.Sprintf("/etc/nginx/sites-enabled/%s.conf", domain)

	// Sanitize pool name for PHP-FPM
	poolName := domain

	vhostConfig := fmt.Sprintf(`# Nginx configuration for %s
server {
    listen 80;
    listen [::]:80;
    server_name %s;

    root %s;
    index index.php index.html index.htm;

    # Set pool variable for PHP-FPM
    set $pool "%s";

    # Logging
    access_log /var/log/nginx/%s-access.log;
    error_log /var/log/nginx/%s-error.log;

    # Security headers
    include snippets/security-headers.conf;

    # Main location block
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    # PHP processing
    include snippets/php%s-fpm.conf;

    # Deny access to hidden files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
}
`, domain, domain, webroot, poolName, domain, domain, phpVersion)

	if utils.CheckFileExists(vhostPath) {
		utils.Log("Updating Nginx vhost for %s", domain)
	} else {
		utils.Log("Creating Nginx vhost for %s", domain)
	}
	
	_, err := utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", vhostPath, vhostConfig))
	if err != nil {
		return fmt.Errorf("failed to create vhost config: %v", err)
	}

	// Enable site
	if !utils.CheckFileExists(vhostLink) {
		utils.Log("Enabling site %s", domain)
		_, err := utils.RunCommand("ln", "-sf", vhostPath, vhostLink)
		if err != nil {
			return fmt.Errorf("failed to enable site: %v", err)
		}
	}

	return nil
}
