package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func defaultUpdate(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.contentHeight = max(6, m.height-6) // leave room for title+footer
		m.tbl.SetHeight(min(18, m.contentHeight))
		cols := createColumns(m.width)
		m.tbl.SetColumns(cols)
		return m, nil

	case gitPushResultMsg:
		return m, gitRefreshRepo(m)

	case gitPullResultMsg:
		if len(msg.Err) != 0 {
			m.addWarning(msg.Err)
			return m, nil
		}

		rs := m.getReportAtCursor()

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

		rs := m.getReportAtCursor()

		index := getRepoIndex(m.reposBeingUpdated, rs.ID)
		if index != -1 {
			m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
		}

		return m, gitRefreshRepo(m)

	case gitRefreshRepoResultMsg:
		m.report.RepoStates[msg.index] = msg.newRepoState

		rows := createRows(m.report.RepoStates)
		m.tbl.SetRows(rows)

		return m, nil
	}

	return nil, nil
}
