package themeswitcher

import tea "charm.land/bubbletea/v2"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)

	return m, cmd
}
