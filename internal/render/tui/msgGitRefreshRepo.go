package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type gitRefreshRepoResultMsg struct {
	newRepoState report.RepoState
	index        int
}

func gitRefreshRepo(m Model) tea.Cmd {
	index := m.tbl.Cursor()
	repoPath := m.report.RepoStates[index].Path

	return func() tea.Msg {
		newRepoState, _ := gitx.CheckRepoState(repoPath)

		return gitRefreshRepoResultMsg{
			newRepoState: newRepoState,
			index:        index,
		}
	}
}
