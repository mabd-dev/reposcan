package stdout

import (
	"fmt"
	"strings"

	"github.com/mabd-dev/reposcan/pkg/report"
)

// RenderReposTable renders the per-repository rows for a ScanReport as a table.
func RenderReposTable(r report.ScanReport) {
	// Table header
	fmt.Printf("%s %s %s %s\n",
		CyanBold("%-*s", RepoW, "Repo"),
		CyanBold("%-*s", BranchW, "Branch"),
		CyanBold("%-*s", VCSW, "VCS"),
		CyanBold("%-*s", RemoteStateW, "State"),
	)
	fmt.Println(strings.Repeat("─", RepoW+1+BranchW+1+VCSW+1+RemoteStateW))

	for _, rs := range r.RepoStates {
		renderRepoState(rs)
	}
}

func renderRepoState(rs report.RepoState) {
	repoCell := fmt.Sprintf("%-*s", RepoW, truncateRunes(rs.Repo, RepoW))
	vcsCell := fmt.Sprintf("%-*s", VCSW, truncateRunes(rs.VCSType, VCSW))
	branchCell := BlueS("%-*s", BranchW, truncateRunes(rs.Branch, BranchW))

	remoteStateStr := getStateColumnStr(rs)

	fmt.Printf("%s %s %s %s\n",
		repoCell,
		branchCell,
		vcsCell,
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

	for i, remoteStatus := range rs.RemoteStatus {
		if i > 0 {
			stateStr.WriteString(" | ")
		}

		if remoteStatus.Ahead > 0 {
			stateStr.WriteString(GreenS("↑%-*d", AheadW, remoteStatus.Ahead))
		} else if remoteStatus.Ahead < 0 {
			stateStr.WriteString(RedS("%-*s", AheadW, "x"))
		} else {
			stateStr.WriteString(GrayS("↑%-*d", AheadW, 0))
		}

		if remoteStatus.Behind > 0 {
			stateStr.WriteString(YellowS("↓%-*d", BehindW, remoteStatus.Behind))
		} else if remoteStatus.Behind < 0 {
			stateStr.WriteString(RedS("%-*s", BehindW, "x"))
		} else {
			stateStr.WriteString(GrayS("↓%-*d", BehindW, 0))
		}

		if remoteStatus.Remote != "" && !(len(rs.RemoteStatus) == 1 && remoteStatus.Remote == "origin") {
			stateStr.WriteString(GrayS("(%s)", remoteStatus.Remote))
		}
	}

	return stateStr.String()
}
