package internal

import (
	"time"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/scan"
	"github.com/mabd-dev/reposcan/internal/vcs"
	vcsgit "github.com/mabd-dev/reposcan/internal/vcs/git"
	vcsjj "github.com/mabd-dev/reposcan/internal/vcs/jj"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func GenerateScanReport(
	configs config.Config,
) report.ScanReport {
	reportWarnings := []string{}

	registry := NewVCSRegistry()

	repoPaths, warnings := scan.FindRepos(configs.Roots, configs.DirIgnore)

	reportWarnings = append(reportWarnings, warnings...)

	repoStates := make([]report.RepoState, 0, len(repoPaths))

	allRepoStates, warnings := vcs.GetRepoStatesConcurrent(
		repoPaths,
		registry,
		configs.MaxWorkers,
		ScanOptions(configs),
	)
	reportWarnings = append(reportWarnings, warnings...)

	now := time.Now()

	// filter repo states based on config OnlyFilter and StaleDays
	for _, repoState := range allRepoStates {
		if !filter(configs.Only, repoState, configs.CountStashAsDirty) {
			continue
		}
		if !staleFilter(configs.Only, configs.StaleDays, repoState, now) {
			continue
		}
		repoStates = append(repoStates, repoState)
	}

	return report.ScanReport{
		Version:           configs.Version,
		GeneratedAt:       now,
		RepoStates:        repoStates,
		TotalScannedRepos: len(allRepoStates),
		Warnings:          reportWarnings,
	}
}

// ScanOptions returns the per-repo scan options implied by configs.
func ScanOptions(configs config.Config) vcs.ScanOptions {
	return vcs.ScanOptions{
		ComputeActivity: staleFilterEnabled(configs.Only, configs.StaleDays),
	}
}

// staleFilterEnabled reports whether stale filtering applies. It is off when
// staleDays <= 0 and for OnlyUnpulled, whose state reflects remote activity
// rather than local work.
func staleFilterEnabled(f config.OnlyFilter, staleDays int) bool {
	return staleDays > 0 && f != config.OnlyUnpulled
}

// staleFilter returns true if repoState should be in output based on its last
// activity. Repos whose activity is unknown (zero LastActivity) are kept so
// that the stale filter never hides a repo it cannot date.
func staleFilter(f config.OnlyFilter, staleDays int, repoState report.RepoState, now time.Time) bool {
	if !staleFilterEnabled(f, staleDays) {
		return true
	}
	if repoState.LastActivity.IsZero() {
		return true
	}

	threshold := time.Duration(staleDays) * 24 * time.Hour
	return now.Sub(repoState.LastActivity) >= threshold
}

func NewVCSRegistry() *vcs.Registry {
	return vcs.NewRegistry(
		vcsgit.New(),
		vcsjj.New(),
	)
}

// Filter repoState based on config only filter
// Returns true if repoState should be in output, false otherwise.
// countStashAsDirty only affects the OnlyDirty case; OnlyStash is independent.
func filter(f config.OnlyFilter, repoState report.RepoState, countStashAsDirty bool) bool {
	switch f {
	case config.OnlyAll:
		return true
	case config.OnlyDirty:
		if repoState.IsDirty(countStashAsDirty) {
			return true
		}
	case config.OnlyUncommitted:
		if len(repoState.UncommitedFiles) > 0 {
			return true
		}
	case config.OnlyUnpushed:
		return repoState.HaveUnpushedCommits()
	case config.OnlyUnpulled:
		return repoState.HaveUnpulledCommits()
	case config.OnlyStash:
		return repoState.HaveStashes()
	}

	return false
}
