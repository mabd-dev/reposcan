package themeswitcher

import (
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type Model struct {
	theme        theme.Theme
	schemeNames  []string
	colorSchemes []theme.Base24ColorSchema

	tbl       table.Model
	textInput textinput.Model

	wantsClose         bool
	selectedSchemeName string
}

type ThemeSelectedMsg struct {
	SchemeName string
}
