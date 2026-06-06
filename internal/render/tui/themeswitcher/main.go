package themeswitcher

import (
	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func New(
	theme theme.Theme,
) Model {
	model := Model{
		theme: theme,
	}

	return model
}

func (m Model) Init() tea.Cmd { return nil }
