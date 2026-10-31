package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
			name: "non-ASCII path in subdirectory",
			setup: func(t *testing.T, repo string) {
				writeFileAt(t, filepath.Join(repo, "a.txt"), "a", oldFileAt)
				runGitAt(t, repo, commitAt, "add", ".")
				runGitAt(t, repo, commitAt, "commit", "-m", "init")
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

			_, paths, err := getUncommittedChanges(repo)
			if err != nil {
				t.Fatalf("getUncommittedChanges: %v", err)
			}

			got, warnings := LastActivity(repo, paths)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("LastActivity = %v, want %v (paths=%q)", got, tt.want, paths)
			}
		})
	}
}

// File names that look like porcelain syntax must still be dated from their
// mtime: a leading quote, or " -> " on either side of a rename.
func TestLastActivity_NamesWithPorcelainSyntax(t *testing.T) {
	gitOrSkip(t)
	if runtime.GOOS == "windows" {
		t.Skip(`'"' and '>' are not allowed in Windows file names`)
	}

	commitAt := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	newFileAt := commitAt.Add(72 * time.Hour)

	tests := []struct {
		name     string
		original string
		renameTo string // empty: modify original in place
	}{
		{name: "modified file with leading quote", original: `"quoted"`},
		{name: "rename destination containing arrow", original: "src.txt", renameTo: "dst -> final.txt"},
		{name: "rename source containing arrow", original: "a -> b", renameTo: "new.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := t.TempDir()
			runGit(t, repo, "init", "-q")
			writeFileAt(t, filepath.Join(repo, tt.original), "a", commitAt)
			runGitAt(t, repo, commitAt, "add", ".")
			runGitAt(t, repo, commitAt, "commit", "-m", "init")

			changed := tt.original
			if tt.renameTo != "" {
				runGitAt(t, repo, commitAt, "mv", tt.original, tt.renameTo)
				changed = tt.renameTo
			}
			writeFileAt(t, filepath.Join(repo, changed), "changed", newFileAt)

			// Goes through the provider so it also covers which paths reach LastActivity.
			state, _ := New().CheckRepoStateWithActivity(repo)
			if !state.LastActivity.Equal(newFileAt) {
				t.Fatalf("LastActivity = %v, want %v (files=%q)", state.LastActivity, newFileAt, state.UncommitedFiles)
			}
		})
	}
}

func TestProviderCheckRepoStateWithActivity(t *testing.T) {
	gitOrSkip(t)

	commitAt := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		corruptIndex bool
		want         time.Time
		wantWarning  bool
	}{
		{name: "clean repo is dated from HEAD", want: commitAt},
		{name: "failing status leaves activity unknown", corruptIndex: true, want: time.Time{}, wantWarning: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := t.TempDir()
			runGit(t, repo, "init", "-q")
			writeFileAt(t, filepath.Join(repo, "a.txt"), "a", commitAt)
			runGitAt(t, repo, commitAt, "add", ".")
			runGitAt(t, repo, commitAt, "commit", "-m", "init")
			if tt.corruptIndex {
				// git status reads the index and fails; git log does not.
				if err := os.WriteFile(filepath.Join(repo, ".git", "index"), []byte("garbage"), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			state, warnings := New().CheckRepoStateWithActivity(repo)
			got := state.LastActivity

			if state.VCSType != "git" {
				t.Fatalf("VCSType = %q, want git", state.VCSType)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("LastActivity = %v, want %v", got, tt.want)
			}
			if (len(warnings) > 0) != tt.wantWarning {
				t.Fatalf("warnings = %v, wantWarning %v", warnings, tt.wantWarning)
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
