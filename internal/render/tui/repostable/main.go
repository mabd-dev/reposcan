// Package repostable is a Model that renders git repo states in a table. Providing functionality like filterning
package repostable

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func New(
	theme theme.Theme,
	report report.ScanReport,
	width int,
	height int,
) Model {
	model := Model{
		width:         width,
		height:        height,
		theme:         theme,
		report:        report,
		filteredRepos: report.RepoStates,
		filterQuery:   "",
	}

	cols := createColumns(width)
	rows := createRows(model.report.RepoStates)

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(height),
	)
	t.Focus()

	km := table.DefaultKeyMap()
	setKeymaps(km)

	// if no repos, show an empty placeholder row so the table renders nicely
	if len(rows) == 0 {
		t.SetRows([]table.Row{{"", "", ""}})
	}

	t.SetStyles(table.Styles{
		Header:   model.theme.Styles.TableHeader,
		Selected: model.theme.Styles.TableSelectedRow,
		Cell:     model.theme.Styles.TableRow,
	})
	model.tbl = t

	return model
}

func (rt Model) Init() tea.Cmd { return nil }

func (m *Model) SetReport(report report.ScanReport) {
	m.report = report
	m.Filter(m.filterQuery)
}

func (m *Model) UpdateWindowSize(width int, height int) Model {
	m.width = width - 2   // border corners
	m.height = height - 2 // border corners

	m.tbl.SetHeight(m.height)
	cols := createColumns(m.width)
	m.tbl.SetColumns(cols)

	return *m
}

// Filter filters repo states based on repo name. Then update table based on filtered repos
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

	rows := createRows(m.filteredRepos)
	m.tbl.SetRows(rows)

	if cursorPosition < len(m.filteredRepos) {
		m.tbl.SetCursor(cursorPosition)
	} else {
		m.tbl.SetCursor(0)
	}

}

func (m *Model) UpdateRepoState(index int, newState report.RepoState) {
	m.filteredRepos[index] = newState

	originalIndex := getRepoIndex(m.report.RepoStates, newState.ID)
	if originalIndex != -1 {
		m.report.RepoStates[originalIndex] = newState
	}

	rows := createRows(m.filteredRepos)
	m.tbl.SetRows(rows)
}

// Blur removes focus from table
func (m *Model) Blur() {
	m.tbl.Blur()
}

// Focus bring focus to table
func (m *Model) Focus() {
	m.tbl.Focus()
}

// Cursor returns the index of the selected row.
func (m *Model) Cursor() int {
	return m.tbl.Cursor()
}

func (rt *Model) ReposCount() int {
	return len(rt.filteredRepos)
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
