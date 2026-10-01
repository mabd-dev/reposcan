//go:build !windows

package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCheckRepoStateWithActivity_RunsStatusOnce guards against dating a clean
// repo with a second `git status` call. A shim git on PATH logs each
// invocation's arguments before running the real git.
func TestCheckRepoStateWithActivity_RunsStatusOnce(t *testing.T) {
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not on PATH")
	}

	repo := t.TempDir()
	commitAt := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	runGit(t, repo, "init", "-q")
	writeFileAt(t, filepath.Join(repo, "a.txt"), "a", commitAt)
	runGitAt(t, repo, commitAt, "add", ".")
	runGitAt(t, repo, commitAt, "commit", "-m", "init")

	shimDir := t.TempDir()
	logPath := filepath.Join(shimDir, "calls.log")
	shim := fmt.Sprintf("#!/bin/sh\necho \"$*\" >> '%s'\nexec '%s' \"$@\"\n", logPath, realGit)
	if err := os.WriteFile(filepath.Join(shimDir, "git"), []byte(shim), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	state, _ := New().CheckRepoStateWithActivity(repo)
	if !state.LastActivity.Equal(commitAt) {
		t.Fatalf("LastActivity = %v, want %v", state.LastActivity, commitAt)
	}

	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	statusCalls := 0
	for _, line := range strings.Split(string(log), "\n") {
		if strings.Contains(line, " status ") {
			statusCalls++
		}
	}
	if statusCalls != 1 {
		t.Fatalf("git status ran %d times, want 1:\n%s", statusCalls, log)
	}
}
