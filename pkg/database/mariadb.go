package database

import (
	"crypto/rand"
	"fmt"
	"svp/pkg/system"
	"svp/pkg/utils"
	"math/big"
	"strings"
)

// InstallMariaDB installs and configures MariaDB
func InstallMariaDB(dbEngine string, verifyOnly bool) error {
	if dbEngine != "mariadb" {
		utils.Skip("Database installation disabled by config")
		return nil
	}

	if utils.CheckPackageInstalled("mariadb-server") {
		utils.Verify("MariaDB already installed")

		// Ensure service is running
		if err := system.EnsureServiceRunning("mariadb", verifyOnly); err != nil {
			return err
		}
	} else {
		if verifyOnly {
			utils.Fail("MariaDB not installed")
			return fmt.Errorf("mariadb not installed")
		}

		utils.Log("Installing MariaDB...")
		_, err := utils.RunCommand("apt-get", "install", "-y", "mariadb-server", "mariadb-client")
		if err != nil {
			return fmt.Errorf("failed to install MariaDB: %v", err)
		}

		if err := system.EnableService("mariadb"); err != nil {
			return err
		}
		if err := system.StartService("mariadb"); err != nil {
			return err
		}

		// Harden MariaDB
		utils.Log("Hardening MariaDB...")
		secureSQL := `DELETE FROM mysql.user WHERE User='';
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test_%';
FLUSH PRIVILEGES;`

		_, err = utils.RunShell(fmt.Sprintf("mariadb -e \"%s\"", secureSQL))
		if err != nil {
			utils.Warn("MariaDB hardening partially failed: %v", err)
		}

		utils.Ok("MariaDB installed and hardened")
	}

	return nil
}

// GeneratePassword generates a random password
func GeneratePassword(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	password := make([]byte, length)

	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[n.Int64()]
	}

	return string(password), nil
}

// CreateDatabase creates a database and user for a domain
func CreateDatabase(domain string, sitesDir string) (dbName, dbUser, dbPass string, err error) {
	// Sanitize database name (remove dots and dashes, keep only alphanumeric and underscore)
	dbName = "drupal_" + strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, domain)

	dbUser = dbName
	dbPass, err = GeneratePassword(24)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate password: %v", err)
	}

	utils.Log("Creating database and user for %s...", domain)

	// Create database
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS \\`%s\\` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;", dbName)
	_, err = utils.RunShell(fmt.Sprintf("mariadb -e '%s'", createDBSQL))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create database: %v", err)
	}
	utils.Ok("Database created: %s", dbName)

	// Create user and grant privileges
	createUserSQL := fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';", dbUser, dbPass)
	_, err = utils.RunShell(fmt.Sprintf("mariadb -e '%s'", createUserSQL))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create database user: %v", err)
	}

	grantSQL := fmt.Sprintf("GRANT ALL PRIVILEGES ON \\`%s\\`.* TO '%s'@'localhost'; FLUSH PRIVILEGES;", dbName, dbUser)
	_, err = utils.RunShell(fmt.Sprintf("mariadb -e '%s'", grantSQL))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to grant privileges: %v", err)
	}
	utils.Ok("Database user created: %s", dbUser)

	// Save credentials to secure file
	credsFile := fmt.Sprintf("%s/%s.db.txt", sitesDir, domain)
	credsContent := fmt.Sprintf(`Database: %s
Username: %s
Password: %s
Host: localhost
Port: 3306
`, dbName, dbUser, dbPass)

	_, err = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", credsFile, credsContent))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to save credentials: %v", err)
	}

	_, _ = utils.RunCommand("chmod", "600", credsFile)
	_, _ = utils.RunCommand("chown", "admin:www-data", credsFile)

	return dbName, dbUser, dbPass, nil
}
