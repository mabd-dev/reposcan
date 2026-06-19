package themeswitcher

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func New(
	t theme.Theme,
) Model {
	textInput := createTextInput()

	schemesNames := theme.Schemes
	colorSchemes := make([]colorSchemeData, len(schemesNames))

	for i, schemeName := range theme.Schemes {
		path := fmt.Sprintf("%v%v", theme.SchemesDir, schemeName)
		schema, err := theme.LoadBase24Schema(path)
		if err == nil {
			colorSchemes[i] = colorSchemeData{schemeName, schema}
		} else {
			logger.Error(fmt.Sprintf("failed to parse color scheme %v, error=%v", schemeName, err.Error()))
		}
	}

	cols := createColumns()
	rows := createRows(0, colorSchemes, t)

	tbl := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithWidth(totalTableWidth),
		table.WithHeight(20),
	)
	tbl.Focus()

	tbl.SetStyles(getTableStyles(t))

	model := Model{
		theme:                t,
		textInput:            textInput,
		tbl:                  tbl,
		colorSchemes:         colorSchemes,
		filteredColorSchemes: colorSchemes,
		selectedSchemeName:   t.Colors.Name,
	}
	return model
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) WantsClose() bool           { return m.wantsClose }
func (m Model) SelectedSchemeName() string { return m.selectedSchemeName }

func (m *Model) Reset() {
	m.wantsClose = false
}

func createTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "search schemes..."
	ti.CharLimit = min(totalTableWidth, 100)
	ti.SetWidth(totalTableWidth)
	return ti
}

func (m *Model) UpdateTheme(newTheme theme.Theme) {
	m.theme = newTheme

	m.tbl.SetStyles(getTableStyles(newTheme))
	m.tbl.SetRows(createRows(m.tbl.Cursor(), m.filteredColorSchemes, m.theme))
}

func getTableStyles(t theme.Theme) table.Styles {
	return table.Styles{
		Header:   t.Styles.TableHeader,
		Selected: lipgloss.NewStyle().Foreground(t.Colors.Accent).Bold(true),
		Cell:     t.Styles.TableRow,
	}
}
