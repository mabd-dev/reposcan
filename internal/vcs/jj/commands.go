package jj

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrJJActionNotImplemented = errors.New("jj action semantics are not implemented")

// JJFetch fetches remote bookmark state using jj's Git interop.
func JJFetch(path string) (string, error) {
	return RunJJCommand(path, "git", "fetch")
}

// JJPush is intentionally not wired into vcs.ActionProvider yet. Define the
// bookmark selection/update semantics before enabling this operation.
func JJPush(path string) (string, error) {
	return "", fmt.Errorf("%w: push bookmark behavior needs to be defined", ErrJJActionNotImplemented)
}

// JJPull is intentionally not wired into vcs.ActionProvider yet. jj does not
// have a direct Git-equivalent pull operation, so the desired behavior needs to
// be defined before enabling this operation.
func JJPull(path string) (string, error) {
	return "", fmt.Errorf("%w: pull has no direct Git-equivalent jj operation", ErrJJActionNotImplemented)
}

// RunJJCommand executes a jj command in repoPath and returns stdout.
func RunJJCommand(repoPath string, args ...string) (string, error) {
	return runJJCommand("jj", repoPath, args...)
}

func runJJCommand(binary string, repoPath string, args ...string) (string, error) {
	cmd := exec.Command(binary, append([]string{"-R", repoPath}, args...)...)

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
