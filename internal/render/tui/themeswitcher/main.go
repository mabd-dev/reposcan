package themeswitcher

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/theme"
)

func New(
	t theme.Theme,
) Model {
	model := Model{
		theme: t,
	}

	schemesNames := theme.Schemes
	colorSchemes := make([]theme.Base24Palette, len(schemesNames))

	for i, schemeName := range schemesNames {
		path := fmt.Sprintf("%v%v", theme.SchemesDir, schemeName)
		palette, err := theme.LoadBase24Schema(path)
		if err == nil {
			colorSchemes[i] = palette.Palette
		} else {
			logger.Error(fmt.Sprintf("failed to parse color scheme %v, error=%v", schemeName, err.Error()))
		}
	}

	cols := createColumns()
	rows := createRows(schemesNames, colorSchemes, t)

	tbl := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithWidth(66),
		table.WithHeight(20),
	)
	tbl.Focus()
	tbl.SetStyles(table.Styles{
		Header:   model.theme.Styles.TableHeader,
		Selected: model.theme.Styles.TableSelectedRow,
		Cell:     model.theme.Styles.TableRow,
	})
	model.tbl = tbl

	return model
}

func (m Model) Init() tea.Cmd { return nil }
