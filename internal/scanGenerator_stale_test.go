package internal

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mabd-dev/reposcan/internal/config"
)

// initDatedGitRepo creates a git repo under root with a single commit dated at,
// plus an uncommitted change whose mtime is also at.
func initDatedGitRepo(t *testing.T, root string, name string, at time.Time) string {
	t.Helper()

	repo := filepath.Join(root, name)
	date := at.Format(time.RFC3339)
	git := func(args ...string) {
		t.Helper()
		full := append([]string{"-C", repo,
			"-c", "user.email=test@example.com",
			"-c", "user.name=test",
			"-c", "commit.gpgsign=false",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+date, "GIT_AUTHOR_DATE="+date)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	git("init", "-q")
	file := filepath.Join(repo, "a.txt")
	if err := os.WriteFile(file, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "init")
	if err := os.WriteFile(file, []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(file, at, at); err != nil {
		t.Fatal(err)
	}

	return repo
}

func TestGenerateScanReport_StaleDays(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	root := t.TempDir()
	now := time.Now()
	initDatedGitRepo(t, root, "stale", now.Add(-30*24*time.Hour))
	initDatedGitRepo(t, root, "fresh", now.Add(-1*time.Hour))

	tests := []struct {
		name      string
		only      config.OnlyFilter
		staleDays int
		wantRepos []string
		wantDated bool
	}{
		{name: "disabled", only: config.OnlyDirty, staleDays: 0, wantRepos: []string{"fresh", "stale"}},
		{name: "dirty stale", only: config.OnlyDirty, staleDays: 7, wantRepos: []string{"stale"}, wantDated: true},
		{name: "uncommitted stale", only: config.OnlyUncommitted, staleDays: 7, wantRepos: []string{"stale"}, wantDated: true},
		{name: "threshold above both", only: config.OnlyDirty, staleDays: 60, wantRepos: []string{}, wantDated: true},
		{name: "unpulled ignores stale", only: config.OnlyUnpulled, staleDays: 7, wantRepos: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Defaults()
			cfg.Roots = []string{root}
			cfg.DirIgnore = nil
			cfg.Only = tt.only
			cfg.StaleDays = tt.staleDays

			r := GenerateScanReport(cfg)

			// Compare by directory name: temp paths can differ in form on Windows.
			gotRepos := []string{}
			for _, s := range r.RepoStates {
				gotRepos = append(gotRepos, filepath.Base(s.Path))
				if s.LastActivity.IsZero() == tt.wantDated {
					t.Fatalf("%s: LastActivity=%v, wantDated=%v", s.Path, s.LastActivity, tt.wantDated)
				}
			}
			if strings.Join(gotRepos, ",") != strings.Join(tt.wantRepos, ",") {
				t.Fatalf("repos = %v, want %v", gotRepos, tt.wantRepos)
			}
		})
	}
}

func TestGenerateScanReport_LastActivityJSONOnlyWhenEnabled(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}

	root := t.TempDir()
	initDatedGitRepo(t, root, "repo", time.Now().Add(-30*24*time.Hour))

	for _, staleDays := range []int{0, 7} {
		cfg := config.Defaults()
		cfg.Roots = []string{root}
		cfg.DirIgnore = nil
		cfg.Only = config.OnlyAll
		cfg.StaleDays = staleDays

		b, err := json.Marshal(GenerateScanReport(cfg))
		if err != nil {
			t.Fatal(err)
		}

		want := staleDays > 0
		if got := strings.Contains(string(b), `"lastActivity"`); got != want {
			t.Fatalf("staleDays=%d: lastActivity present=%v, want %v\n%s", staleDays, got, want, b)
		}
	}
}
