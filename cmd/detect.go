package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/nownow-labs/nownow/internal/detect"
	"github.com/spf13/cobra"
)

var detectJSON bool

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect current context and print it",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := detect.Detect()

		if detectJSON {
			data, err := json.MarshalIndent(ctx, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		// Human-readable output
		if ctx.App != "" {
			fmt.Printf("app:     %s\n", ctx.App)
		}
		if ctx.WindowTitle != "" {
			fmt.Printf("title:   %s\n", ctx.WindowTitle)
		}
		if ctx.Project != "" {
			fmt.Printf("project: %s\n", ctx.Project)
		}
		if ctx.Branch != "" {
			fmt.Printf("branch:  %s\n", ctx.Branch)
		}
		if ctx.HasMusic() {
			fmt.Printf("music:   %s\n", ctx.Music())
		}
		if ctx.HasWatching() {
			fmt.Printf("watching: %s\n", ctx.Watching)
		}
		return nil
	},
}

func init() {
	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(detectCmd)
}
