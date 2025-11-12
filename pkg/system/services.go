package system

import (
	"fmt"
	"svp/pkg/utils"
)

// EnableService enables a systemd service
func EnableService(service string) error {
	_, err := utils.RunCommand("systemctl", "enable", service)
	if err != nil {
		return fmt.Errorf("failed to enable %s: %v", service, err)
	}
	return nil
}

// StartService starts a systemd service
func StartService(service string) error {
	_, err := utils.RunCommand("systemctl", "start", service)
	if err != nil {
		return fmt.Errorf("failed to start %s: %v", service, err)
	}
	return nil
}

// RestartService restarts a systemd service
func RestartService(service string) error {
	_, err := utils.RunCommand("systemctl", "restart", service)
	if err != nil {
		return fmt.Errorf("failed to restart %s: %v", service, err)
	}
	return nil
}

// StopService stops a systemd service
func StopService(service string) error {
	_, err := utils.RunCommand("systemctl", "stop", service)
	// Ignore errors when stopping (service might not be running)
	return err
}

// DisableService disables a systemd service
func DisableService(service string) error {
	_, err := utils.RunCommand("systemctl", "disable", service)
	// Ignore errors when disabling
	return err
}

// ReloadService reloads a systemd service
func ReloadService(service string) error {
	_, err := utils.RunCommand("systemctl", "reload", service)
	if err != nil {
		return fmt.Errorf("failed to reload %s: %v", service, err)
	}
	return nil
}

// EnsureServiceRunning ensures a service is enabled and running
func EnsureServiceRunning(service string, verifyOnly bool) error {
	enabled := utils.CheckServiceEnabled(service)
	active := utils.CheckServiceActive(service)

	if !enabled {
		if verifyOnly {
			utils.Fail("%s not enabled", service)
			return fmt.Errorf("%s not enabled", service)
		}
		utils.Fix("Enabling %s service", service)
		if err := EnableService(service); err != nil {
			return err
		}
	}

	if !active {
		if verifyOnly {
			utils.Fail("%s not running", service)
			return fmt.Errorf("%s not running", service)
		}
		utils.Fix("Starting %s service", service)
		if err := StartService(service); err != nil {
			return err
		}
	}

	if enabled && active {
		utils.Ok("%s running", service)
	}

	return nil
}
