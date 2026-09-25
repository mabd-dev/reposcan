package vcs

import (
	"time"

	"github.com/mabd-dev/reposcan/pkg/report"
)

type Provider interface {
	Type() Type
	CheckRepoState(path string) (report.RepoState, []string)
}

type ActionProvider interface {
	Fetch(path string) (string, error)
	Push(path string) (string, error)
	Pull(path string) (string, error)
}

// ActivityProvider is implemented by providers that can report when a repo was
// last worked on locally. state is the result of CheckRepoState for the same
// path, so implementations can reuse data it already collected.
// A zero time means the activity is unknown.
type ActivityProvider interface {
	LastActivity(path string, state report.RepoState) (time.Time, []string)
}
