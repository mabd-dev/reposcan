package gitx

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func gitOrSkip(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir,
		"-c", "user.email=test@example.com",
		"-c", "user.name=test",
		"-c", "commit.gpgsign=false",
	}, args...)
	cmd := exec.Command("git", full...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestGetStashes_NoStashes(t *testing.T) {
	gitOrSkip(t)
	repo := t.TempDir()
	runGit(t, repo, "init")

	stashes, err := GetStashes(repo)
	if err != nil {
		t.Fatalf("GetStashes: %v", err)
	}
	if len(stashes) != 0 {
		t.Fatalf("expected 0 stashes, got %d: %v", len(stashes), stashes)
	}
}

// A stash lives in the shared refs/stash of the common dir, so a stash made in
// the main worktree must also be visible from a linked worktree.
func TestGetStashes_SharedAcrossWorktrees(t *testing.T) {
	gitOrSkip(t)
	base := t.TempDir()
	repo := filepath.Join(base, "main")
	runGit(t, base, "init", "main")

	file := filepath.Join(repo, "f.txt")
	if err := os.WriteFile(file, []byte("v1\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	runGit(t, repo, "add", "f.txt")
	runGit(t, repo, "commit", "-m", "initial")

	// linked worktree on a new branch
	wt := filepath.Join(base, "wt")
	runGit(t, repo, "worktree", "add", "-b", "feature", wt)

	// create a stash in the main worktree
	if err := os.WriteFile(file, []byte("v2\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	runGit(t, repo, "stash")

	for _, path := range []string{repo, wt} {
		stashes, err := GetStashes(path)
		if err != nil {
			t.Fatalf("GetStashes(%s): %v", path, err)
		}
		if len(stashes) != 1 {
			t.Fatalf("expected 1 stash at %s, got %d: %v", path, len(stashes), stashes)
		}
	}
}
