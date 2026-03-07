package daemon

import (
	"fmt"
	"os"
	"os/exec"

	"fyne.io/systray"
)

// Restart spawns a new daemon process and quits the current one.
func Restart() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating executable: %w", err)
	}

	cmd := exec.Command(exe, "start", "--foreground")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = detachedProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting new daemon: %w", err)
	}

	// Quit current systray (triggers cleanup via onExit)
	systray.Quit()
	return nil
}
