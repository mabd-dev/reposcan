package repostableheader

import (
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestHeaderSetReportCountsRepositoriesAndDirtyStates(t *testing.T) {
	reportWithStash := report.ScanReport{RepoStates: []report.RepoState{
		{UncommitedFiles: []string{"changed.txt"}},
		{Stashes: []string{"stash@{0}"}},
		{},
	}}

	for _, tc := range []struct {
		name              string
		countStashAsDirty bool
		wantDirty         int
	}{
		{name: "ignore stashes", countStashAsDirty: false, wantDirty: 1},
		{name: "count stashes", countStashAsDirty: true, wantDirty: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			header := Header{}
			header.SetReport(reportWithStash, tc.countStashAsDirty)

			if header.repoStatesCount != 3 {
				t.Fatalf("repoStatesCount = %d, want 3", header.repoStatesCount)
			}
			if header.dirtyRepos != tc.wantDirty {
				t.Fatalf("dirtyRepos = %d, want %d", header.dirtyRepos, tc.wantDirty)
			}
		})
	}
}

func TestHeaderViewIncludesApplicationNameAndRepositoryCount(t *testing.T) {
	header := Header{}
	header.SetReport(report.ScanReport{RepoStates: make([]report.RepoState, 2)}, false)

	view := header.View()
	for _, want := range []string{"reposcan", "• 2 repos"} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() = %q, want it to contain %q", view, want)
		}
	}
}
