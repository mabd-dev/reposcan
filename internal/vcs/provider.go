package vcs

import "github.com/mabd-dev/reposcan/pkg/report"

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
// last worked on locally. CheckRepoStateWithActivity returns the same state as
// Provider.CheckRepoState with LastActivity filled in, in a single pass so the
// provider can reuse (and know the success of) the commands it already ran.
// A zero LastActivity means the activity is unknown.
type ActivityProvider interface {
	CheckRepoStateWithActivity(path string) (report.RepoState, []string)
}
