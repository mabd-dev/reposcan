package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/render/tui/common"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type gitFetchResultMsg struct {
	Err    string
	Output string
}

func gitFetch(m Model) tea.Cmd {
	rs := m.reposTable.GetCurrentWorktreeState()
	if rs == nil {
		return nil
	}
	repoPath := rs.Path

	m.reposBeingUpdated = append(m.reposBeingUpdated, rs.RepoID)

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
	rs := m.reposTable.GetCurrentWorktreeState()
	if rs == nil {
		return nil
	}

	repoPath := rs.Path
	m.reposBeingUpdated = append(m.reposBeingUpdated, rs.RepoID)

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
	rs := m.reposTable.GetCurrentWorktreeState()
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
	newWorktreeState common.WorktreeState
	index            int
}

func gitRefreshRepo(m Model) tea.Cmd {
	index := m.reposTable.Cursor()

	rs := m.reposTable.GetWorktreeStateAt(index)
	if rs == nil {
		return nil
	}

	worktreePath := rs.Path

	return func() tea.Msg {
		newWorktreeState, _ := gitx.GetWorktreeState(worktreePath)
		repoState, _ := gitx.GetRepoState(worktreePath, []report.Worktree{newWorktreeState})
		worktreeState := common.MapToWorktreeStates(repoState)[0]

		return gitRefreshRepoResultMsg{
			newWorktreeState: worktreeState,
			index:            index,
		}
	}
}
