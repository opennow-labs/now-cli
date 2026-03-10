//go:build darwin

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"text/template"
)

const launchdLabel = "dev.opennow.cli"

var launchdPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>{{.Label}}</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.Exe}}</string>
    <string>start</string>
    <string>--foreground</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key>
    <false/>
  </dict>
  <key>StandardOutPath</key>
  <string>{{.LogDir}}/now.log</string>
  <key>StandardErrorPath</key>
  <string>{{.LogDir}}/now.err</string>
</dict>
</plist>
`

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
}

// IsAutostartInstalled returns true if the launchd plist exists.
func IsAutostartInstalled() bool {
	_, err := os.Stat(plistPath())
	return err == nil
}

// InstallAutostart creates a launchd plist for login startup.
func InstallAutostart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	logDir := filepath.Join(home, "Library", "Logs", "now")
	os.MkdirAll(logDir, 0700)

	tmpl, err := template.New("plist").Parse(launchdPlist)
	if err != nil {
		return err
	}

	f, err := os.Create(plistPath())
	if err != nil {
		return fmt.Errorf("creating plist: %w", err)
	}

	err = tmpl.Execute(f, map[string]string{
		"Label":  launchdLabel,
		"Exe":    exe,
		"LogDir": logDir,
	})
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	// Try to load the service into launchd so it takes effect immediately.
	// This is best-effort: bootstrap may fail in non-GUI sessions (SSH, CI)
	// or when the daemon is already running. The plist with RunAtLoad=true
	// guarantees it will load on next login regardless.
	if domain, err := guiDomain(); err == nil && !isServiceLoaded() {
		if err := exec.Command("launchctl", "bootstrap", domain, plistPath()).Run(); err != nil {
			fmt.Fprintf(os.Stderr, "note: launchctl bootstrap skipped (%v), will activate on next login\n", err)
		}
	}

	fmt.Printf("autostart installed: %s\n", plistPath())
	return nil
}

func guiDomain() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}
	return "gui/" + u.Uid, nil
}

func isServiceLoaded() bool {
	domain, err := guiDomain()
	if err != nil {
		return false
	}
	return exec.Command("launchctl", "print", domain+"/"+launchdLabel).Run() == nil
}

// BootoutService unloads the service from launchd without deleting the plist.
// The service can be re-loaded later via BootstrapService or on next login.
func BootoutService() error {
	if !isServiceLoaded() {
		return nil
	}
	domain, err := guiDomain()
	if err != nil {
		return fmt.Errorf("getting gui domain: %w", err)
	}
	if err := exec.Command("launchctl", "bootout", domain+"/"+launchdLabel).Run(); err != nil {
		return fmt.Errorf("launchctl bootout: %w", err)
	}
	return nil
}

// BootstrapService loads an existing plist into launchd.
func BootstrapService() error {
	if isServiceLoaded() {
		return nil
	}
	domain, err := guiDomain()
	if err != nil {
		return fmt.Errorf("getting gui domain: %w", err)
	}
	if err := exec.Command("launchctl", "bootstrap", domain, plistPath()).Run(); err != nil {
		return fmt.Errorf("launchctl bootstrap: %w", err)
	}
	return nil
}

// IsServiceLoaded returns true if the launchd service is currently loaded.
func IsServiceLoaded() bool {
	return isServiceLoaded()
}

// launchdRestart spawns a detached subprocess that does bootout+bootstrap.
// This must be detached because bootout sends SIGTERM to the caller process.
// Uses environment variables instead of shell interpolation to avoid escaping issues.
func launchdRestart() error {
	domain, err := guiDomain()
	if err != nil {
		return err
	}
	target := domain + "/" + launchdLabel
	plist := plistPath()

	// Pass arguments via env vars to avoid shell escaping issues with paths.
	cmd := exec.Command("/bin/sh", "-c",
		`/bin/launchctl bootout "$LAUNCHD_TARGET" && sleep 1 && /bin/launchctl bootstrap "$LAUNCHD_DOMAIN" "$LAUNCHD_PLIST"`,
	)
	cmd.Env = append(os.Environ(),
		"LAUNCHD_TARGET="+target,
		"LAUNCHD_DOMAIN="+domain,
		"LAUNCHD_PLIST="+plist,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = detachedProcAttr()
	return cmd.Start()
}

// startViaServiceManager tries to start/ensure the daemon via launchd.
// Returns (true, nil) if launchd is now managing the daemon.
// Returns (false, nil) if launchd is not available and caller should fall back.
func startViaServiceManager() (managed bool, err error) {
	if isServiceLoaded() {
		if running, pid := IsRunning(); running {
			return true, fmt.Errorf("daemon already running (pid %d), managed by launchd", pid)
		}
		// Service loaded but process not running — launchd will restart it
		fmt.Println("daemon is managed by launchd and will restart automatically")
		return true, nil
	}

	// Plist exists or first install — update plist and bootstrap
	if err := InstallAutostart(); err != nil {
		return false, nil // can't install, let caller fall back
	}
	if isServiceLoaded() {
		fmt.Println("daemon started via launchd")
		return true, nil
	}
	// Bootstrap failed (SSH, CI, etc.) — let caller fall back to StartDetached
	return false, nil
}

// stopViaServiceManager unloads the service from launchd if loaded.
// Returns true if launchd was managing the service and bootout was performed.
func stopViaServiceManager() (bool, error) {
	if !IsAutostartInstalled() || !isServiceLoaded() {
		return false, nil
	}
	if err := BootoutService(); err != nil {
		return true, fmt.Errorf("bootout failed: %w", err)
	}
	return true, nil
}

// LogDir returns the daemon log directory on macOS.
func LogDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "Logs", "now")
}

// UninstallAutostart removes the launchd plist.
func UninstallAutostart() error {
	p := plistPath()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	}
	// Unload the service from launchd before removing the plist.
	_ = BootoutService()
	if err := os.Remove(p); err != nil {
		return err
	}
	fmt.Println("autostart removed")
	return nil
}
