package themeswitcher

import (
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"github.com/mabd-dev/reposcan/internal/theme"
)

type colorSchemaData struct {
	id     string
	schema theme.Base24ColorSchema
}

type Model struct {
	theme                theme.Theme
	colorSchemes         []colorSchemaData
	filteredColorSchemes []colorSchemaData

	tbl       table.Model
	textInput textinput.Model

	wantsClose         bool
	selectedSchemeName string
}

type ThemeSelectedMsg struct {
	SchemeName string
}
