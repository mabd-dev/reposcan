package themeswitcher

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func createColumns() []table.Column {
	return []table.Column{
		{Title: "-", Width: 5},       // indicator
		{Title: "Name", Width: 40},   // colorScheme name
		{Title: "Colors", Width: 21}, // colors 7 colors each 3
	}
}

func createRows(
	colorSchemeNames []string,
	palettes []theme.Base24Palette,
	theme theme.Theme,
) []table.Row {
	rows := make([]table.Row, 0, len(palettes))

	for i, p := range palettes {
		colors := createColors(p, theme)
		rows = append(rows, table.Row{
			"",
			colorSchemeNames[i],
			colors,
		})
	}

	return rows
}

func createColors(palette theme.Base24Palette, theme theme.Theme) string {
	dots := make([]string, 7)

	colors := []string{
		palette.Base0D,
		palette.Base08,
		palette.Base09,
		palette.Base0A,
		palette.Base0B,
		palette.Base0C,
		palette.Base0E,
	}

	for i, color := range colors {
		dotStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(color))

		dots[i] = dotStyle.Render("● ")
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, dots...)
}
