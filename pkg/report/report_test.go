package report

import "testing"

func TestHaveStashesAndStashCount(t *testing.T) {
	none := RepoState{}
	if none.HaveStashes() {
		t.Fatalf("HaveStashes should be false for empty Stashes")
	}
	if none.StashCount() != 0 {
		t.Fatalf("StashCount should be 0 for empty Stashes")
	}

	with := RepoState{Stashes: []string{"stash@{0}: WIP", "stash@{1}: WIP"}}
	if !with.HaveStashes() {
		t.Fatalf("HaveStashes should be true when Stashes present")
	}
	if with.StashCount() != 2 {
		t.Fatalf("StashCount should be 2, got %d", with.StashCount())
	}
}

func TestIsDirty_StashOnlyRespectsFlag(t *testing.T) {
	rs := RepoState{Stashes: []string{"stash@{0}: WIP"}}
	if rs.IsDirty(false) {
		t.Fatalf("IsDirty(false) must not consider stashes; stash-only repo should be clean")
	}
	if !rs.IsDirty(true) {
		t.Fatalf("IsDirty(true) must consider stash-only repo as dirty")
	}
}

func TestDirtyReposCount_StashRespectsFlag(t *testing.T) {
	stashOnly := RepoState{Stashes: []string{"stash@{0}: WIP"}}
	uncommitted := RepoState{UncommitedFiles: []string{"file.txt"}}
	clean := RepoState{}

	sc := ScanReport{RepoStates: []RepoState{stashOnly, uncommitted, clean}}

	if got := sc.DirtyReposCount(false); got != 1 {
		t.Fatalf("DirtyReposCount(false) should count only the uncommitted repo, got %d", got)
	}
	if got := sc.DirtyReposCount(true); got != 2 {
		t.Fatalf("DirtyReposCount(true) should count uncommitted + stash-only, got %d", got)
	}
}
