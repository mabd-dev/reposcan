package jj

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mabd-dev/reposcan/internal/vcs"
)

// runJJAt runs jj in repo with commit timestamps pinned to at.
func runJJAt(t *testing.T, repo string, at time.Time, args ...string) {
	t.Helper()
	cmd := exec.Command("jj", append([]string{"-R", repo}, args...)...)
	cmd.Env = append(os.Environ(),
		"JJ_TIMESTAMP="+at.Format(time.RFC3339),
		"JJ_USER=test",
		"JJ_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("jj %v: %v\n%s", args, err, out)
	}
}

func writeFileAt(t *testing.T, path string, content string, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

// parentCommitDateCommandKey records the exact jj log invocation used to date
// the working-copy commit's parent.
func parentCommitDateCommandKey() string {
	return fakeJJCommandKey(
		"log",
		"-r",
		"latest(@- ~ root())",
		"--no-graph",
		"-T",
		`committer.timestamp().format("%Y-%m-%dT%H:%M:%S%:z")`,
	)
}

func TestJJProviderImplementsVcsActivityProvider(t *testing.T) {
	var _ vcs.ActivityProvider = (*Provider)(nil)
}

func TestSummaryPath(t *testing.T) {
	tests := []struct {
		line   string
		want   string
		wantOK bool
	}{
		{line: "M a.txt", want: "a.txt", wantOK: true},
		{line: "A dir/br{ace}.txt", want: "dir/br{ace}.txt", wantOK: true},
		{line: "D gone.txt", want: "gone.txt", wantOK: true},
		{line: "R {top.txt => e/top.txt}", want: "e/top.txt", wantOK: true},
		{line: "R e/{f => h}/g.txt", want: "e/h/g.txt", wantOK: true},
		{line: "C {a.txt => b.txt}", want: "b.txt", wantOK: true},
		// The leading slash is harmless: the path is joined under the repo.
		{line: "R {e/f => }/g.txt", want: "/g.txt", wantOK: true},
		{line: "M", want: "", wantOK: false},
	}

	for _, tt := range tests {
		got, ok := summaryPath(tt.line)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("summaryPath(%q) = %q, %v, want %q, %v", tt.line, got, ok, tt.want, tt.wantOK)
		}
	}
}

// TestCheckRepoStateWithActivity dates real jj repos.
func TestCheckRepoStateWithActivity(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj binary not available")
	}

	base := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	commitAt := base
	oldFileAt := base.Add(-48 * time.Hour)
	newFileAt := base.Add(72 * time.Hour)

	commit := func(t *testing.T, repo string) {
		writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
		runJJAt(t, repo, commitAt, "commit", "-m", "init")
	}

	tests := []struct {
		name  string
		setup func(t *testing.T, repo string)
		want  time.Time
	}{
		{
			name:  "no commits and clean",
			setup: func(t *testing.T, repo string) {},
			want:  time.Time{},
		},
		{
			name: "no commits with new file",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "new.txt"), "x", newFileAt)
			},
			want: newFileAt,
		},
		{
			name:  "commit only",
			setup: commit,
			want:  commitAt,
		},
		{
			name: "edited file newer than commit",
			setup: func(t *testing.T, repo string) {
				commit(t, repo)
				writeFileAt(t, filepath.Join(repo, "a.txt"), "edited", newFileAt)
			},
			want: newFileAt,
		},
		{
			// Scanning snapshots this edit, which re-dates @ to the scan time.
			// Only @- may date the repo, or it would always look fresh.
			name: "edited file older than commit",
			setup: func(t *testing.T, repo string) {
				commit(t, repo)
				writeFileAt(t, filepath.Join(repo, "a.txt"), "edited", oldFileAt.Add(-time.Hour))
			},
			want: commitAt,
		},
		{
			name: "renamed file",
			setup: func(t *testing.T, repo string) {
				commit(t, repo)
				writeFileAt(t, filepath.Join(repo, "sub", "b.txt"), "a", newFileAt)
				if err := os.Remove(filepath.Join(repo, "a.txt")); err != nil {
					t.Fatal(err)
				}
			},
			want: newFileAt,
		},
		{
			name: "deleted file",
			setup: func(t *testing.T, repo string) {
				commit(t, repo)
				if err := os.Remove(filepath.Join(repo, "a.txt")); err != nil {
					t.Fatal(err)
				}
			},
			want: commitAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := initJJRepo(t, t.TempDir(), "repo")
			tt.setup(t, repo)

			state, warnings := New().CheckRepoStateWithActivity(repo)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}
			if !state.LastActivity.Equal(tt.want) {
				t.Fatalf("LastActivity = %v, want %v (uncommitted files %v)", state.LastActivity, tt.want, state.UncommitedFiles)
			}
		})
	}
}

// TestCheckRepoStateWithActivityFailures verifies that a repo is never dated
// from partial signals when a jj command fails.
func TestCheckRepoStateWithActivityFailures(t *testing.T) {
	commitAt := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	fileAt := commitAt.Add(time.Hour)

	tests := []struct {
		name          string
		failedCommand string
		want          time.Time
		wantWarning   string
	}{
		{
			name: "no failure",
			want: fileAt,
		},
		{
			name:          "uncommitted files",
			failedCommand: fakeJJCommandKey("diff", "--summary"),
			wantWarning:   "Last activity unknown",
		},
		{
			name:          "parent commit date",
			failedCommand: parentCommitDateCommandKey(),
			wantWarning:   "Failed to get parent commit date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := fakeRepoDir(t, "repo")
			writeFileAt(t, filepath.Join(repoPath, "a.txt"), "a", fileAt)
			responses := map[string]fakeJJResponse{
				fakeJJCommandKey("git", "remote", "list"): {},
				branchCommandKey():                        {Stdout: "main*?|abc123\n"},
				fakeJJCommandKey("diff", "--summary"):     {Stdout: "M a.txt\n"},
				parentCommitDateCommandKey():              {Stdout: commitAt.Format(time.RFC3339)},
			}
			if tt.failedCommand != "" {
				responses[tt.failedCommand] = fakeJJResponse{Stderr: "inspection failed", ExitCode: 1}
			}
			binary := useFakeJJ(t, responses)

			state, warnings := (&Provider{binary: binary}).CheckRepoStateWithActivity(repoPath)
			if !state.LastActivity.Equal(tt.want) {
				t.Fatalf("LastActivity = %v, want %v (warnings %v)", state.LastActivity, tt.want, warnings)
			}

			joined := strings.Join(warnings, "\n")
			if tt.wantWarning == "" && len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}
			if !strings.Contains(joined, tt.wantWarning) {
				t.Fatalf("warnings = %v, want one containing %q", warnings, tt.wantWarning)
			}
		})
	}
}
