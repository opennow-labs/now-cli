//go:build darwin

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const launchdLabel = "st.ctx.now.nownow"

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
  <true/>
  <key>StandardOutPath</key>
  <string>{{.LogDir}}/nownow.log</string>
  <key>StandardErrorPath</key>
  <string>{{.LogDir}}/nownow.err</string>
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
	logDir := filepath.Join(home, "Library", "Logs", "nownow")
	os.MkdirAll(logDir, 0700)

	tmpl, err := template.New("plist").Parse(launchdPlist)
	if err != nil {
		return err
	}

	f, err := os.Create(plistPath())
	if err != nil {
		return fmt.Errorf("creating plist: %w", err)
	}
	defer f.Close()

	err = tmpl.Execute(f, map[string]string{
		"Label":  launchdLabel,
		"Exe":    exe,
		"LogDir": logDir,
	})
	if err != nil {
		return err
	}

	fmt.Printf("autostart installed: %s\n", plistPath())
	return nil
}

// UninstallAutostart removes the launchd plist.
func UninstallAutostart() error {
	p := plistPath()
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	fmt.Println("autostart removed")
	return nil
}
