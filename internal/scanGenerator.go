package internal

import (
	"time"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/internal/ds/mslice"
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
	// for _, path := range gitReposPaths {
	// 	logger.Debug("Found repo at", logger.StringAttr("path", path))
	// }

	reportWarnings = append(reportWarnings, warnings...)

	repoStates := make([]report.RepoState, 0, len(gitReposPaths))

	allRepoStates, warnings := gitx.GetGitRepoStatesConcurrent(gitReposPaths, configs.MaxWorkers)
	reportWarnings = append(reportWarnings, warnings...)

	// filter repo states based on config OnlyFilter
	for _, repoState := range allRepoStates {
		newWorktrees := filter(configs.Only, repoState.Worktrees)
		if len(newWorktrees) > 0 {
			repoState.Worktrees = newWorktrees
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
func filter(f config.OnlyFilter, worktrees []report.Worktree) []report.Worktree {
	switch f {
	case config.OnlyAll:
		return worktrees
	case config.OnlyDirty:
		return mslice.Filter(worktrees, report.IsDirty)
	case config.OnlyUncommitted:
		return mslice.Filter(worktrees, func(w report.Worktree) bool {
			return len(w.UncommitedFiles) > 0
		})
	case config.OnlyUnpushed:
		return mslice.Filter(worktrees, report.HaveUnpushedCommits)
	case config.OnlyUnpulled:
		return mslice.Filter(worktrees, report.HaveUnpulledCommits)
	}

	return []report.Worktree{}
}
