package config

import "testing"

func TestOnlyFilterIsValid(t *testing.T) {
	tests := []struct {
		name   string
		filter OnlyFilter
		want   bool
	}{
		{name: "all", filter: OnlyAll, want: true},
		{name: "dirty", filter: OnlyDirty, want: true},
		{name: "uncommitted", filter: OnlyUncommitted, want: true},
		{name: "unpushed", filter: OnlyUnpushed, want: true},
		{name: "unpulled", filter: OnlyUnpulled, want: true},
		{name: "stash", filter: OnlyStash, want: true},
		{name: "unknown", filter: OnlyFilter("broken"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.IsValid(); got != tt.want {
				t.Fatalf("expected IsValid() = %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCreateOnlyFilter(t *testing.T) {
	got, err := CreateOnlyFilter(" stash ")
	if err != nil {
		t.Fatalf("create only filter: %v", err)
	}
	if got != OnlyStash {
		t.Fatalf("expected %q, got %q", OnlyStash, got)
	}
}
