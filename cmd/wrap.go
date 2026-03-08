package cmd

import (
	"fmt"
	"os"

	"github.com/nownow-labs/nownow/internal/api"
	"github.com/nownow-labs/nownow/internal/config"
	"github.com/nownow-labs/nownow/internal/wrap"
	"github.com/spf13/cobra"
)

var wrapCmd = &cobra.Command{
	Use:                "wrap [flags] -- command [args...]",
	Short:              "Run a command and push its result as status",
	DisableFlagParsing: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no command specified — usage: nownow wrap -- command [args...]")
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if !cfg.HasToken() {
			return fmt.Errorf("not logged in — run: nownow login")
		}

		name, _ := cmd.Flags().GetString("name")
		onSuccess, _ := cmd.Flags().GetString("on-success")
		onFailure, _ := cmd.Flags().GetString("on-failure")
		quiet, _ := cmd.Flags().GetBool("quiet")

		client := api.NewClient(cfg.Endpoint, cfg.Token)
		client.Version = Version

		opts := wrap.Options{
			Args:      args,
			Name:      name,
			OnSuccess: onSuccess,
			OnFailure: onFailure,
			Quiet:     quiet,
			PushFn: func(msg string) error {
				return client.PushStatus(api.StatusRequest{Content: msg})
			},
		}

		exitCode := wrap.Run(opts)
		if exitCode != 0 {
			// Exit directly to preserve the wrapped command's exit code.
			// RunE errors would print cobra usage, which is not wanted here.
			os.Exit(exitCode)
		}
		return nil
	},
}

func init() {
	wrapCmd.Flags().String("name", "", "human-readable name for the command")
	wrapCmd.Flags().String("on-success", "", "custom message on success (supports {cmd}, {name}, {duration})")
	wrapCmd.Flags().String("on-failure", "", "custom message on failure (supports {cmd}, {name}, {exit_code}, {duration})")
	wrapCmd.Flags().Bool("quiet", false, "suppress nownow output (only push, don't print)")
	rootCmd.AddCommand(wrapCmd)
}
