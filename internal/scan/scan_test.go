package scan

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/vcs"
)

func makeDir(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	return path
}

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func repoTypeByPath(repos []vcs.RepoInfo) map[string]vcs.Type {
	result := map[string]vcs.Type{}
	for _, repo := range repos {
		result[repo.Path] = repo.Type
	}
	return result
}

func TestFindRepos_DetectsSupportedRepoTypes(t *testing.T) {
	root := t.TempDir()

	gitRepo := filepath.Join(root, "git-repo")
	makeDir(t, filepath.Join(gitRepo, ".git"))

	jjRepo := filepath.Join(root, "jj-repo")
	makeDir(t, filepath.Join(jjRepo, ".jj"))

	gitFileRepo := filepath.Join(root, "git-file-repo")
	makeDir(t, gitFileRepo)
	writeFile(t, filepath.Join(gitFileRepo, ".git"), "gitdir: /tmp/worktrees/git-file-repo")

	repos, warnings := FindRepos([]string{root}, nil)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	got := repoTypeByPath(repos)
	want := map[string]vcs.Type{
		gitRepo:     vcs.TypeGit,
		jjRepo:      vcs.TypeJJ,
		gitFileRepo: vcs.TypeGit,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d repos, got %d: %v", len(want), len(got), repos)
	}

	for path, repoType := range want {
		if got[path] != repoType {
			t.Fatalf("expected %s to be %s, got %s", path, repoType, got[path])
		}
	}
}

func TestFindRepos_RespectsDirIgnoreForGitAndJJ(t *testing.T) {
	root := t.TempDir()

	keepRepo := filepath.Join(root, "keep")
	makeDir(t, filepath.Join(keepRepo, ".git"))

	ignoredJJRepo := filepath.Join(root, "node_modules", "jj-repo")
	makeDir(t, filepath.Join(ignoredJJRepo, ".jj"))

	ignoredGitRepo := filepath.Join(root, "ignored", "git-repo")
	makeDir(t, filepath.Join(ignoredGitRepo, ".git"))

	patterns := []string{"**/node_modules/**", "/ignored/**"}
	repos, warnings := FindRepos([]string{root}, patterns)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(repos), repos)
	}

	if repos[0].Path != keepRepo || repos[0].Type != vcs.TypeGit {
		t.Fatalf("expected only %s as git repo, got %v", keepRepo, repos)
	}
}

func TestFindRepos_PrefersJJWhenGitAndJJAreCoLocated(t *testing.T) {
	root := t.TempDir()
	mixedRepo := filepath.Join(root, "mixed")

	makeDir(t, filepath.Join(mixedRepo, ".git"))
	makeDir(t, filepath.Join(mixedRepo, ".jj"))

	repos, warnings := FindRepos([]string{root}, nil)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(repos), repos)
	}

	if repos[0].Path != mixedRepo || repos[0].Type != vcs.TypeJJ {
		t.Fatalf("expected %s to be detected as jj repo, got %v", mixedRepo, repos[0])
	}
}

func TestFindRepos_GitWorktree(t *testing.T) {
	root := t.TempDir()

	worktree := filepath.Join(root, "worktree1")
	makeDir(t, worktree)

	gitFilepath := filepath.Join(worktree, ".git")

	gitFileContent := []byte("gitdir: /Users/someone/worktreeRepo.git/worktrees/worktree1\n")
	err := os.WriteFile(gitFilepath, gitFileContent, 0o644)

	if err != nil {
		t.Fatalf("error writing to file, error=%v", err.Error())
	}

	patterns := []string{"**/node_modules/**", "/ignored/**"}
	reposInfo, _ := FindRepos([]string{root}, patterns)

	if len(reposInfo) != 1 || reposInfo[0].Path != worktree {
		t.Fatalf("expected only %s, got %v", worktree, reposInfo)
	}
}

func TestFindRepos_AddsWalkWarningForNonexistentRoot(t *testing.T) {
	root := t.TempDir()
	// Root that does not exist -> WalkDir returns an lstat error.
	missing := filepath.Join(root, "does-not-exist")

	_, warnings := FindRepos([]string{missing}, nil)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning for nonexistent root, got %v", warnings)
	}
	if len(warnings[0]) == 0 {
		t.Fatalf("expected non-empty warning, got %q", warnings[0])
	}
}

func TestFindRepos_SkipsNonDirectoryEntries(t *testing.T) {
	root := t.TempDir()

	// A loose, non-.git file directly under the root is visited by WalkDir
	// with IsDir()==false and must be skipped without being treated as a repo.
	writeFile(t, filepath.Join(root, "notes.txt"), "hello")

	gitRepo := filepath.Join(root, "git-repo")
	makeDir(t, filepath.Join(gitRepo, ".git"))

	repos, warnings := FindRepos([]string{root}, nil)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(repos) != 1 || repos[0].Path != gitRepo {
		t.Fatalf("expected only %s, got %v", gitRepo, repos)
	}
}

func TestFindRepos_OverlappingRootsAreDeduplicated(t *testing.T) {
	root := t.TempDir()
	gitRepo := filepath.Join(root, "git-repo")
	makeDir(t, filepath.Join(gitRepo, ".git"))

	// Passing the same dir twice reaches the same path on the second walk,
	// which must be skipped via the visited set and collapsed in the output.
	repos, warnings := FindRepos([]string{root, root}, nil)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(repos) != 1 || repos[0].Path != gitRepo {
		t.Fatalf("expected single deduplicated repo %s, got %v", gitRepo, repos)
	}
}

// setHome points os.UserHomeDir at dir on every platform (HOME on unix,
// USERPROFILE on Windows).
func setHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

// TestFindRepos_RootFormsWithIgnorePatterns covers every way a root can be
// written (~, $VAR, trailing slash, relative) against root-anchored "/..."
// and home-relative "~/..." ignore patterns. Roots are made absolute and
// clean before reaching the matcher, or these patterns silently stop matching.
func TestFindRepos_RootFormsWithIgnorePatterns(t *testing.T) {
	// Resolve symlinks (macOS /var -> /private/var) so the home dir and the
	// working dir set by t.Chdir name the same path.
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	setHome(t, home)
	code := filepath.Join(home, "code")
	t.Setenv("REPOSCAN_TEST_ROOT", code)

	makeDir(t, filepath.Join(code, "app", ".git"))
	makeDir(t, filepath.Join(code, "vendor", "lib", ".git"))
	makeDir(t, filepath.Join(home, "skip", "old", ".git"))

	tests := []struct {
		name   string
		cwd    string // relative to home; empty keeps the working dir
		roots  []string
		ignore []string
		want   []string // repo paths relative to home
	}{
		{name: "tilde root", roots: []string{"~/code"}, ignore: []string{"/vendor/**"}, want: []string{"code/app"}},
		{name: "tilde root with trailing slash", roots: []string{"~/code/"}, ignore: []string{"/vendor/**"}, want: []string{"code/app"}},
		{name: "absolute root with trailing slash", roots: []string{code + string(filepath.Separator)}, ignore: []string{"/vendor/**"}, want: []string{"code/app"}},
		{name: "env var root", roots: []string{"$REPOSCAN_TEST_ROOT"}, ignore: []string{"/vendor/**"}, want: []string{"code/app"}},
		{name: "tilde root with tilde ignore", roots: []string{"~"}, ignore: []string{"~/skip/**"}, want: []string{"code/app", "code/vendor/lib"}},
		{name: "tilde ignore is not anchored to the root", roots: []string{"~/code"}, ignore: []string{"~/vendor/**"}, want: []string{"code/app", "code/vendor/lib"}},
		{name: "relative root with tilde ignore", cwd: ".", roots: []string{"."}, ignore: []string{"~/skip/**"}, want: []string{"code/app", "code/vendor/lib"}},
		{name: "relative root with anchored ignore", cwd: "code", roots: []string{"."}, ignore: []string{"/vendor/**"}, want: []string{"code/app"}},
		{name: "relative and absolute spellings of a root are deduplicated", cwd: "code", roots: []string{".", code}, want: []string{"code/app", "code/vendor/lib"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cwd != "" {
				t.Chdir(filepath.Join(home, tt.cwd))
			}

			repos, warnings := FindRepos(tt.roots, tt.ignore)
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}

			got := make([]string, 0, len(repos))
			for _, repo := range repos {
				got = append(got, repo.Path)
			}
			want := make([]string, 0, len(tt.want))
			for _, w := range tt.want {
				want = append(want, filepath.Join(home, filepath.FromSlash(w)))
			}
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("FindRepos(%q, %q)\n got: %q\nwant: %q", tt.roots, tt.ignore, got, want)
			}
		})
	}
}

func TestFindRepos_TildeRootWithoutHomeIsAWarning(t *testing.T) {
	setHome(t, "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")

	repos, warnings := FindRepos([]string{"~/code"}, nil)
	if len(repos) != 0 {
		t.Fatalf("expected no repos, got %v", repos)
	}
	if len(warnings) != 1 || !strings.HasPrefix(warnings[0], "root '~/code'") {
		t.Fatalf("expected 1 warning naming the unexpandable root, got %v", warnings)
	}
}
