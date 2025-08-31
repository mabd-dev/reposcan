package gitx

import (
	"bytes"
	"errors"
	"github.com/MABD-dev/RepoScan/internal/render"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func CreateGitRepoFrom(path string) (gitRepo GitRepo) {
	repoName, err := getGitRepoName(path)
	if err != nil {
		msg := "Failed to get repo name, path=" + path + " error=" + err.Error() + "\n"
		render.Warning(msg)
	}

	branch, err := getGitRepoBranch(path)
	if err != nil {
		msg := "Failed to get branch name, path=" + path + ", error=" + err.Error() + "\n"
		render.Warning(msg)
	}

	ahead, behind, err := getUpstreamStatus(path)
	if err != nil {
		msg := "Failed to get upstream status, path=" + path + ", error=" + err.Error() + "\n"
		render.Warning(msg)
	}

	return GitRepo{
		Path:     path,
		RepoName: repoName,
		Branch:   branch,
		Ahead:    ahead,
		Behind:   behind,
	}
}

// getGitRepoName tries to extract the repo name from a remote URL,
// falling back to first remote or local folder name if needed.
func getGitRepoName(repoPath string) (string, error) {
	// 1. Try "origin" first
	remote, err := runGitCommand(repoPath, "remote", "get-url", "origin")
	if err != nil {
		// 2. If "origin" not found, list remotes
		remotes, rErr := runGitCommand(repoPath, "remote")
		if rErr == nil {
			names := strings.Fields(remotes)
			if len(names) > 0 {
				remote, err = runGitCommand(repoPath, "remote", "get-url", names[0])
				if err != nil {
					remote = ""
				}
			}
		}
	}

	remote = strings.TrimSpace(remote)
	if remote != "" {
		if name, ok := parseRepoName(remote); ok {
			return name, nil
		}
	}

	// 3. Fallback to repo folder name
	top, err := runGitCommand(repoPath, "rev-parse", "--show-toplevel")
	if err == nil {
		return filepath.Base(strings.TrimSpace(top)), nil
	}

	return "", errors.New("could not determine repo name")
}

// parseRepoName extracts the repo name from a remote URL or path.
func parseRepoName(remote string) (string, bool) {
	// handle scp-like: git@host:org/repo.git
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

	// fallback regex
	re := regexp.MustCompile(`([^/\\]+?)(?:\.git)?[/\\]?$`)
	if match := re.FindStringSubmatch(remote); len(match) > 1 {
		return match[1], true
	}

	return "", false
}

func getGitRepoBranch(path string) (branchName string, err error) {
	str, err := runGitCommand(path, "branch", "--show-current")
	if err != nil {
		return "-", err
	}
	return strings.TrimSpace(str), nil
}

func getUpstreamStatus(path string) (ahead int, behind int, err error) {
	lrc, err := runGitCommand(path, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	if err != nil {
		return -1, -1, err
	}
	parts := strings.Fields(strings.TrimSpace(lrc))
	if len(parts) == 2 {
		behind = atoiSafe(parts[0])
		ahead = atoiSafe(parts[1])
	}

	return ahead, behind, nil
}

// check https://git-scm.com/docs/git-status/2.11.4.html for file states
func (r GitRepo) UncommitedFiles() (changes []string, err error) {
	str, err := runGitCommand(r.Path, "status", "--porcelain=v1", "-uall")
	if err != nil {
		return []string{}, err
	}

	changes = strings.Split(strings.TrimRight(str, "\n"), "\n")
	changes = removeEmptyStrings(changes)

	return changes, nil
}

func runGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func removeEmptyStrings(input []string) []string {
	result := []string{}
	for _, s := range input {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

func atoiSafe(s string) int {
	var n int
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}
