package jj

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	cloneCmd := append([]string{"git", "clone"}, cloneArgs...)
	if err := exec.Command("jj", append(cloneCmd, remotePath, workPath)...).Run(); err != nil {
		t.Fatalf("jj git clone: %v", err)
	}

	return trackedJJRepo{
		SeedPath: seedPath,
		WorkPath: workPath,
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

	if state.RemoteStatus[0].Ahead != 0 || state.RemoteStatus[0].Behind != 0 {
		t.Fatalf(
			"expected jj ahead/behind defaults to be 0/0, got %d/%d",
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

	run := func(name string, args ...string) {
		t.Helper()
		if output, err := exec.Command(name, args...).CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, output)
		}
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

	run("git", "init", "--bare", originPath)
	run("git", "init", "--bare", upstreamPath)
	run("git", "clone", originPath, seedPath)
	run("git", "-C", seedPath, "config", "user.name", "test")
	run("git", "-C", seedPath, "config", "user.email", "test@example.com")
	run("git", "-C", seedPath, "remote", "add", "upstream", upstreamPath)

	// A is on both remotes' history; origin has advanced to X before the clone.
	appendLine(seedFile, "A")
	run("git", "-C", seedPath, "add", "README.md")
	run("git", "-C", seedPath, "commit", "-m", "A")
	run("git", "-C", seedPath, "branch", "-M", "main")
	run("git", "-C", seedPath, "push", "upstream", "main")
	appendLine(seedFile, "X")
	run("git", "-C", seedPath, "commit", "-am", "X")
	run("git", "-C", seedPath, "push", "origin", "main")

	run("jj", "git", "clone", "--no-colocate", originPath, workPath)
	run("jj", "-R", workPath, "git", "remote", "add", "upstream", upstreamPath)
	run("jj", "-R", workPath, "git", "fetch", "--remote", "upstream")
	run("jj", "-R", workPath, "bookmark", "track", "main@upstream")

	// Local main moves to L on top of X.
	run("jj", "-R", workPath, "new", "main")
	appendLine(filepath.Join(workPath, "README.md"), "L")
	run("jj", "-R", workPath, "describe", "-m", "L")
	run("jj", "-R", workPath, "bookmark", "move", "main", "-t", "@")
	run("jj", "-R", workPath, "new")

	// origin moves X -> R and upstream moves A -> U.
	appendLine(seedFile, "R")
	run("git", "-C", seedPath, "commit", "-am", "R")
	run("git", "-C", seedPath, "push", "origin", "main")
	run("git", "-C", seedPath, "checkout", "-b", "upstream-work", "HEAD~2")
	appendLine(filepath.Join(seedPath, "UPSTREAM.md"), "U")
	run("git", "-C", seedPath, "add", "UPSTREAM.md")
	run("git", "-C", seedPath, "commit", "-m", "U")
	run("git", "-C", seedPath, "push", "upstream", "upstream-work:main")

	run("jj", "-R", workPath, "git", "fetch", "--remote", "origin")
	run("jj", "-R", workPath, "git", "fetch", "--remote", "upstream")

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
			name:          "repo name",
			failedCommand: fakeJJCommandKey("git", "remote", "list"),
			wantOperation: "get repo name",
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
			failedCommand: trackedBookmarksCommandKey(),
			wantOperation: "get remote status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := filepath.Join(t.TempDir(), "repo")
			responses := map[string]fakeJJResponse{
				fakeJJCommandKey("git", "remote", "list"): {},
				branchCommandKey():                        {Stdout: "main*?|abc123\n"},
				fakeJJCommandKey("diff", "--summary"):     {},
				trackedBookmarksCommandKey():              {},
				untrackedRemotesCommandKey("main"):        {},
			}
			responses[tt.failedCommand] = fakeJJResponse{Stderr: "inspection failed", ExitCode: 1}
			if tt.failedCommand == branchCommandKey() {
				// A failed branch lookup uses "-" for the follow-up remote lookup.
				responses[untrackedRemotesCommandKey("-")] = fakeJJResponse{}
			}
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
