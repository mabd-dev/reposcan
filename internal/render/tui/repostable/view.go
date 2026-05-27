package repostable

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/render/tui/overlay"
)

func (m Model) View() string {
	// When no repos match the current filter, show a friendly empty-state
	// message instead of an empty table. See issue #41.
	if m.ReposCount() == 0 {
		return m.renderEmptyState()
	}

	body := m.theme.Styles.
		BoxFor(m.tbl.Focused()).
		Render(m.tbl.View())

	body = m.addIndicator(body)
	body = m.addSectionTitle(body)

	return body
}

// renderEmptyState returns a styled, centered message when the repo list is empty.
func (m Model) renderEmptyState() string {
	msg := "✨ No repositories found — your workspace is spotless!"
	styled := m.theme.Styles.Base.
		Foreground(m.theme.Colors.Accent).
		Render(msg)

	// Center the message horizontally and vertically inside the box area.
	// Use the table height as the target so the layout stays consistent.
	centered := lipgloss.Place(
		m.width-4,  // account for box borders + padding
		m.height-2, // account for box borders
		lipgloss.Center,
		lipgloss.Center,
		styled,
	)

	return m.theme.Styles.
		BoxFor(m.tbl.Focused()).
		Render(centered)
}

func (m Model) addIndicator(body string) string {
	var repoIndicator string
	if m.ReposCount() == 0 {
		repoIndicator = "0/0"
	} else {
		repoIndicator = strconv.Itoa(m.Cursor()+1) + "/" + strconv.Itoa(m.ReposCount())
	}

	return overlay.PlaceOverlayWithPositionAndPadding(
		overlay.OverlayPositionBottomRight,
		lipgloss.Width(body), lipgloss.Height(body),
		2, 0,
		repoIndicator, body,
		false,
		overlay.WithWhitespaceChars(" "),
	)
}

func (m Model) addSectionTitle(body string) string {
	return overlay.PlaceOverlayWithPositionAndPadding(
		overlay.OverlayPositionTopLeft,
		lipgloss.Width(body), lipgloss.Height(body),
		2, 0,
		"repos", body,
		false,
		overlay.WithWhitespaceChars(" "),
	)
}
