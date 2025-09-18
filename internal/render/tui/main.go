package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/pkg/report"
	"golang.design/x/clipboard"
)

type Model struct {
	report            report.ScanReport
	tbl               table.Model
	showDetails       bool
	isPushing         bool
	width             int
	height            int
	contentHeight     int
	reposBeingUpdated []string
	warnings          []string
	showHelp          bool
}

func (m *Model) addWarning(msg string) {
	m.warnings = append(m.warnings, msg)
}

func (m Model) getReportAtCursor() report.RepoState {
	idx := m.tbl.Cursor()
	return m.report.RepoStates[idx]
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
		t.SetRows([]table.Row{{"", "", ""}})
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
		warnings:      []string{},
	}

	err := clipboard.Init()
	if err != nil {
		m.warnings = append(m.warnings, err.Error())
	}

	p := tea.NewProgram(m, tea.WithOutput(os.Stdout), tea.WithAltScreen())
	_, err = p.Run()
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
	return handleMsg(m, msg)
}

func (m Model) View() string {
	if m.showHelp {
		return generateHelpPopup(m.width, m.height)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		TitleStyle.Render("reposcan"),
		" ",
		SubtleStyle.Render(fmt.Sprintf("• %d repos • generated %s",
			len(m.report.RepoStates), m.report.GeneratedAt.Format(time.RFC3339))),
	)

	dirtyRepos := m.report.DirtyReposCount()
	summary := fmt.Sprintf("Total: %d  |  Uncommitted: %d", len(m.report.RepoStates), dirtyRepos)
	if dirtyRepos > 0 {
		summary = DirtyStyle.Render(summary)
	} else {
		summary = CleanStyle.Render(summary)
	}

	body := m.tbl.View()
	if m.showDetails {
		body = lipgloss.JoinVertical(lipgloss.Left, body, m.detailsView())
	}

	// TODO: show most important keybindings here as well
	footer := FooterStyle.Render("↑/↓ to move • ? keybindings")

	var messages strings.Builder
	for _, msg := range m.warnings {
		messages.WriteString(msg)
		messages.WriteString("\n")
	}
	stdMessages := FooterStyle.Render(messages.String())

	base := lipgloss.JoinVertical(lipgloss.Left,
		header,
		summary,
		body,
		footer,
		stdMessages,
	)
	return base
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
