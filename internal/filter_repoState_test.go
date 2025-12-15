package internal

import (
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/pkg/report"
)

// helper to build a RepoState with desired dirty state
func makeWorktree(uncommited, unpushed, unpulled bool) []report.Worktree {
	worktree := report.Worktree{}

	if uncommited {
		// Mark as dirty by adding an uncommitted file
		worktree.UncommitedFiles = []string{"file.txt"}
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

	worktree.RemoteStatus = append(worktree.RemoteStatus, remoteStatus)
	return []report.Worktree{worktree}
}

func TestFilter_OnlyAll_AllowsAnyRepo(t *testing.T) {
	clean := makeWorktree(false, false, false)
	dirty := makeWorktree(true, false, false)

	if len(filter(config.OnlyAll, clean)) == 0 {
		t.Fatalf("OnlyAll should include clean repos")
	}
	if len(filter(config.OnlyAll, dirty)) == 0 {
		t.Fatalf("OnlyAll should include dirty repos")
	}
}

func TestFilter_OnlyUncommitted_AllowsOnlyReposWithUncommitedChanges(t *testing.T) {
	clean := makeWorktree(false, false, false)

	dirty1 := makeWorktree(true, false, false)
	dirty2 := makeWorktree(false, true, false)
	dirty3 := makeWorktree(false, false, true)

	if len(filter(config.OnlyUncommitted, clean)) != 0 {
		t.Fatalf("OnlyUncommitted should exclude clean repos")
	}

	if len(filter(config.OnlyUncommitted, dirty1)) == 0 {
		t.Fatalf("OnlyUncommitted should include dirty repos, 1")
	}

	if len(filter(config.OnlyUncommitted, dirty2)) != 0 {
		t.Fatalf("OnlyUncommitted should include dirty repos, 2")
	}

	if len(filter(config.OnlyUncommitted, dirty3)) != 0 {
		t.Fatalf("OnlyUncommitted should include dirty repos, 3")
	}
}

func TestFilter_OnlyUnpushed_AllowsOnlyReposWithUnpushedCommits(t *testing.T) {
	clean := makeWorktree(false, false, false)

	dirty1 := makeWorktree(true, false, false)
	dirty2 := makeWorktree(false, true, false)
	dirty3 := makeWorktree(false, false, true)

	if len(filter(config.OnlyUnpushed, clean)) != 0 {
		t.Fatalf("OnlyUnpushed should exclude clean repos")
	}

	if len(filter(config.OnlyUnpushed, dirty1)) != 0 {
		t.Fatalf("OnlyUnpushed should include dirty repos, 1")
	}

	if len(filter(config.OnlyUnpushed, dirty2)) == 0 {
		t.Fatalf("OnlyUnpushed should include dirty repos, 2")
	}

	if len(filter(config.OnlyUnpushed, dirty3)) != 0 {
		t.Fatalf("OnlyUnpushed should include dirty repos, 3")
	}
}

func TestFilter_OnlyUnpulled_AllowsOnlyReposWithUnpulledCommits(t *testing.T) {
	clean := makeWorktree(false, false, false)

	dirty1 := makeWorktree(true, false, false)
	dirty2 := makeWorktree(false, true, false)
	dirty3 := makeWorktree(false, false, true)

	if len(filter(config.OnlyUnpulled, clean)) != 0 {
		t.Fatalf("OnlyUnpulled should exclude clean repos")
	}

	if len(filter(config.OnlyUnpulled, dirty1)) != 0 {
		t.Fatalf("OnlyUnpulled should include dirty repos, 1")
	}

	if len(filter(config.OnlyUnpulled, dirty2)) != 0 {
		t.Fatalf("OnlyUnpulled should include dirty repos, 2")
	}

	if len(filter(config.OnlyUnpulled, dirty3)) == 0 {
		t.Fatalf("OnlyUnpulled should include dirty repos, 3")
	}
}

func TestFilter_OnlyDirty_AllowsOnlyDirtyRepos(t *testing.T) {
	clean := makeWorktree(false, false, false)
	dirty1 := makeWorktree(true, false, false)
	dirty2 := makeWorktree(false, true, false)
	dirty3 := makeWorktree(false, false, true)

	if len(filter(config.OnlyDirty, clean)) != 0 {
		t.Fatalf("OnlyDirty should exclude clean repos")
	}

	if len(filter(config.OnlyDirty, dirty1)) == 0 {
		t.Fatalf("OnlyDirty should include dirty repos")
	}

	if len(filter(config.OnlyDirty, dirty2)) == 0 {
		t.Fatalf("OnlyDirty should include dirty repos")
	}

	if len(filter(config.OnlyDirty, dirty3)) == 0 {
		t.Fatalf("OnlyDirty should include dirty repos")
	}
}
