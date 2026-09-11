package repostable

import (
	"reflect"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func testReport() report.ScanReport {
	return report.ScanReport{
		RepoStates: []report.RepoState{
			{ID: "alpha", Repo: "Alpha", Branch: "main"},
			{ID: "beta", Repo: "Beta", Branch: "Feature/One"},
		},
		TotalScannedRepos: 2,
	}
}

func TestUpdateWindowSizeResizesTableAndColumns(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{ShowVCS: true})

	got := m.UpdateWindowSize(102, 32)

	if got.width != 100 || got.height != 30 {
		t.Fatalf("model size = %dx%d, want 100x30", got.width, got.height)
	}
	if got.tbl.Width() != 100 || got.tbl.Height() != 29 {
		t.Fatalf("table viewport size = %dx%d, want 100x29", got.tbl.Width(), got.tbl.Height())
	}
	wantColumns := createColumns(100, Options{ShowVCS: true})
	if !reflect.DeepEqual(got.tbl.Columns(), wantColumns) {
		t.Fatalf("columns = %#v, want %#v", got.tbl.Columns(), wantColumns)
	}
}

func TestFilterMatchesRepoAndBranchIgnoringCaseAndWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "repo", query: "  ALP  ", want: "alpha"},
		{name: "branch", query: "feature", want: "beta"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := New(stubTheme(), testReport(), 80, 20, Options{})
			m.Filter(tc.query)

			if m.ReposCount() != 1 || m.filteredRepos[0].ID != tc.want {
				t.Fatalf("filtered repos = %#v, want only %q", m.filteredRepos, tc.want)
			}
			if m.filterQuery != tc.query {
				t.Fatalf("filterQuery = %q, want original query %q", m.filterQuery, tc.query)
			}
		})
	}
}

func TestFilterPreservesValidCursorAndResetsInvalidCursor(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{})
	m.tbl.SetCursor(1)

	m.Filter("")
	if m.Cursor() != 1 {
		t.Fatalf("cursor after unfiltered refresh = %d, want 1", m.Cursor())
	}

	m.Filter("Alpha")
	if m.Cursor() != 0 {
		t.Fatalf("cursor after result set shrinks = %d, want 0", m.Cursor())
	}
}

func TestSetReportReappliesCurrentFilter(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{})
	m.Filter("feature")

	m.SetReport(report.ScanReport{
		RepoStates: []report.RepoState{
			{ID: "gamma", Repo: "Gamma", Branch: "feature/two"},
			{ID: "delta", Repo: "Delta", Branch: "main"},
		},
		TotalScannedRepos: 2,
	})

	if m.ReposCount() != 1 || m.filteredRepos[0].ID != "gamma" {
		t.Fatalf("filtered repos = %#v, want updated matching repo", m.filteredRepos)
	}
}

func TestUpdateRepoStateUpdatesFilteredAndOriginalReports(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{})
	updated := report.RepoState{ID: "beta", Repo: "Beta", Branch: "updated"}

	m.UpdateRepoState(1, updated)

	if got := m.filteredRepos[1]; !reflect.DeepEqual(got, updated) {
		t.Fatalf("filtered state = %#v, want %#v", got, updated)
	}
	if got := m.report.RepoStates[1]; !reflect.DeepEqual(got, updated) {
		t.Fatalf("report state = %#v, want %#v", got, updated)
	}
	if got := m.tbl.Rows()[0][0]; got != "Alpha" {
		t.Fatalf("unchanged row repo = %q, want Alpha", got)
	}
	if got := m.tbl.Rows()[1][1]; got != "updated" {
		t.Fatalf("rendered branch = %q, want updated", got)
	}
}

func TestUpdateRepoStatePersistsFilteredRepoUpdate(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{})
	m.Filter("Beta")
	updated := report.RepoState{ID: "beta", Repo: "Beta", Branch: "updated"}

	m.UpdateRepoState(0, updated)

	if got := m.report.RepoStates[1]; !reflect.DeepEqual(got, updated) {
		t.Fatalf("report state = %#v, want %#v", got, updated)
	}

	m.Filter("")
	if got := m.filteredRepos[1]; !reflect.DeepEqual(got, updated) {
		t.Fatalf("state after clearing filter = %#v, want %#v", got, updated)
	}
	if got := m.tbl.Rows()[1][1]; got != "updated" {
		t.Fatalf("rendered branch after clearing filter = %q, want updated", got)
	}
}

func TestUpdateRepoStateLeavesOriginalReportWhenIDIsUnknown(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{})
	m.Filter("Alpha")
	wantReport := append([]report.RepoState(nil), m.report.RepoStates...)
	wantFilteredRepos := append([]report.RepoState(nil), m.filteredRepos...)
	wantRows := m.tbl.Rows()

	m.UpdateRepoState(0, report.RepoState{ID: "unknown", Repo: "Unknown"})

	if !reflect.DeepEqual(m.report.RepoStates, wantReport) {
		t.Fatalf("report states changed for unknown ID: %#v", m.report.RepoStates)
	}
	if !reflect.DeepEqual(m.filteredRepos, wantFilteredRepos) {
		t.Fatalf("filtered states changed for unknown ID: %#v", m.filteredRepos)
	}
	if !reflect.DeepEqual(m.tbl.Rows(), wantRows) {
		t.Fatalf("table rows changed for unknown ID: %#v", m.tbl.Rows())
	}
}

func TestFocusAndBlur(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{})
	if !m.tbl.Focused() {
		t.Fatal("new table is not focused")
	}

	m.Blur()
	if m.tbl.Focused() {
		t.Fatal("Blur() left table focused")
	}

	m.Focus()
	if !m.tbl.Focused() {
		t.Fatal("Focus() left table blurred")
	}
}

func TestGetRepoState(t *testing.T) {
	m := New(stubTheme(), testReport(), 80, 20, Options{})
	m.tbl.SetCursor(1)

	if got := m.GetCurrentRepoState(); got == nil || got.ID != "beta" {
		t.Fatalf("GetCurrentRepoState() = %#v, want beta", got)
	}
	if got := m.GetRepoStateAt(0); got == nil || got.ID != "alpha" {
		t.Fatalf("GetRepoStateAt(0) = %#v, want alpha", got)
	}
	if got := m.GetRepoStateAt(-1); got != nil {
		t.Fatalf("GetRepoStateAt(-1) = %#v, want nil", got)
	}
	if got := m.GetRepoStateAt(m.ReposCount()); got != nil {
		t.Fatalf("GetRepoStateAt(count) = %#v, want nil", got)
	}
}

func TestUpdateThemeChangesThemeAndTableStyles(t *testing.T) {
	reportWithRemote := testReport()
	reportWithRemote.RepoStates[0].RemoteStatus = []report.RemoteStatus{{Remote: "upstream"}}
	m := New(stubTheme(), reportWithRemote, 80, 20, Options{})
	before := m.tbl.View()
	beforeStateCell := m.tbl.Rows()[0][3]
	updated := stubTheme()
	updated.Colors.Accent = lipgloss.Color("#123456")
	updated.Styles.Base = lipgloss.NewStyle().PaddingLeft(2)
	updated.Styles.TableHeader = lipgloss.NewStyle().PaddingLeft(3)
	updated.Styles.TableRow = lipgloss.NewStyle().PaddingLeft(2)
	updated.Styles.TableSelectedRow = lipgloss.NewStyle().PaddingLeft(1)

	m.UpdateTheme(updated)

	if !reflect.DeepEqual(m.theme, updated) {
		t.Fatal("model theme was not updated")
	}
	if after := m.tbl.View(); after == before {
		t.Fatal("rendered table did not change after table styles were updated")
	}
	if got := m.tbl.Rows()[0][3]; got == beforeStateCell {
		t.Fatalf("state cell retained old theme styling: %q", got)
	} else if want := getStateColumnStr(reportWithRemote.RepoStates[0], updated); got != want {
		t.Fatalf("state cell = %q, want %q", got, want)
	}
}
