//go:build linux

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
)

const desktopEntry = `[Desktop Entry]
Type=Application
Name=nownow
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
	return filepath.Join(dir, "nownow.desktop"), nil
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
