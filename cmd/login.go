package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/nownow-labs/nownow/internal/api"
	"github.com/nownow-labs/nownow/internal/config"
	"github.com/spf13/cobra"
)

var tokenFlag string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with now.ctx.st",
	RunE: func(cmd *cobra.Command, args []string) error {
		if tokenFlag != "" {
			return loginWithToken(tokenFlag)
		}

		return loginWithDeviceFlow()
	},
}

func init() {
	loginCmd.Flags().StringVar(&tokenFlag, "token", "", "Login with a token directly (non-interactive)")
	rootCmd.AddCommand(loginCmd)
}

func loginWithToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	if !strings.HasPrefix(token, "now_") {
		return fmt.Errorf("invalid token format (should start with now_)")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Print("Verifying... ")
	client := api.NewClient(cfg.Endpoint, token)
	me, err := client.VerifyToken()
	if err != nil {
		fmt.Println("failed")
		return fmt.Errorf("token verification failed: %w", err)
	}
	fmt.Printf("ok (%s)\n", me.User.Name)

	cfg.Token = token
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	p, _ := config.Path()
	fmt.Printf("Token saved to %s\n", p)
	return nil
}

func loginWithDeviceFlow() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	client := api.NewClient(cfg.Endpoint, "")
	deviceResp, err := client.RequestDeviceCode()
	if err != nil {
		// Fall back to manual token entry
		fmt.Fprintf(os.Stderr, "Device flow unavailable: %v\n", err)
		fmt.Print("Paste your now.ctx.st token: ")
		reader := bufio.NewReader(os.Stdin)
		token, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		return loginWithToken(strings.TrimSpace(token))
	}

	fmt.Printf("Your code: %s\n", deviceResp.UserCode)
	fmt.Printf("Opening %s in your browser...\n\n", deviceResp.VerificationURL)

	if err := openBrowser(deviceResp.VerificationURL); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please open the URL above manually.\n\n")
	}

	fmt.Println("Waiting for authorization...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	interval := time.Duration(deviceResp.Interval) * time.Second
	if interval < time.Second {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(deviceResp.ExpiresIn) * time.Second)
	spinner := []rune{'|', '/', '-', '\\'}
	tick := 0

	for {
		if err := waitWithContext(ctx, interval); err != nil {
			fmt.Print("\r  \n")
			return fmt.Errorf("cancelled")
		}

		if time.Now().After(deadline) {
			fmt.Print("\r  \n")
			return fmt.Errorf("code expired, please run `nownow login` again")
		}

		fmt.Printf("\r %c ", spinner[tick%len(spinner)])
		tick++

		tokenResp, err := client.PollDeviceToken(deviceResp.DeviceCode)
		if err != nil {
			var pending *api.AuthPendingError
			if errors.As(err, &pending) {
				continue
			}
			var rle *api.RateLimitError
			if errors.As(err, &rle) {
				if err := waitWithContext(ctx, rle.RetryAfter); err != nil {
					fmt.Print("\r  \n")
					return fmt.Errorf("cancelled")
				}
				continue
			}
			fmt.Print("\r  \n")
			return fmt.Errorf("authentication failed: %w", err)
		}

		fmt.Print("\r  \n")
		cfg.Token = tokenResp.Token
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Logged in as %s\n", tokenResp.User.Name)
		return nil
	}
}

// waitWithContext waits for the given duration or until the context is cancelled.
func waitWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return cmd.Start()
}
