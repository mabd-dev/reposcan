package git

import (
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
	state, warnings := CheckRepoState(path)
	state.VCSType = string(vcs.TypeGit)

	return state, warnings
}

func (p *Provider) Fetch(path string) (string, error) {
	str, err := RunGitCommand(path, "fetch", "--porcelain")
	if err != nil {
		return "", err
	}
	return str, nil
}

// GitPush pushed git repo at given path using `git push` command and returns stdout of the command + error if any
func (p *Provider) Push(path string) (string, error) {
	str, err := RunGitCommand(path, "push", "--porcelain")
	if err != nil {
		return "", err
	}
	return str, nil
}

func (p *Provider) Pull(path string) (string, error) {
	str, err := RunGitCommand(path, "pull")
	if err != nil {
		return "", err
	}
	return str, nil
}
