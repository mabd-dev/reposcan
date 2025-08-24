package gitx

import (
	"bytes"
	//"fmt"
	"os/exec"
	"strings"
)

// check https://git-scm.com/docs/git-status/2.11.4.html for file states
func UncommitedFiles(gitDir string) (changes []string, err error) {
	str, err := runGitCommand(gitDir, "status", "--porcelain=v1", "-uall")
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
