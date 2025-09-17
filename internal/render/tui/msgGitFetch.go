package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
)

type gitFetchResultMsg struct {
	Err    string
	Output string
}

func gitFetch(m Model) tea.Cmd {
	idx := m.tbl.Cursor()

	rs := m.report.RepoStates[idx]
	repoPath := rs.Path

	m.reposBeingUpdated = append(m.reposBeingUpdated, rs.ID)

	return func() tea.Msg {
		stdout, err := gitx.GitPull(repoPath)

		errMessage := ""
		if err != nil {
			errMessage = err.Error()
		}

		return gitPullResultMsg{
			Err:    errMessage,
			Output: stdout,
		}
	}
}
