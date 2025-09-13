package tui

import (
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mabd-dev/reposcan/pkg/report"
)

type Model struct {
	report        report.ScanReport
	tbl           table.Model
	showDetails   bool
	isPushing     bool
	width         int
	height        int
	contentHeight int
}

// ShowReportTUI runs a Bubble Tea UI that renders the ScanReport in a table.
func ShowReportTUI(r report.ScanReport) error {
	rows := createRows(r)
	cols := createColumns()

	// Now create the table with columns BEFORE rows
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(12),
	)
	t.Focus()

	km := table.DefaultKeyMap()
	setKeymaps(km)

	// if no repos, show an empty placeholder row so the table renders nicely
	if len(rows) == 0 {
		t.SetRows([]table.Row{{"", "", "", ""}})
	}

	t.SetStyles(table.Styles{
		Header:   HeaderStyle,
		Selected: SelectedStyle,
		Cell:     lipgloss.NewStyle(),
	})

	m := Model{
		report:        r,
		tbl:           t,
		showDetails:   false,
		width:         100,
		height:        30,
		contentHeight: 18,
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stdout), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func createRows(r report.ScanReport) []table.Row {
	rows := make([]table.Row, 0, len(r.RepoStates))
	for _, rs := range r.RepoStates {
		uc := len(rs.UncommitedFiles)
		ucStr := CleanStyle.Render(strconv.Itoa(uc))
		if uc > 0 {
			ucStr = DirtyStyle.Render(strconv.Itoa(uc))
		}
		rows = append(rows, table.Row{
			RepoStyle.Render(rs.Repo),
			BranchStyle.Render(rs.Branch),
			ucStr,
			rs.Path,
		})
	}
	return rows
}

func createColumns() []table.Column {
	return []table.Column{
		{Title: HeaderStyle.Render("Repo"), Width: 22},
		{Title: HeaderStyle.Render("Branch"), Width: 18},
		{Title: HeaderStyle.Render("Uncommitted"), Width: 12},
		{Title: HeaderStyle.Render("Path"), Width: 60},
	}
}

func setKeymaps(km table.KeyMap) {
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", "ctrl+u")
	km.PageDown.SetKeys("pgdn", "ctrl+d")
	km.GotoTop.SetKeys("home", "g")
	km.GotoBottom.SetKeys("end", "G")
}
