package tui

import (
	"github.com/mabd-dev/reposcan/pkg/report"
	"strings"
)

func getRepoIndex(repoIds []string, id string) int {
	for i, x := range repoIds {
		if x == id {
			return i
		}
	}
	return -1
}

func deleteRepo(repoIds []string, index int) []string {
	return append(repoIds[:index], repoIds[index+1:]...)
}

// filterRepos filter list of repos based on git repo name. Case insensitive
func filterRepos(repos []report.RepoState, query string) []report.RepoState {
	result := []report.RepoState{}

	q := strings.ToLower(strings.TrimSpace(query))
	if len(q) == 0 {
		return repos
	}

	for _, rs := range repos {
		if strings.Contains(strings.ToLower(rs.Repo), q) {
			result = append(result, rs)
		}
	}
	return result
}
