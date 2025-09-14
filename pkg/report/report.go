// Package report defines public types representing the output of a repository
// scan. Renderers and external tools consume these types to display or persist
// results.
package report

import (
	"time"
)

// RepoState describes the state of a single Git repository discovered during a scan.
type RepoState struct {
	ID              string   `json:"id"`
	Path            string   `json:"path"`
	Repo            string   `json:"repo"`
	Branch          string   `json:"branch"`
	UncommitedFiles []string `json:"uncommitedFiles"`
	Ahead           int      `json:"ahead"`
	Behind          int      `json:"behind"`
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
func (r *RepoState) IsDirty() bool {
	return len(r.UncommitedFiles) > 0 || r.Ahead > 0 || r.Behind > 0
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
