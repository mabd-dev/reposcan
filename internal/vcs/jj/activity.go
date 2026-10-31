package jj

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mabd-dev/reposcan/pkg/report"
)

// CheckRepoStateWithActivity implements vcs.ActivityProvider. It checks the
// repo like CheckRepoState and sets LastActivity to the newest of:
//   - the mtime of any file changed in the working-copy commit (@)
//   - the committer date of @'s newest parent (@-), git's HEAD equivalent
//
// jj has no stash. @'s own committer date and the operation log are not used:
// every jj command, including the ones run by CheckRepoState, snapshots the
// working copy, which re-dates @ and records an operation at scan time.
//
// If the changed files or the parent date could not be read, the activity is
// left unknown (zero) rather than dated from partial signals, which could
// wrongly mark a recently used repo stale.
func (p *Provider) CheckRepoStateWithActivity(path string) (report.RepoState, []string) {
	state, warnings, diffErr := p.checkRepoState(path)
	if diffErr != nil {
		warnings = append(warnings, "Last activity unknown because uncommitted files could not be read, path="+path)
		return state, warnings
	}

	parentDate, err := getParentCommitDate(p.binary, path)
	if err != nil {
		warnings = append(warnings, jjWarning("get parent commit date", path, err))
		return state, warnings
	}

	state.LastActivity = newestFileModTime(path, state.UncommitedFiles)
	if parentDate.After(state.LastActivity) {
		state.LastActivity = parentDate
	}
	return state, warnings
}

// getParentCommitDate returns the committer date of @'s newest parent, or the
// zero time when @ sits on the root commit (a repo without commits).
func getParentCommitDate(binary string, repoPath string) (time.Time, error) {
	out, err := runJJCommand(
		binary,
		repoPath,
		"log",
		"-r",
		"latest(@- ~ root())",
		"--no-graph",
		"-T",
		`committer.timestamp().format("%Y-%m-%dT%H:%M:%S%:z")`,
	)
	if err != nil {
		return time.Time{}, err
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, out)
}

// newestFileModTime returns the newest mtime among `jj diff --summary` entries.
// Entries that no longer exist on disk (e.g. deletions) are skipped.
func newestFileModTime(repoPath string, summaryLines []string) time.Time {
	var latest time.Time
	for _, line := range summaryLines {
		rel, ok := summaryPath(line)
		if !ok {
			continue
		}
		info, err := os.Lstat(filepath.Join(repoPath, rel))
		if err != nil {
			continue
		}
		if mt := info.ModTime(); mt.After(latest) {
			latest = mt
		}
	}
	return latest
}

var renamePattern = regexp.MustCompile(`\{[^{}]* => ([^{}]*)\}`)

// summaryPath extracts the current path from a `jj diff --summary` line: "M
// path", or "R dir/{old => new}/file" for renames and copies.
//
// ponytail: a rename of a file whose name contains braces is not parsed; it is
// skipped like a deletion and the repo is dated from its other signals.
func summaryPath(line string) (string, bool) {
	if len(line) < 3 {
		return "", false
	}
	p := line[2:]
	if line[0] == 'R' || line[0] == 'C' {
		p = renamePattern.ReplaceAllString(p, "${1}")
	}
	return p, p != ""
}
