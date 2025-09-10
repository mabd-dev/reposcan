package gitx

import (
	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/mabd-dev/reposcan/pkg/report"
	"strings"
)

// CheckRepoState inspects the Git repository at path and returns its RepoState
// along with any non-fatal warnings encountered while collecting information.
func CheckRepoState(path string) (repoState report.RepoState, warnings []string) {
	repoName, err := GetRepoName(path)
	if err != nil {
		msg := "Failed to get repo name, path=" + path
		warnings = append(warnings, msg)
	}

	uncommitedFiles, err := GetUncommitedFiles(path)
	if err != nil {
		msg := "Failed to get uncommited files, path=" + path
		warnings = append(warnings, msg)
	}

	branch, err := GetRepoBranch(path)
	if err != nil {
		msg := "Failed to get branch name, path=" + path
		warnings = append(warnings, msg)
	}

	ahead, behind, err := GetUpstreamStatus(path)
	if err != nil {
		msg := "Failed to get upstream status, path=" + path
		warnings = append(warnings, msg)
	}

	return report.RepoState{
		ID:              utils.Hash(path),
		Path:            path,
		Repo:            repoName,
		Branch:          branch,
		UncommitedFiles: uncommitedFiles,
		Ahead:           ahead,
		Behind:          behind,
	}, warnings
}

func removeEmptyStrings(input []string) []string {
	result := []string{}
	for _, s := range input {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

func atoiSafe(s string) int {
	var n int
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
