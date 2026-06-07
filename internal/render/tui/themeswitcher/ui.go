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
	colorSchemas []colorSchemaData,
	t theme.Theme,
) []table.Row {
	rows := make([]table.Row, 0, len(colorSchemas))

	currentLabel := t.Styles.Base.
		Foreground(t.Colors.Success).
		Render("[current]")

	for _, s := range colorSchemas {
		colors := createColors(s.schema.Palette, t)
		name := s.schema.Name
		if name == t.Colors.Name {
			name = name + " " + currentLabel
		}
		rows = append(rows, table.Row{
			emptyChar,
			name,
			t.Styles.Base.Foreground(t.Colors.Muted).Render(s.schema.Variant),
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
