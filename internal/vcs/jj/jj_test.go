package jj

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/vcs"
)

func initJJRepo(t *testing.T, root string, name string) string {
	t.Helper()

	repoPath := filepath.Join(root, name)
	if err := exec.Command("jj", "git", "init", repoPath).Run(); err != nil {
		t.Fatalf("jj git init: %v", err)
	}

	return repoPath
}

func initTrackedJJRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	remotePath := filepath.Join(root, "remote.git")
	seedPath := filepath.Join(root, "seed")
	workPath := filepath.Join(root, "work")

	if err := exec.Command("git", "init", "--bare", remotePath).Run(); err != nil {
		t.Fatalf("git init --bare: %v", err)
	}
	if err := exec.Command("git", "clone", remotePath, seedPath).Run(); err != nil {
		t.Fatalf("git clone: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "config", "user.name", "test").Run(); err != nil {
		t.Fatalf("git config user.name: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "config", "user.email", "test@example.com").Run(); err != nil {
		t.Fatalf("git config user.email: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seedPath, "README.md"), []byte("one\n"), 0o644); err != nil {
		t.Fatalf("write seed README: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "add", "README.md").Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "commit", "-m", "initial").Run(); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "branch", "-M", "main").Run(); err != nil {
		t.Fatalf("git branch -M main: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "push", "origin", "main").Run(); err != nil {
		t.Fatalf("git push origin main: %v", err)
	}
	if err := exec.Command("jj", "git", "clone", remotePath, workPath).Run(); err != nil {
		t.Fatalf("jj git clone: %v", err)
	}

	return workPath
}

func TestProviderCheckRepoStateHandlesMissingRemotesAndBookmarks(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}

	root := t.TempDir()
	repoPath := initJJRepo(t, root, "repo")

	state, warnings := New().CheckRepoState(repoPath)

	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if state.Path != repoPath {
		t.Fatalf("expected path %s, got %s", repoPath, state.Path)
	}

	if state.Repo != "repo" {
		t.Fatalf("expected repo name %q, got %q", "repo", state.Repo)
	}

	if state.ID == "" {
		t.Fatalf("expected non-empty repo id")
	}

	if state.VCSType != string(vcs.TypeJJ) {
		t.Fatalf("expected vcs type %q, got %q", vcs.TypeJJ, state.VCSType)
	}

	if strings.TrimSpace(state.Branch) == "" || state.Branch == "-" {
		t.Fatalf("expected branch display to fall back to a change id, got %q", state.Branch)
	}

	if len(state.RemoteStatus) != 1 {
		t.Fatalf("expected one jj remote status entry, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
	}

	if state.RemoteStatus[0].Ahead != 0 || state.RemoteStatus[0].Behind != 0 {
		t.Fatalf(
			"expected jj ahead/behind defaults to be 0/0, got %d/%d",
			state.RemoteStatus[0].Ahead,
			state.RemoteStatus[0].Behind,
		)
	}

	if len(state.OutgoingCommits) != 0 {
		t.Fatalf("expected no outgoing commits, got %v", state.OutgoingCommits)
	}
}

func TestProviderCheckRepoStateCollectsUncommittedFiles(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}

	root := t.TempDir()
	repoPath := initJJRepo(t, root, "dirty")

	filePath := filepath.Join(repoPath, "hello.txt")
	if err := os.WriteFile(filePath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}

	state, warnings := New().CheckRepoState(repoPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if len(state.UncommitedFiles) != 1 {
		t.Fatalf("expected 1 uncommitted file, got %d: %v", len(state.UncommitedFiles), state.UncommitedFiles)
	}

	if !strings.Contains(state.UncommitedFiles[0], "hello.txt") {
		t.Fatalf("expected summary to mention hello.txt, got %v", state.UncommitedFiles)
	}
}

func TestProviderCheckRepoStateCollectsTrackedBookmarkOutgoingCommits(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	repoPath := initTrackedJJRepo(t)

	filePath := filepath.Join(repoPath, "README.md")
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open repo file: %v", err)
	}
	if _, err := f.WriteString("two\n"); err != nil {
		t.Fatalf("append change 1: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("jj", "-R", repoPath, "describe", "-m", "change 1").Run(); err != nil {
		t.Fatalf("jj describe change 1: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "bookmark", "move", "main", "-t", "@").Run(); err != nil {
		t.Fatalf("jj bookmark move main: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "new").Run(); err != nil {
		t.Fatalf("jj new: %v", err)
	}

	f, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open repo file for change 2: %v", err)
	}
	if _, err := f.WriteString("three\n"); err != nil {
		t.Fatalf("append change 2: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("jj", "-R", repoPath, "describe", "-m", "change 2").Run(); err != nil {
		t.Fatalf("jj describe change 2: %v", err)
	}

	state, warnings := New().CheckRepoState(repoPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if len(state.RemoteStatus) != 1 {
		t.Fatalf("expected one jj remote status entry, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
	}

	if state.RemoteStatus[0].Ahead != 1 {
		t.Fatalf("expected ahead count 1 from tracked bookmark commits, got %d", state.RemoteStatus[0].Ahead)
	}

	if len(state.OutgoingCommits) != 1 {
		t.Fatalf("expected exactly 1 outgoing commit, got %d: %v", len(state.OutgoingCommits), state.OutgoingCommits)
	}

	if !strings.Contains(state.OutgoingCommits[0], "change 1") {
		t.Fatalf("expected outgoing commit list to mention tracked bookmark commit, got %v", state.OutgoingCommits)
	}

	if strings.Contains(state.OutgoingCommits[0], "change 2") {
		t.Fatalf("did not expect working-copy descendant to be treated as tracked-bookmark outgoing commit, got %v", state.OutgoingCommits)
	}
}

func TestProviderCheckRepoStateWarnsWhenBinaryMissing(t *testing.T) {
	repoPath := filepath.Join(t.TempDir(), "repo")

	state, warnings := (&Provider{binary: "jj-does-not-exist"}).CheckRepoState(repoPath)

	if state.Path != repoPath {
		t.Fatalf("expected path %s, got %s", repoPath, state.Path)
	}

	if state.VCSType != string(vcs.TypeJJ) {
		t.Fatalf("expected vcs type %q, got %q", vcs.TypeJJ, state.VCSType)
	}

	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}

	if !strings.Contains(warnings[0], "Failed to inspect jj repo") {
		t.Fatalf("expected missing binary warning, got %v", warnings)
	}
}
