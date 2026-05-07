package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mabd-dev/reposcan/internal/render/tui/alerts"
	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type vcsAction string

const (
	vcsActionFetch vcsAction = "fetch"
	vcsActionPull  vcsAction = "pull"
	vcsActionPush  vcsAction = "push"
)

type vcsActionResultMsg struct {
	Action vcsAction
	RepoID string
	Index  int
	Err    string
	Output string
}

func (m Model) runVCSAction(action vcsAction) (Model, tea.Cmd) {
	rs := m.reposTable.GetCurrentRepoState()
	if rs == nil {
		return m, nil
	}

	actionProvider, ok := m.actionProvider(*rs)
	if !ok {
		return m, unsupportedVCSActionAlert(action, rs.VCSType)
	}

	repoID := rs.ID
	repoPath := rs.Path
	index := m.reposTable.Cursor()
	m.reposBeingUpdated = append(m.reposBeingUpdated, repoID)

	return m, func() tea.Msg {
		stdout, err := runAction(actionProvider, action, repoPath)

		errMessage := ""
		if err != nil {
			errMessage = err.Error()
		}

		return vcsActionResultMsg{
			Action: action,
			RepoID: repoID,
			Index:  index,
			Err:    errMessage,
			Output: stdout,
		}
	}
}

func (m Model) actionProvider(rs report.RepoState) (vcs.ActionProvider, bool) {
	if m.vcsRegistry == nil {
		return nil, false
	}

	return m.vcsRegistry.GetActionProvider(vcs.Type(rs.VCSType))
}

func runAction(provider vcs.ActionProvider, action vcsAction, repoPath string) (string, error) {
	switch action {
	case vcsActionFetch:
		return provider.Fetch(repoPath)
	case vcsActionPull:
		return provider.Pull(repoPath)
	case vcsActionPush:
		return provider.Push(repoPath)
	default:
		return "", fmt.Errorf("unsupported VCS action %q", action)
	}
}

func unsupportedVCSActionAlert(action vcsAction, vcsType string) tea.Cmd {
	if vcsType == "" {
		vcsType = "unknown"
	}

	return func() tea.Msg {
		return alerts.AddAlertMsg{
			Msg: alerts.Alert{
				Type:    alerts.AlertTypeWarning,
				Title:   "Unsupported action",
				Message: fmt.Sprintf("%s is not supported for %s repositories", action, vcsType),
			},
		}
	}
}

type vcsRefreshRepoResultMsg struct {
	newRepoState report.RepoState
	index        int
}

func refreshRepo(m Model, index int) tea.Cmd {
	rs := m.reposTable.GetRepoStateAt(index)
	if rs == nil {
		return nil
	}

	if m.vcsRegistry == nil {
		return unsupportedVCSActionAlert("refresh", rs.VCSType)
	}

	provider, ok := m.vcsRegistry.Get(vcs.Type(rs.VCSType))
	if !ok {
		return unsupportedVCSActionAlert("refresh", rs.VCSType)
	}

	repoPath := rs.Path

	return func() tea.Msg {
		newRepoState, _ := provider.CheckRepoState(repoPath)

		return vcsRefreshRepoResultMsg{
			newRepoState: newRepoState,
			index:        index,
		}
	}
}
