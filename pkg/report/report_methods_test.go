package report

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestIsDirty_RemoteAheadBehind(t *testing.T) {
	tests := []struct {
		name       string
		rs         RepoState
		countStash bool
		wantDirty  bool
	}{
		{name: "remote ahead makes dirty", rs: RepoState{RemoteStatus: []RemoteStatus{{Remote: "origin", Ahead: 1}}}, wantDirty: true},
		{name: "remote behind makes dirty", rs: RepoState{RemoteStatus: []RemoteStatus{{Remote: "origin", Behind: 1}}}, wantDirty: true},
		{name: "clean remote stays clean", rs: RepoState{RemoteStatus: []RemoteStatus{{Remote: "origin", Ahead: 0, Behind: 0}}}, wantDirty: false},
		{name: "one clean one dirty remote", rs: RepoState{RemoteStatus: []RemoteStatus{
			{Remote: "origin", Ahead: 0, Behind: 0},
			{Remote: "upstream", Ahead: 2, Behind: 0},
		}}, wantDirty: true},
		{name: "uncommitted overrides clean remote with stash flag", rs: RepoState{
			UncommitedFiles: []string{"a.txt"},
			Stashes:         []string{"stash@{0}"},
			RemoteStatus:    []RemoteStatus{{Remote: "origin", Ahead: 0, Behind: 0}},
		}, countStash: true, wantDirty: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rs.IsDirty(tc.countStash); got != tc.wantDirty {
				t.Fatalf("IsDirty(%v) = %v, want %v", tc.countStash, got, tc.wantDirty)
			}
		})
	}
}

func TestHaveUnpushedCommits(t *testing.T) {
	tests := []struct {
		name string
		rs   RepoState
		want bool
	}{
		{name: "ahead means unpushed", rs: RepoState{RemoteStatus: []RemoteStatus{{Remote: "origin", Ahead: 3}}}, want: true},
		{name: "behind only means no unpushed", rs: RepoState{RemoteStatus: []RemoteStatus{{Remote: "origin", Behind: 1}}}, want: false},
		{name: "no remotes", rs: RepoState{}, want: false},
		{name: "later remote has unpushed", rs: RepoState{RemoteStatus: []RemoteStatus{
			{Remote: "origin", Ahead: 0},
			{Remote: "upstream", Ahead: 1},
		}}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rs.HaveUnpushedCommits(); got != tc.want {
				t.Fatalf("HaveUnpushedCommits() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHaveUnpulledCommits(t *testing.T) {
	tests := []struct {
		name string
		rs   RepoState
		want bool
	}{
		{name: "behind means unpulled", rs: RepoState{RemoteStatus: []RemoteStatus{{Remote: "origin", Behind: 2}}}, want: true},
		{name: "ahead only means no unpulled", rs: RepoState{RemoteStatus: []RemoteStatus{{Remote: "origin", Ahead: 1}}}, want: false},
		{name: "no remotes", rs: RepoState{}, want: false},
		{name: "later remote has unpulled", rs: RepoState{RemoteStatus: []RemoteStatus{
			{Remote: "origin", Behind: 0},
			{Remote: "upstream", Behind: 4},
		}}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rs.HaveUnpulledCommits(); got != tc.want {
				t.Fatalf("HaveUnpulledCommits() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOutgoingCommits_SingleRemote(t *testing.T) {
	rs := RepoState{RemoteStatus: []RemoteStatus{{
		Remote:          "origin",
		OutgoingCommits: []string{"abc123", "def456"},
	}}}

	got := rs.OutgoingCommits()
	want := []string{"abc123", "def456"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OutgoingCommits() = %v, want %v", got, want)
	}
}

func TestOutgoingCommits_MultipleRemotesPrefixName(t *testing.T) {
	rs := RepoState{RemoteStatus: []RemoteStatus{
		{Remote: "origin", OutgoingCommits: []string{"aaa"}},
		{Remote: "upstream", OutgoingCommits: []string{"bbb", "ccc"}},
	}}

	got := rs.OutgoingCommits()
	want := []string{"origin: aaa", "upstream: bbb", "upstream: ccc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OutgoingCommits() = %v, want %v", got, want)
	}
}

func TestOutgoingCommits_EmptyRemoteNameInMultiNotPrefixed(t *testing.T) {
	rs := RepoState{RemoteStatus: []RemoteStatus{
		{Remote: "", OutgoingCommits: []string{"aaa"}},
		{Remote: "upstream", OutgoingCommits: []string{"bbb"}},
	}}

	got := rs.OutgoingCommits()
	want := []string{"aaa", "upstream: bbb"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OutgoingCommits() = %v, want %v", got, want)
	}
}

func TestOutgoingCommits_NoRemotes(t *testing.T) {
	rs := RepoState{}

	got := rs.OutgoingCommits()
	if len(got) != 0 {
		t.Fatalf("OutgoingCommits() = %v, want empty", got)
	}
}

func TestScanReport_MethodsAndFields(t *testing.T) {
	sc := ScanReport{
		Version:           1,
		RepoStates:        []RepoState{{}, {UncommitedFiles: []string{"f"}}},
		TotalScannedRepos: 2,
		GeneratedAt:       time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Warnings:          []string{"w1", "w2"},
	}

	if sc.DirtyReposCount(false) != 1 {
		t.Fatalf("DirtyReposCount(false) = %d, want 1", sc.DirtyReposCount(false))
	}
	if sc.Version != 1 || sc.TotalScannedRepos != 2 || len(sc.Warnings) != 2 {
		t.Fatalf("ScanReport fields not set correctly: %+v", sc)
	}
	if !sc.GeneratedAt.Equal(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)) {
		t.Fatalf("GeneratedAt mismatch: %v", sc.GeneratedAt)
	}
}

func TestScanReport_ZeroValueWarningsIsNil(t *testing.T) {
	// A zero-value ScanReport has a nil Warnings slice, which marshals to
	// JSON "null". Consumers merging/reading reports should not assume "[]".
	sc := ScanReport{}
	if sc.Warnings != nil {
		t.Fatalf("expected nil Warnings on zero-value ScanReport, got %v", sc.Warnings)
	}
	if strings.Join(sc.Warnings, ",") != "" {
		t.Fatalf("unexpected warnings content: %v", sc.Warnings)
	}
}
