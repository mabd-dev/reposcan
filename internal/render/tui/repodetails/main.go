package repodetails

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func New(
	repoState *report.RepoState,
	theme theme.Theme,
) Model {
	return Model{
		theme:     theme,
		repoState: repoState,
	}
}

func (m *Model) UpdateSize(height int) {
	m.height = height
}

func (m *Model) UpdateData(repoState *report.RepoState) {
	m.repoState = repoState
}

func (m Model) Init() tea.Cmd { return nil }
