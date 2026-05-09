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

	outgoingCommits, err := getOutgoingCommits(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get outgoing commits", path, err))
	} else {
		state.RemoteStatus[0].Ahead = len(outgoingCommits)
		state.RemoteStatus[0].OutgoingCommits = outgoingCommits
	}

	incomingCommits, err := getIncomingCommits(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get incoming commits", path, err))
	} else {
		state.RemoteStatus[0].Behind = len(incomingCommits)
	}

	return state, warnings
}

func jjWarning(operation string, path string, err error) string {
	return fmt.Sprintf("Failed to %s for jj repo, path=%s: %v", operation, path, err)
}
