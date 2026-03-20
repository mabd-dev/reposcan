package jj

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mabd-dev/reposcan/internal/utils"
	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

type Provider struct {
	binary string
}

type trackedBookmark struct {
	Name   string
	Remote string
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

	repoName, err := p.getRepoName(path)
	if err != nil {
		warnings = append(warnings, "Failed to get jj repo name, path="+path)
	} else if strings.TrimSpace(repoName) != "" {
		state.Repo = repoName
	}

	branch, err := p.getBranchDisplay(path)
	if err != nil {
		warnings = append(warnings, "Failed to get jj branch display, path="+path)
	} else if strings.TrimSpace(branch) != "" {
		state.Branch = branch
	}

	uncommittedFiles, err := p.getUncommittedFiles(path)
	if err != nil {
		warnings = append(warnings, "Failed to get jj uncommitted files, path="+path)
	} else {
		state.UncommitedFiles = uncommittedFiles
	}

	outgoingCommits, err := p.getOutgoingCommits(path)
	if err != nil {
		warnings = append(warnings, "Failed to get jj outgoing commits, path="+path)
	} else {
		state.OutgoingCommits = outgoingCommits
		state.RemoteStatus[0].Ahead = len(outgoingCommits)
	}

	return state, warnings
}

func (p *Provider) getRepoName(repoPath string) (string, error) {
	remoteURL, err := p.getFirstRemoteURL(repoPath)
	if err != nil {
		return "", err
	}

	if remoteURL == "" {
		return filepath.Base(repoPath), nil
	}

	if repoName, ok := parseRepoName(remoteURL); ok {
		return repoName, nil
	}

	return filepath.Base(repoPath), nil
}

func (p *Provider) getFirstRemoteURL(repoPath string) (string, error) {
	output, err := p.run(repoPath, "git", "remote", "list")
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Done importing changes") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			return parts[1], nil
		}
	}

	return "", nil
}

func (p *Provider) getBranchDisplay(repoPath string) (string, error) {
	output, err := p.run(
		repoPath,
		"log",
		"-r",
		"@",
		"--no-graph",
		"-T",
		`bookmarks.join(",") ++ "|" ++ change_id.short() ++ "\n"`,
	)
	if err != nil {
		return "", err
	}

	line := strings.TrimSpace(output)
	if line == "" {
		return "-", nil
	}

	parts := strings.SplitN(line, "|", 2)
	if len(parts) != 2 {
		return line, nil
	}

	if strings.TrimSpace(parts[0]) != "" {
		return strings.TrimSpace(parts[0]), nil
	}

	if strings.TrimSpace(parts[1]) != "" {
		return strings.TrimSpace(parts[1]), nil
	}

	return "-", nil
}

func (p *Provider) getUncommittedFiles(repoPath string) ([]string, error) {
	output, err := p.run(repoPath, "diff", "--summary")
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

func (p *Provider) getOutgoingCommits(repoPath string) ([]string, error) {
	trackedBookmarks, err := p.getTrackedBookmarks(repoPath)
	if err != nil {
		return nil, err
	}

	if len(trackedBookmarks) == 0 {
		return []string{}, nil
	}

	revset := buildTrackedOutgoingRevset(trackedBookmarks)
	output, err := p.run(
		repoPath,
		"log",
		"-r",
		revset,
		"--no-graph",
		"-T",
		`commit_id.short() ++ "|" ++ description.first_line() ++ "\n"`,
	)
	if err != nil {
		return nil, err
	}

	commits := []string{}
	seen := map[string]struct{}{}

	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		commitID := strings.TrimSpace(parts[0])
		if commitID == "" {
			continue
		}
		if _, ok := seen[commitID]; ok {
			continue
		}
		seen[commitID] = struct{}{}

		description := ""
		if len(parts) == 2 {
			description = strings.TrimSpace(parts[1])
		}
		if description == "" {
			description = "(no description set)"
		}

		commits = append(commits, commitID+" "+description)
	}

	return commits, nil
}

func (p *Provider) getTrackedBookmarks(repoPath string) ([]trackedBookmark, error) {
	output, err := p.run(
		repoPath,
		"bookmark",
		"list",
		"--tracked",
		"-T",
		`name ++ "|" ++ remote ++ "\n"`,
	)
	if err != nil {
		return nil, err
	}

	bookmarks := []trackedBookmark{}
	seen := map[string]struct{}{}

	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		remote := strings.TrimSpace(parts[1])
		if name == "" || remote == "" {
			continue
		}

		key := name + "|" + remote
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		bookmarks = append(bookmarks, trackedBookmark{
			Name:   name,
			Remote: remote,
		})
	}

	return bookmarks, nil
}

func (p *Provider) run(repoPath string, args ...string) (string, error) {
	cmd := exec.Command(p.binary, append([]string{"-R", repoPath}, args...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}

	return stdout.String(), nil
}

func buildTrackedOutgoingRevset(bookmarks []trackedBookmark) string {
	parts := make([]string, 0, len(bookmarks))

	for _, bookmark := range bookmarks {
		parts = append(parts, fmt.Sprintf(
			`(remote_bookmarks("%s", remote="%s")..bookmarks("%s"))`,
			escapeRevsetString(bookmark.Name),
			escapeRevsetString(bookmark.Remote),
			escapeRevsetString(bookmark.Name),
		))
	}

	return strings.Join(parts, " | ")
}

func escapeRevsetString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)

	return s
}

func parseRepoName(remote string) (string, bool) {
	if strings.Contains(remote, ":") && strings.Contains(remote, "@") && !strings.Contains(remote, "://") {
		parts := strings.SplitN(remote, ":", 2)
		if len(parts) == 2 {
			remote = "ssh://" + parts[0] + "/" + parts[1]
		}
	}

	if u, err := url.Parse(remote); err == nil && u.Path != "" {
		base := path.Base(u.Path)
		base = strings.TrimSuffix(base, ".git")
		return base, true
	}

	re := regexp.MustCompile(`([^/\\]+?)(?:\.git)?[/\\]?$`)
	if match := re.FindStringSubmatch(remote); len(match) > 1 {
		return match[1], true
	}

	return "", false
}
