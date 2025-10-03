package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type gitFetchResultMsg struct {
	Err    string
	Output string
}

func gitFetch(m Model) tea.Cmd {
	rs := m.reposTable.GetCurrentRepoState()
	if rs == nil {
		return nil
	}
	repoPath := rs.Path

	m.reposBeingUpdated = append(m.reposBeingUpdated, rs.ID)

	return func() tea.Msg {
		stdout, err := gitx.GitFetch(repoPath)

		errMessage := ""
		if err != nil {
			errMessage = err.Error()
		}

		return gitFetchResultMsg{
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
	rs := m.reposTable.GetCurrentRepoState()
	if rs == nil {
		return nil
	}

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

type gitPushResultMsg struct {
	Err    string
	Output string
}

func gitPush(m Model) tea.Cmd {
	rs := m.reposTable.GetCurrentRepoState()
	if rs == nil {
		return nil
	}

	repoPath := rs.Path

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

type gitRefreshRepoResultMsg struct {
	newRepoState report.RepoState
	index        int
}

func gitRefreshRepo(m Model) tea.Cmd {
	index := m.reposTable.Cursor()

	rs := m.reposTable.GetRepoStateAt(index)
	if rs == nil {
		return nil
	}

	repoPath := rs.Path

	return func() tea.Msg {
		newRepoState, _ := gitx.CheckRepoState(repoPath)

		return gitRefreshRepoResultMsg{
			newRepoState: newRepoState,
			index:        index,
		}
	}
}
