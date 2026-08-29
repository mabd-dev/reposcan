//go:build !windows

// This test deterministically replays the race documented in state.go: the
// remote-tracking ref is deleted by an external actor between the status
// check and the outgoing-commit lookup. It works by putting a shim git on
// PATH that removes refs/remotes/origin/main right before the git log
// invocation, then runs the real git log, which exits 128.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeShimGit(t *testing.T, realGit, dir string) string {
	t.Helper()
	shimDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(shimDir, 0o755); err != nil {
		t.Fatalf("mkdir shim dir: %v", err)
	}
	// Use single-quoted shell string for the real git path.
	shim := fmt.Sprintf(`#!/bin/sh
REAL_GIT='%s'
dir=""
prev=""
for a in "$@"; do
	if [ "$prev" = "-C" ]; then
		dir="$a"
	fi
	prev="$a"
done
trig=0
for a in "$@"; do
	if [ "$a" = "--format=%%h %%s" ]; then
		trig=1
	fi
done
if [ "$trig" = "1" ] && [ -n "$dir" ]; then
	"$REAL_GIT" -C "$dir" update-ref -d refs/remotes/origin/main
fi
exec "$REAL_GIT" "$@"
`, realGit)
	shimPath := filepath.Join(shimDir, "git")
	if err := os.WriteFile(shimPath, []byte(shim), 0o755); err != nil {
		t.Fatalf("write shim: %v", err)
	}
	return shimDir
}

func TestCheckRepoState_OutgoingCommitsRace(t *testing.T) {
	gitOrSkip(t)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("lookpath git: %v", err)
	}

	root := t.TempDir()
	bare := initBareMain(t, "remote.git")

	seed := filepath.Join(root, "seed")
	runGit(t, root, "init", "seed")
	runGit(t, seed, "commit", "--allow-empty", "-m", "base")
	runGit(t, seed, "branch", "-M", "main")
	runGit(t, seed, "remote", "add", "origin", bare)
	runGit(t, seed, "push", "-u", "origin", "main")

	local := filepath.Join(root, "local")
	runGit(t, root, "clone", bare, local)
	if err := os.WriteFile(filepath.Join(local, "extra.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	runGit(t, local, "add", "extra.txt")
	runGit(t, local, "commit", "-m", "extra")

	// Baseline without the shim: outgoing commit present, no warning.
	state, warnings := CheckRepoState(local)
	if len(state.RemoteStatus) != 1 || len(state.RemoteStatus[0].OutgoingCommits) != 1 {
		t.Fatalf("baseline: expected 1 outgoing commit, got RemoteStatus %+v", state.RemoteStatus)
	}
	for _, w := range warnings {
		if strings.Contains(w, "outgoing commits") {
			t.Fatalf("baseline: unexpected warning: %s", w)
		}
	}

	// Replay the race: shim git on PATH deletes the tracking ref right
	// before the git log lookup. The upstream status check still succeeds
	// (ref present), so ahead/behind survive; only OutgoingCommits fail.
	shimDir := writeShimGit(t, realGit, root)
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	state, warnings = CheckRepoState(local)
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "Failed to get outgoing commits for remote=origin") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected race warning, got warnings: %v", warnings)
	}
	if len(state.RemoteStatus) != 1 {
		t.Fatalf("expected 1 remote status entry, got %+v", state.RemoteStatus)
	}
	rs := state.RemoteStatus[0]
	if rs.Ahead != 1 || rs.Behind != 0 {
		t.Fatalf("expected ahead=1 behind=0 preserved from status check, got ahead=%d behind=%d", rs.Ahead, rs.Behind)
	}
	if len(rs.OutgoingCommits) != 0 {
		t.Fatalf("expected no outgoing commits after disrupted lookup, got %v", rs.OutgoingCommits)
	}
}
