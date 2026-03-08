package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = "# nownow:start"
	markerEnd   = "# nownow:end"
	shebang     = "#!/bin/sh"
)

// DefaultTemplates maps hook names to their default push message templates.
var DefaultTemplates = map[string]string{
	"post-commit": `nownow push "Just committed: $(git log -1 --format=%s)"`,
	"pre-push":    `nownow push "Pushing to $(git rev-parse --abbrev-ref HEAD)"`,
}

// Install adds nownow hooks to the given git repository.
// hooks is a list of hook names (e.g. ["post-commit", "pre-push"]).
// templates overrides the default push command per hook (key = hook name).
func Install(gitDir string, hooks []string, templates map[string]string) error {
	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("creating hooks dir: %w", err)
	}

	for _, name := range hooks {
		if err := validateHookName(name); err != nil {
			return err
		}

		pushCmd := templates[name]
		if pushCmd == "" {
			pushCmd = DefaultTemplates[name]
		}
		if pushCmd == "" {
			return fmt.Errorf("no template for hook %q", name)
		}

		hookPath := filepath.Join(hooksDir, name)
		if err := installOne(hookPath, pushCmd); err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}
	}
	return nil
}

func validateHookName(name string) error {
	if name == "" {
		return fmt.Errorf("empty hook name")
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") || filepath.IsAbs(name) {
		return fmt.Errorf("invalid hook name %q: must not contain path separators", name)
	}
	return nil
}

func installOne(hookPath, pushCmd string) error {
	existing, err := os.ReadFile(hookPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(existing)

	// If already installed, remove old block first
	if strings.Contains(content, markerStart) {
		content = removeBlock(content)
	}

	block := fmt.Sprintf("%s\n%s\n%s", markerStart, pushCmd, markerEnd)

	if len(existing) == 0 {
		// New file
		content = shebang + "\n" + block + "\n"
	} else {
		// Append to existing
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += block + "\n"
	}

	return os.WriteFile(hookPath, []byte(content), 0o755)
}

// Remove uninstalls nownow hooks from the given git repository.
func Remove(gitDir string) error {
	hooksDir := filepath.Join(gitDir, "hooks")
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var errs []error
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		hookPath := filepath.Join(hooksDir, e.Name())
		data, err := os.ReadFile(hookPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("reading %s: %w", e.Name(), err))
			continue
		}
		content := string(data)
		if !strings.Contains(content, markerStart) {
			continue
		}

		cleaned := removeBlock(content)
		// If only shebang (and whitespace) remains, delete the file
		trimmed := strings.TrimSpace(cleaned)
		if trimmed == "" || trimmed == shebang {
			if err := os.Remove(hookPath); err != nil {
				errs = append(errs, fmt.Errorf("removing %s: %w", e.Name(), err))
			}
		} else {
			if err := os.WriteFile(hookPath, []byte(cleaned), 0o755); err != nil {
				errs = append(errs, fmt.Errorf("writing %s: %w", e.Name(), err))
			}
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// List returns the names of hooks that contain nownow blocks.
func List(gitDir string) ([]string, error) {
	hooksDir := filepath.Join(gitDir, "hooks")
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		hookPath := filepath.Join(hooksDir, e.Name())
		data, err := os.ReadFile(hookPath)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), markerStart) {
			result = append(result, e.Name())
		}
	}
	return result, nil
}

// FindGitDir walks up from dir looking for a .git directory or file (worktree).
func FindGitDir(dir string) (string, error) {
	for {
		gitPath := filepath.Join(dir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil {
			if info.IsDir() {
				return gitPath, nil
			}
			// Worktree/submodule: .git is a file containing "gitdir: <path>"
			resolved, err := resolveGitFile(gitPath)
			if err != nil {
				return "", fmt.Errorf("parsing .git file: %w", err)
			}
			return resolved, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository (no .git found)")
		}
		dir = parent
	}
}

// resolveGitFile reads a .git file (used by worktrees/submodules) and returns
// the absolute path to the actual git directory.
func resolveGitFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(string(data))
	if !strings.HasPrefix(line, "gitdir: ") {
		return "", fmt.Errorf("unexpected .git file format: %s", line)
	}
	target := strings.TrimPrefix(line, "gitdir: ")
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return filepath.Clean(target), nil
}

func removeBlock(content string) string {
	var lines []string
	inBlock := false
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == markerStart {
			inBlock = true
			continue
		}
		if strings.TrimSpace(line) == markerEnd {
			inBlock = false
			continue
		}
		if !inBlock {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// BuildTemplate creates a push command from a user-provided template string.
// Supported variables: {commit_msg}, {branch}.
func BuildTemplate(hookName, tmpl string) string {
	msg := tmpl
	msg = strings.ReplaceAll(msg, "{commit_msg}", "$(git log -1 --format=%s)")
	msg = strings.ReplaceAll(msg, "{branch}", "$(git rev-parse --abbrev-ref HEAD)")
	return fmt.Sprintf(`nownow push "%s"`, msg)
}
