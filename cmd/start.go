package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ctx-st/nownow/internal/api"
	"github.com/ctx-st/nownow/internal/config"
	"github.com/ctx-st/nownow/internal/detect"
	"github.com/ctx-st/nownow/internal/template"
	"github.com/spf13/cobra"
)

var startInterval string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Watch and auto-push status on an interval",
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

		client := api.NewClient(cfg.Endpoint, cfg.Token)

		fmt.Printf("nownow watching (every %s) — Ctrl+C to stop\n", interval)

		// Run once immediately
		pushOnce(cfg, client)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		for {
			select {
			case <-ticker.C:
				pushOnce(cfg, client)
			case <-sig:
				fmt.Println("\nstopping")
				return nil
			}
		}
	},
}

func pushOnce(cfg config.Config, client *api.Client) {
	ctx := detect.Detect()

	if cfg.IsIgnored(ctx.App) {
		return
	}

	emoji := cfg.EmojiFor(ctx.App, "")
	if ctx.HasMusic() && emoji == "" {
		emoji = "\U0001F3B5"
	}

	content := template.Render(cfg.Template, ctx, emoji)
	if content == "" {
		return
	}

	err := client.PushStatus(content, emoji)
	if err != nil {
		var rle *api.RateLimitError
		if errors.As(err, &rle) {
			fmt.Printf("[%s] rate limited, waiting %s\n", timeNow(), rle.RetryAfter)
			time.Sleep(rle.RetryAfter)
			return
		}
		fmt.Printf("[%s] push error: %s\n", timeNow(), err)
		return
	}
	fmt.Printf("[%s] %s\n", timeNow(), content)
}

func timeNow() string {
	return time.Now().Format("15:04:05")
}

func init() {
	startCmd.Flags().StringVar(&startInterval, "interval", "", "push interval (default from config, e.g. 5m)")
	rootCmd.AddCommand(startCmd)
}
