package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/nownow-labs/nownow/internal/hook"
	"github.com/spf13/cobra"
)

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage git hooks for automatic status updates",
}

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install nownow git hooks in the current repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		gitDir, err := hook.FindGitDir(cwd)
		if err != nil {
			return err
		}

		hooksFlag, _ := cmd.Flags().GetString("hooks")
		templateFlag, _ := cmd.Flags().GetString("template")

		hookNames := []string{"post-commit"}
		if hooksFlag != "" {
			hookNames = nil
			for _, h := range strings.Split(hooksFlag, ",") {
				h = strings.TrimSpace(h)
				if h != "" {
					hookNames = append(hookNames, h)
				}
			}
			if len(hookNames) == 0 {
				return fmt.Errorf("no valid hook names provided")
			}
		}

		var templates map[string]string
		if templateFlag != "" {
			templates = make(map[string]string)
			for _, name := range hookNames {
				templates[name] = hook.BuildTemplate(name, templateFlag)
			}
		}

		if err := hook.Install(gitDir, hookNames, templates); err != nil {
			return err
		}

		for _, name := range hookNames {
			fmt.Printf("installed: %s\n", name)
		}
		return nil
	},
}

var hookRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove nownow git hooks from the current repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		gitDir, err := hook.FindGitDir(cwd)
		if err != nil {
			return err
		}

		if err := hook.Remove(gitDir); err != nil {
			return err
		}
		fmt.Println("nownow hooks removed")
		return nil
	},
}

var hookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed nownow git hooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		gitDir, err := hook.FindGitDir(cwd)
		if err != nil {
			return err
		}

		hooks, err := hook.List(gitDir)
		if err != nil {
			return err
		}

		if len(hooks) == 0 {
			fmt.Println("no nownow hooks installed")
			return nil
		}
		for _, h := range hooks {
			fmt.Println(h)
		}
		return nil
	},
}

func init() {
	hookInstallCmd.Flags().String("hooks", "", "comma-separated list of hooks to install (default: post-commit)")
	hookInstallCmd.Flags().String("template", "", "custom message template (supports {commit_msg}, {branch})")

	hookCmd.AddCommand(hookInstallCmd)
	hookCmd.AddCommand(hookRemoveCmd)
	hookCmd.AddCommand(hookListCmd)
	rootCmd.AddCommand(hookCmd)
}
