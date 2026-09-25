package internal

import (
	"testing"
	"time"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestStaleFilter(t *testing.T) {
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour

	fresh := report.RepoState{LastActivity: now.Add(-2 * day)}
	exactlyStale := report.RepoState{LastActivity: now.Add(-7 * day)}
	stale := report.RepoState{LastActivity: now.Add(-30 * day)}
	unknown := report.RepoState{}

	tests := []struct {
		name      string
		only      config.OnlyFilter
		staleDays int
		state     report.RepoState
		want      bool
	}{
		{name: "disabled keeps fresh", only: config.OnlyDirty, staleDays: 0, state: fresh, want: true},
		{name: "fresh is dropped", only: config.OnlyDirty, staleDays: 7, state: fresh, want: false},
		{name: "exactly N days is kept", only: config.OnlyDirty, staleDays: 7, state: exactlyStale, want: true},
		{name: "stale is kept", only: config.OnlyDirty, staleDays: 7, state: stale, want: true},
		{name: "unknown activity is kept", only: config.OnlyDirty, staleDays: 7, state: unknown, want: true},
		{name: "unpulled ignores stale days", only: config.OnlyUnpulled, staleDays: 7, state: fresh, want: true},
		{name: "all applies stale days", only: config.OnlyAll, staleDays: 7, state: fresh, want: false},
		{name: "uncommitted applies stale days", only: config.OnlyUncommitted, staleDays: 7, state: fresh, want: false},
		{name: "unpushed applies stale days", only: config.OnlyUnpushed, staleDays: 7, state: fresh, want: false},
		{name: "stash applies stale days", only: config.OnlyStash, staleDays: 7, state: fresh, want: false},
		{name: "negative is treated as disabled", only: config.OnlyDirty, staleDays: -1, state: fresh, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := staleFilter(tt.only, tt.staleDays, tt.state, now); got != tt.want {
				t.Fatalf("staleFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanOptions_ComputeActivity(t *testing.T) {
	tests := []struct {
		only      config.OnlyFilter
		staleDays int
		want      bool
	}{
		{only: config.OnlyDirty, staleDays: 0, want: false},
		{only: config.OnlyDirty, staleDays: 3, want: true},
		{only: config.OnlyUnpulled, staleDays: 3, want: false},
	}

	for _, tt := range tests {
		cfg := config.Config{Only: tt.only, StaleDays: tt.staleDays}
		if got := ScanOptions(cfg).ComputeActivity; got != tt.want {
			t.Fatalf("ScanOptions(only=%s, staleDays=%d).ComputeActivity = %v, want %v", tt.only, tt.staleDays, got, tt.want)
		}
	}
}
