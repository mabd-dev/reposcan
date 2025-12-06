package repostable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mabd-dev/reposcan/internal/logger"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
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

func createRows(repoStates []report.RepoState, theme theme.Theme) []table.Row {
	rows := make([]table.Row, 0, len(repoStates))
	for _, rs := range repoStates {
		state := getStateColumnStr(rs, theme)

		rows = append(rows, table.Row{
			rs.Repo,
			rs.Branch,
			state,
		})
	}
	return rows
}

func getStateColumnStr(rs report.RepoState, theme theme.Theme) string {
	lines := []string{}
	var stateStr strings.Builder

	for i, remoteStatus := range rs.RemoteStatus {
		stateStr.Reset()

		uc := len(rs.UncommitedFiles)
		if uc > 0 {
			stateStr.WriteString(fmt.Sprintf("⏳%-d", uc))
		} else if uc == 0 {
			stateStr.WriteString(fmt.Sprintf("⏳%-d", uc))
		}

		if remoteStatus.Ahead > 0 {
			stateStr.WriteString(fmt.Sprintf(" ↑%-d", remoteStatus.Ahead))
		} else if remoteStatus.Ahead < 0 {
			stateStr.WriteString(fmt.Sprintf(" %-s ", "x"))
		} else {
			stateStr.WriteString(fmt.Sprintf(" ↑%-d", 0))
		}

		if remoteStatus.Behind > 0 {
			stateStr.WriteString(fmt.Sprintf(" ↓%-d", remoteStatus.Behind))
		} else if remoteStatus.Behind < 0 {
			stateStr.WriteString(fmt.Sprintf(" %-s", "x"))
		} else {
			stateStr.WriteString(fmt.Sprintf(" ↓%-d", 0))
		}

		remoteName := theme.Styles.Muted.Render(fmt.Sprintf(" (%s)", remoteStatus.Remote))
		stateStr.WriteString(remoteName)

		if i < len(rs.RemoteStatus)-1 { // not the last element
			stateStr.WriteString("\n")
		}

		lines = append(lines, stateStr.String())
	}

	s := lipgloss.JoinVertical(lipgloss.Center, lines...)

	logger.Debug("staus output=", logger.StringAttr("s=", s))

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

func getRepoIndex(repos []report.RepoState, id string) int {
	for i, s := range repos {
		if s.ID == id {
			return i
		}
	}
	return -1
}
