package git

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mabd-dev/reposcan/internal/vcs"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// CheckRepoStateWithActivity implements vcs.ActivityProvider. It checks the
// repo like CheckRepoState and sets LastActivity to the newest of:
//   - the mtime of any uncommitted file
//   - the committer date of HEAD
//   - the committer date of the newest stash
//
// Signals that are unavailable (deleted files, a repo without commits, no
// stashes) are skipped. If `git status` failed, recent edits cannot be seen,
// so the activity is left unknown (zero) rather than dated from HEAD alone,
// which could wrongly mark a recently edited repo stale.
func (p *Provider) CheckRepoStateWithActivity(path string) (report.RepoState, []string) {
	state, warnings, changedPaths, statusErr := checkRepoState(path)
	state.VCSType = string(vcs.TypeGit)

	if statusErr != nil {
		warnings = append(warnings, "Last activity unknown because uncommitted files could not be read, path="+path)
		return state, warnings
	}

	lastActivity, activityWarnings := LastActivity(path, changedPaths)
	state.LastActivity = lastActivity
	return state, append(warnings, activityWarnings...)
}

// LastActivity returns the most recent local activity in the Git repository at
// path. changedPaths are the repo-relative paths of uncommitted files, as
// returned by getUncommittedChanges.
func LastActivity(path string, changedPaths []string) (latest time.Time, warnings []string) {
	latest = newestFileModTime(path, changedPaths)

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

// newestFileModTime returns the newest mtime among the changed paths. Paths
// that no longer exist on disk (e.g. deletions) are skipped.
func newestFileModTime(repoPath string, changedPaths []string) time.Time {
	var latest time.Time
	for _, rel := range changedPaths {
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
