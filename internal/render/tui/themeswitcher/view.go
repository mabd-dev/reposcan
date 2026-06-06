package themeswitcher

func (m Model) View() string {
	return m.theme.Styles.Box.Render("tetttt")
}
