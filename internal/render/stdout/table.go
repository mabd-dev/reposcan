package stdout

import (
	"fmt"
	"github.com/mabd-dev/reposcan/pkg/report"
	"strings"
)

// RenderReposTable renders the per-repository rows for a ScanReport as a table.
func RenderReposTable(r report.ScanReport) {
	// Table header
	fmt.Printf("%s %s %s\n",
		CyanBold("%-*s", RepoW, "Repo"),
		CyanBold("%-*s", BranchW, "Branch"),
		CyanBold("%-*s", RemoteStateW, "State"),
	)
	fmt.Println(strings.Repeat("─", RepoW+1+BranchW+RemoteStateW+1))

	for _, rs := range r.RepoStates {
		renderRepoState(rs)
	}
}

func renderRepoState(rs report.RepoState) {
	repoCell := fmt.Sprintf("%-*s", RepoW, truncateRunes(rs.Repo, RepoW))
	branchCell := BlueS("%-*s", BranchW, truncateRunes(rs.Branch, BranchW))

	remoteStateStr := getStateColumnStr(rs)

	fmt.Printf("%s %s %s\n",
		repoCell,
		branchCell,
		remoteStateStr,
	)
}

func getStateColumnStr(rs report.RepoState) string {
	var stateStr strings.Builder

	uc := len(rs.UncommitedFiles)
	if uc > 0 {
		stateStr.WriteString(RedS("⏳%-*d", UncommW, uc))
	} else if uc == 0 {
		stateStr.WriteString(GrayS("⏳%-*d", UncommW, uc))
	}

	if rs.Ahead > 0 {
		stateStr.WriteString(GreenS("↑%-*d", AheadW, rs.Ahead))
	} else if rs.Ahead < 0 {
		stateStr.WriteString(RedS("%-*s ", AheadW, "x"))
	} else {
		stateStr.WriteString(GrayS("↑%-*d", AheadW, 0))
	}

	if rs.Behind > 0 {
		stateStr.WriteString(GreenS("↓%-*d", BehindW, rs.Behind))
	} else if rs.Behind < 0 {
		stateStr.WriteString(RedS("%-*s ", BehindW, "x"))
	} else {
		stateStr.WriteString(GrayS("↓%-*d", BehindW, 0))
	}

	return stateStr.String()
}
