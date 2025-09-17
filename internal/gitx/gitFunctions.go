package gitx

import (
	"bytes"
	"errors"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// GitPush pushed git repo at given path using `git push` command and returns stdout of the command + error if any
func GitPush(path string) (string, error) {
	str, err := RunGitCommand(path, "push", "--porcelain")
	if err != nil {
		return "", err
	}
	return str, nil
}

func GitPull(path string) (string, error) {
	str, err := RunGitCommand(path, "pull")
	if err != nil {
		return "", err
	}
	return str, nil
}

func GitFetch(path string) (string, error) {
	str, err := RunGitCommand(path, "fetch", "--porcelain")
	if err != nil {
		return "", err
	}
	return str, nil
}

// GetRepoBranch returns the current branch name for the Git repository at path.
func GetRepoBranch(path string) (branchName string, err error) {
	str, err := RunGitCommand(path, "branch", "--show-current")
	if err != nil {
		return "-", err
	}
	return strings.TrimSpace(str), nil
}

// GetUncommitedFiles returns the list of uncommitted files (status porcelain)
// for the Git repository at path.
func GetUncommitedFiles(path string) (changes []string, err error) {
	str, err := RunGitCommand(path, "status", "--porcelain=v1", "-uall")
	if err != nil {
		return []string{}, err
	}

	changes = strings.Split(strings.TrimRight(str, "\n"), "\n")
	changes = removeEmptyStrings(changes)

	return changes, nil
}

// GetUpstreamStatus returns the ahead/behind counts relative to the upstream
// tracking branch for the repository at path.
func GetUpstreamStatus(path string) (ahead int, behind int, err error) {
	lrc, err := RunGitCommand(path, "rev-list", "--left-right", "--count", "@{u}...HEAD")
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

// GetRepoName tries to extract the repository name from its remote URL,
// falling back to the first remote name or the local folder name if needed.
func GetRepoName(repoPath string) (string, error) {
	// 1. Try "origin" first
	remote, err := RunGitCommand(repoPath, "remote", "get-url", "origin")
	if err != nil {
		// 2. If "origin" not found, list remotes
		remotes, rErr := RunGitCommand(repoPath, "remote")
		if rErr == nil {
			names := strings.Fields(remotes)
			if len(names) > 0 {
				remote, err = RunGitCommand(repoPath, "remote", "get-url", names[0])
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
	top, err := RunGitCommand(repoPath, "rev-parse", "--show-toplevel")
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

// RunGitCommand executes a git command in dir and returns its stdout as a string.
// Stderr is discarded, and non-zero exit codes are returned as errors.
func RunGitCommand(dir string, args ...string) (string, error) {
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
