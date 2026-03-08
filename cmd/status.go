package cmd

import (
	"fmt"
	"sync"

	"github.com/opennow-labs/now-cli/internal/api"
	"github.com/opennow-labs/now-cli/internal/config"
	"github.com/opennow-labs/now-cli/internal/daemon"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show your current status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if !cfg.HasToken() {
			return fmt.Errorf("not logged in — run: now login")
		}

		// Daemon status
		if running, pid := daemon.IsRunning(); running {
			fmt.Printf("daemon: running (pid %d)\n", pid)
		} else {
			fmt.Println("daemon: not running")
		}

		client := api.NewClient(cfg.Endpoint, cfg.Token)

		// Fetch identity and live view concurrently
		var (
			me      *api.MeResponse
			live    *api.LiveResponse
			meErr   error
			liveErr error
			wg      sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			me, meErr = client.VerifyToken()
		}()
		go func() {
			defer wg.Done()
			live, liveErr = client.GetLive()
		}()
		wg.Wait()

		if meErr != nil {
			return fmt.Errorf("auth failed: %w", meErr)
		}
		if liveErr != nil {
			return fmt.Errorf("fetching live status: %w", liveErr)
		}

		// Find ourselves
		for _, entry := range live.Feed {
			if entry.ID == me.User.ID {
				if entry.Status != "" {
					fmt.Printf("%s %s\n", entry.Emoji, entry.Status)
				} else {
					fmt.Println("(no status set)")
				}
				if entry.LastSeenAt != "" {
					fmt.Printf("last seen: %s\n", entry.LastSeenAt)
				}
				return nil
			}
		}

		fmt.Println("(no status set)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
