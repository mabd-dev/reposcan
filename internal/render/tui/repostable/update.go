package repostable

import tea "charm.land/bubbletea/v2"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			if m.tbl.Cursor() == len(m.tbl.Rows())-1 {
				m.tbl.SetCursor(0)
				return m, nil
			}

		case "k":
			if m.tbl.Cursor() == 0 {
				m.tbl.SetCursor(len(m.tbl.Rows()) - 1)
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)

	return m, cmd
}
