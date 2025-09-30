package config

import "testing"

func TestDefaults_SensibleValues(t *testing.T) {
	cfg := Defaults()
	if cfg.Output.Type != OutputTable {
		t.Fatalf("expected default output=table, got %v", cfg.Output.Type)
	}
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
