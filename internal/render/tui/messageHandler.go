package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

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
		// Clear warning on each keybinding click
		m.warnings = []string{}

		if m.showHelp {
			return handlePopupKeymaps(m, msg)
		} else {
			return handleRepoTableKeymaps(m, msg)
		}

	// Git functions

	case gitPushResultMsg:
		return m, gitRefreshRepo(m)

	case gitPullResultMsg:
		if len(msg.Err) != 0 {
			m.addWarning(msg.Err)
			return m, nil
		}

		idx := m.tbl.Cursor()
		rs := m.report.RepoStates[idx]

		index := getRepoIndex(m.reposBeingUpdated, rs.ID)
		if index != -1 {
			m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
		}
		return m, gitRefreshRepo(m)

	case gitFetchResultMsg:
		if len(msg.Err) != 0 {
			m.addWarning(msg.Err)
			return m, nil
		}

		idx := m.tbl.Cursor()
		rs := m.report.RepoStates[idx]

		index := getRepoIndex(m.reposBeingUpdated, rs.ID)
		if index != -1 {
			m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
		}

		return m, gitRefreshRepo(m)

	case gitRefreshRepoResultMsg:
		m.report.RepoStates[msg.index] = msg.newRepoState

		rows := createRows(m.report)
		m.tbl.SetRows(rows)

		return m, nil
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func handlePopupKeymaps(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.showHelp = false
		return m, nil
	}
	return defaultUpdateReturn(m, msg)
}

func handleRepoTableKeymaps(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "enter":
		m.showDetails = !m.showDetails
		return m, nil
	case "p":
		return m, gitPull(m)
	case "P":
		return m, gitPush(m)
	case "f":
		return m, gitFetch(m)
	case "?":
		m.showHelp = true
		return m, nil
	}
	return defaultUpdateReturn(m, msg)
}

func defaultUpdateReturn(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}
