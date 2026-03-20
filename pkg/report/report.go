// Package report defines public types representing the output of a repository
// scan. Renderers and external tools consume these types to display or persist
// results.
package report

import (
	"time"
)

type RemoteStatus struct {
	Remote string `json:"remote"`
	Ahead  int    `json:"ahead"`
	Behind int    `json:"behind"`
}

// RepoState describes the state of a single repository discovered during a scan.
type RepoState struct {
	ID              string         `json:"id"`
	Path            string         `json:"path"`
	Repo            string         `json:"repo"`
	VCSType         string         `json:"vcsType"`
	Branch          string         `json:"branch"`
	UncommitedFiles []string       `json:"uncommitedFiles"`
	OutgoingCommits []string       `json:"outgoingCommits"`
	RemoteStatus    []RemoteStatus `json:"remoteStatus"`
}

// ScanReport aggregates the results of scanning one or more directories for
// repositories and summarizing their status.
type ScanReport struct {
	Version     int         `json:"version"`
	RepoStates  []RepoState `json:"repoStates"`
	GeneratedAt time.Time   `json:"generatedAt"`
	Warnings    []string    `json:"warnings"`
}

// IsDirty reports whether the repository has uncommitted changes or is ahead/behind.
func (r *RepoState) IsDirty() bool {
	atLeastOneDirtyRemote := false
	for _, remoteStatus := range r.RemoteStatus {
		if remoteStatus.Ahead > 0 || remoteStatus.Behind > 0 {
			atLeastOneDirtyRemote = true
		}
	}
	return len(r.UncommitedFiles) > 0 || atLeastOneDirtyRemote
}

func (r *RepoState) HaveUnpushedCommits() bool {
	for _, remoteStatus := range r.RemoteStatus {
		if remoteStatus.Ahead > 0 {
			return true
		}
	}
	return false
}

func (r *RepoState) HaveUnpulledCommits() bool {
	for _, remoteStatus := range r.RemoteStatus {
		if remoteStatus.Behind > 0 {
			return true
		}
	}
	return false
}

// DirtyReposCount count all dirty repos based on [IsDirty] function on RepoState struct
func (sc *ScanReport) DirtyReposCount() int {
	dirtyRepos := 0
	for _, rs := range sc.RepoStates {
		if rs.IsDirty() {
			dirtyRepos++
		}
	}
	return dirtyRepos
}
