package repostable

import (
	"github.com/mabd-dev/reposcan/pkg/report"
)

func getRepoIndex(repos []report.RepoState, id string) int {
	for i, s := range repos {
		if s.ID == id {
			return i
		}
	}
	return -1
}
