package git

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mabd-dev/reposcan/internal/vcs"
)

func TestProviderCheckRepoStateSetsVCSType(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	root := t.TempDir()
	repoPath := filepath.Join(root, "repo")

	if err := exec.Command("git", "init", repoPath).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	state, _ := New().CheckRepoState(repoPath)

	if state.Path != repoPath {
		t.Fatalf("expected path %s, got %s", repoPath, state.Path)
	}

	if state.Repo != "repo" {
		t.Fatalf("expected repo name %q, got %q", "repo", state.Repo)
	}

	if state.ID == "" {
		t.Fatalf("expected non-empty repo id")
	}

	if state.VCSType != string(vcs.TypeGit) {
		t.Fatalf("expected vcs type %q, got %q", vcs.TypeGit, state.VCSType)
	}
}
