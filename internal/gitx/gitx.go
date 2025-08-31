package gitx

import (
	"bytes"
	"errors"
	"github.com/MABD-dev/RepoScan/internal/render"
	"os/exec"
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

	return GitRepo{
		Path:     path,
		RepoName: repoName,
		Branch:   branch,
	}
}

func getGitRepoName(path string) (repoName string, err error) {
	remote, err := runGitCommand(path, "remote", "get-url", "origin")
	if err != nil {
		return "-", err
	}

	remote = strings.TrimSpace(remote)

	re := regexp.MustCompile(`([^/]+?)(?:\.git)?$`)
	match := re.FindStringSubmatch(remote)
	if len(match) > 1 {
		repoName = match[1]
	} else {
		return "", errors.New("repo name cannot be found")
	}

	return repoName, nil
}

func getGitRepoBranch(path string) (branchName string, err error) {
	str, err := runGitCommand(path, "branch", "--show-current")
	if err != nil {
		return "-", err
	}
	return strings.TrimSpace(str), nil
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
