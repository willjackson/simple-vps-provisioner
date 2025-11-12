package utils

import (
	"os"
	"strings"
)

// CheckPackageInstalled checks if a Debian package is installed
func CheckPackageInstalled(pkg string) bool {
	output, err := RunCommand("dpkg", "-l", pkg)
	if err != nil {
		return false
	}
	return strings.Contains(output, "ii  "+pkg)
}

// CheckServiceEnabled checks if a systemd service is enabled
func CheckServiceEnabled(service string) bool {
	_, err := RunCommand("systemctl", "is-enabled", service)
	return err == nil
}

// CheckServiceActive checks if a systemd service is active
func CheckServiceActive(service string) bool {
	_, err := RunCommand("systemctl", "is-active", service)
	return err == nil
}

// CheckFileExists checks if a file exists
func CheckFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// CheckDirExists checks if a directory exists
func CheckDirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// RequireRoot ensures the program is running as root
func RequireRoot() {
	if os.Geteuid() != 0 {
		Err("This program must be run as root")
		os.Exit(1)
	}
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	if !CheckDirExists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}
