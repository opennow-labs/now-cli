package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ctx-st/nownow/internal/api"
	"github.com/ctx-st/nownow/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your API token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		fmt.Print("Paste your now.ctx.st token: ")
		reader := bufio.NewReader(os.Stdin)
		token, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		token = strings.TrimSpace(token)

		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		if !strings.HasPrefix(token, "now_") {
			return fmt.Errorf("invalid token format (should start with now_)")
		}

		// Verify the token
		fmt.Print("Verifying... ")
		client := api.NewClient(cfg.Endpoint, token)
		me, err := client.VerifyToken()
		if err != nil {
			fmt.Println("failed")
			return fmt.Errorf("token verification failed: %w", err)
		}
		fmt.Printf("ok (%s)\n", me.User.Name)

		// Save
		cfg.Token = token
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		p, _ := config.Path()
		fmt.Printf("Token saved to %s\n", p)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
