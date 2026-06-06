package themeswitcher

import (
	"charm.land/bubbles/v2/table"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type Model struct {
	theme       theme.Theme
	schemeNames []string

	tbl table.Model

	filterFocused bool
	filterQuery   string
}

type ThemeSelectedMsg struct {
	SchemeName string
}
