package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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

func TestGitProviderImplementsVcsProvider(t *testing.T) {
	var _ vcs.Provider = (*Provider)(nil)
}

func TestGitProviderImplementsVcsActionProvider(t *testing.T) {
	var _ vcs.ActionProvider = (*Provider)(nil)
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

func TestProviderType(t *testing.T) {
	if got := New().Type(); got != vcs.TypeGit {
		t.Fatalf("expected type %q, got %q", vcs.TypeGit, got)
	}
}

func TestProviderFetchPushPull(t *testing.T) {
	gitOrSkip(t)
	root := t.TempDir()
	bare := initBareMain(t, "remote.git")

	// Seed the bare remote so clones get branch tracking for origin.
	seed := filepath.Join(root, "seed")
	runGit(t, root, "init", "seed")
	runGit(t, seed, "commit", "--allow-empty", "-m", "base")
	runGit(t, seed, "branch", "-M", "main")
	runGit(t, seed, "remote", "add", "origin", bare)
	runGit(t, seed, "push", "-u", "origin", "main")

	local := filepath.Join(root, "local")
	runGit(t, root, "clone", bare, local)
	if err := os.WriteFile(filepath.Join(local, "f.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	runGit(t, local, "add", "f.txt")
	runGit(t, local, "commit", "-m", "local change")

	if out, err := New().Fetch(local); err != nil {
		t.Fatalf("Fetch: %v (%s)", err, out)
	}
	if out, err := New().Push(local); err != nil {
		t.Fatalf("Push: %v (%s)", err, out)
	}

	// Second clone pushes new work so `local` has something real to pull.
	peer := filepath.Join(root, "peer")
	runGit(t, root, "clone", bare, peer)
	if err := os.WriteFile(filepath.Join(peer, "g.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	runGit(t, peer, "add", "g.txt")
	runGit(t, peer, "commit", "-m", "peer change")
	if _, err := New().Push(peer); err != nil {
		t.Fatalf("Push(peer): %v", err)
	}

	if out, err := New().Pull(local); err != nil {
		t.Fatalf("Pull: %v (%s)", err, out)
	}
	if _, err := os.Stat(filepath.Join(local, "g.txt")); err != nil {
		t.Fatalf("expected pulled file g.txt: %v", err)
	}

	// Error paths: not a git repository.
	notRepo := filepath.Join(root, "not-repo")
	if err := os.Mkdir(notRepo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, action := range map[string]func(string) (string, error){
		"Fetch": New().Fetch,
		"Push":  New().Push,
		"Pull":  New().Pull,
	} {
		if _, err := action(notRepo); err == nil {
			t.Fatalf("expected %s to fail on non-repo dir", name)
		}
	}
}

func TestCheckRepoState_WarnsOnNonGitDir(t *testing.T) {
	gitOrSkip(t)
	dir := t.TempDir()

	state, warnings := CheckRepoState(dir)

	if state.Branch != "-" {
		t.Fatalf("expected branch %q on error, got %q", "-", state.Branch)
	}
	if state.Repo != "" {
		t.Fatalf("expected empty repo name, got %q", state.Repo)
	}
	if len(state.UncommitedFiles) != 0 {
		t.Fatalf("expected no uncommitted files, got %v", state.UncommitedFiles)
	}
	if len(state.RemoteStatus) != 1 ||
		state.RemoteStatus[0].Remote != "" ||
		state.RemoteStatus[0].Ahead != -1 ||
		state.RemoteStatus[0].Behind != -1 {
		t.Fatalf("expected empty-remote fallback status, got %v", state.RemoteStatus)
	}

	wantWarnings := []string{
		"Failed to get branch name, path=" + dir,
		"Failed to get git remotes, path=" + dir,
		"Failed to get repo name, path=" + dir,
		"Failed to get uncommited files, path=" + dir,
		"Failed to get stashes, path=" + dir,
	}
	if !reflect.DeepEqual(warnings, wantWarnings) {
		t.Fatalf("expected warnings %v, got %v", wantWarnings, warnings)
	}
}

func TestCheckRepoState_WarnsWhenRemoteBranchMissing(t *testing.T) {
	gitOrSkip(t)
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	repo := filepath.Join(root, "repo")

	runGit(t, root, "init", "--bare", bare)
	runGit(t, root, "init", "repo")
	runGit(t, repo, "commit", "--allow-empty", "-m", "base")
	runGit(t, repo, "branch", "-M", "main")
	runGit(t, repo, "remote", "add", "origin", bare)

	state, warnings := CheckRepoState(repo)

	want := "Failed to get upstream status for remote=origin, path=" + repo
	if len(warnings) != 1 || warnings[0] != want {
		t.Fatalf("expected warning %q, got %v", want, warnings)
	}

	if len(state.RemoteStatus) != 1 {
		t.Fatalf("expected 1 remote status, got %d: %v", len(state.RemoteStatus), state.RemoteStatus)
	}
	rs := state.RemoteStatus[0]
	if rs.Remote != "origin" || rs.Ahead != -1 || rs.Behind != -1 {
		t.Fatalf("expected origin status with ahead/behind -1, got %+v", rs)
	}
}
