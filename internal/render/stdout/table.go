package stdout

import (
	"fmt"
	"github.com/MABD-dev/reposcan/pkg/report"
	"strings"
)

func RenderReposTable(r report.ScanReport) {
	// Table header
	fmt.Printf("%s %s %s %s %s\n",
		CyanBold("%-*s", RepoW, "Repo"),
		CyanBold("%-*s", BranchW, "Branch"),
		CyanBold("%-*s", UncommW, "Not-Commited"),
		CyanBold("%-*s", AheadW, "State"),
		//CyanBold("%-*s", BehindW, "Behind"),
		CyanBold("%s", "Path"),
	)
	fmt.Println(strings.Repeat("─", RepoW+1+BranchW+AheadW+BehindW+1+UncommW+1+60-2))

	// Table row
	for _, rs := range r.RepoStates {
		renderRepoState(rs)
	}
}

func renderRepoState(rs report.RepoState) {
	uc := len(rs.UncommitedFiles)
	ucCell := GrayS("%-*d", UncommW, uc)
	if uc > 0 {
		ucCell = RedS("%-*d", UncommW, uc)
	}

	var remoteState strings.Builder
	//aheadCell := GrayS("%-*d", AheadW, rs.Ahead)
	if rs.Ahead > 0 {
		remoteState.WriteString(GreenS("↑%-*d", AheadW, rs.Ahead))
		//aheadCell = GreenS("↑%-*d", AheadW, rs.Ahead)
	} else if rs.Ahead < 0 {
		//aheadCell = RedS("%-*d", AheadW, rs.Ahead)
		remoteState.WriteString(GrayS("↑%-*d", 3, -1))
	} else {
		remoteState.WriteString(GrayS("↑%-*d", 3, 0))
	}

	//behindCell := GrayS("%-*d", BehindW, rs.Behind)
	if rs.Behind > 0 {
		remoteState.WriteString(GreenS("↑%-*d", BehindW, rs.Behind))
		//behindCell = GreenS("↓%-*d", BehindW, rs.Behind)
	} else if rs.Behind < 0 {
		//behindCell = RedS("%-*d", BehindW, rs.Behind)
		remoteState.WriteString(GrayS("↓%-*d", 2, -1))
	} else {
		remoteState.WriteString(GrayS("↓%-*d", 2, 0))
	}

	repoCell := fmt.Sprintf("%-*s", RepoW, truncateRunes(rs.Repo, RepoW))
	branchCell := BlueS("%-*s", BranchW, truncateRunes(rs.Branch, BranchW))

	fmt.Printf("%s %s %s %s %s\n",
		repoCell,
		branchCell,
		ucCell,
		remoteState.String(),
		// aheadCell,
		// behindCell,
		rs.Path, // full path, no truncation
	)
}
