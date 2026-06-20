package colorschemeswitcher

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	styles := m.theme.Styles
	colors := m.theme.Colors

	// header
	header := styles.Base.Italic(true).PaddingLeft(1).Width(totalTableWidth).Align(lipgloss.Center).Render("Themes")

	// text input + counter
	counter := styles.Base.Foreground(colors.Muted).Render(fmt.Sprintf("%v/%v", len(m.tbl.Rows()), len(m.colorSchemes)))

	textInputView := styles.BoxFor(m.textInput.Focused()).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.textInput.View(),
			counter,
		),
	)

	// putting everything together
	return styles.Box.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		textInputView,
		m.tbl.View(),
	))
}
