//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// removeSelf renames the running .exe (Windows allows this) and spawns a
// detached cleanup process that deletes the renamed file after we exit.
// Returns ("", nil) on full success, or (renamedPath, nil) when the binary
// was renamed but async cleanup could not be scheduled.
func removeSelf(exe string) (pendingCleanup string, err error) {
	renamed := exe + fmt.Sprintf(".uninstall.%d", os.Getpid())

	// Clean up any leftover from a previous failed uninstall.
	_ = os.Remove(renamed)

	if err := os.Rename(exe, renamed); err != nil {
		return "", err
	}

	// Spawn a detached batch script to delete the renamed file after we exit.
	bat := filepath.Join(os.TempDir(), fmt.Sprintf("now-uninstall-%d.bat", os.Getpid()))
	script := fmt.Sprintf(
		"@echo off\r\ntimeout /t 2 /nobreak >nul 2>&1\r\ndel /f /q \"%s\"\r\ndel /f /q \"%%~f0\"\r\n",
		renamed,
	)
	if err := os.WriteFile(bat, []byte(script), 0700); err != nil {
		return renamed, nil
	}
	cmd := exec.Command("cmd", "/C", bat)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000008} // DETACHED_PROCESS
	if err := cmd.Start(); err != nil {
		_ = os.Remove(bat)
		return renamed, nil
	}
	return "", nil
}
