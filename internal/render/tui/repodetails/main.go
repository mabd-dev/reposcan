package repodetails

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func New(
	worktreeState *common.WorktreeState,
	theme theme.Theme,
) Model {
	return Model{
		theme:         theme,
		worktreeState: worktreeState,
	}
}

func (m *Model) UpdateSize(height int) {
	m.height = height
}

func (m *Model) UpdateData(worktreeState *common.WorktreeState) {
	m.worktreeState = worktreeState
}

func (m Model) Init() tea.Cmd { return nil }
