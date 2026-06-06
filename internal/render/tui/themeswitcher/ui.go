package themeswitcher

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
)

const (
	cursorChar   = "›"
	selectedChar = "✓"
	emptyChar    = " "
)

const (
	indicatorWidth int = 2
	nameWidth          = 40
	variantWidth       = 6
	colorsWidth        = 14
	totalWidth         = indicatorWidth + nameWidth + variantWidth + colorsWidth
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
	palettes []theme.Base24ColorSchema,
	t theme.Theme,
) []table.Row {
	rows := make([]table.Row, 0, len(palettes))

	currentLabel := t.Styles.Base.
		Foreground(t.Colors.Success).
		Render("[current]")

	for _, p := range palettes {
		colors := createColors(p.Palette, t)
		name := p.Name
		if name == t.Colors.Name {
			name = name + " " + currentLabel
		}
		rows = append(rows, table.Row{
			emptyChar,
			name,
			p.Variant,
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
