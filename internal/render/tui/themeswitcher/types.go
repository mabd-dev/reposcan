package themeswitcher

import (
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type Model struct {
	theme       theme.Theme
	schemeNames []string

	tbl table.Model

	textInput textinput.Model
}

type ThemeSelectedMsg struct {
	SchemeName string
}
