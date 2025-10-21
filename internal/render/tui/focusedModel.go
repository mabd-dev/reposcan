package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"golang.design/x/clipboard"
)

// focusModel defined how each group of ui-elements handles tui.Update function
type focusedModel interface {
	update(m Model, msg tea.Msg) (tea.Model, tea.Cmd)
	keybindings() []common.Keybinding
}

type popupFM struct{}

func (r popupFM) update(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.reposTable.Focus()
			m.showHelp = false
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

func (r popupFM) keybindings() []common.Keybinding {
	return helpPopupKeybindings
}

type reposTableFM struct{}

func (r reposTableFM) update(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		// case "p":
		// 	return m, gitPull(m)
		// case "P":
		// 	return m, gitPush(m)
		// case "f":
		// 	return m, gitFetch(m)
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
		case "/":
			m.reposFilter.show = true
			m.reposFilter.textInput.Focus()
			m.reposTable.Blur()
			return m, nil
		case "?":
			m.showHelp = true
			m.reposTable.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	nm, cmd := defaultUpdate(m, msg)

	if nm != nil {
		return nm, cmd
	}

	m.reposTable, cmd = m.reposTable.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.repoDetails.UpdateData(m.reposTable.GetCurrentRepoState())

		case "k", "up":
			m.repoDetails.UpdateData(m.reposTable.GetCurrentRepoState())
		}
	}
	return m, cmd
}

func (r reposTableFM) keybindings() []common.Keybinding {
	return reposTableKeybindings
}

type reposFilterTextFieldFM struct{}

func (r reposFilterTextFieldFM) update(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			// refresh to original state
			m.reposFilter.show = false
			m.reposFilter.textInput.SetValue("")

			m.reposTable.Filter("")
			m.reposTable.Focus()
			m.repoDetails.UpdateData(m.reposTable.GetCurrentRepoState())

			return m, nil
		case "enter":
			if len(strings.TrimSpace(m.reposFilter.textInput.Value())) == 0 {
				m.reposFilter.show = false
			}

			m.reposFilter.textInput.Blur()
			m.reposTable.Focus()
			m.repoDetails.UpdateData(m.reposTable.GetCurrentRepoState())

			return m, nil
		}
	}

	// on each keystorke, update repos list
	var cmd tea.Cmd
	m.reposFilter.textInput, cmd = m.reposFilter.textInput.Update(msg)

	m.reposTable.Filter(m.reposFilter.textInput.Value())
	m.repoDetails.UpdateData(m.reposTable.GetCurrentRepoState())

	return m, cmd
}

func (r reposFilterTextFieldFM) keybindings() []common.Keybinding {
	return reposTableFilterKeybindings
}
