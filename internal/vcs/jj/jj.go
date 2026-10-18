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
		// -1 renders as "x", as for a git repo without a remote.
		RemoteStatus: []report.RemoteStatus{{
			Remote: "",
			Ahead:  -1,
			Behind: -1,
		}},
	}

	if _, err := exec.LookPath(p.binary); err != nil {
		return state, []string{
			fmt.Sprintf("Failed to inspect jj repo, path=%s: %v", path, err),
		}
	}

	warnings := []string{}

	remotes, remotesErr := getRemotes(p.binary, path)
	if remotesErr != nil {
		warnings = append(warnings, jjWarning("list remotes", path, remotesErr))
	} else {
		state.Repo = repoNameFromRemotes(remotes, path)
	}

	// A change ID fallback is not a bookmark name, so it is never compared with
	// remote bookmarks.
	var bookmarkNames []string
	branch, isBookmark, err := getBranch(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get branch display", path, err))
	} else if strings.TrimSpace(branch) != "" {
		state.Branch = branch
		if isBookmark {
			bookmarkNames = strings.Split(branch, ",")
		}
	}

	uncommittedFiles, err := getUncommittedFiles(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get uncommitted files", path, err))
	} else {
		state.UncommitedFiles = uncommittedFiles
	}

	if remotesErr != nil || len(remotes) == 0 {
		return state, warnings
	}

	remoteNames := make([]string, 0, len(remotes))
	for _, remote := range remotes {
		remoteNames = append(remoteNames, remote.Name)
	}

	remoteStatuses, err := getBookmarkRemoteStatuses(p.binary, path, bookmarkNames, remoteNames)
	if err != nil {
		warnings = append(warnings, jjWarning("get remote status", path, err))
	} else if len(remoteStatuses) > 0 {
		state.RemoteStatus = make([]report.RemoteStatus, 0, len(remoteStatuses))
		for _, remoteStatus := range remoteStatuses {
			if remoteStatus.NoRemoteBookmark {
				state.RemoteStatus = append(state.RemoteStatus, report.RemoteStatus{
					Remote: remoteStatus.Remote,
					Ahead:  -1,
					Behind: -1,
				})
				continue
			}

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
