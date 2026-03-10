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

// BootoutService is a no-op on Windows.
func BootoutService() error { return nil }

// BootstrapService is a no-op on Windows.
func BootstrapService() error { return nil }

// IsServiceLoaded always returns false on Windows.
func IsServiceLoaded() bool { return false }

// launchdRestart is a no-op on Windows.
func launchdRestart() error { return fmt.Errorf("launchd not available on Windows") }

// startViaServiceManager is a no-op on Windows — no service manager integration.
func startViaServiceManager() (bool, error) { return false, nil }

// stopViaServiceManager is a no-op on Windows.
func stopViaServiceManager() (bool, error) { return false, nil }

// LogDir returns "" on Windows (logs are in the config directory).
func LogDir() string { return "" }

// UninstallAutostart is a no-op on Windows.
func UninstallAutostart() error {
	return nil
}
