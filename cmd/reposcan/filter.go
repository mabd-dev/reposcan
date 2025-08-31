package main

import (
	"github.com/MABD-dev/RepoScan/internal/config"
	"github.com/MABD-dev/RepoScan/pkg/report"
)

// Filter repoState based on config only filter
// Returns true if repoState should be in output, false otherwise
func Filter(f config.OnlyFilter, repoState report.RepoState) bool {
	switch f {
	case config.OnlyAll:
		return true
	case config.OnlyDirty:
		if repoState.IsDirty() {
			return true
		}
	}

	return false
}
