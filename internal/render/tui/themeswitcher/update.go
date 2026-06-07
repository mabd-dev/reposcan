package themeswitcher

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/mabd-dev/reposcan/internal/logger"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.textInput.Focused() {
		return m.updateTextInput(msg)
	}

	return m.updateSchemasTable(msg)
}

func (m Model) updateTextInput(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctr+c":
			m.textInput.SetValue("")
			m.focusSchemasTable()
		case "enter":
			m.focusSchemasTable()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	filteredColorSchemes := make([]colorSchemaData, 0, len(m.colorSchemes))
	for _, s := range m.colorSchemes {
		if strings.Contains(strings.ToLower(s.schema.Name), strings.ToLower(m.textInput.Value())) {
			filteredColorSchemes = append(filteredColorSchemes, s)
		}
	}
	m.filteredColorSchemes = filteredColorSchemes
	m.tbl.SetRows(createRows(filteredColorSchemes, m.theme))

	return m, cmd
}

func (m Model) updateSchemasTable(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			m.wantsClose = true
			return m, nil
		case "/":
			m.focusTextInput()
			return m, nil
		case "enter":
			cursor := m.tbl.Cursor()
			if cursor >= 0 && cursor < len(m.filteredColorSchemes) {
				m.selectedSchemeName = m.filteredColorSchemes[cursor].id
				logger.Debug(fmt.Sprintf("selected scheme name=%v", m.selectedSchemeName))
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	m.updateCursorInRows()

	return m, cmd
}

func (m *Model) focusSchemasTable() {
	m.tbl.Focus()
	m.textInput.Blur()
}

func (m *Model) focusTextInput() {
	m.tbl.Blur()
	m.textInput.Focus()
}

func (m *Model) updateCursorInRows() {
	cursor := m.tbl.Cursor()
	rows := m.tbl.Rows()
	for i := range rows {
		if i == cursor {
			rows[i][0] = cursorChar
		} else {
			rows[i][0] = emptyChar
		}
		if strings.HasPrefix(rows[i][1], m.theme.Colors.Name) {
			rows[i][0] = selectedChar
		}
	}
	m.tbl.SetRows(rows)
}
