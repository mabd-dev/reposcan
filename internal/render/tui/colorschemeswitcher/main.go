package colorschemeswitcher

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func New(t theme.Theme, height int) Model {
	colorSchemes := make([]colorSchemeData, 0, len(theme.Schemes))
	for _, schemeID := range theme.Schemes {
		path := fmt.Sprintf("%v%v", theme.SchemesDir, schemeID)
		schema, err := theme.LoadBase24Schema(path)
		if err == nil {
			colorSchemes = append(colorSchemes, colorSchemeData{schemeID, schema})
		} else {
			logger.Error(fmt.Sprintf("failed to parse color scheme %v, error=%v", schemeID, err.Error()))
		}
	}

	maxCounterWidth := len(fmt.Sprintf("%v/%v", len(colorSchemes), len(colorSchemes)))
	tiWidth := totalTableWidth - maxCounterWidth
	textInput := createTextInput(tiWidth)

	cols := createColumns()
	rows := createRows(0, colorSchemes, t)

	tableHeight := max(5, height-10)
	tbl := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithWidth(totalTableWidth),
		table.WithHeight(tableHeight),
	)
	tbl.Focus()

	tbl.SetStyles(getTableStyles(t))

	model := Model{
		theme:                t,
		textInput:            textInput,
		tbl:                  tbl,
		colorSchemes:         colorSchemes,
		filteredColorSchemes: colorSchemes,
		selectedSchemeID:     t.Colors.ID,
	}
	return model
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) WantsClose() bool         { return m.wantsClose }
func (m Model) SelectedSchemeID() string { return m.selectedSchemeID }

func (m *Model) Reset() {
	m.wantsClose = false
}

func createTextInput(width int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "search schemes..."
	ti.CharLimit = min(width, 100)
	ti.SetWidth(width)
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
