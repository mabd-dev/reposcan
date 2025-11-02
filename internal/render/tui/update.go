package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.getFocusedModel().update(m, msg)
}

func defaultUpdate(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

		m.reposTable.UpdateWindowSize(
			m.width*sizeReposTableWidthPercent/100,
			m.height*sizeReposTableHeightPercent/100,
		)
		return m, nil

	case gitPushResultMsg:
		return m, gitRefreshRepo(m)

	case gitPullResultMsg:
		if len(msg.Err) != 0 {
			m.addWarning(msg.Err)
			return m, nil
		}

		rs := m.reposTable.GetCurrentRepoState()
		if rs == nil {
			return m, nil
		}

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

		rs := m.reposTable.GetCurrentRepoState()
		if rs == nil {
			return m, nil
		}

		index := getRepoIndex(m.reposBeingUpdated, rs.ID)
		if index != -1 {
			m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
		}

		return m, gitRefreshRepo(m)

	case gitRefreshRepoResultMsg:
		m.reposTable.UpdateRepoState(msg.index, msg.newRepoState)

		return m, nil

	case alerts.AddAlertMsg, alerts.TickMsg:
		var cmd tea.Cmd
		m.alerts, cmd = m.alerts.Update(msg)
		return m, cmd
	}

	return nil, nil
}
