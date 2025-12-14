package gitx

import "strings"

func getWorktreesPaths(path string) ([]string, error) {
	str, err := RunGitCommand(path, "worktree", "list", "--porcelain")
	if err != nil {
		return []string{}, nil
	}

	paths := []string{}
	lines := strings.SplitSeq(str, "\n")
	for line := range lines {
		path, found := strings.CutPrefix(line, "worktree ")
		if found {
			paths = append(paths, path)
		}
	}
	return paths, nil
}
