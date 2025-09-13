package tui

import (
	"fmt"
	"os"
	"time"

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
	cols := createColumns(100)
	rows := createRows(r)

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

func setKeymaps(km table.KeyMap) {
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", "ctrl+u")
	km.PageDown.SetKeys("pgdn", "ctrl+d")
	km.GotoTop.SetKeys("home", "g")
	km.GotoBottom.SetKeys("end", "G")
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.contentHeight = max(6, m.height-6) // leave room for title+footer
		m.tbl.SetHeight(min(18, m.contentHeight))
		cols := createColumns(m.width)
		m.tbl.SetColumns(cols)
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

func (m Model) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		TitleStyle.Render("reposcan"),
		" ",
		SubtleStyle.Render(fmt.Sprintf("• %d repos • generated %s",
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
		summary = DirtyStyle.Render(summary)
	} else {
		summary = CleanStyle.Render(summary)
	}

	body := m.tbl.View()
	if m.showDetails {
		body = lipgloss.JoinVertical(lipgloss.Left, body, m.detailsView())
	}

	footer := FooterStyle.Render("↑/↓ to move • enter to toggle details • q to quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		summary,
		body,
		footer,
	)
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
