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
// - Or a `.git` file whose contents include "gitdir:" (worktrees/submodules).
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
			}
			return nil
		})
	}

	return removeDuplicates(gitReposPaths), warnings
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if file, err := os.Lstat(gitPath); err == nil {
		if file.IsDir() {
			return true
		} else {
			// git worktrees has gitdir: folder
			b, err := os.ReadFile(gitPath)
			if err != nil {
				return false
			}
			return strings.Contains(string(b), "gitdir:")
		}
	}

	// Check if this is a bare repository
	// Bare repos don't have .git folder, but have HEAD, refs/, and objects/ directly
	return isBareRepo(path)
}

// isBareRepo checks if a directory is a bare git repository
// A bare repo has the git contents directly in the directory (no .git folder)
// Key indicators: HEAD file, refs/ directory, objects/ directory
func isBareRepo(path string) bool {
	// Check for required bare repo markers
	headPath := filepath.Join(path, "HEAD")
	refsPath := filepath.Join(path, "refs")
	objectsPath := filepath.Join(path, "objects")

	// All three must exist
	if !fileExists(headPath) {
		return false
	}
	if !dirExists(refsPath) {
		return false
	}
	if !dirExists(objectsPath) {
		return false
	}

	// Additional validation: check if config file has bare = true (optional but recommended)
	configPath := filepath.Join(path, "config")
	if fileExists(configPath) {
		if b, err := os.ReadFile(configPath); err == nil {
			content := string(b)
			// Look for bare = true in the config
			if strings.Contains(content, "bare = true") || strings.Contains(content, "bare=true") {
				return true
			}
		}
	}

	// If we have HEAD, refs, and objects, it's likely a bare repo even without config check
	// This handles edge cases where config might not explicitly say bare = true
	return true
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
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
