package cmd

import (
	"fmt"
	"os"

	"github.com/nownow-labs/nownow/internal/daemon"
	"github.com/nownow-labs/nownow/internal/upgrade"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade nownow to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Checking for updates... ")

		release, err := upgrade.CheckLatest()
		if err != nil {
			return err
		}

		latest := upgrade.NormalizeVersion(release.TagName)
		current := upgrade.NormalizeVersion(Version)

		if !upgrade.IsNewer(current, latest) {
			fmt.Printf("already up to date (%s).\n", current)
			return nil
		}

		fmt.Printf("%s -> %s\n", current, latest)

		asset, err := upgrade.FindAsset(release)
		if err != nil {
			return err
		}

		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("locating current binary: %w", err)
		}

		fmt.Printf("Downloading %s...\n", asset.Name)

		if err := upgrade.Download(asset, execPath); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("permission denied. Try: sudo nownow upgrade")
			}
			return err
		}

		fmt.Printf("Upgraded to %s.\n", latest)

		if running, _ := daemon.IsRunning(); running {
			fmt.Println("Daemon is running. Restart to apply: nownow stop && nownow start")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
