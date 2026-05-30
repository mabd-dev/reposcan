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

	allRepoStates, warnings := vcs.GetRepoStatesConcurrent(repoPaths, registry, configs.MaxWorkers)
	reportWarnings = append(reportWarnings, warnings...)

	// filter repo states based on config OnlyFilter
	for _, repoState := range allRepoStates {
		if filter(configs.Only, repoState, configs.CountStashAsDirty) {
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
