package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestProviderCheckRepoStateCollectsOutgoingCommits(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	root := t.TempDir()
	remotePath := filepath.Join(root, "remote.git")
	repoPath := filepath.Join(root, "repo")

	if err := exec.Command("git", "init", "--bare", remotePath).Run(); err != nil {
		t.Fatalf("git init --bare: %v", err)
	}
	if err := exec.Command("git", "init", repoPath).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "config", "user.name", "test").Run(); err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "config", "user.email", "test@example.com").Run(); err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("one\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "add", "README.md").Run(); err != nil {
		t.Fatalf("git add initial: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "commit", "-m", "initial").Run(); err != nil {
		t.Fatalf("git commit initial: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "branch", "-M", "main").Run(); err != nil {
		t.Fatalf("git branch -M main: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "remote", "add", "origin", remotePath).Run(); err != nil {
		t.Fatalf("git remote add origin: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "push", "-u", "origin", "main").Run(); err != nil {
		t.Fatalf("git push origin main: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatalf("update README: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "add", "README.md").Run(); err != nil {
		t.Fatalf("git add outgoing: %v", err)
	}
	if err := exec.Command("git", "-C", repoPath, "commit", "-m", "local change").Run(); err != nil {
		t.Fatalf("git commit outgoing: %v", err)
	}

	state, warnings := New().CheckRepoState(repoPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if len(state.RemoteStatus) != 1 {
		t.Fatalf("expected one remote status, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
	}

	if state.RemoteStatus[0].Remote != "origin" {
		t.Fatalf("expected origin remote, got %q", state.RemoteStatus[0].Remote)
	}

	if state.RemoteStatus[0].Ahead != 1 {
		t.Fatalf("expected ahead count 1, got %d", state.RemoteStatus[0].Ahead)
	}

	if len(state.RemoteStatus[0].OutgoingCommits) != 1 {
		t.Fatalf("expected 1 outgoing commit, got %d: %v", len(state.RemoteStatus[0].OutgoingCommits), state.RemoteStatus[0].OutgoingCommits)
	}

	if !strings.Contains(state.RemoteStatus[0].OutgoingCommits[0], "local change") {
		t.Fatalf("expected outgoing commit summary to include commit message, got %v", state.RemoteStatus[0].OutgoingCommits)
	}
}
