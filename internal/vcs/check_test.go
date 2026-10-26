package vcs

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/mabd-dev/reposcan/pkg/report"
)

type activityStubProvider struct {
	stubProvider
	at    time.Time
	calls *atomic.Int32
}

func (p activityStubProvider) CheckRepoStateWithActivity(path string) (report.RepoState, []string) {
	p.calls.Add(1)
	state, warnings := p.CheckRepoState(path)
	state.LastActivity = p.at
	return state, append(warnings, "activity warning")
}

func TestCheckRepoState_ComputeActivity(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		activity     bool
		opts         ScanOptions
		wantActivity time.Time
		wantCalls    int32
		wantWarnings int
	}{
		{name: "disabled", activity: true, opts: ScanOptions{}, wantCalls: 0},
		{name: "enabled", activity: true, opts: ScanOptions{ComputeActivity: true}, wantActivity: at, wantCalls: 1, wantWarnings: 1},
		{name: "enabled without ActivityProvider", activity: false, opts: ScanOptions{ComputeActivity: true}, wantCalls: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := &atomic.Int32{}
			var provider Provider = stubProvider{repoType: TypeGit}
			if tt.activity {
				provider = activityStubProvider{stubProvider: stubProvider{repoType: TypeGit}, at: at, calls: calls}
			}

			state, warnings := CheckRepoState(provider, "/tmp/repo", tt.opts)

			if !state.LastActivity.Equal(tt.wantActivity) {
				t.Fatalf("LastActivity = %v, want %v", state.LastActivity, tt.wantActivity)
			}
			if got := calls.Load(); got != tt.wantCalls {
				t.Fatalf("CheckRepoStateWithActivity calls = %d, want %d", got, tt.wantCalls)
			}
			if len(warnings) != tt.wantWarnings {
				t.Fatalf("warnings = %v, want %d", warnings, tt.wantWarnings)
			}
		})
	}
}

func TestGetRepoStatesConcurrent_ComputesActivityOnlyWhenEnabled(t *testing.T) {
	repos := []RepoInfo{
		{Path: "/tmp/a", Type: TypeGit},
		{Path: "/tmp/b", Type: TypeGit},
	}

	for _, enabled := range []bool{false, true} {
		calls := &atomic.Int32{}
		registry := NewRegistry(activityStubProvider{
			stubProvider: stubProvider{repoType: TypeGit},
			at:           time.Now(),
			calls:        calls,
		})

		states, _ := GetRepoStatesConcurrent(repos, registry, 2, ScanOptions{ComputeActivity: enabled})

		want := int32(0)
		if enabled {
			want = int32(len(repos))
		}
		if got := calls.Load(); got != want {
			t.Fatalf("enabled=%v: CheckRepoStateWithActivity calls = %d, want %d", enabled, got, want)
		}
		for _, s := range states {
			if s.LastActivity.IsZero() == enabled {
				t.Fatalf("enabled=%v: unexpected LastActivity %v for %s", enabled, s.LastActivity, s.Path)
			}
		}
	}
}
