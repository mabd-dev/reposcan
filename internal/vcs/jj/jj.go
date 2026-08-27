package jj

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Provider struct {
	binary string
}

func New() *Provider {
	return &Provider{binary: "jj"}
}

func (p *Provider) Type() vcs.Type {
	return vcs.TypeJJ
}

func (p *Provider) CheckRepoState(path string) (report.RepoState, []string) {
	state := report.RepoState{
		ID:      utils.Hash(path),
		Path:    path,
		Repo:    filepath.Base(path),
		Branch:  "-",
		VCSType: string(vcs.TypeJJ),
		RemoteStatus: []report.RemoteStatus{{
			Remote: "",
			Ahead:  0,
			Behind: 0,
		}},
	}

	if _, err := exec.LookPath(p.binary); err != nil {
		return state, []string{
			fmt.Sprintf("Failed to inspect jj repo, path=%s: %v", path, err),
		}
	}

	warnings := []string{}

	repoName, err := getRepoName(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get repo name", path, err))
	} else if strings.TrimSpace(repoName) != "" {
		state.Repo = repoName
	}

	branch, err := getBranchDisplay(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get branch display", path, err))
	} else if strings.TrimSpace(branch) != "" {
		state.Branch = branch
	}

	uncommittedFiles, err := getUncommittedFiles(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get uncommitted files", path, err))
	} else {
		state.UncommitedFiles = uncommittedFiles
	}

	remoteStatuses, err := getBookmarkRemoteStatuses(p.binary, path, strings.Split(state.Branch, ","))
	if err != nil {
		warnings = append(warnings, jjWarning("get remote status", path, err))
	} else if len(remoteStatuses) > 0 {
		state.RemoteStatus = make([]report.RemoteStatus, 0, len(remoteStatuses))
		for _, remoteStatus := range remoteStatuses {
			state.RemoteStatus = append(state.RemoteStatus, report.RemoteStatus{
				Remote:          remoteStatus.Remote,
				Ahead:           len(remoteStatus.OutgoingCommits),
				Behind:          len(remoteStatus.IncomingCommits),
				OutgoingCommits: remoteStatus.OutgoingCommits,
			})
		}
	}

	return state, warnings
}

func jjWarning(operation string, path string, err error) string {
	return fmt.Sprintf("Failed to %s for jj repo, path=%s: %v", operation, path, err)
}

// JJFetch fetches remote bookmark state using jj's Git interop.
func (p *Provider) Fetch(path string) (string, error) {
	return RunJJCommand(path, "git", "fetch")
}

// JJPush is intentionally not wired into vcs.ActionProvider yet. Define the
// bookmark selection/update semantics before enabling this operation.
func (p *Provider) Push(path string) (string, error) {
	return "", fmt.Errorf("%w: push bookmark behavior needs to be defined", ErrJJActionNotImplemented)
}

// JJPull is intentionally not wired into vcs.ActionProvider yet. jj does not
// have a direct Git-equivalent pull operation, so the desired behavior needs to
// be defined before enabling this operation.
func (p *Provider) Pull(path string) (string, error) {
	return "", fmt.Errorf("%w: pull has no direct Git-equivalent jj operation", ErrJJActionNotImplemented)
}
