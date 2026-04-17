package repostable

import (
	"slices"

	"github.com/charmbracelet/bubbles/table"
	"github.com/mabd-dev/reposcan/internal/ds/mmap"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type Model struct {
	width  int
	height int
	theme  theme.Theme

	tbl table.Model

	allRows      []tableRow
	filteredRows []tableRow

	allGroups      []worktreesGroup
	filteredGroups []worktreesGroup

	allWorktreeStates []common.WorktreeState
	filterQuery       string
}

type worktreesGroup struct {
	repoName  string
	worktrees []tableRow
}

type tableRow struct {
	Repo     string
	Branch   string
	State    string
	IsHeader bool
	WtIndex  int
}

func convertToGroups(
	worktreeStates []common.WorktreeState,
	theme theme.Theme,
) (allGroups []worktreesGroup) {

	wtIndex := 0

	worktreesMap := toMap(worktreeStates)
	sortedRepoNames := mmap.Keys(worktreesMap)
	slices.Sort(sortedRepoNames)

	for _, repoName := range sortedRepoNames {
		wtStates := worktreesMap[repoName]

		rows := []tableRow{}

		for i, wt := range wtStates {
			rows = append(rows, tableRow{
				Repo:     generateRepoColumn(wt, i, len(wtStates)),
				Branch:   wt.Branch,
				State:    getStateColumnStr(wt, theme),
				IsHeader: false,
				WtIndex:  wtIndex,
			})
			wtIndex++
		}

		allGroups = append(allGroups, worktreesGroup{
			repoName:  repoName,
			worktrees: rows,
		})
	}

	return allGroups
}

func toMap(worktreeStates []common.WorktreeState) map[string][]common.WorktreeState {
	worktreesMap := map[string][]common.WorktreeState{}

	for _, wt := range worktreeStates {
		worktreesMap[wt.RepoName] = append(worktreesMap[wt.RepoName], wt)
	}
	return worktreesMap
}

func generateRepoColumn(wt common.WorktreeState, i, totalRows int) string {
	col := wt.RepoName
	if totalRows > 1 {
		if i < totalRows-1 { // last index
			col = "  ├ " + wt.WorktreeName
		} else {
			col = "  ┕ " + wt.WorktreeName
		}
	}
	return col
}
