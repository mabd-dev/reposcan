// Package repostable is a Model that renders git repo states in a table. Providing functionality like filterning
package repostable

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/ds/mmap"
	"github.com/mabd-dev/reposcan/internal/ds/mslice"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func New(
	theme theme.Theme,
	worktreeStates []common.WorktreeState,
	width int,
	height int,
) Model {

	model := Model{
		width:                  width,
		height:                 height,
		theme:                  theme,
		allWorktreeStates:      worktreeStates,
		filteredWorktreeStates: worktreeStates,
		filterQuery:            "",
	}

	cols := createColumns(width)
	rows := createRows(worktreeStates, theme)

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
	m.allWorktreeStates = mslice.Flatten(mmap.Map(report.RepoStates, common.MapToWorktreeStates))
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
		m.filteredWorktreeStates = m.allWorktreeStates
	} else {
		m.filteredWorktreeStates = []common.WorktreeState{}
		for _, rs := range m.allWorktreeStates {
			if strings.Contains(strings.ToLower(rs.RepoName), q) ||
				strings.Contains(strings.ToLower(rs.Branch), q) {
				m.filteredWorktreeStates = append(m.filteredWorktreeStates, rs)
			}
		}
	}

	cursorPosition := m.tbl.Cursor()

	rows := createRows(m.filteredWorktreeStates, m.theme)
	m.tbl.SetRows(rows)

	if cursorPosition < len(m.filteredWorktreeStates) {
		m.tbl.SetCursor(cursorPosition)
	} else {
		m.tbl.SetCursor(0)
	}

}

func (m *Model) UpdateRepoState(index int, newState common.WorktreeState) {
	m.filteredWorktreeStates[index] = newState

	originalIndex := getWorktreeIndex(m.allWorktreeStates, newState.RepoID)
	if originalIndex != -1 {
		m.allWorktreeStates[originalIndex] = newState
	}

	rows := createRows(m.filteredWorktreeStates, m.theme)
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
	return len(rt.filteredWorktreeStates)
}

func (m *Model) GetCurrentWorktreeState() *common.WorktreeState {
	return m.GetWorktreeStateAt(m.Cursor())
}

func (m *Model) GetWorktreeStateAt(index int) *common.WorktreeState {
	if index < 0 {
		return nil
	}
	if index >= len(m.filteredWorktreeStates) {
		return nil
	}
	return &m.filteredWorktreeStates[index]
}
