package cmd

import (
	"fmt"
	"strings"

	"github.com/nownow-labs/nownow/internal/api"
	"github.com/nownow-labs/nownow/internal/config"
	"github.com/nownow-labs/nownow/internal/detect"
	"github.com/nownow-labs/nownow/internal/template"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [message]",
	Short: "Detect context and push status (or push a manual message)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if !cfg.HasToken() {
			return fmt.Errorf("not logged in — run: nownow login")
		}

		var req api.StatusRequest

		if len(args) > 0 {
			// Manual message
			req.Content = strings.Join(args, " ")
		} else {
			// Auto-detect
			ctx := detect.Detect()

			if cfg.IsIgnored(ctx.App) {
				fmt.Println("ignored app, skipping")
				return nil
			}

			// Sanitize context before rendering so privacy-disabled fields
			// never leak into activity/content strings.
			if !cfg.SendMusicEnabled() {
				ctx.MusicArtist = ""
				ctx.MusicTrack = ""
			}
			if !cfg.SendWatchingEnabled() {
				ctx.Watching = ""
			}
			if !cfg.SendAppEnabled() {
				ctx.App = ""
			}

			activity := cfg.ResolveActivity(ctx.App, ctx.Watching)

			req.App = ctx.App
			req.Activity = activity
			req.MusicArtist = ctx.MusicArtist
			req.MusicTrack = ctx.MusicTrack
			req.Watching = ctx.Watching
			req.Content = template.Render(cfg.Template, ctx, activity)
		}

		if req.Content == "" {
			fmt.Println("nothing to push")
			return nil
		}

		client := api.NewClient(cfg.Endpoint, cfg.Token)
		client.Version = Version
		client.Telemetry = cfg.TelemetryEnabled()
		client.SendApp = cfg.SendAppEnabled()
		client.SendMusic = cfg.SendMusicEnabled()
		client.SendWatching = cfg.SendWatchingEnabled()
		if err := client.PushStatus(req); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Printf("pushed: %s\n", req.Content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
