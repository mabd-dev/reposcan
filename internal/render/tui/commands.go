package tui

import (
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

type gitPullResultMsg struct {
	Err    string
	Output string
}

func gitPull(m Model) tea.Cmd {
	idx := m.tbl.Cursor()
	repoPath := m.report.RepoStates[idx].Path

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
