// Package repostable renders repository states in an interactive table.
package repostable

import (
	"strings"

	"charm.land/bubbles/v2/table"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func New(
	theme theme.Theme,
	report report.ScanReport,
	width int,
	height int,
	options Options,
) Model {
	model := Model{
		width:         width,
		height:        height,
		theme:         theme,
		report:        report,
		filteredRepos: report.RepoStates,
		filterQuery:   "",
		options:       options,
	}

	cols := createColumns(width, model.options)
	rows := createRows(model.report.RepoStates, theme, model.options)

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithWidth(width),
		table.WithHeight(height),
	)
	t.Focus()

	t.SetStyles(table.Styles{
		Header:   model.theme.Styles.TableHeader,
		Selected: model.theme.Styles.TableSelectedRow,
		Cell:     model.theme.Styles.TableRow,
	})
	model.tbl = t

	return model
}

func (m *Model) SetReport(report report.ScanReport) {
	m.report = report
	m.Filter(m.filterQuery)
}

func (m *Model) UpdateWindowSize(width int, height int) Model {
	m.width = width - 2   // border corners
	m.height = height - 2 // border corners

	m.tbl.SetWidth(m.width)
	m.tbl.SetHeight(m.height)
	cols := createColumns(m.width, m.options)
	m.tbl.SetColumns(cols)

	return *m
}

// Filter limits repository states by repository or branch name and refreshes the table.
func (m *Model) Filter(query string) {
	m.filterQuery = query
	q := strings.ToLower(strings.TrimSpace(query))
	if len(q) == 0 {
		m.filteredRepos = m.report.RepoStates
	} else {
		m.filteredRepos = []report.RepoState{}
		for _, rs := range m.report.RepoStates {
			if strings.Contains(strings.ToLower(rs.Repo), q) ||
				strings.Contains(strings.ToLower(rs.Branch), q) {
				m.filteredRepos = append(m.filteredRepos, rs)
			}
		}
	}

	cursorPosition := m.tbl.Cursor()

	rows := createRows(m.filteredRepos, m.theme, m.options)
	m.tbl.SetRows(rows)

	if cursorPosition < len(m.filteredRepos) {
		m.tbl.SetCursor(cursorPosition)
	} else {
		m.tbl.SetCursor(0)
	}
}

func (m *Model) UpdateRepoState(index int, newState report.RepoState) {
	if index < 0 || index >= len(m.filteredRepos) || m.filteredRepos[index].ID != newState.ID {
		return
	}

	originalIndex := getRepoIndex(m.report.RepoStates, newState.ID)
	if originalIndex == -1 {
		return
	}

	m.filteredRepos[index] = newState
	m.report.RepoStates[originalIndex] = newState

	rows := createRows(m.filteredRepos, m.theme, m.options)
	m.tbl.SetRows(rows)
}

// Blur removes focus from the table.
func (m *Model) Blur() {
	m.tbl.Blur()
}

// Focus gives focus to the table.
func (m *Model) Focus() {
	m.tbl.Focus()
}

// Cursor returns the index of the selected row.
func (m *Model) Cursor() int {
	return m.tbl.Cursor()
}

func (m *Model) ReposCount() int {
	return len(m.filteredRepos)
}

func (m *Model) GetCurrentRepoState() *report.RepoState {
	return m.GetRepoStateAt(m.Cursor())
}

func (m *Model) GetRepoStateAt(index int) *report.RepoState {
	if index < 0 {
		return nil
	}
	if index >= len(m.filteredRepos) {
		return nil
	}
	return &m.filteredRepos[index]
}

func (m *Model) UpdateTheme(newTheme theme.Theme) {
	m.theme = newTheme
	m.tbl.SetStyles(table.Styles{
		Header:   m.theme.Styles.TableHeader,
		Selected: m.theme.Styles.TableSelectedRow,
		Cell:     m.theme.Styles.TableRow,
	})
	m.tbl.SetRows(createRows(m.filteredRepos, m.theme, m.options))
}
