package cmd

import (
	"svp/pkg/updater"
)

// Update checks for and installs updates
func Update(currentVersion string) error {
	return updater.Update(currentVersion)
}
