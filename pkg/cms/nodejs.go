package cms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"svp/pkg/utils"
)

// NodeApp represents a detected Node.js application
type NodeApp struct {
	Path      string // Relative path from repository root
	Type      string // next, nuxt, svelte, react, etc.
	Name      string // Application name from package.json
	HasBuild  bool   // Whether it has a build script
	Port      int    // Suggested port number
}

// DetectNodeApps searches for Node.js applications in the repository
// It searches the project directory up to 2 levels deep for Node-based apps
func DetectNodeApps(projectDir string) ([]NodeApp, error) {
	var apps []NodeApp
	portCounter := 3000 // Start from port 3000

	// Directories to skip during search
	skipDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		".next":        true,
		".nuxt":        true,
		"dist":         true,
		"build":        true,
		"out":          true,
		".output":      true,
		"vendor":       true,
		".svelte-kit":  true,
		"coverage":     true,
		".cache":       true,
		"tmp":          true,
		"temp":         true,
		"public":       true,
		"static":       true,
		"assets":       true,
	}

	// Walk the directory tree up to 2 levels deep
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip paths with errors
		}

		// Skip files, only process directories
		if !info.IsDir() {
			return nil
		}

		// Calculate depth relative to projectDir
		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return nil
		}

		// Skip if path is the project root (we'll check it separately)
		if relPath == "." {
			return nil
		}

		// Calculate depth (number of directory separators)
		depth := strings.Count(relPath, string(os.PathSeparator))

		// Skip directories deeper than 2 levels
		if depth > 2 {
			return filepath.SkipDir
		}

		// Skip excluded directories
		dirName := filepath.Base(path)
		if skipDirs[dirName] {
			return filepath.SkipDir
		}

		// Check if this directory contains a Node app
		app := detectNodeAppInDir(path, projectDir, portCounter)
		if app != nil {
			apps = append(apps, *app)
			portCounter++
			// Skip subdirectories of detected apps to avoid duplicates
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Also check the root directory
	rootApp := detectNodeAppInDir(projectDir, projectDir, portCounter)
	if rootApp != nil {
		apps = append(apps, *rootApp)
	}

	return apps, nil
}

// detectNodeAppInDir checks if a specific directory contains a Node.js application
func detectNodeAppInDir(dir, projectRoot string, port int) *NodeApp {
	packageJSONPath := filepath.Join(dir, "package.json")

	if !utils.CheckFileExists(packageJSONPath) {
		return nil
	}

	// Read package.json
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return nil
	}

	// Determine app type
	appType := detectFrameworkType(dir, packageJSON)
	if appType == "" {
		return nil // Not a recognized Node framework
	}

	// Get relative path from project root
	relPath, err := filepath.Rel(projectRoot, dir)
	if err != nil {
		relPath = filepath.Base(dir)
	}
	if relPath == "." {
		relPath = ""
	}

	// Check for build script
	hasBuild := false
	if scripts, ok := packageJSON["scripts"].(map[string]interface{}); ok {
		if _, exists := scripts["build"]; exists {
			hasBuild = true
		}
	}

	// Get app name
	name := "node-app"
	if pkgName, ok := packageJSON["name"].(string); ok {
		name = pkgName
	}

	return &NodeApp{
		Path:     relPath,
		Type:     appType,
		Name:     name,
		HasBuild: hasBuild,
		Port:     port,
	}
}

// detectFrameworkType identifies the Node.js framework being used
func detectFrameworkType(dir string, packageJSON map[string]interface{}) string {
	// Check for framework-specific config files
	frameworkMarkers := map[string]string{
		"next.config.js":     "next",
		"next.config.mjs":    "next",
		"next.config.ts":     "next",
		"nuxt.config.js":     "nuxt",
		"nuxt.config.ts":     "nuxt",
		"svelte.config.js":   "svelte",
		"remix.config.js":    "remix",
		"astro.config.mjs":   "astro",
		"vite.config.js":     "vite",
		"vite.config.ts":     "vite",
		"gatsby-config.js":   "gatsby",
	}

	for marker, appType := range frameworkMarkers {
		if utils.CheckFileExists(filepath.Join(dir, marker)) {
			return appType
		}
	}

	// Check package.json dependencies
	deps := make(map[string]interface{})
	if d, ok := packageJSON["dependencies"].(map[string]interface{}); ok {
		deps = d
	}
	if d, ok := packageJSON["devDependencies"].(map[string]interface{}); ok {
		for k, v := range d {
			deps[k] = v
		}
	}

	// Check for framework dependencies
	frameworkDeps := map[string]string{
		"next":           "next",
		"nuxt":           "nuxt",
		"@sveltejs/kit":  "svelte",
		"@remix-run/react": "remix",
		"astro":          "astro",
		"gatsby":         "gatsby",
		"react-scripts":  "react",
		"vue":            "vue",
	}

	for dep, appType := range frameworkDeps {
		if _, exists := deps[dep]; exists {
			return appType
		}
	}

	// If has build script and dependencies, it's a generic Node app
	if scripts, ok := packageJSON["scripts"].(map[string]interface{}); ok {
		if _, exists := scripts["build"]; exists {
			return "node"
		}
		if _, exists := scripts["start"]; exists {
			return "node"
		}
	}

	return ""
}

// InstallNodeApp sets up a Node.js application
func InstallNodeApp(app NodeApp, domain, webroot, gitRepo, gitBranch, adminUser string) error {
	utils.Log("Setting up Node.js application: %s (%s)", app.Name, app.Type)

	projectDir := filepath.Join(webroot, domain)
	appDir := filepath.Join(projectDir, app.Path)

	// Ensure directory exists
	if !utils.CheckDirExists(appDir) {
		return fmt.Errorf("node app directory does not exist: %s", appDir)
	}

	// Install Node.js if not already installed
	if err := installNodeJS(adminUser); err != nil {
		return fmt.Errorf("failed to install Node.js: %v", err)
	}

	// Install dependencies
	utils.Log("Installing Node.js dependencies...")
	installCmd := fmt.Sprintf("cd %s && sudo -u %s npm install", appDir, adminUser)
	if _, err := utils.RunShell(installCmd); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	// Build if build script exists
	if app.HasBuild {
		utils.Log("Building Node.js application...")
		buildCmd := fmt.Sprintf("cd %s && sudo -u %s npm run build", appDir, adminUser)
		if _, err := utils.RunShell(buildCmd); err != nil {
			return fmt.Errorf("failed to build application: %v", err)
		}
	}

	utils.Ok("Node.js application %s setup complete", app.Name)
	return nil
}

// installNodeJS installs Node.js using NodeSource repository
func installNodeJS(adminUser string) error {
	// Check if Node.js is already installed
	if _, err := utils.RunCommand("node", "--version"); err == nil {
		utils.Log("Node.js is already installed")
		return nil
	}

	utils.Log("Installing Node.js...")

	// Install NodeSource repository (Node.js 20 LTS)
	// Using the new setup script
	setupCmd := "curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -"
	if _, err := utils.RunShell(setupCmd); err != nil {
		return fmt.Errorf("failed to setup NodeSource repository: %v", err)
	}

	// Install Node.js
	if _, err := utils.RunCommand("apt-get", "install", "-y", "nodejs"); err != nil {
		return fmt.Errorf("failed to install Node.js: %v", err)
	}

	utils.Ok("Node.js installed successfully")
	return nil
}

// CreateNodeSystemdService creates a systemd service for a Node.js application
func CreateNodeSystemdService(app NodeApp, domain, webroot, adminUser string) error {
	projectDir := filepath.Join(webroot, domain)
	appDir := filepath.Join(projectDir, app.Path)

	serviceName := fmt.Sprintf("node-%s", domain)
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)

	// Determine start command based on framework
	var execCmd string
	switch app.Type {
	case "next":
		execCmd = "/usr/bin/npm run start"
	case "nuxt":
		execCmd = "/usr/bin/npm run start"
	case "svelte":
		execCmd = "/usr/bin/npm run preview"
	case "astro":
		execCmd = "/usr/bin/npm run preview"
	default:
		execCmd = "/usr/bin/npm run start"
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=Node.js application for %s
After=network.target

[Service]
Type=simple
User=%s
WorkingDirectory=%s
ExecStart=%s
Restart=always
RestartSec=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=%s
Environment=NODE_ENV=production
Environment=PORT=%d

[Install]
WantedBy=multi-user.target
`, domain, adminUser, appDir, execCmd, serviceName, app.Port)

	// Write service file
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write systemd service file: %v", err)
	}

	// Reload systemd
	if _, err := utils.RunCommand("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	// Enable and start service
	if _, err := utils.RunCommand("systemctl", "enable", serviceName); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	if _, err := utils.RunCommand("systemctl", "start", serviceName); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	utils.Ok("Systemd service %s created and started", serviceName)
	return nil
}

// GetNodeAppSummary returns a human-readable summary of detected Node apps
func GetNodeAppSummary(apps []NodeApp) string {
	if len(apps) == 0 {
		return "No Node.js applications detected"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Found %d Node.js application(s):\n\n", len(apps)))

	for i, app := range apps {
		summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, app.Name))
		summary.WriteString(fmt.Sprintf("   Type: %s\n", app.Type))
		summary.WriteString(fmt.Sprintf("   Path: %s\n", app.Path))
		if app.HasBuild {
			summary.WriteString("   Build: Yes\n")
		}
		summary.WriteString(fmt.Sprintf("   Suggested Port: %d\n\n", app.Port))
	}

	return summary.String()
}
