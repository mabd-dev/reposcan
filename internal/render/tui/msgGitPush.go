package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
)

type gitPushResultMsg struct {
	Err    string
	Output string
}

func gitPush(m Model) tea.Cmd {
	idx := m.tbl.Cursor()
	repoPath := m.report.RepoStates[idx].Path

	return func() tea.Msg {
		stdout, err := gitx.GitPush(repoPath)

		errMessage := ""
		if err != nil {
			errMessage = err.Error()
		}

		return gitPushResultMsg{
			Err:    errMessage,
			Output: stdout,
		}
	}
}

func (msg gitPushResultMsg) updateUi(m Model) Model {
	m.messages = append(m.messages, fmt.Sprintf("git push result msg=%s", msg))
	return m
}
