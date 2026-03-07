//go:build windows

package daemon

import "fmt"

// IsAutostartInstalled returns false on Windows (not supported).
func IsAutostartInstalled() bool {
	return false
}

// InstallAutostart is a stub on Windows — users should use Task Scheduler.
func InstallAutostart() error {
	return fmt.Errorf("autostart not supported on Windows yet — use Task Scheduler")
}

// UninstallAutostart is a no-op on Windows.
func UninstallAutostart() error {
	return nil
}
