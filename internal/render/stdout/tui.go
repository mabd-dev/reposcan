package stdout

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MABD-dev/reposcan/pkg/report"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	subtleStyle   = lipgloss.NewStyle().Faint(true)
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	repoStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	branchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	cleanStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	dirtyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	footerStyle   = lipgloss.NewStyle().Faint(true)
	sectionStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12"))
)

type model struct {
	report        report.ScanReport
	tbl           table.Model
	showDetails   bool
	width         int
	height        int
	contentHeight int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.contentHeight = max(6, m.height-6) // leave room for title+footer
		m.tbl.SetHeight(min(18, m.contentHeight))
		m.reflowColumns()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.showDetails = !m.showDetails
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func (m model) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		titleStyle.Render("reposcan"),
		" ",
		subtleStyle.Render(fmt.Sprintf("• %d repos • generated %s",
			len(m.report.RepoStates), m.report.GeneratedAt.Format(time.RFC3339))),
	)

	var dirty int
	for _, rs := range m.report.RepoStates {
		if len(rs.UncommitedFiles) > 0 {
			dirty++
		}
	}
	summary := fmt.Sprintf("Total: %d  |  Uncommitted: %d", len(m.report.RepoStates), dirty)
	if dirty > 0 {
		summary = dirtyStyle.Render(summary)
	} else {
		summary = cleanStyle.Render(summary)
	}

	body := m.tbl.View()
	if m.showDetails {
		body = lipgloss.JoinVertical(lipgloss.Left, body, m.detailsView())
	}

	footer := footerStyle.Render("↑/↓ to move • enter to toggle details • q to quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		summary,
		body,
		footer,
	)
}

func (m model) detailsView() string {
	if len(m.report.RepoStates) == 0 {
		return ""
	}
	idx := m.tbl.Cursor()
	if idx < 0 || idx >= len(m.report.RepoStates) {
		return ""
	}
	rs := m.report.RepoStates[idx]

	uc := len(rs.UncommitedFiles)
	ucStr := cleanStyle.Render(strconv.Itoa(uc))
	if uc > 0 {
		ucStr = dirtyStyle.Render(strconv.Itoa(uc))
	}

	lines := []string{
		sectionStyle.Render("\nDetails"),
		fmt.Sprintf("%s %s", headerStyle.Render("Repo:"), rs.Repo),
		fmt.Sprintf("%s %s", headerStyle.Render("Branch:"), rs.Branch),
		fmt.Sprintf("%s %s", headerStyle.Render("Uncommitted:"), ucStr),
		fmt.Sprintf("%s %s", headerStyle.Render("Path:"), rs.Path),
	}
	if uc > 0 {
		lines = append(lines, headerStyle.Render("Files:"))
		for _, f := range rs.UncommitedFiles {
			lines = append(lines, "  "+lipgloss.NewStyle().Faint(true).Render(f))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *model) reflowColumns() {
	// Compute widths based on terminal width, prioritizing Path.
	w := max(60, m.width)
	repoW := 22
	branchW := 18
	ucW := 12
	pathW := w - (repoW + 1 + branchW + 1 + ucW + 3)
	if pathW < 20 {
		// shrink others proportionally
		delta := 20 - pathW
		repoW = max(12, repoW-delta/2)
		branchW = max(10, branchW-delta/2)
		pathW = 20
	}

	cols := []table.Column{
		{Title: headerStyle.Render("Repo"), Width: repoW},
		{Title: headerStyle.Render("Branch"), Width: branchW},
		{Title: headerStyle.Render("Uncommitted"), Width: ucW},
		{Title: headerStyle.Render("Path"), Width: pathW},
	}
	m.tbl.SetColumns(cols)
}

// ShowReportTUI runs a Bubble Tea UI that renders the ScanReport in a table.
func ShowReportTUI(r report.ScanReport) error {
	rows := make([]table.Row, 0, len(r.RepoStates))
	for _, rs := range r.RepoStates {
		uc := len(rs.UncommitedFiles)
		ucStr := cleanStyle.Render(strconv.Itoa(uc))
		if uc > 0 {
			ucStr = dirtyStyle.Render(strconv.Itoa(uc))
		}
		rows = append(rows, table.Row{
			repoStyle.Render(rs.Repo),
			branchStyle.Render(rs.Branch),
			ucStr,
			rs.Path, // full path
		})
	}

	// 1) Define columns FIRST
	cols := []table.Column{
		{Title: headerStyle.Render("Repo"), Width: 22},
		{Title: headerStyle.Render("Branch"), Width: 18},
		{Title: headerStyle.Render("Uncommitted"), Width: 12},
		{Title: headerStyle.Render("Path"), Width: 60},
	}

	// 2) Now create the table with columns BEFORE rows
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(12),
	)

	t.Focus()

	// (optional) Make sure arrows + j/k are bound (these are defaults, but explicit is nice)
	km := table.DefaultKeyMap()
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", "ctrl+u")
	km.PageDown.SetKeys("pgdn", "ctrl+d")
	km.GotoTop.SetKeys("home", "g")
	km.GotoBottom.SetKeys("end", "G")
	//t.SetKeyMap(km)

	// Optional: if no repos, show an empty placeholder row so the table renders nicely
	if len(rows) == 0 {
		t.SetRows([]table.Row{{"", "", "", ""}})
	}

	t.SetStyles(table.Styles{
		Header:   headerStyle,
		Selected: selectedStyle,
		Cell:     lipgloss.NewStyle(),
	})

	m := model{
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
