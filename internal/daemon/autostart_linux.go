//go:build linux

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

const desktopEntry = `[Desktop Entry]
Type=Application
Name=now
Exec=%s start --foreground
Hidden=false
NoDisplay=true
X-GNOME-Autostart-enabled=true
`

func autostartPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "autostart")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "now.desktop"), nil
}

// IsAutostartInstalled returns true if the .desktop autostart entry exists.
func IsAutostartInstalled() bool {
	p, err := autostartPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(p)
	return err == nil
}

// InstallAutostart creates a .desktop autostart entry.
func InstallAutostart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	p, err := autostartPath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(desktopEntry, exe)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("autostart installed: %s\n", p)
	return nil
}

// BootoutService is a no-op on Linux (launchd is macOS-only).
func BootoutService() error { return nil }

// BootstrapService is a no-op on Linux (launchd is macOS-only).
func BootstrapService() error { return nil }

// IsServiceLoaded always returns false on Linux.
func IsServiceLoaded() bool { return false }

// launchdRestart is a no-op on Linux.
func launchdRestart() error { return fmt.Errorf("launchd not available on Linux") }

// startViaServiceManager is a no-op on Linux — no service manager integration.
func startViaServiceManager() (bool, error) { return false, nil }

// stopViaServiceManager is a no-op on Linux.
func stopViaServiceManager() (bool, error) { return false, nil }

// LogDir returns "" on Linux (logs are in the config directory).
func LogDir() string { return "" }

// UninstallAutostart removes the .desktop autostart entry.
func UninstallAutostart() error {
	p, err := autostartPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	fmt.Println("autostart removed")
	return nil
}
