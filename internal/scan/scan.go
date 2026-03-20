package scan

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mabd-dev/reposcan/internal/vcs"
)

// FindRepos walks each root and returns directories that look like supported repos.
// Simple rules:
// - A directory containing `.jj` is treated as a jj repo.
// - A directory containing `.git` (directory) is treated as a Git repo.
// - Or a `.git` file whose contents include "gitdir:" (worktrees/submodules).
// - When we find a repo root, we SkipDir to avoid descending into nested repos (for now).
func FindRepos(
	roots []string,
	dirignore []string,
) (repoPaths []vcs.RepoPath, warnings []string) {
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

			if repoType, ok := detectRepoType(path); ok {
				repoPaths = append(repoPaths, vcs.RepoPath{
					Path: path,
					Type: repoType,
				})
				return fs.SkipDir
			}
			return nil
		})
	}

	return removeDuplicateRepoPaths(repoPaths), warnings
}

func detectRepoType(path string) (vcs.Type, bool) {
	if isJJRepo(path) {
		return vcs.TypeJJ, true
	}

	if isGitRepo(path) {
		return vcs.TypeGit, true
	}

	return "", false
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if file, err := os.Lstat(gitPath); err == nil {
		if file.IsDir() {
			return true
		}

		// git worktrees/submodules use a `.git` file containing `gitdir: ...`
		b, err := os.ReadFile(gitPath)
		if err != nil {
			return false
		}
		return strings.Contains(string(b), "gitdir:")
	}

	return false
}

func isJJRepo(path string) bool {
	jjPath := filepath.Join(path, ".jj")
	info, err := os.Lstat(jjPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func removeDuplicateRepoPaths(repoPaths []vcs.RepoPath) []vcs.RepoPath {
	seen := make(map[string]struct{}, len(repoPaths))
	distinct := make([]vcs.RepoPath, 0, len(repoPaths))

	for _, repoPath := range repoPaths {
		if _, ok := seen[repoPath.Path]; !ok {
			seen[repoPath.Path] = struct{}{}
			distinct = append(distinct, repoPath)
		}
	}

	return distinct
}
