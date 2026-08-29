package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestGetUpstreamStatus_AheadBehind(t *testing.T) {
	gitOrSkip(t)
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	runGit(t, root, "init", "--bare", bare)

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

	ahead, behind, err := GetUpstreamStatus(local)
	if err != nil {
		t.Fatalf("GetUpstreamStatus: %v", err)
	}
	if ahead != 1 || behind != 0 {
		t.Fatalf("expected ahead=1 behind=0, got ahead=%d behind=%d", ahead, behind)
	}
}

func TestGetUpstreamStatus_NoUpstream(t *testing.T) {
	gitOrSkip(t)
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "commit", "--allow-empty", "-m", "c")

	ahead, behind, err := GetUpstreamStatus(repo)
	if err == nil {
		t.Fatal("expected error for repo without upstream")
	}
	if ahead != -1 || behind != -1 {
		t.Fatalf("expected ahead=-1 behind=-1 on error, got ahead=%d behind=%d", ahead, behind)
	}
}

func TestGetUpstreamStatusForRemote_RevListError(t *testing.T) {
	gitOrSkip(t)
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "commit", "--allow-empty", "-m", "init")

	// Point the remote-tracking ref at a tree object: rev-parse resolves it
	// but rev-list rejects it, mirroring a corrupt/foreign remote ref.
	treeOut, err := exec.Command("git", "-C", repo, "rev-parse", "HEAD^{tree}").Output()
	if err != nil {
		t.Fatalf("rev-parse tree: %v", err)
	}
	runGit(t, repo, "update-ref", "refs/remotes/origin/main", strings.TrimSpace(string(treeOut)))

	if _, err := GetUpstreamStatusForRemote(repo, "origin", "main"); err == nil {
		t.Fatal("expected error when remote branch ref is not a commit")
	}
}

func TestGetOutgoingCommitsForRemote_NonGitDir(t *testing.T) {
	gitOrSkip(t)
	dir := t.TempDir()

	commits, err := GetOutgoingCommitsForRemote(dir, "origin", "main")
	if err == nil {
		t.Fatal("expected error for non-git dir")
	}
	if len(commits) != 0 {
		t.Fatalf("expected no commits, got %v", commits)
	}
}

func TestGetRepoName_NonGitDir(t *testing.T) {
	gitOrSkip(t)
	dir := t.TempDir()

	if _, err := GetRepoName(dir); err == nil {
		t.Fatal("expected error for non-git dir")
	}
}

func TestGetRepoName_UsesFirstNonOriginRemote(t *testing.T) {
	gitOrSkip(t)
	root := t.TempDir()
	bare := filepath.Join(root, "service.git")
	runGit(t, root, "init", "--bare", bare)

	repo := filepath.Join(root, "repo")
	runGit(t, root, "init", "repo")
	runGit(t, repo, "remote", "add", "upstream", bare)

	name, err := GetRepoName(repo)
	if err != nil {
		t.Fatalf("GetRepoName: %v", err)
	}
	if name != "service" {
		t.Fatalf("expected repo name %q, got %q", "service", name)
	}
}

func TestFirstRemoteURL(t *testing.T) {
	gitOrSkip(t)
	root := t.TempDir()
	bare := filepath.Join(root, "service.git")
	runGit(t, root, "init", "--bare", bare)

	repo := filepath.Join(root, "repo")
	runGit(t, root, "init", "repo")
	runGit(t, repo, "remote", "add", "upstream", bare)

	url, err := firstRemoteURL(repo)
	if err != nil {
		t.Fatalf("firstRemoteURL: %v", err)
	}
	if strings.TrimSpace(url) != bare {
		t.Fatalf("expected url %q, got %q", bare, url)
	}

	noRemote := filepath.Join(root, "no-remote")
	runGit(t, root, "init", "no-remote")
	if _, err := firstRemoteURL(noRemote); err == nil {
		t.Fatal("expected error for repo without remotes")
	}

	notRepo := filepath.Join(root, "not-repo")
	if err := os.Mkdir(notRepo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := firstRemoteURL(notRepo); err == nil {
		t.Fatal("expected error for non-git dir")
	}
}

func TestParseRepoName(t *testing.T) {
	cases := []struct {
		name   string
		remote string
		want   string
		ok     bool
	}{
		{name: "scp style", remote: "git@gitlab.com:group/sub/repo.git", want: "repo", ok: true},
		{name: "https", remote: "https://github.com/org/repo.git", want: "repo", ok: true},
		{name: "local path with .git suffix", remote: "/srv/git/repo.git", want: "repo", ok: true},
		{name: "regex fallback for host-only URL", remote: "ssh://git@example.com", want: "git@example.com", ok: true},
		{name: "empty", remote: "", want: "", ok: false},
		{name: "unparseable with trailing separators", remote: "http://", want: "", ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseRepoName(tc.remote)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("parseRepoName(%q) = (%q, %v), want (%q, %v)", tc.remote, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestAtoiSafe(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{in: "12", want: 12},
		{in: "012", want: 12},
		{in: "12a", want: 12},
		{in: "-5", want: 0},
		{in: "abc", want: 0},
		{in: "", want: 0},
	}

	for _, tc := range cases {
		if got := atoiSafe(tc.in); got != tc.want {
			t.Fatalf("atoiSafe(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
