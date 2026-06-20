package colorschemeswitcher

import (
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type colorSchemeData struct {
	id     string
	scheme theme.Base24Scheme
}

type Model struct {
	theme                theme.Theme
	colorSchemes         []colorSchemeData
	filteredColorSchemes []colorSchemeData

	tbl       table.Model
	textInput textinput.Model

	wantsClose         bool
	selectedSchemeName string
}

type ThemeSelectedMsg struct {
	SchemeName string
}
