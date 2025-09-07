package scan

import (
    "os"
    "path/filepath"
    "testing"
)

func makeDir(t *testing.T, path string) string {
    t.Helper()
    if err := os.MkdirAll(path, 0o755); err != nil {
        t.Fatalf("mkdir %s: %v", path, err)
    }
    return path
}

func touchDir(t *testing.T, path string) string { return makeDir(t, path) }

func TestFindGitRepos_FindsDotGitDirs(t *testing.T) {
    root := t.TempDir()
    // repo1 at root/repo1/.git
    repo1 := filepath.Join(root, "repo1")
    touchDir(t, filepath.Join(repo1, ".git"))

    // repo2 nested in a subdir
    repo2 := filepath.Join(root, "parent", "repo2")
    touchDir(t, filepath.Join(repo2, ".git"))

    repos, warnings := FindGitRepos([]string{root}, nil)
    if len(warnings) != 0 {
        t.Fatalf("unexpected warnings: %v", warnings)
    }
    if len(repos) != 2 {
        t.Fatalf("expected 2 repos, got %d: %v", len(repos), repos)
    }
}

func TestFindGitRepos_RespectsDirIgnore(t *testing.T) {
    root := t.TempDir()
    // repo outside ignored path
    repo1 := filepath.Join(root, "repo1")
    touchDir(t, filepath.Join(repo1, ".git"))

    // repo under node_modules should be ignored
    nmRepo := filepath.Join(root, "node_modules", "x")
    touchDir(t, filepath.Join(nmRepo, ".git"))

    // repo under absolute ignored folder
    ignoredAbs := filepath.Join(root, "ignored")
    repo3 := filepath.Join(ignoredAbs, "repo3")
    touchDir(t, filepath.Join(repo3, ".git"))

    patterns := []string{"**/node_modules/**", "/ignored/**"}
    repos, _ := FindGitRepos([]string{root}, patterns)

    if len(repos) != 1 || repos[0] != repo1 {
        t.Fatalf("expected only %s, got %v", repo1, repos)
    }
}

