package git

import (
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Type() vcs.Type {
	return vcs.TypeGit
}

func (p *Provider) CheckRepoState(path string) (report.RepoState, []string) {
	state, warnings := gitx.CheckRepoState(path)
	state.VCSType = string(vcs.TypeGit)

	return state, warnings
}
