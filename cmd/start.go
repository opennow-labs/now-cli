package cmd

import (
	"fmt"
	"time"

	"github.com/nownow-labs/nownow/internal/config"
	"github.com/nownow-labs/nownow/internal/daemon"
	"github.com/nownow-labs/nownow/internal/tray"
	"github.com/spf13/cobra"
)

var (
	startInterval   string
	startForeground bool
	startNoAutostart bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the nownow daemon (background by default)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if !cfg.HasToken() {
			return fmt.Errorf("not logged in — run: nownow login")
		}

		intervalStr := startInterval
		if intervalStr == "" {
			intervalStr = cfg.Interval
		}
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return fmt.Errorf("invalid interval %q: %w", intervalStr, err)
		}

		if startForeground {
			// Run in foreground (used by detached process and launchd)
			tray.Version = Version
			tray.RestartFunc = daemon.Restart
			return daemon.RunForeground(interval)
		}

		// Launch as background daemon
		if err := daemon.StartDetached(); err != nil {
			return err
		}

		// Install autostart on first run
		if !startNoAutostart {
			if err := daemon.InstallAutostart(); err != nil {
				// Non-fatal — just warn
				fmt.Printf("note: autostart setup skipped (%s)\n", err)
			}
		}

		return nil
	},
}

func init() {
	startCmd.Flags().StringVar(&startInterval, "interval", "", "push interval (default from config, e.g. 5m)")
	startCmd.Flags().BoolVar(&startForeground, "foreground", false, "run in foreground (used internally)")
	startCmd.Flags().BoolVar(&startNoAutostart, "no-autostart", false, "skip autostart installation")
	startCmd.Flags().MarkHidden("foreground")
	rootCmd.AddCommand(startCmd)
}
