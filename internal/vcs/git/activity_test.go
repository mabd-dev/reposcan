package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// runGitAt runs git in dir with committer/author dates pinned to at.
func runGitAt(t *testing.T, dir string, at time.Time, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir,
		"-c", "user.email=test@example.com",
		"-c", "user.name=test",
		"-c", "commit.gpgsign=false",
	}, args...)
	cmd := exec.Command("git", full...)
	date := at.Format(time.RFC3339)
	cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+date, "GIT_AUTHOR_DATE="+date)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
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

func TestLastActivity(t *testing.T) {
	gitOrSkip(t)

	base := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	commitAt := base
	oldFileAt := base.Add(-48 * time.Hour)
	newFileAt := base.Add(72 * time.Hour)
	stashAt := base.Add(120 * time.Hour)

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
			name: "no commits with untracked file",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "new.txt"), "x", newFileAt)
			},
			want: newFileAt,
		},
		{
			name: "commit only",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
			},
			want: commitAt,
		},
		{
			name: "uncommitted file newer than commit",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
				writeFileAt(t, filepath.Join(repo, "a.txt"), "changed", newFileAt)
			},
			want: newFileAt,
		},
		{
			name: "commit newer than uncommitted file",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
				writeFileAt(t, filepath.Join(repo, "a.txt"), "changed", oldFileAt)
			},
			want: commitAt,
		},
		{
			name: "stash newer than commit",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
				writeFileAt(t, filepath.Join(repo, "a.txt"), "wip", oldFileAt)
				runGitAt(t, repo, stashAt, "stash")
			},
			want: stashAt,
		},
		{
			name: "deleted file is skipped",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
				if err := os.Remove(filepath.Join(repo, "a.txt")); err != nil {
					t.Fatal(err)
				}
			},
			want: commitAt,
		},
		{
			name: "renamed file uses new path",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
				runGitAt(t, repo, commitAt, "mv", "a.txt", "b.txt")
				if err := os.Chtimes(filepath.Join(repo, "b.txt"), newFileAt, newFileAt); err != nil {
					t.Fatal(err)
				}
			},
			want: newFileAt,
		},
		{
			name: "quoted path with special characters",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
				// Non-ASCII names are C-quoted by git (core.quotePath defaults to true).
				writeFileAt(t, filepath.Join(repo, "dir", "café.txt"), "x", newFileAt)
			},
			want: newFileAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := t.TempDir()
			runGit(t, repo, "init", "-q")
			tt.setup(t, repo)

			files, err := GetUncommitedFiles(repo)
			if err != nil {
				t.Fatalf("GetUncommitedFiles: %v", err)
			}

			got, warnings := LastActivity(repo, files)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("LastActivity = %v, want %v (files=%q)", got, tt.want, files)
			}
		})
	}
}

func TestLastActivity_WarnsOnNonGitDir(t *testing.T) {
	gitOrSkip(t)

	got, warnings := LastActivity(t.TempDir(), nil)
	if !got.IsZero() {
		t.Fatalf("LastActivity = %v, want zero", got)
	}
	if len(warnings) == 0 {
		t.Fatalf("expected warnings for non-git dir")
	}
}

func TestPorcelainPath(t *testing.T) {
	tests := []struct {
		line   string
		want   string
		wantOk bool
	}{
		{line: " M a.txt", want: "a.txt", wantOk: true},
		{line: "?? dir/new file.txt", want: "dir/new file.txt", wantOk: true},
		{line: "R  old.txt -> new.txt", want: "new.txt", wantOk: true},
		{line: "RM old.txt -> new.txt", want: "new.txt", wantOk: true},
		{line: "C  src.txt -> copy.txt", want: "copy.txt", wantOk: true},
		{line: `?? "caf\303\251.txt"`, want: "café.txt", wantOk: true},
		{line: `R  "a\"b.txt" -> "c\"d.txt"`, want: `c"d.txt`, wantOk: true},
		{line: " M a -> b.txt", want: "a -> b.txt", wantOk: true},
		{line: `?? "broken`, wantOk: false},
		{line: "?? ", wantOk: false},
		{line: "", wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got, ok := porcelainPath(tt.line)
			if ok != tt.wantOk || got != tt.want {
				t.Fatalf("porcelainPath(%q) = (%q, %v), want (%q, %v)", tt.line, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}
