package internal

import (
	"time"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/gitx"
	"github.com/mabd-dev/reposcan/internal/scan"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func GenerateScanReport(
	configs config.Config,
) report.ScanReport {
	reportWarnings := []string{}

	// Find git repos at defined configs.Roots
	gitReposPaths, warnings := scan.FindGitRepos(configs.Roots, configs.DirIgnore)

	reportWarnings = append(reportWarnings, warnings...)

	repoStates := make([]report.RepoState, 0, len(gitReposPaths))

	allRepoStates, warnings := gitx.GetGitRepoStatesConcurrent(gitReposPaths, configs.MaxWorkers)
	reportWarnings = append(reportWarnings, warnings...)

	// filter repo states based on config OnlyFilter
	for _, repoState := range allRepoStates {
		if filter(configs.Only, repoState) {
			repoStates = append(repoStates, repoState)
		}
	}

	return report.ScanReport{
		Version:     configs.Version,
		GeneratedAt: time.Now(),
		RepoStates:  repoStates,
		Warnings:    reportWarnings,
	}
}

// Filter repoState based on config only filter
// Returns true if repoState should be in output, false otherwise
func filter(f config.OnlyFilter, repoState report.RepoState) bool {
	switch f {
	case config.OnlyAll:
		return true
	case config.OnlyDirty:
		if repoState.IsDirty() {
			return true
		}
	case config.OnlyUncommitted:
		if len(repoState.UncommitedFiles) > 0 {
			return true
		}
	case config.OnlyUnpushed:
		if repoState.Ahead > 0 {
			return true
		}
	case config.OnlyUnpulled:
		if repoState.Behind > 0 {
			return true
		}
	}

	return false
}
