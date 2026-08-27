package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/theme"
	"golang.design/x/clipboard"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.currentFocus() {
	case FocusReposTable:
		return m.updateReposTable(msg)
	case FocusReposFilter:
		return m.updateReposFilter(msg)
	case FocusHelpPopup:
		return m.keybindingPopup(msg)
	case FocusThemeSwitcher:
		return m.updateColorSchemeSwitcher(msg)
	}
	return m, nil
}

func (m Model) updateReposTable(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		// case "p":
		// 	return m.runVCSAction(vcsActionPull)
		// case "P":
		// 	return m.runVCSAction(vcsActionPush)
		// case "f":
		// 	return m.runVCSAction(vcsActionFetch)
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "c":
			rs := m.reposTable.GetCurrentRepoState()
			if rs == nil {
				return m, nil
			}

			path := shellEscapePath(rs.Path)
			clipboard.Write(clipboard.FmtText, []byte(path))

			return m, func() tea.Msg {
				return alerts.AddAlertMsg{
					Msg: alerts.Alert{
						Type:    alerts.AlertTypeInfo,
						Title:   "",
						Message: "Path copied to clipboard",
					},
				}
			}
		case "right", "l", "left", "h":
			var cmd tea.Cmd
			m.repoDetails, cmd = m.repoDetails.Update(msg)
			return m, cmd
		case "r":
			m.loading = true
			request := generateReport{configs: m.configs}
			return m, request.Cmd()
		case "ctrl+t":
			m.pushFocus(FocusThemeSwitcher)
			return m, nil
		case "/":
			m.pushFocus(FocusReposFilter)
			return m, nil
		case "?":
			m.pushFocus(FocusHelpPopup)
			return m, nil
		}
	}

	var cmd tea.Cmd
	nm, cmd := defaultUpdate(m, msg)

	if nm != nil {
		return nm, cmd
	}

	m.reposTable, cmd = m.reposTable.Update(msg)
	return m, cmd
}

func (m Model) updateReposFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.popFocus(true)

			return m, nil
		case "enter":
			emptyQuery := len(strings.TrimSpace(m.reposFilter.Value())) == 0

			m.popFocus(emptyQuery)

			return m, nil
		}
	}

	// on each keystorke, update repos list
	var cmd tea.Cmd
	m.reposFilter, cmd = m.reposFilter.Update(msg)

	m.reposTable.Filter(m.reposFilter.Value())

	return m, cmd
}

func (m Model) updateColorSchemeSwitcher(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.colorSchemeSwitcher, cmd = m.colorSchemeSwitcher.Update(msg)

	if m.colorSchemeSwitcher.WantsClose() {
		m.popFocus(true)
		return m, nil
	}

	if schemeID := m.colorSchemeSwitcher.SelectedSchemeID(); schemeID != "" && schemeID != m.theme.Colors.ID {
		newColors, err := theme.CreateColors(schemeID)
		if err != nil {
			logger.Error("Failed to create new theme from name '%v': %v", schemeID, err)
			return m, nil
		}
		cfgFullPath, err := config.DefaultPaths().GetConfigFullPath()
		if err != nil {
			logger.Error("Failed to get config full path : %v", err)
			return m, nil
		}
		oldColorSchemeID := m.theme.Colors.ID
		m.configs.Output.ColorSchemeName = schemeID
		err = config.UpdateConfigs(m.configs, cfgFullPath)
		if err != nil {
			m.configs.Output.ColorSchemeName = oldColorSchemeID
			logger.Error("failed to save new configs in config.toml: %v", err)
			return m, nil
		}

		m.theme = theme.Theme{
			Colors: newColors,
			Styles: theme.CreateStyles(newColors),
		}
		m.reposTable.UpdateTheme(m.theme)
		m.repoDetails.UpdateTheme(m.theme)
		m.alerts.UpdateTheme(m.theme)
		m.colorSchemeSwitcher.UpdateTheme(m.theme)
		return m, nil
	}

	return m, cmd
}

func (m Model) keybindingPopup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc":
			m.popFocus(true)
			return m, nil
		}
	}

	var cmd tea.Cmd
	nm, cmd := defaultUpdate(m, msg)

	if nm != nil {
		return nm, cmd
	}
	return m, nil
}

func defaultUpdate(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case vcsActionResultMsg:
		if len(msg.Err) != 0 {
			logger.Warn(msg.Err)
			m.removeRepoBeingUpdated(msg.RepoID)
			return m, func() tea.Msg {
				return alerts.AddAlertMsg{
					Msg: alerts.Alert{
						Type:    alerts.MsgTypeError,
						Title:   "Action failed",
						Message: msg.Err,
					},
				}
			}
		}

		m.removeRepoBeingUpdated(msg.RepoID)
		return m, refreshRepo(m, msg.Index)

	case vcsRefreshRepoResultMsg:
		m.reposTable.UpdateRepoState(msg.index, msg.newRepoState)

		return m, nil
	case generateReportResponse:
		m.loading = false
		m.reposTable.SetReport(msg.report)
		return m, nil

	case alerts.AddAlertMsg, alerts.TickMsg:
		var cmd tea.Cmd
		m.alerts, cmd = m.alerts.Update(msg)
		return m, cmd
	}

	return nil, nil
}

func (m *Model) removeRepoBeingUpdated(repoID string) {
	index := getRepoIndex(m.reposBeingUpdated, repoID)
	if index != -1 {
		m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
	}
}
