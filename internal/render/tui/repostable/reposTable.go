// Package repostable is a Model that renders git repo states in a table. Providing functionality like filterning
package repostable

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
	"strings"
)

type Table struct {
	tbl           table.Model
	report        report.ScanReport
	filteredRepos []report.RepoState
	Theme         theme.Theme
}

func (rt *Table) InitUI(
	width int,
	height int,
) {
	cols := createColumns(width)
	rows := createRows(rt.report.RepoStates)

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
		Header:   rt.Theme.Styles.TableHeader,
		Selected: rt.Theme.Styles.TableSelectedRow,
		Cell:     rt.Theme.Styles.TableRow,
	})
	rt.tbl = t
}

func setKeymaps(km table.KeyMap) {
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", tea.KeyCtrlU.String())
	km.PageDown.SetKeys("pgdn", tea.KeyCtrlD.String())
	km.GotoTop.SetKeys("home", "g")
	km.GotoBottom.SetKeys("end", "G")
}

func (rt Table) Init() tea.Cmd { return nil }

func (rt Table) Update(msg tea.Msg) (Table, tea.Cmd) {
	var cmd tea.Cmd
	rt.tbl, cmd = rt.tbl.Update(msg)

	return rt, cmd
}

func (rt Table) View() string {
	return rt.Theme.Styles.BoxFor(rt.tbl.Focused()).Render(rt.tbl.View())
}

func (rt *Table) UpdateWindowSize(width int, height int) {
	rt.tbl.SetHeight(height)
	cols := createColumns(width)
	rt.tbl.SetColumns(cols)
}

func (rt *Table) SetReport(report report.ScanReport) {
	rt.report = report
	rt.filteredRepos = report.RepoStates
}

// Filter filters repo states based on repo name. Then update table based on filtered repos
func (rt *Table) Filter(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	if len(q) == 0 {
		rt.filteredRepos = rt.report.RepoStates
	} else {
		rt.filteredRepos = []report.RepoState{}
		for _, rs := range rt.report.RepoStates {
			if strings.Contains(strings.ToLower(rs.Repo), q) ||
				strings.Contains(strings.ToLower(rs.Branch), q) {
				rt.filteredRepos = append(rt.filteredRepos, rs)
			}
		}
	}

	rows := createRows(rt.filteredRepos)
	rt.tbl.SetRows(rows)

	if len(rows) > 0 {
		rt.tbl.SetCursor(0)
	}

}

func (rt *Table) UpdateRepoState(index int, newState report.RepoState) {
	rt.filteredRepos[index] = newState

	originalIndex := getRepoIndex(rt.report.RepoStates, newState.ID)
	if originalIndex != -1 {
		rt.report.RepoStates[originalIndex] = newState
	}

	rows := createRows(rt.filteredRepos)
	rt.tbl.SetRows(rows)
}

// Blur removes focus from table
func (rt *Table) Blur() {
	rt.tbl.Blur()
}

// Focus bring focus to table
func (rt *Table) Focus() {
	rt.tbl.Focus()
}

// Cursor returns the index of the selected row.
func (rt *Table) Cursor() int {
	return rt.tbl.Cursor()
}

func (rt *Table) GetCurrentRepoState() *report.RepoState {
	return rt.GetRepoStateAt(rt.Cursor())
}

func (rt *Table) GetRepoStateAt(index int) *report.RepoState {
	if index < 0 {
		return nil
	}
	if index >= len(rt.filteredRepos) {
		return nil
	}
	return &rt.filteredRepos[index]
}
