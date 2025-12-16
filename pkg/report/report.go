// Package report defines public types representing the output of a repository
// scan. Renderers and external tools consume these types to display or persist
// results.
package report

import (
	"time"
)

// RepoState describes the state of a single Git repository discovered during a scan.
type RepoState struct {
	ID        string     `json:"id"`
	Repo      string     `json:"repo"`
	Worktrees []Worktree `json:"worktrees"`
}

type Worktree struct {
	Name            string         `json:"name"`
	Path            string         `json:"path"`
	Branch          string         `json:"branch"`
	UncommitedFiles []string       `json:"uncommitedFiles"`
	RemoteStatus    []RemoteStatus `json:"remoteStatus"`
}

type RemoteStatus struct {
	Remote string `json:"remote"`
	Ahead  int    `json:"ahead"`
	Behind int    `json:"behind"`
}

// ScanReport aggregates the results of scanning one or more directories for
// Git repositories and summarizing their status.
type ScanReport struct {
	Version     int         `json:"version"`
	RepoStates  []RepoState `json:"repoStates"`
	GeneratedAt time.Time   `json:"generatedAt"`
	Warnings    []string    `json:"warnings"`
}

// IsDirty reports whether the repository has uncommitted changes or is ahead/behind.
func IsDirty(r Worktree) bool {
	atLeastOneDirtyRemote := false
	for _, remoteStatus := range r.RemoteStatus {
		if remoteStatus.Ahead > 0 || remoteStatus.Behind > 0 {
			atLeastOneDirtyRemote = true
		}
	}
	return len(r.UncommitedFiles) > 0 || atLeastOneDirtyRemote
}

func HaveUnpushedCommits(r Worktree) bool {
	for _, remoteStatus := range r.RemoteStatus {
		if remoteStatus.Ahead > 0 {
			return true
		}
	}
	return false
}

func HaveUnpulledCommits(r Worktree) bool {
	for _, remoteStatus := range r.RemoteStatus {
		if remoteStatus.Behind > 0 {
			return true
		}
	}
	return false
}

// DirtyWorktreeCount count all dirty repos based on [IsDirty] function on RepoState struct
func (sc *ScanReport) DirtyWorktreeCount() int {
	count := 0
	for _, rs := range sc.RepoStates {
		for _, wt := range rs.Worktrees {
			if IsDirty(wt) {
				count++
			}
		}
	}
	return count
}
