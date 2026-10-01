package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestValidate_StaleDays(t *testing.T) {
	tests := []struct {
		name         string
		staleDays    int
		only         OnlyFilter
		wantErrors   []string
		wantWarnings []string
	}{
		{name: "disabled", staleDays: 0, only: OnlyDirty},
		{name: "enabled", staleDays: 7, only: OnlyDirty},
		{name: "negative", staleDays: -1, only: OnlyDirty, wantErrors: []string{"staleDays"}},
		{name: "ignored for unpulled", staleDays: 7, only: OnlyUnpulled, wantWarnings: []string{"staleDays"}},
		{name: "disabled with unpulled", staleDays: 0, only: OnlyUnpulled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Defaults()
			cfg.Roots = []string{t.TempDir()}
			cfg.StaleDays = tt.staleDays
			cfg.Only = tt.only

			result := Validate(cfg)

			if got := issueFields(result.Errors); !reflect.DeepEqual(got, orEmpty(tt.wantErrors)) {
				t.Fatalf("errors = %v, want %v", got, tt.wantErrors)
			}
			if got := issueFields(result.Warnings); !reflect.DeepEqual(got, orEmpty(tt.wantWarnings)) {
				t.Fatalf("warnings = %v, want %v", got, tt.wantWarnings)
			}
		})
	}
}

func TestLoad_ReadsStaleDaysFromTOML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("staleDays = 14\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := Load(&cfg, path); err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.StaleDays != 14 {
		t.Fatalf("StaleDays = %d, want 14", cfg.StaleDays)
	}
}

func TestDefaults_StaleDaysDisabled(t *testing.T) {
	if got := Defaults().StaleDays; got != 0 {
		t.Fatalf("Defaults().StaleDays = %d, want 0", got)
	}
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
