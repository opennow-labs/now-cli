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

		var content, emoji, app string

		if len(args) > 0 {
			// Manual message
			content = strings.Join(args, " ")
		} else {
			// Auto-detect
			ctx := detect.Detect()

			if cfg.IsIgnored(ctx.App) {
				fmt.Println("ignored app, skipping")
				return nil
			}

			app = ctx.App
			emoji = cfg.EmojiFor(ctx.App, "")
			if ctx.HasMusic() && emoji == "" {
				emoji = "\U0001F3B5"
			}

			content = template.Render(cfg.Template, ctx, emoji)
		}

		if content == "" {
			fmt.Println("nothing to push")
			return nil
		}

		client := api.NewClient(cfg.Endpoint, cfg.Token)
		client.Version = Version
		client.Telemetry = cfg.TelemetryEnabled()
		if err := client.PushStatus(content, emoji, app); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Printf("pushed: %s\n", content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
