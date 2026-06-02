package repodetails

import (
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func createRepoState(
	uncommitedFiles []string,
	stashes []string,
) report.RepoState {
	return report.RepoState{
		UncommitedFiles: uncommitedFiles,
		Stashes:         stashes,
	}
}

func createModel(
	selectedTabIndex int,
	repoState report.RepoState,
) Model {
	return Model{
		height:           100,
		theme:            theme.Theme{},
		repoState:        &repoState,
		selectedTabIndex: selectedTabIndex,
		tabs:             createTabsList(repoState),
	}
}
