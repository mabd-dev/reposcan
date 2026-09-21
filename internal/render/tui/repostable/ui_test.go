package repostable

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestCreateColumnsIncludesVCSColumn(t *testing.T) {
	columns := createColumns(100, Options{ShowVCS: true})

	if len(columns) != 5 {
		t.Fatalf("expected 5 columns, got %d: %v", len(columns), columns)
	}
	if columns[2].Title != "VCS" {
		t.Fatalf("expected third column to be VCS, got %q", columns[2].Title)
	}
}

func TestCreateColumnsOmitsVCSColumnWhenDisabled(t *testing.T) {
	columns := createColumns(100, Options{ShowVCS: false})

	if len(columns) != 4 {
		t.Fatalf("expected 4 columns, got %d: %v", len(columns), columns)
	}
	for _, c := range columns {
		if c.Title == "VCS" {
			t.Fatalf("did not expect a VCS column, got columns: %v", columns)
		}
	}
	if columns[3].Title != "State" || columns[3].Width != RemoteStateW+VCSW {
		t.Fatalf("expected State column to reclaim VCS width, got %+v", columns[3])
	}
}

func TestActiveColumnDefsAllCellsDefined(t *testing.T) {
	for _, options := range []Options{
		{ShowVCS: true},
		{ShowVCS: false},
	} {
		defs := activeColumnDefs(options)

		for _, def := range defs {
			if def.cell == nil {
				t.Fatalf("expected %q column to define a cell renderer", def.title)
			}
		}
	}
}

func TestCreateRowsIncludesVCSValue(t *testing.T) {
	rows := createRows([]report.RepoState{
		{
			Repo:    "jj-repo",
			VCSType: "jj",
			Branch:  "@",
			RemoteStatus: []report.RemoteStatus{
				{Remote: "origin", Ahead: 1, Behind: 2},
			},
		},
	}, theme.Theme{}, Options{ShowVCS: true})

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) != 5 {
		t.Fatalf("expected 5 cells, got %d: %v", len(rows[0]), rows[0])
	}
	if rows[0][2] != "jj" {
		t.Fatalf("expected VCS cell to be jj, got %q", rows[0][2])
	}
	if rows[0][4] != "⏳0 ↑1 ↓2" {
		t.Fatalf("unexpected state cell: %q", rows[0][4])
	}
}

func TestCreateRowsOmitsVCSValueWhenDisabled(t *testing.T) {
	rows := createRows([]report.RepoState{
		{
			Repo:    "jj-repo",
			VCSType: "jj",
			Branch:  "@",
			RemoteStatus: []report.RemoteStatus{
				{Remote: "origin", Ahead: 1, Behind: 2},
			},
		},
	}, theme.Theme{}, Options{ShowVCS: false})

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) != 4 {
		t.Fatalf("expected 4 cells, got %d: %v", len(rows[0]), rows[0])
	}
	if rows[0][3] != "⏳0 ↑1 ↓2" {
		t.Fatalf("unexpected state cell: %q", rows[0][3])
	}
}

func TestRedistributeHiddenWidthFallsBackToLastColumn(t *testing.T) {
	active := []columnDef{
		{title: "Repo", widthPercent: RepoW},
		{title: "Branch", widthPercent: BranchW},
	}

	redistributeHiddenWidth(active, VCSW)

	if active[0].widthPercent != RepoW {
		t.Fatalf("expected first column width to stay %d, got %d", RepoW, active[0].widthPercent)
	}
	if active[1].widthPercent != BranchW+VCSW {
		t.Fatalf("expected last column to reclaim hidden width, got %d", active[1].widthPercent)
	}
}

func TestRedistributeHiddenWidthPrefersExpandableColumn(t *testing.T) {
	active := []columnDef{
		{title: "Repo", widthPercent: RepoW},
		{title: "State", widthPercent: RemoteStateW, expand: true},
		{title: "Branch", widthPercent: BranchW},
	}

	redistributeHiddenWidth(active, VCSW)

	if active[1].widthPercent != RemoteStateW+VCSW {
		t.Fatalf("expected expandable column to reclaim hidden width, got %d", active[1].widthPercent)
	}
	if active[2].widthPercent != BranchW {
		t.Fatalf("expected last column width to stay %d, got %d", BranchW, active[2].widthPercent)
	}
}

func TestStashColumnStr(t *testing.T) {
	tests := []struct {
		name string
		rs   report.RepoState
		want string
	}{
		{name: "empty", rs: report.RepoState{}, want: ""},
		{name: "two stashes", rs: report.RepoState{Stashes: []string{"one", "two"}}, want: "2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := stashColumnStr(tc.rs); got != tc.want {
				t.Fatalf("stashColumnStr() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetStateColumnStr(t *testing.T) {
	tests := []struct {
		name string
		rs   report.RepoState
		want string
	}{
		{name: "no remote", rs: report.RepoState{UncommitedFiles: []string{"one"}}, want: "⏳1 "},
		{name: "clean origin", rs: report.RepoState{RemoteStatus: []report.RemoteStatus{{Remote: "origin"}}}, want: "⏳0 ↑0 ↓0"},
		{name: "ahead and behind", rs: report.RepoState{RemoteStatus: []report.RemoteStatus{{Remote: "origin", Ahead: 2, Behind: 3}}}, want: "⏳0 ↑2 ↓3"},
		{name: "unknown counts", rs: report.RepoState{RemoteStatus: []report.RemoteStatus{{Remote: "origin", Ahead: -1, Behind: -1}}}, want: "⏳0 x x"},
		{name: "named remote", rs: report.RepoState{RemoteStatus: []report.RemoteStatus{{Remote: "upstream"}}}, want: "⏳0 ↑0 ↓0 (upstream)"},
		{name: "multiple remotes", rs: report.RepoState{RemoteStatus: []report.RemoteStatus{{Remote: "origin", Ahead: 1}, {Remote: "", Behind: 1}}}, want: "⏳0 ↑1 ↓0 (origin) | ↑0 ↓1"},
	}

	testTheme := theme.Theme{Styles: theme.Styles{Base: lipgloss.NewStyle()}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := getStateColumnStr(tc.rs, testTheme); got != tc.want {
				t.Fatalf("getStateColumnStr() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetRepoIndex(t *testing.T) {
	repos := []report.RepoState{{ID: "alpha"}, {ID: "beta"}}
	if got := getRepoIndex(repos, "beta"); got != 1 {
		t.Fatalf("getRepoIndex(beta) = %d, want 1", got)
	}
	if got := getRepoIndex(repos, "missing"); got != -1 {
		t.Fatalf("getRepoIndex(missing) = %d, want -1", got)
	}
}
