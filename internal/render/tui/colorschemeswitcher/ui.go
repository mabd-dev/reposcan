package colorschemeswitcher

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
)

const (
	cursorChar = "›"
	emptyChar  = " "
)

const (
	indicatorWidth  int = 2
	nameWidth           = 40
	variantWidth        = 6
	colorsWidth         = 14
	totalTableWidth     = indicatorWidth + nameWidth + variantWidth + colorsWidth
)

func createColumns() []table.Column {
	return []table.Column{
		{Title: " ", Width: indicatorWidth},
		{Title: "", Width: nameWidth},
		{Title: "", Width: variantWidth},
		{Title: "", Width: colorsWidth},
	}
}

func createRows(
	selectedIndex int,
	colorSchemes []colorSchemeData,
	t theme.Theme,
) []table.Row {
	rows := make([]table.Row, 0, len(colorSchemes))

	currentLabel := t.Styles.Base.
		Foreground(t.Colors.Success).
		Render("[current]")

	for i, s := range colorSchemes {
		colors := createColors(s.scheme.Palette, t)
		name := s.scheme.Name
		if name == t.Colors.Name {
			name = name + " " + currentLabel
		}

		cursorIndicator := emptyChar
		if i == selectedIndex {
			cursorIndicator = cursorChar
		}

		rows = append(rows, table.Row{
			cursorIndicator,
			name,
			t.Styles.Base.Foreground(t.Colors.Muted).Render(s.scheme.Variant),
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
