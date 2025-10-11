package repostable

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	"github.com/mabd-dev/reposcan/pkg/report"
	"strings"
)

const (
	RepoW        = 40
	BranchW      = 40
	RemoteStateW = 20 //(uncommited files count + aheadW + behindW + 4 space)
)

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
