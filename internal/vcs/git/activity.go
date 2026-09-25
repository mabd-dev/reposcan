package git

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mabd-dev/reposcan/pkg/report"
)

// LastActivity implements vcs.ActivityProvider. It returns the newest of:
//   - the mtime of any uncommitted file listed in state.UncommitedFiles
//   - the committer date of HEAD
//   - the committer date of the newest stash
//
// Signals that are unavailable (deleted files, a repo without commits, no
// stashes) are skipped. The zero time is returned when no signal is available.
//
// CheckRepoState reports an empty file list both for a clean repo and when
// `git status` failed. Dating a repo from HEAD alone after a failed status
// could hide recent edits and mark it stale, so an empty list is re-checked
// and a failing status leaves the activity unknown (zero).
func (p *Provider) LastActivity(path string, state report.RepoState) (time.Time, []string) {
	files := state.UncommitedFiles
	if len(files) == 0 {
		var err error
		files, err = GetUncommitedFiles(path)
		if err != nil {
			return time.Time{}, []string{"Failed to get uncommitted files, last activity unknown, path=" + path}
		}
	}
	return LastActivity(path, files)
}

// LastActivity returns the most recent local activity in the Git repository at
// path. uncommittedFiles are `git status --porcelain=v1` lines, as returned by
// GetUncommitedFiles.
func LastActivity(path string, uncommittedFiles []string) (latest time.Time, warnings []string) {
	latest = newestFileModTime(path, uncommittedFiles)

	headDate, err := getHeadCommitDate(path)
	if err != nil {
		// An unborn branch (no commits yet) has no HEAD date; that is not a failure.
		if _, verifyErr := RunGitCommand(path, "rev-parse", "-q", "--verify", "HEAD"); verifyErr == nil {
			warnings = append(warnings, "Failed to get HEAD commit date, path="+path)
		}
	} else if headDate.After(latest) {
		latest = headDate
	}

	stashDate, err := getNewestStashDate(path)
	if err != nil {
		warnings = append(warnings, "Failed to get newest stash date, path="+path)
	} else if stashDate.After(latest) {
		latest = stashDate
	}

	return latest, warnings
}

func getHeadCommitDate(path string) (time.Time, error) {
	out, err := RunGitCommand(path, "log", "-1", "--format=%cI")
	if err != nil {
		return time.Time{}, err
	}
	return parseGitDate(out)
}

// getNewestStashDate returns the zero time without error when there are no stashes.
func getNewestStashDate(path string) (time.Time, error) {
	out, err := RunGitCommand(path, "stash", "list", "-1", "--format=%cI")
	if err != nil {
		return time.Time{}, err
	}
	return parseGitDate(out)
}

// parseGitDate parses a strict ISO 8601 date (%cI). Empty output yields the zero time.
func parseGitDate(out string) (time.Time, error) {
	out = strings.TrimSpace(out)
	if out == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, out)
}

// newestFileModTime returns the newest mtime among porcelain status entries.
// Entries that no longer exist on disk (e.g. deletions) are skipped.
func newestFileModTime(repoPath string, porcelainLines []string) time.Time {
	var latest time.Time
	for _, line := range porcelainLines {
		rel, ok := porcelainPath(line)
		if !ok {
			continue
		}
		info, err := os.Lstat(filepath.Join(repoPath, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		if mt := info.ModTime(); mt.After(latest) {
			latest = mt
		}
	}
	return latest
}

// porcelainPath extracts the current path from a `git status --porcelain=v1`
// line ("XY path" or "XY orig -> path" for renames/copies). C-quoted paths are
// unquoted.
func porcelainPath(line string) (string, bool) {
	if len(line) < 4 {
		return "", false
	}
	p := line[3:]

	if line[0] == 'R' || line[0] == 'C' || line[1] == 'R' || line[1] == 'C' {
		// Git C-quotes any name containing " -> ", so the separator is the
		// first " -> " after the (possibly quoted) original path.
		start := 0
		if strings.HasPrefix(p, `"`) {
			end := quotedPathEnd(p)
			if end < 0 {
				return "", false
			}
			start = end
		}
		if i := strings.Index(p[start:], " -> "); i >= 0 {
			p = p[start+i+len(" -> "):]
		}
	}

	if strings.HasPrefix(p, `"`) {
		unquoted, err := strconv.Unquote(p)
		if err != nil {
			return "", false
		}
		p = unquoted
	}

	p = strings.TrimSuffix(p, "/")
	if p == "" {
		return "", false
	}
	return p, true
}

// quotedPathEnd returns the index just past the closing quote of the C-quoted
// string at the start of s, or -1 if it is unterminated.
func quotedPathEnd(s string) int {
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '"':
			return i + 1
		}
	}
	return -1
}
