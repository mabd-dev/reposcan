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

func makeBareRepo(t *testing.T, path string) string {
	t.Helper()
	// Create required directories and files for a bare repo
	makeDir(t, path)
	makeDir(t, filepath.Join(path, "refs"))
	makeDir(t, filepath.Join(path, "refs", "heads"))
	makeDir(t, filepath.Join(path, "refs", "tags"))
	makeDir(t, filepath.Join(path, "objects"))
	makeDir(t, filepath.Join(path, "objects", "info"))
	makeDir(t, filepath.Join(path, "objects", "pack"))

	// Create HEAD file
	headPath := filepath.Join(path, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}

	// Create config file with bare = true
	configPath := filepath.Join(path, "config")
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}

func TestFindGitRepos_FindsBareRepos(t *testing.T) {
	root := t.TempDir()

	// Create a regular repo
	regularRepo := filepath.Join(root, "regular-repo")
	touchDir(t, filepath.Join(regularRepo, ".git"))

	// Create a bare repo
	bareRepo := filepath.Join(root, "bare-repo.git")
	makeBareRepo(t, bareRepo)

	repos, warnings := FindGitRepos([]string{root}, nil)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d: %v", len(repos), repos)
	}

	// Check that both repos were found
	foundRegular := false
	foundBare := false
	for _, repo := range repos {
		if repo == regularRepo {
			foundRegular = true
		}
		if repo == bareRepo {
			foundBare = true
		}
	}

	if !foundRegular {
		t.Errorf("regular repo not found in results: %v", repos)
	}
	if !foundBare {
		t.Errorf("bare repo not found in results: %v", repos)
	}
}

func TestIsBareRepo(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
	}{
		{
			name: "valid bare repo with config",
			setup: func(t *testing.T) string {
				return makeBareRepo(t, filepath.Join(t.TempDir(), "bare.git"))
			},
			expected: true,
		},
		{
			name: "bare repo without config",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "bare-no-config.git")
				makeDir(t, path)
				makeDir(t, filepath.Join(path, "refs"))
				makeDir(t, filepath.Join(path, "objects"))
				os.WriteFile(filepath.Join(path, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
				return path
			},
			expected: true,
		},
		{
			name: "not a bare repo - missing HEAD",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "not-bare")
				makeDir(t, path)
				makeDir(t, filepath.Join(path, "refs"))
				makeDir(t, filepath.Join(path, "objects"))
				return path
			},
			expected: false,
		},
		{
			name: "not a bare repo - missing refs",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "not-bare")
				makeDir(t, path)
				makeDir(t, filepath.Join(path, "objects"))
				os.WriteFile(filepath.Join(path, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
				return path
			},
			expected: false,
		},
		{
			name: "not a bare repo - missing objects",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "not-bare")
				makeDir(t, path)
				makeDir(t, filepath.Join(path, "refs"))
				os.WriteFile(filepath.Join(path, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
				return path
			},
			expected: false,
		},
		{
			name: "regular directory",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "regular-dir")
				makeDir(t, path)
				return path
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			result := isBareRepo(path)
			if result != tt.expected {
				t.Errorf("isBareRepo(%s) = %v, want %v", path, result, tt.expected)
			}
		})
	}
}
