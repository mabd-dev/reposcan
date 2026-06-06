package themeswitcher

import "charm.land/lipgloss/v2"

func (m Model) View() string {
	styles := m.theme.Styles
	// colors := m.theme.Colors

	header := styles.Base.Italic(true).PaddingLeft(1).Width(totalTableWidth).Align(lipgloss.Center).Render("Themes")

	textInputView := styles.BoxFor(m.textInput.Focused()).Render(m.textInput.View())

	return styles.Box.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		textInputView,
		m.tbl.View(),
	))
}
