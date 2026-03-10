package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/opennow-labs/now-cli/internal/config"
	"github.com/opennow-labs/now-cli/internal/daemon"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall now: stop daemon, remove autostart and binary",
	Long: `Uninstall now by stopping the daemon, removing autostart configuration,
and deleting the binary.

Use --purge to also remove config and log directories.`,
	RunE: runUninstall,
}

var (
	uninstallPurge bool
	uninstallYes   bool
)

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallPurge, "purge", false, "also remove config and log directories")
	uninstallCmd.Flags().BoolVarP(&uninstallYes, "yes", "y", false, "skip confirmation prompt")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Build the confirmation message
	fmt.Fprintln(os.Stderr, "This will:")
	fmt.Fprintln(os.Stderr, "  - Stop the daemon")
	fmt.Fprintln(os.Stderr, "  - Remove autostart (launchd/desktop entry)")
	fmt.Fprintln(os.Stderr, "  - Remove the now binary")
	if uninstallPurge {
		configDir, _ := config.Dir()
		if configDir != "" {
			fmt.Fprintf(os.Stderr, "  - Remove config: %s\n", configDir)
		}
		if logDir := daemon.LogDir(); logDir != "" {
			fmt.Fprintf(os.Stderr, "  - Remove logs: %s\n", logDir)
		}
	}

	if !uninstallYes {
		fmt.Fprint(os.Stderr, "\nAre you sure? [y/N]: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() || !strings.EqualFold(strings.TrimSpace(scanner.Text()), "y") {
			fmt.Fprintln(os.Stderr, "cancelled")
			return nil
		}
	}

	// 1. Stop daemon
	if err := daemon.Stop(); err != nil {
		if strings.Contains(err.Error(), "not running") {
			fmt.Fprintln(os.Stderr, "✓ daemon not running")
		} else {
			return fmt.Errorf("stopping daemon: %w", err)
		}
	} else {
		fmt.Fprintln(os.Stderr, "✓ daemon stopped")
	}

	// 2. Remove autostart
	if !daemon.IsAutostartInstalled() {
		fmt.Fprintln(os.Stderr, "✓ autostart not installed")
	} else if err := daemon.UninstallAutostart(); err != nil {
		return fmt.Errorf("removing autostart: %w", err)
	}

	// 3. Remove binary BEFORE purge to avoid half-uninstalled state.
	//    On Unix, deleting a running binary is safe (inode stays alive).
	//    On Windows, removeSelf() renames the .exe and spawns a cleanup process.
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding binary path: %w", err)
	}
	pending, err := removeSelf(exe)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing binary: %w", err)
	}
	if pending != "" {
		fmt.Fprintf(os.Stderr, "✓ binary renamed to %s (will be deleted after exit)\n", pending)
	} else {
		fmt.Fprintf(os.Stderr, "✓ binary removed: %s\n", exe)
	}

	// 4. Purge: remove config and log directories (safe now that binary is handled)
	if uninstallPurge {
		configDir, _ := config.Dir()
		if configDir != "" {
			if _, err := os.Stat(configDir); err == nil {
				if err := os.RemoveAll(configDir); err != nil {
					return fmt.Errorf("removing config: %w", err)
				}
				fmt.Fprintf(os.Stderr, "✓ config removed: %s\n", configDir)
			}
		}

		if logDir := daemon.LogDir(); logDir != "" {
			if _, err := os.Stat(logDir); err == nil {
				if err := os.RemoveAll(logDir); err != nil {
					return fmt.Errorf("removing logs: %w", err)
				}
				fmt.Fprintf(os.Stderr, "✓ logs removed: %s\n", logDir)
			}
		}
	}

	if uninstallPurge {
		fmt.Fprintln(os.Stderr, "\nnow has been completely uninstalled.")
	} else {
		fmt.Fprintln(os.Stderr)
		configDir, _ := config.Dir()
		if configDir != "" {
			fmt.Fprintf(os.Stderr, "Config and logs were kept. To remove them manually:\n")
			fmt.Fprintf(os.Stderr, "  rm -rf %s\n", configDir)
			if logDir := daemon.LogDir(); logDir != "" {
				fmt.Fprintf(os.Stderr, "  rm -rf %s\n", logDir)
			}
		}
	}

	return nil
}
