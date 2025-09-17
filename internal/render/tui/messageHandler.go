package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type MsgHandler interface {
	updateUi(m Model) Model
}

func handleMsg(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.contentHeight = max(6, m.height-6) // leave room for title+footer
		m.tbl.SetHeight(min(18, m.contentHeight))
		cols := createColumns(m.width)
		m.tbl.SetColumns(cols)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.showDetails = !m.showDetails
			return m, nil
		case "p":
			return m, gitPush(m)
		case "P":
			return m, gitPull(m)
		case "f":
			return m, gitFetch(m)
		}

	case MsgHandler:
		m = msg.updateUi(m)
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}
