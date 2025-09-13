package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m Model) detailsView() string {
	if len(m.report.RepoStates) == 0 {
		return ""
	}
	idx := m.tbl.Cursor()
	if idx < 0 || idx >= len(m.report.RepoStates) {
		return ""
	}
	rs := m.report.RepoStates[idx]

	uc := len(rs.UncommitedFiles)
	ucStr := CleanStyle.Render(strconv.Itoa(uc))
	if uc > 0 {
		ucStr = DirtyStyle.Render(strconv.Itoa(uc))
	}

	lines := []string{
		SectionStyle.Render("\nDetails"),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Repo:"), rs.Repo),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Branch:"), rs.Branch),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Uncommitted:"), ucStr),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Path:"), rs.Path),
	}
	if uc > 0 {
		lines = append(lines, HeaderStyle.Render("Files:"))
		for _, f := range rs.UncommitedFiles {
			lines = append(lines, "  "+lipgloss.NewStyle().Faint(true).Render(f))
		}
	}
	return strings.Join(lines, "\n")
}

func (m *Model) reflowColumns() {
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
		{Title: HeaderStyle.Render("Repo"), Width: repoW},
		{Title: HeaderStyle.Render("Branch"), Width: branchW},
		{Title: HeaderStyle.Render("Uncommitted"), Width: ucW},
		{Title: HeaderStyle.Render("Path"), Width: pathW},
	}
	m.tbl.SetColumns(cols)
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
