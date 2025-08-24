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
func FindGitRepos(roots []string) (gitRepos []string, warnings []string) {

	//gitRepos := []string{}
	//warnings := []string{}

	for _, root := range roots {
		root = os.ExpandEnv(root)

		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// TODO: return warnings back
				// possible errors: permission denied
				warnings = append(warnings, err.Error())
				// fmt.Println("Warning: " + err.Error())
				return nil
			}

			if !d.IsDir() {
				return nil
			}

			if isGitRepo(path) {
				gitRepos = append(gitRepos, path)
				return fs.SkipDir
			}
			return nil
		})
	}

	return removeDuplicates(gitRepos), warnings

}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	if file, err := os.Lstat(gitPath); err == nil {
		if file.IsDir() {
			return true
		} else {
			// git worktrees has gitdir: folder
			b, err := os.ReadFile(path)
			if err != nil {
				return false
			}
			return strings.Contains(string(b), "gitdir:")
		}
	}
	return false
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
