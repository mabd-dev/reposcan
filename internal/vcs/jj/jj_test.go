package jj

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func initJJRepo(t *testing.T, root string, name string) string {
	t.Helper()

	repoPath := filepath.Join(root, name)
	if err := exec.Command("jj", "git", "init", repoPath).Run(); err != nil {
		t.Fatalf("jj git init: %v", err)
	}

	return repoPath
}

type trackedJJRepo struct {
	SeedPath string
	WorkPath string
}

// initTrackedJJRepo clones a one-commit remote with jj. cloneArgs are passed to
// jj git clone, e.g. --colocate.
func initTrackedJJRepo(t *testing.T, cloneArgs ...string) trackedJJRepo {
	t.Helper()

	root := t.TempDir()
	remotePath := filepath.Join(root, "remote.git")
	seedPath := filepath.Join(root, "seed")
	workPath := filepath.Join(root, "work")

	if err := exec.Command("git", "init", "--bare", "--initial-branch=main", remotePath).Run(); err != nil {
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
	cloneCmd := append([]string{"git", "clone"}, cloneArgs...)
	if err := exec.Command("jj", append(cloneCmd, remotePath, workPath)...).Run(); err != nil {
		t.Fatalf("jj git clone: %v", err)
	}

	return trackedJJRepo{
		SeedPath: seedPath,
		WorkPath: workPath,
	}
}

// runCommand runs a fixture command and fails the test with its output.
func runCommand(t *testing.T, name string, args ...string) {
	t.Helper()
	if output, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
}

func TestJJProviderImplementsVcsProvider(t *testing.T) {
	var _ vcs.Provider = (*Provider)(nil)
}

func TestJJProviderImplementsVcsActionProvider(t *testing.T) {
	var _ vcs.ActionProvider = (*Provider)(nil)
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

	// Like git, a repo without a remote reports -1/-1, rendered as "x".
	if state.RemoteStatus[0].Ahead != -1 || state.RemoteStatus[0].Behind != -1 {
		t.Fatalf(
			"expected jj ahead/behind without a remote to be -1/-1, got %d/%d",
			state.RemoteStatus[0].Ahead,
			state.RemoteStatus[0].Behind,
		)
	}

	if len(state.RemoteStatus[0].OutgoingCommits) != 0 {
		t.Fatalf("expected no outgoing commits, got %v", state.RemoteStatus[0].OutgoingCommits)
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

	repoPath := initTrackedJJRepo(t).WorkPath

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

	if state.RemoteStatus[0].Remote != "origin" {
		t.Fatalf("expected remote status to be scoped to origin, got %q", state.RemoteStatus[0].Remote)
	}

	if state.Branch != "main" {
		t.Fatalf("expected branch display to use the current bookmark, got %q", state.Branch)
	}

	if len(state.RemoteStatus[0].OutgoingCommits) != 1 {
		t.Fatalf("expected exactly 1 outgoing commit, got %d: %v", len(state.RemoteStatus[0].OutgoingCommits), state.RemoteStatus[0].OutgoingCommits)
	}

	if !strings.Contains(state.RemoteStatus[0].OutgoingCommits[0], "change 1") {
		t.Fatalf("expected outgoing commit list to mention tracked bookmark commit, got %v", state.RemoteStatus[0].OutgoingCommits)
	}

	if strings.Contains(state.RemoteStatus[0].OutgoingCommits[0], "change 2") {
		t.Fatalf("did not expect working-copy descendant to be treated as tracked-bookmark outgoing commit, got %v", state.RemoteStatus[0].OutgoingCommits)
	}
}

// TestProviderCheckRepoStateIgnoresBookmarksContainingTheName guards against
// substring bookmark matching, where "main" would also match "maintenance".
func TestProviderCheckRepoStateIgnoresBookmarksContainingTheName(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	repoPath := initTrackedJJRepo(t).WorkPath

	if err := exec.Command("jj", "-R", repoPath, "new", "main", "-m", "maintenance work").Run(); err != nil {
		t.Fatalf("jj new main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoPath, "maintenance.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatalf("write maintenance file: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "bookmark", "create", "maintenance", "-r", "@").Run(); err != nil {
		t.Fatalf("jj bookmark create maintenance: %v", err)
	}
	if err := exec.Command("jj", "-R", repoPath, "new", "main").Run(); err != nil {
		t.Fatalf("jj new main: %v", err)
	}

	state, warnings := New().CheckRepoState(repoPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if state.Branch != "main" {
		t.Fatalf("expected branch display to use main, got %q", state.Branch)
	}

	if len(state.RemoteStatus) != 1 {
		t.Fatalf("expected one jj remote status entry, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
	}

	status := state.RemoteStatus[0]
	if status.Ahead != 0 || len(status.OutgoingCommits) != 0 {
		t.Fatalf("expected no outgoing commits on main, got ahead %d with %v", status.Ahead, status.OutgoingCommits)
	}
}

func TestProviderCheckRepoStateCollectsTrackedBookmarkIncomingCommits(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	repo := initTrackedJJRepo(t)
	seedPath := repo.SeedPath
	workPath := repo.WorkPath

	workFile := filepath.Join(workPath, "README.md")
	f, err := os.OpenFile(workFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open work file: %v", err)
	}
	if _, err := f.WriteString("local\n"); err != nil {
		t.Fatalf("append local change: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("jj", "-R", workPath, "describe", "-m", "local change").Run(); err != nil {
		t.Fatalf("jj describe local change: %v", err)
	}
	if err := exec.Command("jj", "-R", workPath, "bookmark", "move", "main", "-t", "@").Run(); err != nil {
		t.Fatalf("jj bookmark move main: %v", err)
	}
	if err := exec.Command("jj", "-R", workPath, "new").Run(); err != nil {
		t.Fatalf("jj new: %v", err)
	}

	seedFile := filepath.Join(seedPath, "README.md")
	f, err = os.OpenFile(seedFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open seed file: %v", err)
	}
	if _, err := f.WriteString("remote\n"); err != nil {
		t.Fatalf("append remote change: %v", err)
	}
	_ = f.Close()

	if err := exec.Command("git", "-C", seedPath, "add", "README.md").Run(); err != nil {
		t.Fatalf("git add remote: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "commit", "-m", "remote change").Run(); err != nil {
		t.Fatalf("git commit remote: %v", err)
	}
	if err := exec.Command("git", "-C", seedPath, "push", "origin", "main").Run(); err != nil {
		t.Fatalf("git push remote: %v", err)
	}
	if err := exec.Command("jj", "-R", workPath, "git", "fetch").Run(); err != nil {
		t.Fatalf("jj git fetch: %v", err)
	}

	state, warnings := New().CheckRepoState(workPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if len(state.RemoteStatus) != 1 {
		t.Fatalf("expected one jj remote status entry, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
	}

	if state.RemoteStatus[0].Behind != 1 {
		t.Fatalf("expected behind count 1 from tracked bookmark incoming commits, got %d", state.RemoteStatus[0].Behind)
	}

	if state.RemoteStatus[0].Remote != "origin" {
		t.Fatalf("expected remote status to be scoped to origin, got %q", state.RemoteStatus[0].Remote)
	}

	if state.Branch != "main" {
		t.Fatalf("expected branch display to use the current bookmark, got %q", state.Branch)
	}

	if state.RemoteStatus[0].Ahead != 1 {
		t.Fatalf("expected ahead count 1 from local tracked bookmark commit, got %d", state.RemoteStatus[0].Ahead)
	}
}

// TestProviderCheckRepoStateReportsInSyncCloneAsZero guards against counting
// the whole history as incoming when jj has no @git ref for the bookmark.
func TestProviderCheckRepoStateReportsInSyncCloneAsZero(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	tests := []struct {
		name     string
		colocate string
	}{
		{name: "non-colocated", colocate: "--no-colocate"},
		{name: "colocated", colocate: "--colocate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := initTrackedJJRepo(t, tt.colocate).WorkPath

			state, warnings := New().CheckRepoState(repoPath)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}

			if len(state.RemoteStatus) != 1 {
				t.Fatalf("expected one jj remote status entry, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
			}

			status := state.RemoteStatus[0]
			if status.Remote != "origin" || status.Ahead != 0 || status.Behind != 0 {
				t.Fatalf("expected origin 0/0 for an in-sync clone, got %q %d/%d", status.Remote, status.Ahead, status.Behind)
			}
		})
	}
}

// TestProviderCheckRepoStateCountsFromWorkingCopyParent checks that ahead and
// behind match git's HEAD-based counts with @- as HEAD: a commit counts even if
// no bookmark points at it, and edits in @ only show as uncommitted files.
func TestProviderCheckRepoStateCountsFromWorkingCopyParent(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	tests := []struct {
		name            string
		commit          bool
		wantAhead       int
		wantUncommitted int
	}{
		{name: "edits in the working copy", wantUncommitted: 1},
		{name: "commit without moving the bookmark", commit: true, wantAhead: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workPath := initTrackedJJRepo(t).WorkPath
			if err := os.WriteFile(filepath.Join(workPath, "local.txt"), []byte("local\n"), 0o644); err != nil {
				t.Fatalf("write local file: %v", err)
			}
			if tt.commit {
				runCommand(t, "jj", "-R", workPath, "commit", "-m", "local change")
			}

			state, warnings := New().CheckRepoState(workPath)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}

			if state.Branch != "main" {
				t.Fatalf("expected branch display main, got %q", state.Branch)
			}
			if len(state.UncommitedFiles) != tt.wantUncommitted {
				t.Fatalf("expected %d uncommitted files, got %v", tt.wantUncommitted, state.UncommitedFiles)
			}
			if len(state.RemoteStatus) != 1 {
				t.Fatalf("expected one jj remote status entry, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
			}

			status := state.RemoteStatus[0]
			if status.Remote != "origin" || status.Ahead != tt.wantAhead || status.Behind != 0 {
				t.Fatalf("expected origin %d/0, got %q %d/%d", tt.wantAhead, status.Remote, status.Ahead, status.Behind)
			}
		})
	}
}

// TestProviderCheckRepoStateReportsBookmarkMissingOnRemote checks that, like a
// git branch that was never pushed, a local-only bookmark reports -1/-1 for the
// remote instead of looking in sync.
func TestProviderCheckRepoStateReportsBookmarkMissingOnRemote(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	workPath := initTrackedJJRepo(t).WorkPath
	if err := os.WriteFile(filepath.Join(workPath, "feat.txt"), []byte("feat\n"), 0o644); err != nil {
		t.Fatalf("write feature file: %v", err)
	}
	runCommand(t, "jj", "-R", workPath, "commit", "-m", "feature work")
	runCommand(t, "jj", "-R", workPath, "bookmark", "create", "feat", "-r", "@-")

	state, warnings := New().CheckRepoState(workPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if state.Branch != "feat" {
		t.Fatalf("expected branch display feat, got %q", state.Branch)
	}

	want := []report.RemoteStatus{{Remote: "origin", Ahead: -1, Behind: -1}}
	if !reflect.DeepEqual(state.RemoteStatus, want) {
		t.Fatalf("expected remote status %v, got %v", want, state.RemoteStatus)
	}
}

// TestProviderCheckRepoStateReportsDeletedRemoteBookmark checks that a bookmark
// deleted on the remote reports -1/-1, like a git branch whose remote branch
// is gone, instead of counting the whole history as ahead.
func TestProviderCheckRepoStateReportsDeletedRemoteBookmark(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	repo := initTrackedJJRepo(t)
	if err := os.WriteFile(filepath.Join(repo.WorkPath, "local.txt"), []byte("local\n"), 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}
	runCommand(t, "jj", "-R", repo.WorkPath, "commit", "-m", "local change")
	runCommand(t, "jj", "-R", repo.WorkPath, "bookmark", "set", "main", "-r", "@-")
	// A bare remote refuses to delete its HEAD branch, so point HEAD elsewhere.
	remoteURL, err := exec.Command("git", "-C", repo.SeedPath, "remote", "get-url", "origin").Output()
	if err != nil {
		t.Fatalf("git remote get-url origin: %v", err)
	}
	runCommand(t, "git", "-C", repo.SeedPath, "push", "origin", "main:other")
	runCommand(t, "git", "--git-dir", strings.TrimSpace(string(remoteURL)), "symbolic-ref", "HEAD", "refs/heads/other")
	runCommand(t, "git", "-C", repo.SeedPath, "push", "origin", "--delete", "main")
	runCommand(t, "jj", "-R", repo.WorkPath, "git", "fetch")

	state, warnings := New().CheckRepoState(repo.WorkPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if state.Branch != "main" {
		t.Fatalf("expected branch display main, got %q", state.Branch)
	}

	want := []report.RemoteStatus{{Remote: "origin", Ahead: -1, Behind: -1}}
	if !reflect.DeepEqual(state.RemoteStatus, want) {
		t.Fatalf("expected remote status %v, got %v", want, state.RemoteStatus)
	}
}

// TestProviderCheckRepoStateCountsIncomingPerRemoteForMultiTargetConflict
// guards against one remote's conflict target inflating another remote's
// behind count. Fetching two diverged remotes leaves main with targets
// {L, R, U}. origin is behind only by R because X is already in L.
func TestProviderCheckRepoStateCountsIncomingPerRemoteForMultiTargetConflict(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	appendLine := func(path string, line string) {
		t.Helper()
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatalf("open %s: %v", path, err)
		}
		defer f.Close()
		if _, err := f.WriteString(line + "\n"); err != nil {
			t.Fatalf("append to %s: %v", path, err)
		}
	}

	root := t.TempDir()
	originPath := filepath.Join(root, "origin.git")
	upstreamPath := filepath.Join(root, "upstream.git")
	seedPath := filepath.Join(root, "seed")
	workPath := filepath.Join(root, "work")
	seedFile := filepath.Join(seedPath, "README.md")

	runCommand(t, "git", "init", "--bare", "--initial-branch=main", originPath)
	runCommand(t, "git", "init", "--bare", "--initial-branch=main", upstreamPath)
	runCommand(t, "git", "clone", originPath, seedPath)
	runCommand(t, "git", "-C", seedPath, "config", "user.name", "test")
	runCommand(t, "git", "-C", seedPath, "config", "user.email", "test@example.com")
	runCommand(t, "git", "-C", seedPath, "remote", "add", "upstream", upstreamPath)

	// A is on both remotes' history; origin has advanced to X before the clone.
	appendLine(seedFile, "A")
	runCommand(t, "git", "-C", seedPath, "add", "README.md")
	runCommand(t, "git", "-C", seedPath, "commit", "-m", "A")
	runCommand(t, "git", "-C", seedPath, "branch", "-M", "main")
	runCommand(t, "git", "-C", seedPath, "push", "upstream", "main")
	appendLine(seedFile, "X")
	runCommand(t, "git", "-C", seedPath, "commit", "-am", "X")
	runCommand(t, "git", "-C", seedPath, "push", "origin", "main")

	runCommand(t, "jj", "git", "clone", "--no-colocate", originPath, workPath)
	runCommand(t, "jj", "-R", workPath, "git", "remote", "add", "upstream", upstreamPath)
	runCommand(t, "jj", "-R", workPath, "git", "fetch", "--remote", "upstream")
	runCommand(t, "jj", "-R", workPath, "bookmark", "track", "main@upstream")

	// Local main moves to L on top of X.
	runCommand(t, "jj", "-R", workPath, "new", "main")
	appendLine(filepath.Join(workPath, "README.md"), "L")
	runCommand(t, "jj", "-R", workPath, "describe", "-m", "L")
	runCommand(t, "jj", "-R", workPath, "bookmark", "move", "main", "-t", "@")
	runCommand(t, "jj", "-R", workPath, "new")

	// origin moves X -> R and upstream moves A -> U.
	appendLine(seedFile, "R")
	runCommand(t, "git", "-C", seedPath, "commit", "-am", "R")
	runCommand(t, "git", "-C", seedPath, "push", "origin", "main")
	runCommand(t, "git", "-C", seedPath, "checkout", "-b", "upstream-work", "HEAD~2")
	appendLine(filepath.Join(seedPath, "UPSTREAM.md"), "U")
	runCommand(t, "git", "-C", seedPath, "add", "UPSTREAM.md")
	runCommand(t, "git", "-C", seedPath, "commit", "-m", "U")
	runCommand(t, "git", "-C", seedPath, "push", "upstream", "upstream-work:main")

	runCommand(t, "jj", "-R", workPath, "git", "fetch", "--remote", "origin")
	runCommand(t, "jj", "-R", workPath, "git", "fetch", "--remote", "upstream")

	state, warnings := New().CheckRepoState(workPath)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	behind := map[string]int{}
	for _, status := range state.RemoteStatus {
		behind[status.Remote] = status.Behind
	}
	want := map[string]int{"origin": 1, "upstream": 1}
	if !reflect.DeepEqual(behind, want) {
		t.Fatalf("expected behind counts %v, got %v (remote status %v)", want, behind, state.RemoteStatus)
	}
}

// TestProviderCheckRepoStateFollowsBookmarkMovedPastWorkingCopy covers a fetch
// that fast-forwards main while @ stays on the old commit: reposcan should
// still show main and count the fetched commit as behind.
func TestProviderCheckRepoStateFollowsBookmarkMovedPastWorkingCopy(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	tests := []struct {
		name      string
		setup     func(t *testing.T, repo trackedJJRepo)
		wantAhead int
	}{
		{name: "only main"},
		{
			name: "local commit",
			setup: func(t *testing.T, repo trackedJJRepo) {
				workPath := repo.WorkPath
				if err := os.WriteFile(filepath.Join(workPath, "local.txt"), []byte("local\n"), 0o644); err != nil {
					t.Fatalf("write local file: %v", err)
				}
				runCommand(t, "jj", "-R", workPath, "commit", "-m", "local change")
			},
			wantAhead: 1,
		},
		{
			// feat has no remote, so it is not shown next to main.
			name: "local-only bookmark off the same commit",
			setup: func(t *testing.T, repo trackedJJRepo) {
				workPath := repo.WorkPath
				runCommand(t, "jj", "-R", workPath, "new", "main", "-m", "feature work")
				if err := os.WriteFile(filepath.Join(workPath, "feat.txt"), []byte("feat\n"), 0o644); err != nil {
					t.Fatalf("write feature file: %v", err)
				}
				runCommand(t, "jj", "-R", workPath, "bookmark", "create", "feat", "-r", "@")
				runCommand(t, "jj", "-R", workPath, "new", "main")
			},
		},
		{
			// feat branches off @'s own parent C, which is newer than the commit
			// shared with main; it must not hide main. C is unpushed, so it
			// counts as ahead.
			name: "local-only bookmark between working copy and main",
			setup: func(t *testing.T, repo trackedJJRepo) {
				workPath := repo.WorkPath
				runCommand(t, "jj", "-R", workPath, "new", "main", "-m", "C")
				runCommand(t, "jj", "-R", workPath, "new", "-m", "feature work")
				if err := os.WriteFile(filepath.Join(workPath, "feat.txt"), []byte("feat\n"), 0o644); err != nil {
					t.Fatalf("write feature file: %v", err)
				}
				runCommand(t, "jj", "-R", workPath, "bookmark", "create", "feat", "-r", "@")
				runCommand(t, "jj", "-R", workPath, "new", "feat-", "-m", "W")
			},
			wantAhead: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := initTrackedJJRepo(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			seedFile := filepath.Join(repo.SeedPath, "README.md")
			if err := os.WriteFile(seedFile, []byte("one\nremote\n"), 0o644); err != nil {
				t.Fatalf("write seed file: %v", err)
			}
			runCommand(t, "git", "-C", repo.SeedPath, "commit", "-am", "remote change")
			runCommand(t, "git", "-C", repo.SeedPath, "push", "origin", "main")
			runCommand(t, "jj", "-R", repo.WorkPath, "git", "fetch")

			state, warnings := New().CheckRepoState(repo.WorkPath)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}

			if state.Branch != "main" {
				t.Fatalf("expected branch display main, got %q", state.Branch)
			}

			if len(state.RemoteStatus) != 1 {
				t.Fatalf("expected one jj remote status entry, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
			}

			status := state.RemoteStatus[0]
			if status.Remote != "origin" || status.Ahead != tt.wantAhead || status.Behind != 1 {
				t.Fatalf("expected origin %d/1, got %q %d/%d", tt.wantAhead, status.Remote, status.Ahead, status.Behind)
			}
		})
	}
}

// TestProviderCheckRepoStateKeepsChangeIDWhenBookmarkIsUnclear checks that the
// moved-bookmark fallback does not guess: sharing only root() with main, or a
// fetch that leaves main and a tracked feature bookmark both past @, keeps the
// change ID because jj does not record which bookmark @ started from.
func TestProviderCheckRepoStateKeepsChangeIDWhenBookmarkIsUnclear(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	tests := []struct {
		name  string
		setup func(t *testing.T, repo trackedJJRepo)
	}{
		{
			name: "working copy off root",
			setup: func(t *testing.T, repo trackedJJRepo) {
				runCommand(t, "jj", "-R", repo.WorkPath, "new", "root()")
			},
		},
		{
			name: "main and a tracked feature bookmark past the working copy",
			setup: func(t *testing.T, repo trackedJJRepo) {
				runCommand(t, "git", "-C", repo.SeedPath, "checkout", "-b", "feature")
				if err := os.WriteFile(filepath.Join(repo.SeedPath, "feature.txt"), []byte("feature\n"), 0o644); err != nil {
					t.Fatalf("write feature file: %v", err)
				}
				runCommand(t, "git", "-C", repo.SeedPath, "add", "feature.txt")
				runCommand(t, "git", "-C", repo.SeedPath, "commit", "-m", "feature work")
				runCommand(t, "git", "-C", repo.SeedPath, "push", "origin", "feature")
				runCommand(t, "git", "-C", repo.SeedPath, "checkout", "main")
				if err := os.WriteFile(filepath.Join(repo.SeedPath, "README.md"), []byte("one\nremote\n"), 0o644); err != nil {
					t.Fatalf("write seed file: %v", err)
				}
				runCommand(t, "git", "-C", repo.SeedPath, "commit", "-am", "remote change")
				runCommand(t, "git", "-C", repo.SeedPath, "push", "origin", "main")
				runCommand(t, "jj", "-R", repo.WorkPath, "git", "fetch")
				runCommand(t, "jj", "-R", repo.WorkPath, "bookmark", "track", "feature@origin")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := initTrackedJJRepo(t)
			tt.setup(t, repo)

			state, warnings := New().CheckRepoState(repo.WorkPath)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}

			if strings.Contains(state.Branch, "main") || strings.Contains(state.Branch, "feature") {
				t.Fatalf("expected a change id, got %q", state.Branch)
			}
		})
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

// TestProviderCheckRepoStateWarningsIncludeCommandFailureDetails verifies that
// each inspection failure produces one contextual warning and does not prevent
// the remaining repository checks from completing.
func TestProviderCheckRepoStateWarningsIncludeCommandFailureDetails(t *testing.T) {
	tests := []struct {
		name          string
		failedCommand string
		wantOperation string
	}{
		{
			name:          "remotes",
			failedCommand: fakeJJCommandKey("git", "remote", "list"),
			wantOperation: "list remotes",
		},
		{
			name:          "branch display",
			failedCommand: branchCommandKey(),
			wantOperation: "get branch display",
		},
		{
			name:          "uncommitted files",
			failedCommand: fakeJJCommandKey("diff", "--summary"),
			wantOperation: "get uncommitted files",
		},
		{
			name:          "remote status",
			failedCommand: remoteBookmarksCommandKey(),
			wantOperation: "get remote status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := filepath.Join(t.TempDir(), "repo")
			responses := map[string]fakeJJResponse{
				fakeJJCommandKey("git", "remote", "list"): {Stdout: "origin https://example.com/acme/widgets.git\n"},
				branchCommandKey():                        {Stdout: "main*?|abc123\n"},
				fakeJJCommandKey("diff", "--summary"):     {},
				remoteBookmarksCommandKey():               {},
			}
			responses[tt.failedCommand] = fakeJJResponse{Stderr: "inspection failed", ExitCode: 1}
			binary := useFakeJJ(t, responses)

			_, warnings := (&Provider{binary: binary}).CheckRepoState(repoPath)
			if len(warnings) != 1 {
				t.Fatalf("CheckRepoState() warnings = %v, want exactly one", warnings)
			}

			wantParts := []string{
				"Failed to " + tt.wantOperation + " for jj repo",
				"path=" + repoPath,
				"command=",
				"inspection failed",
			}
			for _, want := range wantParts {
				if !strings.Contains(warnings[0], want) {
					t.Fatalf("warning = %q, want it to contain %q", warnings[0], want)
				}
			}
		})
	}
}
