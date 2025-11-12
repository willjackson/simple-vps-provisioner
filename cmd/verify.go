package cmd

import (
	"fmt"
	"svp/pkg/config"
	"svp/pkg/database"
	"svp/pkg/system"
	"svp/pkg/utils"
	"svp/pkg/web"
	"svp/types"
)

// Verify checks the system configuration without making changes
func Verify(cfg *types.Config) error {
	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println("  Configuration Verification")
	fmt.Println("==========================================================")
	fmt.Println()

	var errors []error

	// Check base packages
	utils.Section("Base Packages")
	if err := system.EnsureBasePackages(true); err != nil {
		errors = append(errors, err)
	}

	// Check Nginx
	utils.Section("Nginx")
	if err := web.InstallNginx(true); err != nil {
		errors = append(errors, err)
	}

	// Check PHP
	utils.Section("PHP")
	versions, err := config.ReadPHPVersions()
	if err != nil {
		utils.Warn("Could not read PHP version config: %v", err)
	}

	phpVersion := cfg.PHPVersion
	if phpVersion == "" && versions.Current != "" {
		phpVersion = versions.Current
	}

	if phpVersion != "" {
		if err := web.InstallPHP(phpVersion, true); err != nil {
			errors = append(errors, err)
		}
	} else {
		utils.Warn("No PHP version configured")
	}

	// Check database
	utils.Section("Database")
	if err := database.InstallMariaDB(cfg.DBEngine, true); err != nil {
		errors = append(errors, err)
	}

	// Check Composer
	utils.Section("Composer")
	if err := web.InstallComposer(true); err != nil {
		errors = append(errors, err)
	}

	// Check admin user
	utils.Section("Admin User")
	if err := config.EnsureAdminUser(true); err != nil {
		errors = append(errors, err)
	}

	// Check firewall
	utils.Section("Firewall")
	if err := system.SetupFirewall(cfg.UFWEnable, true); err != nil {
		errors = append(errors, err)
	}

	// Check SSL/Certbot if enabled
	if cfg.SSLEnable || cfg.LEEmail != "" {
		utils.Section("SSL/Certbot")
		if err := ssl.InstallCertbot(true); err != nil {
			errors = append(errors, err)
		}
		if err := ssl.SetupAutoRenewal(true); err != nil {
			errors = append(errors, err)
		}
	}

	// Print summary
	fmt.Println()
	fmt.Println("==========================================================")
	if len(errors) == 0 {
		utils.Ok("All checks passed!")
	} else {
		utils.Fail("%d check(s) failed", len(errors))
		fmt.Println("\nFailed checks:")
		for _, err := range errors {
			fmt.Printf("  • %v\n", err)
		}
	}
	fmt.Println("==========================================================")
	fmt.Println()

	if len(errors) > 0 {
		return fmt.Errorf("verification failed with %d error(s)", len(errors))
	}

	return nil
}
