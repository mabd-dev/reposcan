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
	m.updateCursorInRows()

	return m, cmd
}

func (m *Model) updateCursorInRows() {
	cursor := m.tbl.Cursor()
	rows := m.tbl.Rows()
	for i := range rows {
		if i == cursor {
			rows[i][0] = cursorChar
		} else {
			rows[i][0] = emptyChar
		}
	}
	m.tbl.SetRows(rows)
}
