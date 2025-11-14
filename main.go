package main

import (
	"flag"
	"fmt"
	"os"
	"svp/cmd"
	"svp/pkg/utils"
	"svp/types"
)

// version is set at build time via -ldflags="-X main.version=VERSION"
// Default to dev if not set during build
var version = "dev"

// Documentation URL
const documentationURL = "https://github.com/willjackson/simple-vps-provisioner#readme"

func main() {
	// Ensure running as root
	utils.RequireRoot()

	// Get command from first argument
	if len(os.Args) < 2 {
		showGeneralHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	// Handle global flags
	if command == "-version" || command == "--version" {
		fmt.Printf("Simple VPS Provisioner (svp) version %s\n", version)
		if version == "dev" {
			fmt.Println("This is a development build.")
			fmt.Println("Production builds should use: go build -ldflags=\"-X main.version=VERSION\"")
		}
		os.Exit(0)
	}

	if command == "-help" || command == "--help" || command == "help" {
		showGeneralHelp()
		os.Exit(0)
	}

	// Execute command
	switch command {
	case "setup":
		setupCommand()
	case "verify":
		verifyCommand()
	case "update":
		updateCommand()
	case "php-update":
		phpUpdateCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		showGeneralHelp()
		os.Exit(1)
	}
}

func showGeneralHelp() {
	fmt.Printf("Simple VPS Provisioner (svp) v%s\n\n", version)
	fmt.Println("A Go CLI tool for provisioning Debian VPS with LAMP stack for Drupal or WordPress.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  svp <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  setup        Provision VPS with LAMP stack and CMS")
	fmt.Println("  verify       Verify system configuration without making changes")
	fmt.Println("  update       Check for and install updates to svp")
	fmt.Println("  php-update   Update PHP version for a specific domain")
	fmt.Println()
	fmt.Println("Global Flags:")
	fmt.Println("  --version     Show version information")
	fmt.Println("  --debug       Enable debug mode")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  svp setup --cms drupal --domain example.com --le-email admin@example.com")
	fmt.Println("  svp verify")
	fmt.Println("  svp update")
	fmt.Println()
	fmt.Println("Get help for a specific command:")
	fmt.Println("  svp setup --help")
	fmt.Println("  svp php-update --help")
	fmt.Println()
	fmt.Printf("Documentation: %s\n", documentationURL)
}

func setupCommand() {
	cfg := &types.Config{Mode: "setup"}
	fs := flag.NewFlagSet("setup", flag.ExitOnError)

	// Setup-specific flags
	fs.StringVar(&cfg.CMS, "cms", "drupal", "CMS to install: drupal or wordpress")
	fs.StringVar(&cfg.PHPVersion, "php-version", "8.4", "PHP version to install")
	fs.StringVar(&cfg.PrimaryDomain, "domain", "", "Primary domain name (required)")
	fs.StringVar(&cfg.ExtraDomains, "extra-domains", "", "Extra domains (comma-separated)")
	fs.StringVar(&cfg.LEEmail, "le-email", "", "Let's Encrypt email address for SSL")
	fs.StringVar(&cfg.Webroot, "webroot", "/var/www", "Parent directory for sites")
	fs.StringVar(&cfg.GitRepo, "git-repo", "", "Git repository URL")
	fs.StringVar(&cfg.GitBranch, "git-branch", "", "Git branch (uses repository default if not specified)")
	fs.StringVar(&cfg.DrupalRoot, "drupal-root", "", "Drupal root path (relative to repo)")
	fs.StringVar(&cfg.Docroot, "docroot", "", "Custom docroot path")
	fs.StringVar(&cfg.DBEngine, "db-engine", "mariadb", "Database engine: mariadb or none")
	fs.StringVar(&cfg.DBImport, "db", "", "Path to database file for import")
	fs.StringVar(&cfg.CreateSwap, "create-swap", "auto", "Create swap: yes, no, or auto")
	fs.BoolVar(&cfg.UFWEnable, "firewall", true, "Enable UFW firewall")
	fs.BoolVar(&cfg.SSLEnable, "ssl", true, "Enable SSL/HTTPS with Let's Encrypt")
	fs.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")
	fs.BoolVar(&cfg.KeepExistingDB, "keep-existing-db", false, "Keep existing database and drop tables")

	fs.Usage = func() {
		fmt.Printf("Simple VPS Provisioner (svp) v%s - Setup Command\n\n", version)
		fmt.Println("Usage:")
		fmt.Println("  svp setup [options]")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Provision a fresh VPS with complete LAMP stack and CMS installation.")
		fmt.Println()
		fmt.Println("Required Flags:")
		fmt.Println("  --domain string")
		fmt.Println("        Primary domain name")
		fmt.Println()
		fmt.Println("Common Flags:")
		fmt.Println("  --cms string")
		fmt.Println("        CMS to install: drupal or wordpress (default \"drupal\")")
		fmt.Println("  --le-email string")
		fmt.Println("        Let's Encrypt email for SSL certificates")
		fmt.Println("  --php-version string")
		fmt.Println("        PHP version to install (default \"8.4\")")
		fmt.Println("  --git-repo string")
		fmt.Println("        Git repository URL to clone and deploy")
		fmt.Println("  --git-branch string")
		fmt.Println("        Git branch to checkout (uses repository default if not specified)")
		fmt.Println()
		fmt.Println("Additional Flags:")
		fmt.Println("  --extra-domains string")
		fmt.Println("        Additional domains, comma-separated")
		fmt.Println("  --db string")
		fmt.Println("        Path to database backup file for import")
		fmt.Println("  --keep-existing-db")
		fmt.Println("        Keep existing database and reuse credentials (default: false)")
		fmt.Println("  --webroot string")
		fmt.Println("        Parent directory for sites (default \"/var/www\")")
		fmt.Println("  --drupal-root string")
		fmt.Println("        Drupal root path relative to repository")
		fmt.Println("  --docroot string")
		fmt.Println("        Custom document root path")
		fmt.Println("  --db-engine string")
		fmt.Println("        Database engine: mariadb or none (default \"mariadb\")")
		fmt.Println("  --create-swap string")
		fmt.Println("        Create swap: yes, no, or auto (default \"auto\")")
		fmt.Println("  --firewall")
		fmt.Println("        Enable UFW firewall (default true)")
		fmt.Println("  --ssl")
		fmt.Println("        Enable SSL/HTTPS (default true)")
		fmt.Println("  --debug")
		fmt.Println("        Enable debug mode")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Fresh Drupal site with SSL:")
		fmt.Println("  svp setup --cms drupal --domain example.com --le-email admin@example.com")
		fmt.Println()
		fmt.Println("  # WordPress without SSL:")
		fmt.Println("  svp setup --cms wordpress --domain example.com --ssl=false")
		fmt.Println()
		fmt.Println("  # Deploy from Git with specific PHP version:")
		fmt.Println("  svp setup --cms drupal --domain example.com \\")
		fmt.Println("    --git-repo https://github.com/org/repo.git \\")
		fmt.Println("    --php-version 8.4 --le-email admin@example.com")
		fmt.Println()
		fmt.Println("  # Import existing database:")
		fmt.Println("  svp setup --cms drupal --domain example.com \\")
		fmt.Println("    --db /path/to/backup.sql.gz --le-email admin@example.com")
		fmt.Println()
		fmt.Println("  # Multiple domains:")
		fmt.Println("  svp setup --cms drupal --domain example.com \\")
		fmt.Println("    --extra-domains \"staging.example.com,dev.example.com\" \\")
		fmt.Println("    --le-email admin@example.com")
		fmt.Println()
		fmt.Printf("Documentation: %s\n", documentationURL)
	}

	fs.Parse(os.Args[2:])

	// Enable debug mode if requested
	if cfg.Debug {
		os.Setenv("DEBUG", "1")
		fmt.Println("DEBUG MODE ENABLED")
	}

	// Validate required parameters
	if cfg.PrimaryDomain == "" {
		utils.Err("Primary domain is required for setup")
		fmt.Println("\nUsage: svp setup --domain example.com [options]")
		fmt.Println("Run 'svp setup --help' for more information")
		os.Exit(1)
	}

	// Validate CMS type
	if cfg.CMS != "drupal" && cfg.CMS != "wordpress" {
		utils.Err("Invalid CMS type: %s (must be 'drupal' or 'wordpress')", cfg.CMS)
		os.Exit(1)
	}

	// Warn if SSL is enabled but no email provided
	if cfg.SSLEnable && cfg.LEEmail == "" {
		utils.Warn("SSL is enabled but no Let's Encrypt email provided. SSL will be skipped.")
		utils.Warn("Use --le-email to enable SSL/HTTPS.")
		cfg.SSLEnable = false
	}

	// Execute setup
	if err := cmd.FullSetup(cfg); err != nil {
		utils.Err("Setup failed: %v", err)
		os.Exit(1)
	}
}

func verifyCommand() {
	cfg := &types.Config{
		Mode:       "verify",
		VerifyOnly: true,
	}
	fs := flag.NewFlagSet("verify", flag.ExitOnError)

	fs.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")

	fs.Usage = func() {
		fmt.Printf("Simple VPS Provisioner (svp) v%s - Verify Command\n\n", version)
		fmt.Println("Usage:")
		fmt.Println("  svp verify [options]")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Check system configuration without making any changes.")
		fmt.Println("  Verifies that all components are properly installed and running.")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --debug")
		fmt.Println("        Enable debug mode")
		fmt.Println()
		fmt.Println("What it checks:")
		fmt.Println("  • Base system packages")
		fmt.Println("  • Nginx installation and status")
		fmt.Println("  • PHP-FPM installation and status")
		fmt.Println("  • MariaDB installation and status")
		fmt.Println("  • Composer installation")
		fmt.Println("  • Firewall configuration")
		fmt.Println("  • SSL certificates (if configured)")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  svp verify")
		fmt.Println()
		fmt.Printf("Documentation: %s\n", documentationURL)
	}

	fs.Parse(os.Args[2:])

	// Enable debug mode if requested
	if cfg.Debug {
		os.Setenv("DEBUG", "1")
		fmt.Println("DEBUG MODE ENABLED")
	}

	// Execute verify
	if err := cmd.Verify(cfg); err != nil {
		utils.Err("Verification failed: %v", err)
		os.Exit(1)
	}
}

func updateCommand() {
	fs := flag.NewFlagSet("update", flag.ExitOnError)

	var debug bool
	fs.BoolVar(&debug, "debug", false, "Enable debug mode")

	fs.Usage = func() {
		fmt.Printf("Simple VPS Provisioner (svp) v%s - Update Command\n\n", version)
		fmt.Println("Usage:")
		fmt.Println("  svp update [options]")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Check for and install the latest version of svp from GitHub releases.")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --debug")
		fmt.Println("        Enable debug mode")
		fmt.Println()
		fmt.Println("What it does:")
		fmt.Println("  • Checks GitHub for the latest release")
		fmt.Println("  • Downloads the new binary")
		fmt.Println("  • Verifies checksums")
		fmt.Println("  • Backs up current version")
		fmt.Println("  • Installs the new version")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  svp update")
		fmt.Println()
		fmt.Printf("Documentation: %s\n", documentationURL)
	}

	fs.Parse(os.Args[2:])

	// Enable debug mode if requested
	if debug {
		os.Setenv("DEBUG", "1")
		fmt.Println("DEBUG MODE ENABLED")
	}

	// Execute update
	if err := cmd.Update(version); err != nil {
		utils.Err("Update failed: %v", err)
		os.Exit(1)
	}
}

func phpUpdateCommand() {
	cfg := &types.Config{Mode: "php-update"}
	fs := flag.NewFlagSet("php-update", flag.ExitOnError)

	fs.StringVar(&cfg.PrimaryDomain, "domain", "", "Domain to update (required)")
	fs.StringVar(&cfg.PHPVersion, "php-version", "", "New PHP version (required)")
	fs.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")

	fs.Usage = func() {
		fmt.Printf("Simple VPS Provisioner (svp) v%s - PHP Update Command\n\n", version)
		fmt.Println("Usage:")
		fmt.Println("  svp php-update --domain DOMAIN --php-version VERSION")
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println("  Update the PHP version for a specific domain.")
		fmt.Println()
		fmt.Println("Required Flags:")
		fmt.Println("  --domain string")
		fmt.Println("        Domain to update")
		fmt.Println("  --php-version string")
		fmt.Println("        New PHP version (8.1, 8.2, 8.3, or 8.4)")
		fmt.Println()
		fmt.Println("Optional Flags:")
		fmt.Println("  --debug")
		fmt.Println("        Enable debug mode")
		fmt.Println()
		fmt.Println("What it does:")
		fmt.Println("  • Installs the new PHP version if needed")
		fmt.Println("  • Creates new PHP-FPM pool for the domain")
		fmt.Println("  • Updates Nginx configuration")
		fmt.Println("  • Reconfigures SSL certificates")
		fmt.Println("  • Updates site configuration")
		fmt.Println("  • Removes old PHP-FPM pool")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  # Update to PHP 8.4:")
		fmt.Println("  svp php-update --domain example.com --php-version 8.4")
		fmt.Println()
		fmt.Println("  # Update to PHP 8.3:")
		fmt.Println("  svp php-update --domain mysite.com --php-version 8.3")
		fmt.Println()
		fmt.Printf("Documentation: %s\n", documentationURL)
	}

	fs.Parse(os.Args[2:])

	// Enable debug mode if requested
	if cfg.Debug {
		os.Setenv("DEBUG", "1")
		fmt.Println("DEBUG MODE ENABLED")
	}

	// Validate required parameters
	if cfg.PrimaryDomain == "" {
		utils.Err("Domain is required for php-update")
		fmt.Println("\nUsage: svp php-update --domain example.com --php-version 8.4")
		fmt.Println("Run 'svp php-update --help' for more information")
		os.Exit(1)
	}

	if cfg.PHPVersion == "" {
		utils.Err("PHP version is required for php-update")
		fmt.Println("\nUsage: svp php-update --domain example.com --php-version 8.4")
		fmt.Println("Run 'svp php-update --help' for more information")
		os.Exit(1)
	}

	// Execute PHP update
	if err := cmd.PHPUpdate(cfg); err != nil {
		utils.Err("PHP update failed: %v", err)
		os.Exit(1)
	}
}
