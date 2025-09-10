package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// helper to temporarily change HOME for tests
func withTempHome(t *testing.T) (restore func(), tempHome string) {
	t.Helper()
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	// On Windows, Cobra/Go may rely on USERPROFILE, but this repo is Linux oriented.
	os.Setenv("HOME", dir)
	return func() { os.Setenv("HOME", oldHome) }, dir
}

func TestCreateOrReadConfigs_CreatesWhenMissingAndWritesFile(t *testing.T) {
	restore, home := withTempHome(t)
	defer restore()

	// pick a nested config file path like ~/.config/reposcan/config.toml
	cfgRel := ".config/reposcan/config.toml"

	cfg, err := CreateOrReadConfigs(cfgRel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return defaults and create file
	if len(cfg.Roots) == 0 {
		t.Fatalf("expected defaults with at least one root if HOME is set")
	}

	cfgAbs := filepath.Join(home, cfgRel)
	if _, err := os.Stat(cfgAbs); err != nil {
		t.Fatalf("expected config file created at %s, got err: %v", cfgAbs, err)
	}
}

func TestCreateOrReadConfigs_ReadsExistingFile(t *testing.T) {
	restore, home := withTempHome(t)
	defer restore()

	cfgRel := ".config/reposcan/config.toml"
	cfgAbs := filepath.Join(home, cfgRel)

	// Seed a custom config
	seeded := Defaults()
	seeded.Roots = []string{"/tmp/foo", "/tmp/bar"}
	seeded.Output = OutputJson
	if err := WriteToFile(seeded, cfgAbs); err != nil {
		t.Fatalf("seed write error: %v", err)
	}

	// loader should read without overwriting
	got, err := CreateOrReadConfigs(cfgRel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(got.Roots, ",") != strings.Join(seeded.Roots, ",") || got.Output != OutputJson {
		t.Fatalf("expected to read seeded config, got: %+v", got)
	}
}

func TestValidate(t *testing.T) {
	cfg := Defaults()
	// Force an invalid root path to trigger an error
	cfg.Roots = []string{"/path/that/does/not/exist"}

	v := Validate(cfg)
	if len(v.Errors) == 0 {
		t.Fatalf("expected errors when roots do not exist")
	}

	// JsonOutputPath: expect warning when dir missing (non-empty)
	cfg = Defaults()
	cfg.JsonOutputPath = "/definitely/missing/dir"
	v = Validate(cfg)
	if len(v.Warnings) == 0 {
		t.Fatalf("expected a warning when JsonOutputPath does not exist")
	}
}
