package detect

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// detectGit returns the project name and current branch.
func detectGit() (project, branch string) {
	// Get the top-level directory of the git repo
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", ""
	}
	topLevel := strings.TrimSpace(string(out))
	if topLevel == "" {
		return "", ""
	}
	project = filepath.Base(topLevel)

	// Get current branch
	out, err = exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return project, ""
	}
	branch = strings.TrimSpace(string(out))

	return project, branch
}
