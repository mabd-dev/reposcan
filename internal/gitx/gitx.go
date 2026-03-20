package gitx

import (
	"fmt"
	"strings"

	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// CheckRepoState inspects the Git repository at path and returns its RepoState
// along with any non-fatal warnings encountered while collecting information.
func CheckRepoState(path string) (repoState report.RepoState, warnings []string) {

	branch, err := GetRepoBranch(path)
	if err != nil {
		msg := "Failed to get branch name, path=" + path
		warnings = append(warnings, msg)
	}

	remotes, err := GetGitRemotes(path)
	if err != nil {
		msg := "Failed to get git remotes, path=" + path
		warnings = append(warnings, msg)
	}

	remoteStatuses := []report.RemoteStatus{}

	if len(remotes) == 0 {
		remoteStatuses = append(remoteStatuses, report.RemoteStatus{
			Remote: "",
			Ahead:  -1,
			Behind: -1,
		})
	}

	for _, remote := range remotes {
		remoteStatus, err := GetUpstreamStatusForAllRemotes(path, remote, branch)
		if err != nil {
			msg := fmt.Sprintf("Failed to get upstream status for remote=%s, path=%s", remote, path)
			warnings = append(warnings, msg)
			remoteStatuses = append(remoteStatuses, report.RemoteStatus{
				Remote: remote,
				Ahead:  -1,
				Behind: -1,
			})
		} else {
			remoteStatuses = append(remoteStatuses, report.RemoteStatus{
				Remote: remote,
				Ahead:  remoteStatus.Ahead,
				Behind: remoteStatus.Behind,
			})
		}
	}

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

	return report.RepoState{
		ID:              utils.Hash(path),
		Path:            path,
		Repo:            repoName,
		VCSType:         string(vcs.TypeGit),
		Branch:          branch,
		UncommitedFiles: uncommitedFiles,
		RemoteStatus:    remoteStatuses,
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
