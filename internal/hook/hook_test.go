package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.MkdirAll(filepath.Join(gitDir, "hooks"), 0o755)
	return gitDir
}

func TestInstallNewHook(t *testing.T) {
	gitDir := setupGitDir(t)
	err := Install(gitDir, []string{"post-commit"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(gitDir, "hooks", "post-commit"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.HasPrefix(content, shebang) {
		t.Error("missing shebang")
	}
	if !strings.Contains(content, markerStart) {
		t.Error("missing start marker")
	}
	if !strings.Contains(content, markerEnd) {
		t.Error("missing end marker")
	}
	if !strings.Contains(content, "Just committed") {
		t.Error("missing default template")
	}

	// Verify executable
	info, _ := os.Stat(filepath.Join(gitDir, "hooks", "post-commit"))
	if info.Mode()&0o111 == 0 {
		t.Error("hook not executable")
	}
}

func TestInstallAppendsToExisting(t *testing.T) {
	gitDir := setupGitDir(t)
	hookPath := filepath.Join(gitDir, "hooks", "post-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\necho existing\n"), 0o755)

	err := Install(gitDir, []string{"post-commit"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(hookPath)
	content := string(data)

	if !strings.Contains(content, "echo existing") {
		t.Error("existing content was overwritten")
	}
	if !strings.Contains(content, markerStart) {
		t.Error("nownow block not appended")
	}
}

func TestInstallWithCustomTemplate(t *testing.T) {
	gitDir := setupGitDir(t)
	templates := map[string]string{
		"post-commit": `nownow push "deployed {commit_msg}"`,
	}
	err := Install(gitDir, []string{"post-commit"}, templates)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(gitDir, "hooks", "post-commit"))
	if !strings.Contains(string(data), "deployed {commit_msg}") {
		t.Error("custom template not used")
	}
}

func TestInstallReplacesExistingBlock(t *testing.T) {
	gitDir := setupGitDir(t)
	// Install twice, should not duplicate
	Install(gitDir, []string{"post-commit"}, nil)
	Install(gitDir, []string{"post-commit"}, nil)

	data, _ := os.ReadFile(filepath.Join(gitDir, "hooks", "post-commit"))
	if strings.Count(string(data), markerStart) != 1 {
		t.Error("duplicate nownow blocks found")
	}
}

func TestRemove(t *testing.T) {
	gitDir := setupGitDir(t)
	Install(gitDir, []string{"post-commit"}, nil)

	err := Remove(gitDir)
	if err != nil {
		t.Fatal(err)
	}

	// File should be deleted (only had shebang + nownow block)
	_, err = os.Stat(filepath.Join(gitDir, "hooks", "post-commit"))
	if !os.IsNotExist(err) {
		t.Error("hook file should have been deleted")
	}
}

func TestRemovePreservesOtherContent(t *testing.T) {
	gitDir := setupGitDir(t)
	hookPath := filepath.Join(gitDir, "hooks", "post-commit")
	os.WriteFile(hookPath, []byte("#!/bin/sh\necho existing\n"), 0o755)
	Install(gitDir, []string{"post-commit"}, nil)

	Remove(gitDir)

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal("hook file should still exist")
	}
	content := string(data)
	if strings.Contains(content, markerStart) {
		t.Error("nownow block should be removed")
	}
	if !strings.Contains(content, "echo existing") {
		t.Error("existing content should be preserved")
	}
}

func TestList(t *testing.T) {
	gitDir := setupGitDir(t)
	Install(gitDir, []string{"post-commit", "pre-push"}, nil)

	hooks, err := List(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(hooks))
	}
}

func TestListEmpty(t *testing.T) {
	gitDir := setupGitDir(t)
	hooks, err := List(gitDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 0 {
		t.Fatalf("expected 0 hooks, got %d", len(hooks))
	}
}

func TestFindGitDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.Mkdir(gitDir, 0o755)

	sub := filepath.Join(dir, "a", "b")
	os.MkdirAll(sub, 0o755)

	found, err := FindGitDir(sub)
	if err != nil {
		t.Fatal(err)
	}
	if found != gitDir {
		t.Errorf("expected %s, got %s", gitDir, found)
	}
}

func TestFindGitDirNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindGitDir(dir)
	if err == nil {
		t.Error("expected error for non-git dir")
	}
}

func TestBuildTemplate(t *testing.T) {
	result := BuildTemplate("post-commit", "deployed {commit_msg} to {branch}")
	if !strings.Contains(result, "git log -1") {
		t.Error("commit_msg not expanded")
	}
	if !strings.Contains(result, "git rev-parse") {
		t.Error("branch not expanded")
	}
}

func TestFindGitDirWorktree(t *testing.T) {
	dir := t.TempDir()
	// Simulate a worktree: .git is a file with "gitdir: <path>"
	realGitDir := filepath.Join(dir, "real-gitdir")
	os.MkdirAll(realGitDir, 0o755)

	worktree := filepath.Join(dir, "worktree")
	os.MkdirAll(worktree, 0o755)
	os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: "+realGitDir+"\n"), 0o644)

	found, err := FindGitDir(worktree)
	if err != nil {
		t.Fatal(err)
	}
	if found != realGitDir {
		t.Errorf("expected %s, got %s", realGitDir, found)
	}
}

func TestFindGitDirWorktreeRelative(t *testing.T) {
	dir := t.TempDir()
	realGitDir := filepath.Join(dir, ".real-git")
	os.MkdirAll(realGitDir, 0o755)

	// Relative gitdir path
	os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: .real-git\n"), 0o644)

	found, err := FindGitDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if found != realGitDir {
		t.Errorf("expected %s, got %s", realGitDir, found)
	}
}

func TestValidateHookNameRejectsTraversal(t *testing.T) {
	bad := []string{"../../etc/cron", "../foo", "a/b", "a\\b", "/abs"}
	for _, name := range bad {
		err := validateHookName(name)
		if err == nil {
			t.Errorf("expected error for hook name %q", name)
		}
	}
}

func TestValidateHookNameAcceptsValid(t *testing.T) {
	good := []string{"post-commit", "pre-push", "commit-msg", "pre-rebase"}
	for _, name := range good {
		if err := validateHookName(name); err != nil {
			t.Errorf("unexpected error for %q: %v", name, err)
		}
	}
}
