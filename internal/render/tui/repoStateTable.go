package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func createRows(repoStates []report.RepoState) []table.Row {
	rows := make([]table.Row, 0, len(repoStates))
	for _, rs := range repoStates {
		state := getStateColumnStr(rs)

		rows = append(rows, table.Row{
			rs.Repo,
			rs.Branch,
			state,
		})
	}
	return rows
}

func createColumns(maxWidth int) []table.Column {
	repoW := maxWidth * RepoW / 100
	branchW := maxWidth * BranchW / 100
	remoteStateW := maxWidth * RemoteStateW / 100

	return []table.Column{
		{Title: "Repo", Width: repoW},
		{Title: "Branch", Width: branchW},
		{Title: "State", Width: remoteStateW},
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
		lines = append(lines, HeaderStyle.Render("Uncommited Files:"))
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
		stateStr.WriteString(fmt.Sprintf("⏳%-d", uc))
	} else if uc == 0 {
		stateStr.WriteString(fmt.Sprintf("⏳%-d", uc))
	}

	if rs.Ahead > 0 {
		stateStr.WriteString(fmt.Sprintf(" ↑%-d", rs.Ahead))
	} else if rs.Ahead < 0 {
		stateStr.WriteString(fmt.Sprintf(" %-s ", "x"))
	} else {
		stateStr.WriteString(fmt.Sprintf(" ↑%-d", 0))
	}

	if rs.Behind > 0 {
		stateStr.WriteString(fmt.Sprintf(" ↓%-d", rs.Behind))
	} else if rs.Behind < 0 {
		stateStr.WriteString(fmt.Sprintf(" %-s", "x"))
	} else {
		stateStr.WriteString(fmt.Sprintf(" ↓%-d", 0))
	}

	return stateStr.String()
}
