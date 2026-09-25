package vcs

import "github.com/mabd-dev/reposcan/pkg/report"

// ScanOptions controls optional, potentially costly work done per repository.
type ScanOptions struct {
	// ComputeActivity fills RepoState.LastActivity for providers implementing
	// ActivityProvider. It should only be enabled when stale filtering is on.
	ComputeActivity bool
}

// CheckRepoState checks the repo at path with provider and, when requested by
// opts and supported by the provider, fills in its last activity.
func CheckRepoState(provider Provider, path string, opts ScanOptions) (report.RepoState, []string) {
	if opts.ComputeActivity {
		if activityProvider, ok := provider.(ActivityProvider); ok {
			return activityProvider.CheckRepoStateWithActivity(path)
		}
	}

	return provider.CheckRepoState(path)
}
