package repostable

func (rt Model) View() string {
	return rt.theme.Styles.BoxFor(rt.tbl.Focused()).Render(rt.tbl.View())
}
