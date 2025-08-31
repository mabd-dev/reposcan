package stdout

import (
	"fmt"
	"strings"

	"github.com/MABD-dev/reposcan/pkg/report"
)

func RenderReposTable(r report.ScanReport) {
	// Table header
	fmt.Printf("%s %s %s %s %s\n",
		CyanBold("%-*s", RepoW, "Repo"),
		CyanBold("%-*s", BranchW, "Branch"),
		CyanBold("%-*s", UncommW, "Not-Commited"),
		CyanBold("%-*s", AheadW, "Ahead"),
		//CyanBold("%-*s", behindW, "Behind"),
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

	aheadCell := GrayS("%-*d", AheadW, rs.Ahead)
	if rs.Ahead > 0 {
		aheadCell = GreenS("%-*d", AheadW, rs.Ahead)
	} else if rs.Ahead < 0 {
		aheadCell = RedS("%-*d", AheadW, rs.Ahead)
	}

	// behindCell := GrayS("%-*d", behindW, rs.Behind)
	// if rs.Behind > 0 || rs.Behind < 0 {
	// 	behindCell = RedS("%-*d", behindW, rs.Behind)
	// }

	repoCell := fmt.Sprintf("%-*s", RepoW, truncateRunes(rs.Repo, RepoW))
	branchCell := BlueS("%-*s", BranchW, truncateRunes(rs.Branch, BranchW))

	fmt.Printf("%s %s %s %s %s\n",
		repoCell,
		branchCell,
		ucCell,
		aheadCell,
		//behindCell,
		rs.Path, // full path, no truncation
	)
}
