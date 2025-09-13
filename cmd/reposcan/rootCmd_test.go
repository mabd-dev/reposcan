package reposcan

import (
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// helper to build a RepoState with desired dirty state
func makeRepoState(dirty bool) report.RepoState {
	rs := report.RepoState{}
	if dirty {
		// Mark as dirty by adding an uncommitted file
		rs.UncommitedFiles = []string{"file.txt"}
	}
	return rs
}

func TestFilter_OnlyAll_AllowsAnyRepo(t *testing.T) {
	clean := makeRepoState(false)
	dirty := makeRepoState(true)

	if !filter(config.OnlyAll, clean) {
		t.Fatalf("OnlyAll should include clean repos")
	}
	if !filter(config.OnlyAll, dirty) {
		t.Fatalf("OnlyAll should include dirty repos")
	}
}

func TestFilter_OnlyDirty_AllowsOnlyDirtyRepos(t *testing.T) {
	clean := makeRepoState(false)
	dirty := makeRepoState(true)

	if filter(config.OnlyDirty, clean) {
		t.Fatalf("OnlyDirty should exclude clean repos")
	}
	if !filter(config.OnlyDirty, dirty) {
		t.Fatalf("OnlyDirty should include dirty repos")
	}
}
