package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/render/tui/overlay"
)

func (m Model) View() string {
	if m.loading {
		return "Loading..."
	}

	footer := m.getFooterView()

	// Calculate heights
	footerHeight := lipgloss.Height(footer)
	bodyHeight := m.height - footerHeight

	reposTableHeight := bodyHeight * sizeReposTableHeightPercent / 100
	m.reposTable = m.reposTable.UpdateWindowSize(m.width, reposTableHeight)
	reposTable := m.reposTable.View()

	m.repoDetails.UpdateData(m.reposTable.GetCurrentRepoState())
	m.repoDetails.UpdateSize(bodyHeight - reposTableHeight)
	reposDetails := m.repoDetails.View()

	body := lipgloss.JoinVertical(lipgloss.Left, reposTable, reposDetails)
	body = lipgloss.NewStyle().
		Height(bodyHeight).
		MaxHeight(bodyHeight).
		Render(body)

	view := lipgloss.JoinVertical(lipgloss.Left, body, footer)

	view = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(view)

	view = m.renderAlerts(view, m.alerts.AlertStates(m.width, m.height))

	if m.currentFocus() == FocusKeybindingPopup {
		helpView := generateHelpPopup(m.theme, reposTableKeybindings)

		view = overlay.PlaceOverlayWithPosition(
			overlay.OverlayPositionCenter,
			m.width, m.height,
			helpView, view,
			true,
			overlay.WithWhitespaceChars(" "), // fill empty space
		)
	}

	return view
}

func (m *Model) getFooterView() string {
	if m.IsReposFilterVisible() {
		//if m.reposFilter.show {
		return m.theme.Styles.Base.
			Foreground(m.theme.Colors.Foreground).
			Render(m.reposFilter.View())
	}

	return m.generateKeybindingsFooterView()
}

func (m *Model) generateKeybindingsFooterView() string {
	keybindings := m.keybindings()

	addKeybinding := false
	switch m.currentFocus() {
	case FocusReposTable:
		addKeybinding = true
	}

	if addKeybinding {
		keybindings = append(keybindings, common.Keybinding{
			Key:         "?",
			Description: "More Keybindings",
			ShortDesc:   "Keybindings",
		})
	}

	kbStyle := m.theme.Styles.Base.Foreground(m.theme.Colors.Foreground)
	mutedStyle := m.theme.Styles.Muted

	var sb strings.Builder
	for i, kb := range keybindings {
		sb.WriteString(mutedStyle.Render(kb.ShortDesc))
		sb.WriteString(mutedStyle.Render(": "))
		sb.WriteString(kbStyle.Render(kb.Key))

		if i < len(keybindings)-1 {
			sb.WriteString(mutedStyle.Render(" | "))
		}
	}

	return m.theme.Styles.Muted.Render(sb.String())
}

// renderAlerts take list of alerts, calculate each alert y position and render it (it it's visible). Overlay each alert on top of main [view] (bg view)
func (m *Model) renderAlerts(
	view string,
	alertStates []alerts.AlertState,
) string {
	if len(alertStates) == 0 {
		return view
	}

	for _, alert := range alertStates {
		if alert.IsVisible {
			view = overlay.PlaceOverlay(
				alert.X,
				alert.Y,
				alert.AlertView,
				view,
				false,
				overlay.WithWhitespaceChars(" "),
			)
		}
	}
	return view
}
