package cms

import (
	"fmt"
	"path/filepath"
	"strings"
	"svp/pkg/utils"
)

// UpdateDrushURLToHTTPS updates drush.yml to use HTTPS
func UpdateDrushURLToHTTPS(domain, composerDir string) error {
	drushYmlPath := filepath.Join(composerDir, "drush", "drush.yml")
	
	if !utils.CheckFileExists(drushYmlPath) {
		utils.Warn("drush.yml not found, skipping HTTPS update")
		return nil
	}
	
	// Read current content
	content, err := utils.RunShell(fmt.Sprintf("cat %s", drushYmlPath))
	if err != nil {
		return fmt.Errorf("failed to read drush.yml: %v", err)
	}
	
	// Check if already using https
	if strings.Contains(content, "https://") {
		utils.Verify("drush.yml already using HTTPS")
		return nil
	}
	
	// Replace http with https
	newContent := strings.ReplaceAll(content, fmt.Sprintf("http://%s", domain), fmt.Sprintf("https://%s", domain))
	
	utils.Log("Updating drush.yml to use HTTPS...")
	err = utils.RunShell(fmt.Sprintf("cat > %s <<'EOF'\n%s\nEOF", drushYmlPath, newContent))
	if err != nil {
		return fmt.Errorf("failed to update drush.yml: %v", err)
	}
	
	utils.Ok("Updated drush.yml to use HTTPS")
	return nil
}
