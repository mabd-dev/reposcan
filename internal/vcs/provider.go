package vcs

import "github.com/mabd-dev/reposcan/pkg/report"

type Provider interface {
	Type() Type
	CheckRepoState(path string) (report.RepoState, []string)
}
