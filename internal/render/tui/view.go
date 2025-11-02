package tui

import (
	"strconv"
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

	body := m.reposTable.View()

	var repoIndicator string
	if m.reposTable.ReposCount() == 0 {
		repoIndicator = "0/0"
	} else {
		repoIndicator = strconv.Itoa(m.reposTable.Cursor()+1) + "/" + strconv.Itoa(m.reposTable.ReposCount())
	}

	body = overlay.PlaceOverlayWithPositionAndPadding(
		overlay.OverlayPositionBottomRight,
		lipgloss.Width(body), lipgloss.Height(body),
		2, 0,
		repoIndicator, body,
		false,
		overlay.WithWhitespaceChars(" "),
	)

	body = lipgloss.JoinVertical(lipgloss.Left, body, m.repoDetails.View())

	var footer string
	if m.reposFilter.show {
		textfieldStr := m.theme.Styles.Base.
			Foreground(m.theme.Colors.Foreground).
			Render(m.reposFilter.textInput.View())

		footer = textfieldStr
	} else {
		footer = m.generateFooter()
	}

	// Calculate heights
	footerHeight := lipgloss.Height(footer)
	availableHeight := m.height - footerHeight

	body = lipgloss.NewStyle().
		Height(availableHeight).
		MaxHeight(availableHeight).
		Render(body)

	view := lipgloss.JoinVertical(lipgloss.Left,
		//header,
		body,
		footer,
	)

	view = lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		// Background(m.theme.Colors.Background).
		Render(view)

	view = m.renderAlerts(view, m.alerts.AlertStates(m.width, m.height))

	if m.showHelp {
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

func (m *Model) generateFooter() string {
	focusedModel := m.getFocusedModel()
	keybindings := focusedModel.keybindings()

	addKeybinding := false
	switch focusedModel.(type) {
	case reposTableFM:
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
