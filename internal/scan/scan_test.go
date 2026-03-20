package scan

import (
	"os"
	"path/filepath"
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

func repoTypeByPath(repos []vcs.RepoPath) map[string]vcs.Type {
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
