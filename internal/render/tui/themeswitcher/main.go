package themeswitcher

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func New(
	t theme.Theme,
) Model {
	textInput := createTextInput()
	model := Model{
		theme:     t,
		textInput: textInput,
	}

	schemesNames := theme.Schemes
	colorSchemes := make([]theme.Base24ColorSchema, len(schemesNames))

	for i, schemeName := range theme.Schemes {
		path := fmt.Sprintf("%v%v", theme.SchemesDir, schemeName)
		palette, err := theme.LoadBase24Schema(path)
		if err == nil {
			colorSchemes[i] = palette
		} else {
			logger.Error(fmt.Sprintf("failed to parse color scheme %v, error=%v", schemeName, err.Error()))
		}
	}

	cols := createColumns()
	rows := createRows(colorSchemes, t)

	tbl := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithWidth(totalTableWidth),
		table.WithHeight(20),
	)
	tbl.Focus()
	tbl.SetStyles(table.Styles{
		Header:   model.theme.Styles.TableHeader,
		Selected: model.theme.Styles.TableSelectedRow,
		Cell:     model.theme.Styles.TableRow,
	})
	model.schemeNames = schemesNames
	model.colorSchemes = colorSchemes
	model.tbl = tbl
	model.updateCursorInRows()

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

	m.tbl.SetStyles(table.Styles{
		Header:   m.theme.Styles.TableHeader,
		Selected: m.theme.Styles.TableSelectedRow,
		Cell:     m.theme.Styles.TableRow,
	})
	m.updateCursorInRows()
}
