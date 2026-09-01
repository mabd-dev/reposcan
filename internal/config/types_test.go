package config

import "testing"

func TestDefaults_WhenUserHomeDirFails_UsesEmptyRoot(t *testing.T) {
	setHomeEnvironment(t, "")

	cfg := Defaults()

	// This characterizes the current fallback rather than endorsing an empty root.
	if len(cfg.Roots) != 1 || cfg.Roots[0] != "" {
		t.Fatalf("expected one empty root, got %v", cfg.Roots)
	}
}
