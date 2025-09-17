package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
)

type gitPullResultMsg struct {
	Err    string
	Output string
}

func gitPull(m Model) tea.Cmd {
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

func (msg gitPullResultMsg) updateUi(m Model) Model {
	if len(msg.Err) != 0 {
		// TODO: handle error
		fmt.Println("error pulling git repo=", msg)
		return m
	}

	m.messages = append(m.messages, fmt.Sprintf("git pull result msg=%s", msg))

	idx := m.tbl.Cursor()
	rs := m.report.RepoStates[idx]

	index := getRepoIndex(m.reposBeingUpdated, rs.ID)
	if index != -1 {
		m.reposBeingUpdated = deleteRepo(m.reposBeingUpdated, index)
	}

	return m

}
