package config

import "testing"

func TestDefaults_SensibleValues(t *testing.T) {
	cfg := Defaults()
	if cfg.Only != OnlyDirty {
		t.Fatalf("expected default only=dirty, got %v", cfg.Only)
	}
	if cfg.MaxWorkers <= 0 {
		t.Fatalf("expected maxWorkers > 0, got %d", cfg.MaxWorkers)
	}
	if len(cfg.Roots) == 0 {
		t.Fatalf("expected at least one default root when HOME is set")
	}
}

func TestConfigShowVCSColumn(t *testing.T) {
	tests := []struct {
		name    string
		showVCS *bool
		want    bool
	}{
		{
			name: "unset defaults to shown",
			want: true,
		},
		{
			name:    "explicit true is shown",
			showVCS: boolPtr(true),
			want:    true,
		},
		{
			name:    "explicit false is hidden",
			showVCS: boolPtr(false),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Tui: Tui{ShowVCS: tt.showVCS}}

			if got := cfg.ShowVCSColumn(); got != tt.want {
				t.Fatalf("expected ShowVCSColumn() = %v, got %v", tt.want, got)
			}
		})
	}
}

func boolPtr(v bool) *bool {
	return &v
}
