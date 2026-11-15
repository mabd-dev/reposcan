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

	// On Windows, os.UserHomeDir() resolves the home directory by checking USERPROFILE first, then falling back to HOMEDRIVE+HOMEPATH. It only checks HOME as a last resort.
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", dir)
	os.Setenv("USERPROFILE", dir)

	restoreEnv := func(key, val string) {
		if val == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, val)
		}
	}

	return func() {
		restoreEnv("HOME", oldHome)
		restoreEnv("USERPROFILE", oldUserProfile)
	}, dir
}

func TestCreateOrReadConfigs_CreatesWhenMissingAndWritesFile(t *testing.T) {
	restore, _ := withTempHome(t)
	defer restore()

	// pick a nested config file path like ~/.config/reposcan/config.toml
	cfgPath, err := DefaultPaths().GetConfigFullPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := CreateOrReadConfigs(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return defaults and create file
	if len(cfg.Roots) == 0 {
		t.Fatalf("expected defaults with at least one root if HOME is set")
	}

	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("expected config file created at %s, got err: %v", cfgPath, err)
	}
}

func TestCreateOrReadConfigs_ReadsExistingFile(t *testing.T) {
	restore, _ := withTempHome(t)
	defer restore()

	cfgPath, err := DefaultPaths().GetConfigFullPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Seed a custom config
	seeded := Defaults()
	seeded.Roots = []string{"/tmp/foo", "/tmp/bar"}
	seeded.Output.Type = OutputJson
	if err := writeToFile(seeded, cfgPath); err != nil {
		t.Fatalf("seed write error: %v", err)
	}

	// loader should read without overwriting
	got, err := CreateOrReadConfigs(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(got.Roots, ",") != strings.Join(seeded.Roots, ",") || got.Output.Type != OutputJson {
		t.Fatalf("expected to read seeded config, got: %+v", got)
	}
}

// The config path is resolved from the home dir and read as-is, so writes
// must use it as-is too. A home dir with a literal "$" used to be
// env-expanded on write only, sending new configs and saved themes to a
// different file than the one read back.
func TestCreateOrReadConfigs_HomeWithDollarSign(t *testing.T) {
	home := filepath.Join(t.TempDir(), "dev$REPOSCAN_TEST_VAR")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // os.UserHomeDir on Windows
	t.Setenv("REPOSCAN_TEST_VAR", "expanded")

	cfgPath, err := DefaultPaths().GetConfigFullPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := CreateOrReadConfigs(cfgPath); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("expected config created at %s: %v", cfgPath, err)
	}

	// A saved change (e.g. a theme picked in the TUI) must be read back.
	cfg := Defaults()
	cfg.Output.ColorSchemeName = "dracula"
	if err := UpdateConfigs(cfg, cfgPath); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := CreateOrReadConfigs(cfgPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.Output.ColorSchemeName != "dracula" {
		t.Fatalf("expected saved colorscheme to be read back, got %q", got.Output.ColorSchemeName)
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
	cfg.Output.JSONPath = "/definitely/missing/dir"
	v = Validate(cfg)
	if len(v.Warnings) == 0 {
		t.Fatalf("expected a warning when JsonOutputPath does not exist")
	}
}
