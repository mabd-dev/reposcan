package repodetails

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func New(
	repoState *report.RepoState,
	theme theme.Theme,
) Model {
	return Model{
		theme:            theme,
		repoState:        repoState,
		tabs:             createTabsList(*repoState),
		selectedTabIndex: 0,
	}
}

func (m *Model) UpdateSize(height int) {
	m.height = height
}

func (m *Model) UpdateData(repoState *report.RepoState) {
	m.repoState = repoState
	m.tabs = createTabsList(*repoState)
}

func (m Model) Init() tea.Cmd { return nil }

func (m *Model) UpdateTheme(newTheme theme.Theme) {
	m.theme = newTheme
}

func createTabsList(repoState report.RepoState) []tab {
	return []tab{
		{
			key:             tabChanges,
			name:            "Changes",
			highlightedText: fmt.Sprintf("(%v)", len(repoState.UncommitedFiles)),
		},
		{
			key:             tabStashes,
			name:            "Stashes",
			highlightedText: fmt.Sprintf("(%v)", len(repoState.Stashes)),
		},
	}
}
