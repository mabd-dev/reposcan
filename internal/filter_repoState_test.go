package internal

import (
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// helper to build a RepoState with desired dirty state
func makeRepoState(uncommited, unpushed, unpulled bool) report.RepoState {
	rs := report.RepoState{}
	if uncommited {
		// Mark as dirty by adding an uncommitted file
		rs.UncommitedFiles = []string{"file.txt"}
	}

	remoteStatus := report.RemoteStatus{
		Remote: "something",
		Ahead:  -1,
		Behind: -1,
	}

	if unpushed {
		remoteStatus.Ahead = 1
	}
	if unpulled {
		remoteStatus.Behind = 1
	}

	rs.RemoteStatus = append(rs.RemoteStatus, remoteStatus)

	return rs
}

// makeStashOnlyRepoState builds a clean repo whose only local state is a stash.
func makeStashOnlyRepoState() report.RepoState {
	rs := makeRepoState(false, false, false)
	rs.Stashes = []string{"stash@{0}: WIP on main: abc123 msg"}
	return rs
}

func TestFilter_OnlyStash_AllowsOnlyReposWithStashes(t *testing.T) {
	clean := makeRepoState(false, false, false)
	stash := makeStashOnlyRepoState()

	// flag must not affect OnlyStash in either direction
	for _, flag := range []bool{false, true} {
		if filter(config.OnlyStash, clean, flag) {
			t.Fatalf("OnlyStash should exclude repos without stashes (flag=%v)", flag)
		}
		if !filter(config.OnlyStash, stash, flag) {
			t.Fatalf("OnlyStash should include repos with stashes (flag=%v)", flag)
		}
	}
}

func TestFilter_OnlyDirty_StashRespectsCountStashAsDirty(t *testing.T) {
	stash := makeStashOnlyRepoState()

	if filter(config.OnlyDirty, stash, false) {
		t.Fatalf("OnlyDirty should exclude stash-only repos when countStashAsDirty=false")
	}
	if !filter(config.OnlyDirty, stash, true) {
		t.Fatalf("OnlyDirty should include stash-only repos when countStashAsDirty=true")
	}
}

func TestFilter_StashOnlyRepo_ExcludedFromOtherFilters(t *testing.T) {
	stash := makeStashOnlyRepoState()

	for _, f := range []config.OnlyFilter{config.OnlyUncommitted, config.OnlyUnpushed, config.OnlyUnpulled} {
		if filter(f, stash, true) {
			t.Fatalf("filter %q should exclude stash-only repos", f)
		}
	}
}

func TestFilter_OnlyAll_AllowsAnyRepo(t *testing.T) {
	clean := makeRepoState(false, false, false)
	dirty := makeRepoState(true, false, false)

	if !filter(config.OnlyAll, clean, false) {
		t.Fatalf("OnlyAll should include clean repos")
	}
	if !filter(config.OnlyAll, dirty, false) {
		t.Fatalf("OnlyAll should include dirty repos")
	}
}

func TestFilter_OnlyUncommitted_AllowsOnlyReposWithUncommitedChanges(t *testing.T) {
	clean := makeRepoState(false, false, false)

	dirty1 := makeRepoState(true, false, false)
	dirty2 := makeRepoState(false, true, false)
	dirty3 := makeRepoState(false, false, true)

	if filter(config.OnlyUncommitted, clean, false) {
		t.Fatalf("OnlyUncommitted should exclude clean repos")
	}

	if !filter(config.OnlyUncommitted, dirty1, false) {
		t.Fatalf("OnlyUncommitted should include dirty repos, 1")
	}

	if filter(config.OnlyUncommitted, dirty2, false) {
		t.Fatalf("OnlyUncommitted should include dirty repos, 2")
	}

	if filter(config.OnlyUncommitted, dirty3, false) {
		t.Fatalf("OnlyUncommitted should include dirty repos, 3")
	}
}

func TestFilter_OnlyUnpushed_AllowsOnlyReposWithUnpushedCommits(t *testing.T) {
	clean := makeRepoState(false, false, false)

	dirty1 := makeRepoState(true, false, false)
	dirty2 := makeRepoState(false, true, false)
	dirty3 := makeRepoState(false, false, true)

	if filter(config.OnlyUnpushed, clean, false) {
		t.Fatalf("OnlyUnpushed should exclude clean repos")
	}

	if filter(config.OnlyUnpushed, dirty1, false) {
		t.Fatalf("OnlyUnpushed should include dirty repos, 1")
	}

	if !filter(config.OnlyUnpushed, dirty2, false) {
		t.Fatalf("OnlyUnpushed should include dirty repos, 2")
	}

	if filter(config.OnlyUnpushed, dirty3, false) {
		t.Fatalf("OnlyUnpushed should include dirty repos, 3")
	}
}

func TestFilter_OnlyUnpulled_AllowsOnlyReposWithUnpulledCommits(t *testing.T) {
	clean := makeRepoState(false, false, false)

	dirty1 := makeRepoState(true, false, false)
	dirty2 := makeRepoState(false, true, false)
	dirty3 := makeRepoState(false, false, true)

	if filter(config.OnlyUnpulled, clean, false) {
		t.Fatalf("OnlyUnpulled should exclude clean repos")
	}

	if filter(config.OnlyUnpulled, dirty1, false) {
		t.Fatalf("OnlyUnpulled should include dirty repos, 1")
	}

	if filter(config.OnlyUnpulled, dirty2, false) {
		t.Fatalf("OnlyUnpulled should include dirty repos, 2")
	}

	if !filter(config.OnlyUnpulled, dirty3, false) {
		t.Fatalf("OnlyUnpulled should include dirty repos, 3")
	}
}

func TestFilter_OnlyDirty_AllowsOnlyDirtyRepos(t *testing.T) {
	clean := makeRepoState(false, false, false)
	dirty1 := makeRepoState(true, false, false)
	dirty2 := makeRepoState(false, true, false)
	dirty3 := makeRepoState(false, false, true)

	if filter(config.OnlyDirty, clean, false) {
		t.Fatalf("OnlyDirty should exclude clean repos")
	}

	if !filter(config.OnlyDirty, dirty1, false) {
		t.Fatalf("OnlyDirty should include dirty repos")
	}

	if !filter(config.OnlyDirty, dirty2, false) {
		t.Fatalf("OnlyDirty should include dirty repos")
	}

	if !filter(config.OnlyDirty, dirty3, false) {
		t.Fatalf("OnlyDirty should include dirty repos")
	}
}
