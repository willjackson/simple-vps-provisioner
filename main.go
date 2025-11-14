package main

import (
	"flag"
	"fmt"
	"svp/pkg/utils"
	"os"
	"svp/cmd"
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

	// Define configuration
	cfg := &types.Config{}

	// Define flags (no mode flag - it's now a positional argument)
	flag.StringVar(&cfg.CMS, "cms", "drupal", "CMS to install: drupal or wordpress")
	flag.StringVar(&cfg.PHPVersion, "php-version", "8.4", "PHP version to install")
	flag.StringVar(&cfg.PrimaryDomain, "domain", "", "Primary domain name (required for setup)")
	flag.StringVar(&cfg.ExtraDomains, "extra-domains", "", "Extra domains (comma-separated)")
	flag.StringVar(&cfg.LEEmail, "le-email", "", "Let's Encrypt email address")
	flag.StringVar(&cfg.Webroot, "webroot", "/var/www", "Parent directory for sites")
	flag.StringVar(&cfg.GitRepo, "git-repo", "", "Git repository URL")
	flag.StringVar(&cfg.GitBranch, "git-branch", "", "Git branch to checkout (uses repository default if not specified)")
	flag.StringVar(&cfg.DrupalRoot, "drupal-root", "", "Drupal root path (relative to repo)")
	flag.StringVar(&cfg.Docroot, "docroot", "", "Custom docroot path")
	flag.StringVar(&cfg.DBEngine, "db-engine", "mariadb", "Database engine: mariadb or none")
	flag.StringVar(&cfg.DBImport, "db", "", "Path to database file for import (instead of site-install)")
	flag.StringVar(&cfg.CreateSwap, "create-swap", "auto", "Create swap: yes, no, or auto")
	flag.BoolVar(&cfg.UFWEnable, "firewall", true, "Enable UFW firewall")
	flag.BoolVar(&cfg.SSLEnable, "ssl", true, "Enable SSL/HTTPS with Let's Encrypt (requires -le-email)")
	flag.BoolVar(&cfg.SwitchAll, "switch-all", false, "Switch all sites to new PHP version")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&cfg.KeepExistingDB, "keep-existing-db", false, "Keep existing database and drop tables (default: recreate database)")

	// Version flag
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Simple VPS Provisioner (svp) v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "A Go CLI tool for provisioning Debian VPS with LAMP stack for Drupal or WordPress.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  svp <command> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  setup        Provision VPS with LAMP stack and CMS\n")
		fmt.Fprintf(os.Stderr, "  verify       Verify system configuration without making changes\n")
		fmt.Fprintf(os.Stderr, "  update       Check for and install updates to svp\n")
		fmt.Fprintf(os.Stderr, "  php-update   Update PHP version for a specific domain\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Install Drupal for a domain:\n")
		fmt.Fprintf(os.Stderr, "  svp setup -cms drupal -domain example.com\n\n")
		fmt.Fprintf(os.Stderr, "  # Install WordPress with SSL:\n")
		fmt.Fprintf(os.Stderr, "  svp setup -cms wordpress -domain example.com -le-email admin@example.com\n\n")
		fmt.Fprintf(os.Stderr, "  # Verify system configuration:\n")
		fmt.Fprintf(os.Stderr, "  svp verify\n\n")
		fmt.Fprintf(os.Stderr, "For more examples and documentation, visit:\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", documentationURL)
	}

	// Parse flags
	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("Simple VPS Provisioner (svp) version %s\n", version)
		if version == "dev" {
			fmt.Println("This is a development build.")
			fmt.Println("Production builds should use: go build -ldflags=\"-X main.version=VERSION\"")
		}
		os.Exit(0)
	}

	// Get command from first positional argument
	args := flag.Args()
	if len(args) == 0 {
		// No command provided - show help
		flag.Usage()
		os.Exit(0)
	}

	cfg.Mode = args[0]

	// Enable debug mode if requested
	if cfg.Debug {
		os.Setenv("DEBUG", "1")
		fmt.Println("DEBUG MODE ENABLED")
	}

	// Validate CMS type
	if cfg.CMS != "drupal" && cfg.CMS != "wordpress" {
		utils.Err("Invalid CMS type: %s (must be 'drupal' or 'wordpress')", cfg.CMS)
		os.Exit(1)
	}

	// Set verify only flag
	cfg.VerifyOnly = (cfg.Mode == "verify")

	// Execute command based on mode
	var err error
	switch cfg.Mode {
	case "setup":
		// Validate required parameters
		if cfg.PrimaryDomain == "" {
			utils.Err("Primary domain is required for setup mode")
			utils.Err("Use: svp setup -cms %s -domain your-domain.com", cfg.CMS)
			os.Exit(1)
		}

		// Warn if SSL is enabled but no email provided
		if cfg.SSLEnable && cfg.LEEmail == "" {
			utils.Warn("SSL is enabled but no Let's Encrypt email provided. SSL will be skipped.")
			utils.Warn("Use -le-email to enable SSL/HTTPS.")
			cfg.SSLEnable = false
		}

		err = cmd.FullSetup(cfg)

	case "verify":
		err = cmd.Verify(cfg)

	case "update":
		err = cmd.Update(version)

	case "php-update":
		// Validate required parameters
		if cfg.PrimaryDomain == "" {
			utils.Err("Domain is required for php-update mode")
			utils.Err("Use: svp php-update -domain example.com -php-version 8.4")
			os.Exit(1)
		}
		if cfg.PHPVersion == "" {
			utils.Err("PHP version is required for php-update mode")
			utils.Err("Use: svp php-update -domain example.com -php-version 8.4")
			os.Exit(1)
		}
		err = cmd.PHPUpdate(cfg)

	default:
		utils.Err("Unknown command: %s", cfg.Mode)
		utils.Err("Available commands: setup, verify, update, php-update")
		utils.Err("Run 'svp' for help")
		os.Exit(1)
	}

	// Handle errors
	if err != nil {
		utils.Err("Operation failed: %v", err)
		os.Exit(1)
	}

	// Success
	os.Exit(0)
}
