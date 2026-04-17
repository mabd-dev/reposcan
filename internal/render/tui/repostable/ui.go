package repostable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/internal/theme"
)

const (
	RepoW        = 30
	BranchW      = 30
	RemoteStateW = 40
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

func createRows(
	originalGroups []worktreesGroup,
	groups []worktreesGroup,
	filterQuery string,
	theme theme.Theme,
) []table.Row {
	rows := []table.Row{}
	inFilterMode := len(filterQuery) > 0

	for i, group := range groups {
		hasMultipleWorktreeInOrigianList := len(originalGroups[i].worktrees) > 1

		showHeader := (inFilterMode && hasMultipleWorktreeInOrigianList) || len(group.worktrees) > 1
		if showHeader { // show header
			rows = append(rows, table.Row{
				"📦 " + group.repoName,
				strings.Repeat("┅", 10),
				strings.Repeat("┅", 10),
			})
		}

		for _, r := range group.worktrees {
			rows = append(rows, table.Row{
				r.Repo,
				r.Branch,
				r.State,
			})
		}
	}
	return rows
}

func getStateColumnStr(rs common.WorktreeState, theme theme.Theme) string {
	parts := []string{}

	uc := len(rs.UncommitedFiles)
	ucStr := fmt.Sprintf("⏳%-d ", uc)

	for _, remoteStatus := range rs.RemoteStatus {
		var statusParts []string

		if remoteStatus.Ahead > 0 {
			statusParts = append(statusParts, fmt.Sprintf("↑%-d", remoteStatus.Ahead))
		} else if remoteStatus.Ahead < 0 {
			statusParts = append(statusParts, "x")
		} else {
			statusParts = append(statusParts, fmt.Sprintf("↑%-d", 0))
		}

		if remoteStatus.Behind > 0 {
			statusParts = append(statusParts, fmt.Sprintf("↓%-d", remoteStatus.Behind))
		} else if remoteStatus.Behind < 0 {
			statusParts = append(statusParts, "x")
		} else {
			statusParts = append(statusParts, fmt.Sprintf("↓%-d", 0))
		}

		remoteName := theme.Styles.Base.Render(fmt.Sprintf("(%s)", remoteStatus.Remote))
		statusParts = append(statusParts, remoteName)

		parts = append(parts, strings.Join(statusParts, " "))
	}

	// Combine uncommitted count with all remote statuses, separated by " | "
	s := ucStr
	s += strings.Join(parts, " | ")

	return s
}

func setKeymaps(km table.KeyMap) {
	km.LineUp.SetKeys("up", "k")
	km.LineDown.SetKeys("down", "j")
	km.PageUp.SetKeys("pgup", tea.KeyCtrlU.String())
	km.PageDown.SetKeys("pgdn", tea.KeyCtrlD.String())
	km.GotoTop.SetKeys("home", "g")
	km.GotoBottom.SetKeys("end", "G")
}

func getWorktreeIndex(worktrees []common.WorktreeState, id string) int {
	for i, s := range worktrees {
		if s.RepoID == id {
			return i
		}
	}
	return -1
}
