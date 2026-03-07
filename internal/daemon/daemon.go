package daemon

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nownow-labs/nownow/internal/config"
	"github.com/nownow-labs/nownow/internal/settings"
	"github.com/nownow-labs/nownow/internal/tray"
)

// PidFile returns the path to the daemon PID file.
func PidFile() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "daemon.pid"), nil
}

// IsRunning checks if a daemon process is alive.
func IsRunning() (bool, int) {
	pidPath, err := PidFile()
	if err != nil {
		return false, 0
	}
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false, 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false, 0
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, 0
	}
	// Signal 0 checks if process exists
	if err := process.Signal(syscall.Signal(0)); err != nil {
		os.Remove(pidPath)
		return false, 0
	}
	return true, pid
}

// WritePid writes the current process PID to the pid file.
func WritePid() error {
	pidPath, err := PidFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pidPath), 0700); err != nil {
		return err
	}
	return os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0600)
}

// RemovePid removes the pid file only if it still belongs to the current process.
// This prevents a restarting process from deleting a new process's pid file.
func RemovePid() {
	p, err := PidFile()
	if err != nil {
		return
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		os.Remove(p)
		return
	}
	if pid == os.Getpid() {
		os.Remove(p)
	}
}

// Stop sends SIGTERM to the running daemon and waits for it to exit.
func Stop() error {
	running, pid := IsRunning()
	if !running {
		return fmt.Errorf("daemon is not running")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon (pid %d): %w", pid, err)
	}
	// Wait for process to exit (up to 5 seconds)
	for i := 0; i < 50; i++ {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			// Process gone
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("daemon (pid %d) did not exit within 5 seconds", pid)
}

// StartDetached launches the daemon as a background process.
func StartDetached() error {
	if running, pid := IsRunning(); running {
		return fmt.Errorf("daemon already running (pid %d)", pid)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	cmd := exec.Command(exe, "start", "--foreground")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	// Detach from parent process group
	cmd.SysProcAttr = detachedProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("daemon started (pid %d)\n", cmd.Process.Pid)
	return nil
}

// RunForeground runs the menubar tray + push loop (called by detached process).
func RunForeground(interval time.Duration) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !cfg.HasToken() {
		return fmt.Errorf("not logged in — run: nownow login")
	}

	if err := WritePid(); err != nil {
		return fmt.Errorf("writing pid: %w", err)
	}
	defer RemovePid()

	// Start settings HTTP server
	settings.AutostartIsInstalled = IsAutostartInstalled
	settings.AutostartInstall = InstallAutostart
	settings.AutostartUninstall = UninstallAutostart
	if err := settings.Start(tray.Version); err != nil {
		log.Printf("warning: settings UI unavailable: %v", err)
	} else {
		tray.SettingsAvailable = true
	}

	// Launch systray — this blocks on the main thread
	tray.Run(interval)
	return nil
}

// InstallAutostart and UninstallAutostart are implemented per-platform
// in autostart_darwin.go, autostart_linux.go, autostart_windows.go.
