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

	allRows := []tableRow{}
	lastIndex := len(worktreeStates) - 1
	for i, wt := range worktreeStates {
		previousIsSameRepo := i > 0 && worktreeStates[i-1].RepoName == wt.RepoName
		nextIsSameRepo := i < lastIndex && worktreeStates[i+1].RepoName == wt.RepoName
		createHeader := !previousIsSameRepo && nextIsSameRepo
		if createHeader {
			allRows = append(allRows, tableRow{
				Repo:     "📦 " + wt.RepoName,
				Branch:   "",
				State:    "",
				IsHeader: true,
				WtIndex:  -1,
			})
		}

		repoName := wt.RepoName
		if createHeader || previousIsSameRepo {
			repoName = "  ├ " + wt.WorktreeName
			if !nextIsSameRepo {
				repoName = "  ┕ " + wt.WorktreeName
			}
		}
		allRows = append(allRows, tableRow{
			Repo:     repoName,
			Branch:   wt.Branch,
			State:    getStateColumnStr(wt, theme),
			IsHeader: false,
			WtIndex:  i,
		})
		// if previous is not same repo name && next is same repo name -> create header
	}

	model := Model{
		width:             width,
		height:            height,
		theme:             theme,
		allRows:           allRows,
		filteredRows:      allRows,
		allWorktreeStates: worktreeStates,
		filterQuery:       "",
	}

	cols := createColumns(width)
	rows := createRows(allRows, theme)

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
		m.filteredRows = m.allRows
	} else {
		m.filteredRows = []tableRow{}
		for _, r := range m.allRows {
			if r.IsHeader {
				if strings.Contains(strings.ToLower(r.Repo), q) {
					m.filteredRows = append(m.filteredRows, r)
				}
			} else {
				if strings.Contains(strings.ToLower(r.Repo), q) ||
					strings.Contains(strings.ToLower(r.Branch), q) {
					m.filteredRows = append(m.filteredRows, r)
				}
			}

		}
	}

	cursorPosition := m.tbl.Cursor()

	rows := createRows(m.filteredRows, m.theme)
	m.tbl.SetRows(rows)

	if cursorPosition < len(m.filteredRows) {
		m.tbl.SetCursor(cursorPosition)
	} else {
		m.tbl.SetCursor(0)
	}

}

func (m *Model) UpdateRepoState(index int, newState common.WorktreeState) {
	// TODO: implement later
	// m.filteredWorktreeStates[index] = newState
	//
	// originalIndex := getWorktreeIndex(m.allWorktreeStates, newState.RepoID)
	// if originalIndex != -1 {
	// 	m.allWorktreeStates[originalIndex] = newState
	// }
	//
	// rows := createRows(m.filteredRows, m.theme)
	// m.tbl.SetRows(rows)
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
	return len(rt.filteredRows)
}

func (m *Model) GetCurrentWorktreeState() *common.WorktreeState {
	return m.GetWorktreeStateAt(m.Cursor())
}

func (m *Model) GetWorktreeStateAt(index int) *common.WorktreeState {
	if index < 0 {
		return nil
	}

	if index >= len(m.filteredRows) {
		return nil
	}

	if m.filteredRows[index].IsHeader {
		return nil
	}

	worktree := m.allWorktreeStates[m.filteredRows[index].WtIndex]
	return &worktree
}
