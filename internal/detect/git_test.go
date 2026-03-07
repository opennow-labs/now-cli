package detect

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDetectGitInRepo(t *testing.T) {
	// Create a temp git repo
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "myproject")
	os.MkdirAll(repoDir, 0755)

	// Initialize git repo
	cmds := [][]string{
		{"git", "init"},
		{"git", "checkout", "-b", "feat/test-branch"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %s %v", args, string(out), err)
		}
	}

	// Change to repo dir
	origDir, _ := os.Getwd()
	os.Chdir(repoDir)
	defer os.Chdir(origDir)

	project, branch := detectGit()

	if project != "myproject" {
		t.Errorf("project = %q, want %q", project, "myproject")
	}
	if branch != "feat/test-branch" {
		t.Errorf("branch = %q, want %q", branch, "feat/test-branch")
	}
}

func TestDetectGitOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	project, branch := detectGit()

	if project != "" {
		t.Errorf("project should be empty outside git repo, got %q", project)
	}
	if branch != "" {
		t.Errorf("branch should be empty outside git repo, got %q", branch)
	}
}
