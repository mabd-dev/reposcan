package repodetails

import (
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type tabKey string

const (
	tabChanges tabKey = "changes"
	tabStashes        = "stashes"
)

type tab struct {
	key             tabKey
	name            string
	highlightedText string
}

type Model struct {
	height int
	theme  theme.Theme

	repoState *report.RepoState

	tabs        []tab
	selectedTab tabKey
}
