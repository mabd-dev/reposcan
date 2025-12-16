package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindGitRepos walks each root and returns directories that look like Git worktrees.
// Simple rules:
// - A directory containing `.git` (directory) is a repo root.
// - Or a folder with name suffifx`.git` (worktrees probably).
// - When we find a repo root, we SkipDir to avoid descending into nested repos (for now).
func FindGitRepos(
	roots []string,
	dirignore []string,
) (gitReposPaths []string, warnings []string) {
	matcher := NewIgnoreMatcher(roots, dirignore)

	visitedDir := map[string]struct{}{}

	for _, root := range roots {
		root = os.ExpandEnv(root)

		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// possible errors: permission denied
				warnings = append(warnings, err.Error())
				return nil
			}

			if !d.IsDir() {
				return nil
			}

			if matcher.ShouldIgnore(path) {
				return fs.SkipDir
			}

			if _, visited := visitedDir[path]; visited {
				return fs.SkipDir
			}
			visitedDir[path] = struct{}{}

			if isGitRepo(path) {
				gitReposPaths = append(gitReposPaths, path)
				return fs.SkipDir
			} else if isBareRepo(path) {
				gitReposPaths = append(gitReposPaths, path)
				return fs.SkipDir
			}
			return nil
		})
	}

	return removeDuplicates(gitReposPaths), warnings
}

// isGitRepo checks if [path] contains `.git` folder
func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if file, err := os.Lstat(gitPath); err == nil {
		if file.IsDir() {
			return true
		}
	}
	return false
}

// isBareRepo checks if folder name ends with '.git'
func isBareRepo(path string) bool {
	return strings.HasSuffix(path, ".git")
}

func removeDuplicates(strs []string) []string {
	seen := make(map[string]struct{}, len(strs))
	distinct := make([]string, 0, len(strs))

	for _, s := range strs {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			distinct = append(distinct, s)
		}
	}
	return distinct
}
