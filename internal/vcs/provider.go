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
