package common

import "github.com/mabd-dev/reposcan/pkg/report"

type Keybinding struct {
	Key         string
	Description string
	ShortDesc   string
}

type WorktreeState struct {
	RepoID          string
	RepoName        string
	Path            string
	Branch          string
	UncommitedFiles []string
	RemoteStatus    []report.RemoteStatus
}

func MapToWorktreeStates(rs report.RepoState) []WorktreeState {
	worktreeStates := []WorktreeState{}
	for _, wt := range rs.Worktrees {
		worktreeStates = append(worktreeStates, WorktreeState{
			RepoID:          rs.ID,
			RepoName:        rs.Repo,
			Path:            wt.Path,
			Branch:          wt.Branch,
			UncommitedFiles: wt.UncommitedFiles,
			RemoteStatus:    wt.RemoteStatus,
		})
	}

	return worktreeStates
}
