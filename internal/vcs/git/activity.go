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
func (p *Provider) LastActivity(path string, state report.RepoState) (time.Time, []string) {
	return LastActivity(path, state.UncommitedFiles)
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
		if i := strings.LastIndex(p, " -> "); i >= 0 {
			p = p[i+len(" -> "):]
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
