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
	// NoRemoteBookmark is set when the remote has none of the bookmarks, so
	// there is nothing to count against.
	NoRemoteBookmark bool
}

type remote struct {
	Name string
	URL  string
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

func GetRepoName(repoPath string) (string, error) {
	return getRepoName("jj", repoPath)
}

func getRepoName(binary string, repoPath string) (string, error) {
	remotes, err := getRemotes(binary, repoPath)
	if err != nil {
		return "", err
	}

	return repoNameFromRemotes(remotes, repoPath), nil
}

func repoNameFromRemotes(remotes []remote, repoPath string) string {
	if len(remotes) > 0 {
		if repoName, ok := parseRepoName(remotes[0].URL); ok {
			return repoName
		}
	}

	return filepath.Base(repoPath)
}

func GetFirstRemoteURL(repoPath string) (string, error) {
	return getFirstRemoteURL("jj", repoPath)
}

func getFirstRemoteURL(binary string, repoPath string) (string, error) {
	remotes, err := getRemotes(binary, repoPath)
	if err != nil || len(remotes) == 0 {
		return "", err
	}

	return remotes[0].URL, nil
}

func getRemotes(binary string, repoPath string) ([]remote, error) {
	output, err := runJJCommand(binary, repoPath, "git", "remote", "list")
	if err != nil {
		return nil, err
	}

	remotes := []remote{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Done importing changes") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			remotes = append(remotes, remote{Name: parts[0], URL: parts[1]})
		}
	}

	return remotes, nil
}

func GetBranchDisplay(repoPath string) (string, error) {
	return getBranchDisplay("jj", repoPath)
}

func getBranchDisplay(binary string, repoPath string) (string, error) {
	display, _, err := getBranch(binary, repoPath, true)
	return display, err
}

// getBranch returns the branch display and whether it names bookmarks, as
// opposed to a change ID or "-" fallback. findMoved enables the moved-bookmark
// lookup; callers turn it off when the repo has no remotes, since only tracked
// bookmarks can move past @ and tracking needs a remote.
func getBranch(binary string, repoPath string, findMoved bool) (string, bool, error) {
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
		return "", false, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return "-", false, nil
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
			return bookmarkDisplay, true, nil
		}

		if changeID := strings.TrimSpace(parts[1]); changeID != "" && fallback == "-" {
			fallback = changeID
		}
	}

	if !findMoved {
		return fallback, false, nil
	}

	// The lookup is optional, so a failure still returns the change ID with
	// the error.
	movedDisplay, err := getMovedBookmarkDisplay(binary, repoPath)
	if err != nil {
		return fallback, false, err
	}
	if movedDisplay != "" {
		return movedDisplay, true, nil
	}

	return fallback, false, nil
}

// getMovedBookmarkDisplay finds tracked bookmarks that moved past @ when no
// bookmark sits below it, e.g. after jj git fetch fast-forwards main while the
// working copy stays on the old commit. Candidates are the tracked bookmarks at
// or after the newest commit @ shares with any tracked bookmark. Local-only
// bookmarks are left out of the search so they cannot pull the shared commit
// away from the tracked ones; they have no remote updates to report anyway.
//
// jj does not record which bookmark @ started from, so the result must be
// unambiguous: sharing only root() says nothing, and candidates on more than
// one commit (e.g. main moved and a feature bookmark branched off the shared
// commit) could be either. Both cases return "" so the caller keeps the change
// ID.
func getMovedBookmarkDisplay(binary string, repoPath string) (string, error) {
	trackedBookmarks, err := getTrackedBookmarks(binary, repoPath)
	if err != nil {
		return "", err
	}

	tracked := map[string]struct{}{}
	parts := []string{}
	for _, bookmark := range trackedBookmarks {
		if _, ok := tracked[bookmark.Name]; ok {
			continue
		}
		tracked[bookmark.Name] = struct{}{}
		parts = append(parts, fmt.Sprintf(`bookmarks(exact:"%s")`, escapeRevsetString(bookmark.Name)))
	}
	if len(parts) == 0 {
		return "", nil
	}

	trackedLocal := strings.Join(parts, " | ")
	output, err := runJJCommand(
		binary,
		repoPath,
		"log",
		"-r",
		fmt.Sprintf("(%s) & (heads(::@ & ::(%s)) ~ root())::", trackedLocal, trackedLocal),
		"--no-graph",
		"-T",
		`bookmarks.join(",") ++ "\n"`,
	)
	if err != nil {
		return "", err
	}

	lines := []string{}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) != 1 {
		return "", nil
	}

	// The commit can carry other bookmarks too, so filter names again.
	names := []string{}
	for _, name := range strings.Split(cleanBookmarkDisplay(lines[0]), ",") {
		if _, ok := tracked[name]; ok {
			names = append(names, name)
		}
	}

	return strings.Join(names, ","), nil
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

// getBookmarkRemoteStatuses compares each bookmark with the bookmark of the
// same name on every remote, like the git provider compares the current branch
// with <remote>/<branch>. Tracking does not matter. A remote that has none of
// the bookmarks, or every remote when there are no bookmarks, gets a
// NoRemoteBookmark entry.
func getBookmarkRemoteStatuses(
	binary string,
	repoPath string,
	bookmarkNames []string,
	remoteNames []string,
) ([]bookmarkRemoteStatus, error) {
	if len(remoteNames) == 0 {
		return []bookmarkRemoteStatus{}, nil
	}

	onRemote := map[trackedBookmark]struct{}{}
	if len(bookmarkNames) > 0 {
		remoteBookmarks, err := getRemoteBookmarks(binary, repoPath)
		if err != nil {
			return nil, err
		}
		for _, bookmark := range remoteBookmarks {
			onRemote[bookmark] = struct{}{}
		}
	}

	names := []string{}
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
		names = append(names, bookmarkName)
	}

	// Remotes stay in configured order, as in the git provider.
	statuses := []bookmarkRemoteStatus{}
	for _, remoteName := range remoteNames {
		found := false
		for _, name := range names {
			tb := trackedBookmark{Name: name, Remote: remoteName}
			if _, ok := onRemote[tb]; !ok {
				continue
			}
			found = true

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
				Remote:          remoteName,
				OutgoingCommits: outgoingCommits,
				IncomingCommits: incomingCommits,
			})
		}

		if !found {
			statuses = append(statuses, bookmarkRemoteStatus{Remote: remoteName, NoRemoteBookmark: true})
		}
	}

	return statuses, nil
}

// getRemoteBookmarks lists every remote bookmark, tracked or not, excluding
// jj's synthetic "git" remote. Newer jj keeps a remote bookmark that was
// deleted on the remote as an absent entry until it is untracked; it is left
// out so it reports like a missing branch instead of counting all history.
func getRemoteBookmarks(binary string, repoPath string) ([]trackedBookmark, error) {
	output, err := runJJCommand(
		binary,
		repoPath,
		"bookmark",
		"list",
		"--all",
		// No name argument: its pattern syntax differs across jj versions, so
		// callers match names exactly instead.
		"-T",
		`if(present, name ++ "|" ++ remote ++ "\n")`,
	)
	if err != nil {
		return nil, err
	}

	var bookmarks []trackedBookmark
	for _, line := range strings.Split(output, "\n") {
		name, remote, ok := strings.Cut(line, "|")
		if !ok {
			continue
		}

		name = strings.TrimSpace(name)
		remote = strings.TrimSpace(remote)
		if name == "" || remote == "" || remote == "git" {
			continue
		}

		bookmarks = append(bookmarks, trackedBookmark{Name: name, Remote: remote})
	}

	return bookmarks, nil
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
	// jj prints file paths relative to its working directory, even with -R, so
	// run it from the repo. -R . still pins the repo so jj never falls back to a
	// parent one, and unlike -R repoPath it is not re-resolved against Dir when
	// repoPath is relative.
	cmd := exec.Command(binary, append([]string{"-R", "."}, args...)...)
	cmd.Dir = repoPath

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

// The revset builders count commits the way reposcan's git provider does,
// with @- in place of HEAD: @ is the working copy, whose changes are reported
// as uncommitted files, so the last commit is @-. Counting from @- rather than
// the local bookmark matters because jj git fetch moves a tracked bookmark
// forward without moving @, and because jj commit does not move the bookmark.
// Incoming excludes ::@ instead of ::@-, which only differs when the remote
// bookmark already contains @, e.g. after jj git push -c @; @ is local work,
// not a commit to pull. exact: patterns avoid substring matches on older jj
// (e.g. "main" also matching "maintenance").
func buildTrackedOutgoingRevset(bookmarks []trackedBookmark) string {
	parts := make([]string, 0, len(bookmarks))

	for _, bookmark := range bookmarks {
		parts = append(parts, fmt.Sprintf(`(%s..@-)`, remoteBookmarkRevset(bookmark)))
	}

	return strings.Join(parts, " | ")
}

func buildTrackedIncomingRevset(bookmarks []trackedBookmark) string {
	parts := make([]string, 0, len(bookmarks))

	for _, bookmark := range bookmarks {
		parts = append(parts, fmt.Sprintf(`(@..%s)`, remoteBookmarkRevset(bookmark)))
	}

	return strings.Join(parts, " | ")
}

func remoteBookmarkRevset(bookmark trackedBookmark) string {
	return fmt.Sprintf(
		`remote_bookmarks(exact:"%s", remote=exact:"%s")`,
		escapeRevsetString(bookmark.Name),
		escapeRevsetString(bookmark.Remote),
	)
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
