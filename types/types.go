package types

// Config holds the global configuration for the provisioner
type Config struct {
	// Mode determines which operation to perform
	Mode string

	// CMS type: "drupal" or "wordpress"
	CMS string

	// PHP version to use (e.g., "8.3")
	PHPVersion string

	// Primary domain name
	PrimaryDomain string

	// Extra domains (comma-separated)
	ExtraDomains string

	// Let's Encrypt email
	LEEmail string

	// Webroot parent directory
	Webroot string

	// Git repository URL
	GitRepo string

	// Git branch
	GitBranch string

	// Drupal root path (relative to repo)
	DrupalRoot string

	// Custom docroot path
	Docroot string

	// Database engine: "mariadb" or "none"
	DBEngine string

	// Whether to create swap
	CreateSwap string // "yes", "no", or "auto"

	// Whether to enable firewall
	UFWEnable bool

	// Whether to enable SSL/HTTPS
	SSLEnable bool

	// Switch all sites to new PHP version
	SwitchAll bool

	// Verify only mode
	VerifyOnly bool

	// Debug mode
	Debug bool
}

// SiteConfig represents configuration for a single site
type SiteConfig struct {
	Domain     string
	PHPVersion string
	Webroot    string
	DBName     string
	DBUser     string
	DBPass     string
}

// PHPVersions tracks current and previous PHP versions
type PHPVersions struct {
	Current  string
	Previous string
}
