package git

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// RemoteStatus holds the ahead/behind status for a specific remote branch
type remoteStatus struct {
	Ahead  int
	Behind int
}

func GetGitRemotes(path string) (remotes []string, err error) {
	str, err := RunGitCommand(path, "remote")
	if err != nil {
		return []string{}, err
	}

	str = strings.TrimSpace(str)
	if len(str) == 0 {
		return []string{}, nil
	}

	remotes = strings.Split(strings.TrimRight(str, "\n"), "\n")
	return remotes, nil
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
	changes, _, err = getUncommittedChanges(path)
	return changes, err
}

// getUncommittedChanges is GetUncommitedFiles that also returns the current
// path of each change. Paths are taken from the -z output rather than parsed
// back out of the display lines, where " -> " is ambiguous in file names.
func getUncommittedChanges(path string) (changes []string, paths []string, err error) {
	// -z prints paths verbatim. Without it git C-quotes paths that contain
	// spaces or non-ASCII characters.
	str, err := RunGitCommand(path, "status", "--porcelain=v1", "-z", "-uall")
	if err != nil {
		return []string{}, nil, err
	}

	changes, paths = parsePorcelainZ(str)
	return changes, paths, nil
}

// parsePorcelainZ converts `git status --porcelain=v1 -z` output into the
// "XY path" and "XY orig -> path" lines of the newline-separated format, and
// the current path of each entry.
func parsePorcelainZ(output string) (changes []string, paths []string) {
	changes = []string{}
	fields := strings.Split(output, "\x00")

	for i := 0; i < len(fields); i++ {
		entry := fields[i]
		if len(entry) < 4 {
			continue
		}
		paths = append(paths, entry[3:])

		// A rename or copy is followed by a separate field with the original path.
		if strings.ContainsAny(entry[:2], "RC") && i+1 < len(fields) {
			i++
			entry = entry[:3] + fields[i] + " -> " + entry[3:]
		}

		changes = append(changes, entry)
	}

	return changes, paths
}

// GetStashes returns the list of stash entries (one per `git stash list` line)
// for the Git repository at path. Stashes are stored in the shared refs/stash
// of the common dir, so linked worktrees report the same stashes.
func GetStashes(path string) (stashes []string, err error) {
	str, err := RunGitCommand(path, "stash", "list")
	if err != nil {
		return []string{}, err
	}

	stashes = strings.Split(strings.TrimRight(str, "\n"), "\n")
	stashes = removeEmptyStrings(stashes)

	return stashes, nil
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

// GetUpstreamStatusForRemote returns the ahead/behind counts for the current
// branch against the same branch on the given remote.
func GetUpstreamStatusForRemote(
	path string,
	remote string,
	currentBranch string,
) (remoteStatus, error) {

	// Construct remote branch ref: remote/branch
	remoteBranchRef := remote + "/" + currentBranch

	// Check if remote branch exists
	_, err := RunGitCommand(path, "rev-parse", "--verify", remoteBranchRef)
	if err != nil {
		// Remote branch doesn't exist, skip this remote
		return remoteStatus{}, err
	}

	// Get ahead/behind count for this remote branch
	lrc, err := RunGitCommand(path, "rev-list", "--left-right", "--count", remoteBranchRef+"...HEAD")
	if err != nil {
		return remoteStatus{}, err
	}

	parts := strings.Fields(strings.TrimSpace(lrc))
	var ahead, behind int
	if len(parts) == 2 {
		behind = atoiSafe(parts[0])
		ahead = atoiSafe(parts[1])
	}

	return remoteStatus{
		Ahead:  ahead,
		Behind: behind,
	}, nil
}

func GetOutgoingCommitsForRemote(path string, remote string, currentBranch string) ([]string, error) {
	remoteBranchRef := remote + "/" + currentBranch

	str, err := RunGitCommand(path, "log", "--format=%h %s", remoteBranchRef+"..HEAD")
	if err != nil {
		return []string{}, err
	}

	commits := strings.Split(strings.TrimRight(str, "\n"), "\n")
	commits = removeEmptyStrings(commits)

	return commits, nil
}

// GetRepoName tries to extract the repository name from its remote URL,
// falling back to the first remote name or the local folder name if needed.
func GetRepoName(repoPath string) (string, error) {
	// 1. Try "origin" first
	remote, err := RunGitCommand(repoPath, "remote", "get-url", "origin")
	if err != nil {
		remote, _ = firstRemoteURL(repoPath)
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

func firstRemoteURL(repoPath string) (string, error) {
	remotes, err := RunGitCommand(repoPath, "remote")
	if err != nil {
		return "", err
	}
	remoteNames := strings.Fields(remotes)
	if len(remoteNames) == 0 {
		return "", errors.New("no remotes")
	}

	remote, err := RunGitCommand(repoPath, "remote", "get-url", remoteNames[0])
	return remote, err
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
// Stderr is attached to the returned error so callers can show git's own
// explanation (missing upstream, conflicts, auth failures) to the user.
func RunGitCommand(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}
	return stdout.String(), nil
}
