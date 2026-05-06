package repostable

import (
	"testing"

	"github.com/mabd-dev/reposcan/internal/theme"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestCreateColumnsIncludesVCSColumn(t *testing.T) {
	columns := createColumns(100)

	if len(columns) != 4 {
		t.Fatalf("expected 4 columns, got %d: %v", len(columns), columns)
	}
	if columns[2].Title != "VCS" {
		t.Fatalf("expected third column to be VCS, got %q", columns[2].Title)
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
	}, theme.Theme{})

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) != 4 {
		t.Fatalf("expected 4 cells, got %d: %v", len(rows[0]), rows[0])
	}
	if rows[0][2] != "jj" {
		t.Fatalf("expected VCS cell to be jj, got %q", rows[0][2])
	}
	if rows[0][3] != "⏳0 ↑1 ↓2" {
		t.Fatalf("unexpected state cell: %q", rows[0][3])
	}
}
