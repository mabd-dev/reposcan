package gitx

import (
	"bytes"
	"errors"
	//"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func CreateGitReposFrom(paths []string) (gitRepos []GitRepo) {
	for _, p := range paths {
		gitRepo := GitRepo{
			Path: p,
		}
		gitRepos = append(gitRepos, gitRepo)
	}
	return gitRepos
}

func (r GitRepo) GitRepoName() (repoName string, err error) {
	remote, err := runGitCommand(r.Path, "remote", "get-url", "origin")
	if err != nil {
		return "", err
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

func (r GitRepo) GitRepoBranch() (branchName string, err error) {
	str, err := runGitCommand(r.Path, "branch", "--show-current")
	if err != nil {
		return "", err
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
