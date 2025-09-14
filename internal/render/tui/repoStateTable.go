package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func createRows(r report.ScanReport) []table.Row {
	rows := make([]table.Row, 0, len(r.RepoStates))
	for _, rs := range r.RepoStates {
		state := getStateColumnStr(rs)

		rows = append(rows, table.Row{
			RepoStyle.Render(rs.Repo),
			BranchStyle.Render(rs.Branch),
			state,
			RepoStyle.Render(rs.Path),
		})
	}
	return rows
}

func createColumns(maxWidth int) []table.Column {
	repoW := maxWidth * RepoW / 100
	branchW := maxWidth * BranchW / 100
	remoteStateW := maxWidth * RemoteStateW / 100
	pathW := maxWidth * PathW / 100

	return []table.Column{
		{Title: HeaderStyle.Render("Repo"), Width: repoW},
		{Title: HeaderStyle.Render("Branch"), Width: branchW},
		{Title: HeaderStyle.Render("State"), Width: remoteStateW},
		{Title: HeaderStyle.Render("Path"), Width: pathW},
	}
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

	lines := []string{
		SectionStyle.Render("\nDetails"),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Repo:"), rs.Repo),
		fmt.Sprintf("%s %s", HeaderStyle.Render("Branch:"), rs.Branch),
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

func getStateColumnStr(rs report.RepoState) string {
	var stateStr strings.Builder

	uc := len(rs.UncommitedFiles)
	if uc > 0 {
		stateStr.WriteString(DirtyStyle.Render(fmt.Sprintf("⏳%-d", uc)))
	} else if uc == 0 {
		stateStr.WriteString(FooterStyle.Render(fmt.Sprintf("⏳%-d", uc)))
	}

	if rs.Ahead > 0 {
		stateStr.WriteString(CleanStyle.Render(fmt.Sprintf(" ↑%-d", rs.Ahead)))
	} else if rs.Ahead < 0 {
		stateStr.WriteString(DirtyStyle.Render(fmt.Sprintf(" %-s ", "x")))
	} else {
		stateStr.WriteString(FooterStyle.Render(fmt.Sprintf(" ↑%-d", 0)))
	}

	if rs.Behind > 0 {
		stateStr.WriteString(CleanStyle.Render(fmt.Sprintf(" ↓%-d", rs.Behind)))
	} else if rs.Behind < 0 {
		stateStr.WriteString(DirtyStyle.Render(fmt.Sprintf(" %-s", "x")))
	} else {
		stateStr.WriteString(FooterStyle.Render(fmt.Sprintf(" ↓%-d", 0)))
	}

	return stateStr.String()
}
