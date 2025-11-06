package repostable

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/render/tui/overlay"
)

func (m Model) View() string {
	body := m.theme.Styles.
		BoxFor(m.tbl.Focused()).
		Render(m.tbl.View())

	body = m.addIndicator(body)
	body = m.addSectionTitle(body)

	return body
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
