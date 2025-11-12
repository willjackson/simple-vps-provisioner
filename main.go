package main

import (
	"flag"
	"fmt"
	"svp/pkg/utils"
	"os"
	"svp/cmd"
	"svp/types"
)

const version = "1.0.0"

func main() {
	// Ensure running as root
	utils.RequireRoot()

	// Define configuration
	cfg := &types.Config{}

	// Define flags
	flag.StringVar(&cfg.Mode, "mode", "setup", "Operation mode: setup, verify")
	flag.StringVar(&cfg.CMS, "cms", "drupal", "CMS to install: drupal or wordpress")
	flag.StringVar(&cfg.PHPVersion, "php-version", "8.3", "PHP version to install")
	flag.StringVar(&cfg.PrimaryDomain, "domain", "", "Primary domain name (required for setup)")
	flag.StringVar(&cfg.ExtraDomains, "extra-domains", "", "Extra domains (comma-separated)")
	flag.StringVar(&cfg.LEEmail, "le-email", "", "Let's Encrypt email address")
	flag.StringVar(&cfg.Webroot, "webroot", "/var/www", "Parent directory for sites")
	flag.StringVar(&cfg.GitRepo, "git-repo", "", "Git repository URL")
	flag.StringVar(&cfg.GitBranch, "git-branch", "main", "Git branch to checkout")
	flag.StringVar(&cfg.DrupalRoot, "drupal-root", "", "Drupal root path (relative to repo)")
	flag.StringVar(&cfg.Docroot, "docroot", "", "Custom docroot path")
	flag.StringVar(&cfg.DBEngine, "db-engine", "mariadb", "Database engine: mariadb or none")
	flag.StringVar(&cfg.CreateSwap, "create-swap", "auto", "Create swap: yes, no, or auto")
	flag.BoolVar(&cfg.UFWEnable, "firewall", true, "Enable UFW firewall")
	flag.BoolVar(&cfg.SwitchAll, "switch-all", false, "Switch all sites to new PHP version")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug mode")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Simple VPS Provisioner (svp) v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "A Go CLI tool for provisioning Debian VPS with LAMP stack for Drupal or WordPress.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  svp [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Install Drupal for a domain:\n")
		fmt.Fprintf(os.Stderr, "  svp -mode setup -cms drupal -domain example.com\n\n")
		fmt.Fprintf(os.Stderr, "  # Install WordPress for a domain:\n")
		fmt.Fprintf(os.Stderr, "  svp -mode setup -cms wordpress -domain example.com\n\n")
		fmt.Fprintf(os.Stderr, "  # Install with multiple domains:\n")
		fmt.Fprintf(os.Stderr, "  svp -mode setup -cms drupal -domain example.com -extra-domains 'staging.example.com,dev.example.com'\n\n")
		fmt.Fprintf(os.Stderr, "  # Deploy from Git repository:\n")
		fmt.Fprintf(os.Stderr, "  svp -mode setup -cms drupal -domain example.com -git-repo https://github.com/org/repo.git -git-branch main\n\n")
		fmt.Fprintf(os.Stderr, "  # Verify system configuration:\n")
		fmt.Fprintf(os.Stderr, "  svp -mode verify\n\n")
	}

	// Parse flags
	flag.Parse()

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
			utils.Err("Use: svp -mode setup -cms %s -domain your-domain.com", cfg.CMS)
			os.Exit(1)
		}

		err = cmd.FullSetup(cfg)

	case "verify":
		err = cmd.Verify(cfg)

	default:
		utils.Err("Unknown mode: %s", cfg.Mode)
		utils.Err("Available modes: setup, verify")
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
