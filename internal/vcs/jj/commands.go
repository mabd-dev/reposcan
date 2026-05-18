package jj

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

type trackedBookmark struct {
	Name   string
	Remote string
}

type bookmarkRemoteStatus struct {
	Remote          string
	OutgoingCommits []string
	IncomingCommits []string
}

var ErrJJActionNotImplemented = errors.New("jj action semantics are not implemented")

type commandError struct {
	Binary   string
	RepoPath string
	Args     []string
	Stderr   string
	Err      error
}

func (e commandError) Error() string {
	command := append([]string{e.Binary, "-R", e.RepoPath}, e.Args...)
	message := fmt.Sprintf("command=%q failed: %v", strings.Join(command, " "), e.Err)
	if e.Stderr != "" {
		message += ": " + e.Stderr
	}
	return message
}

func (e commandError) Unwrap() error {
	return e.Err
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

func GetRepoName(repoPath string) (string, error) {
	return getRepoName("jj", repoPath)
}

func getRepoName(binary string, repoPath string) (string, error) {
	remoteURL, err := getFirstRemoteURL(binary, repoPath)
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

func GetFirstRemoteURL(repoPath string) (string, error) {
	return getFirstRemoteURL("jj", repoPath)
}

func getFirstRemoteURL(binary string, repoPath string) (string, error) {
	output, err := runJJCommand(binary, repoPath, "git", "remote", "list")
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

func GetBranchDisplay(repoPath string) (string, error) {
	return getBranchDisplay("jj", repoPath)
}

func getBranchDisplay(binary string, repoPath string) (string, error) {
	output, err := runJJCommand(
		binary,
		repoPath,
		"log",
		"-r",
		"latest(ancestors(@) & bookmarks()) | @",
		"--no-graph",
		"-T",
		`bookmarks.join(",") ++ "|" ++ change_id.short() ++ "\n"`,
	)
	if err != nil {
		return "", err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return "-", nil
	}

	fallback := "-"
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			if fallback == "-" {
				fallback = line
			}
			continue
		}

		if bookmarkDisplay := cleanBookmarkDisplay(parts[0]); bookmarkDisplay != "" {
			return bookmarkDisplay, nil
		}

		if changeID := strings.TrimSpace(parts[1]); changeID != "" && fallback == "-" {
			fallback = changeID
		}
	}

	return fallback, nil
}

func cleanBookmarkDisplay(display string) string {
	bookmarks := []string{}

	for _, bookmark := range strings.Split(display, ",") {
		bookmark = cleanBookmarkName(bookmark)
		if strings.Contains(bookmark, "@") {
			continue
		}
		if bookmark != "" {
			bookmarks = append(bookmarks, bookmark)
		}
	}

	return strings.Join(bookmarks, ",")
}

func cleanBookmarkName(bookmark string) string {
	bookmark = strings.TrimSpace(bookmark)
	bookmark = strings.TrimRight(bookmark, "*?")
	return strings.TrimSpace(bookmark)
}

func GetUncommittedFiles(repoPath string) ([]string, error) {
	return getUncommittedFiles("jj", repoPath)
}

func getUncommittedFiles(binary string, repoPath string) ([]string, error) {
	output, err := runJJCommand(binary, repoPath, "diff", "--summary")
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

func GetOutgoingCommits(repoPath string) ([]string, error) {
	return getOutgoingCommits("jj", repoPath)
}

func getOutgoingCommits(binary string, repoPath string) ([]string, error) {
	trackedBookmarks, err := getTrackedBookmarks(binary, repoPath)
	if err != nil {
		return nil, err
	}

	if len(trackedBookmarks) == 0 {
		return []string{}, nil
	}

	revset := buildTrackedOutgoingRevset(trackedBookmarks)
	output, err := runJJCommand(
		binary,
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

	return parseCommitSummaries(output), nil
}

func GetIncomingCommits(repoPath string) ([]string, error) {
	return getIncomingCommits("jj", repoPath)
}

func getIncomingCommits(binary string, repoPath string) ([]string, error) {
	trackedBookmarks, err := getTrackedBookmarks(binary, repoPath)
	if err != nil {
		return nil, err
	}

	if len(trackedBookmarks) == 0 {
		return []string{}, nil
	}

	revset := buildTrackedIncomingRevset(trackedBookmarks)
	output, err := runJJCommand(
		binary,
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

	return parseCommitSummaries(output), nil
}

func getBookmarkRemoteStatuses(
	binary string,
	repoPath string,
	bookmarkNames []string,
) ([]bookmarkRemoteStatus, error) {
	if len(bookmarkNames) == 0 {
		return []bookmarkRemoteStatus{}, nil
	}

	trackedBookmarks, err := getTrackedBookmarks(binary, repoPath)
	if err != nil {
		return nil, err
	}

	statuses := []bookmarkRemoteStatus{}
	seenBookmarks := map[string]struct{}{}
	for _, bookmarkName := range bookmarkNames {
		bookmarkName = cleanBookmarkName(bookmarkName)
		if bookmarkName == "" {
			continue
		}
		if _, ok := seenBookmarks[bookmarkName]; ok {
			continue
		}
		seenBookmarks[bookmarkName] = struct{}{}

		remotes := matchingRemotes(trackedBookmarks, bookmarkName)
		if len(remotes) == 0 {
			remotes, err = getUntrackedRemotesForBookmark(binary, repoPath, bookmarkName)
			if err != nil {
				return nil, err
			}
		}

		for _, remote := range remotes {
			tb := trackedBookmark{Name: bookmarkName, Remote: remote}

			outgoingCommits, err := getCommitsForRevset(
				binary,
				repoPath,
				buildTrackedOutgoingRevset([]trackedBookmark{tb}),
			)
			if err != nil {
				return nil, err
			}

			incomingCommits, err := getCommitsForRevset(
				binary,
				repoPath,
				buildTrackedIncomingRevset([]trackedBookmark{tb}),
			)
			if err != nil {
				return nil, err
			}

			statuses = append(statuses, bookmarkRemoteStatus{
				Remote:          remote,
				OutgoingCommits: outgoingCommits,
				IncomingCommits: incomingCommits,
			})
		}
	}

	return statuses, nil
}

func matchingRemotes(bookmarks []trackedBookmark, name string) []string {
	var remotes []string
	for _, b := range bookmarks {
		if b.Name == name {
			remotes = append(remotes, b.Remote)
		}
	}
	return remotes
}

func getUntrackedRemotesForBookmark(binary string, repoPath string, bookmarkName string) ([]string, error) {
	output, err := runJJCommand(
		binary,
		repoPath,
		"bookmark",
		"list",
		"--all",
		bookmarkName,
		"-T",
		`name ++ "|" ++ remote ++ "\n"`,
	)
	if err != nil {
		return nil, err
	}

	var remotes []string
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
		if name != bookmarkName || remote == "" || remote == "git" {
			continue
		}

		remotes = append(remotes, remote)
	}

	return remotes, nil
}

func getCommitsForRevset(binary string, repoPath string, revset string) ([]string, error) {
	if strings.TrimSpace(revset) == "" {
		return []string{}, nil
	}

	output, err := runJJCommand(
		binary,
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

	return parseCommitSummaries(output), nil
}

func getTrackedBookmarks(binary string, repoPath string) ([]trackedBookmark, error) {
	output, err := runJJCommand(
		binary,
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
		return "", commandError{
			Binary:   binary,
			RepoPath: repoPath,
			Args:     args,
			Stderr:   strings.TrimSpace(stderr.String()),
			Err:      err,
		}
	}

	return stdout.String(), nil
}

func parseCommitSummaries(output string) []string {
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

	return commits
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

func buildTrackedIncomingRevset(bookmarks []trackedBookmark) string {
	parts := make([]string, 0, len(bookmarks))

	for _, bookmark := range bookmarks {
		parts = append(parts, fmt.Sprintf(
			`(remote_bookmarks("%s", remote="git")..remote_bookmarks("%s", remote="%s"))`,
			escapeRevsetString(bookmark.Name),
			escapeRevsetString(bookmark.Name),
			escapeRevsetString(bookmark.Remote),
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
