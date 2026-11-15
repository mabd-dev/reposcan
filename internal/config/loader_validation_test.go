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

func TestCreateOrReadConfigs_PartialFileKeepsDefaults(t *testing.T) {
	restore, _ := withTempHome(t)
	defer restore()

	cfgPath, err := DefaultPaths().GetConfigFullPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		toml  string
		check func(t *testing.T, got Config)
	}{
		{
			name: "only roots set",
			toml: `roots = ["/x"]`,
			check: func(t *testing.T, got Config) {
				def := Defaults()
				if strings.Join(got.Roots, ",") != "/x" {
					t.Errorf("roots = %v, want [/x]", got.Roots)
				}
				if got.Only != def.Only || got.Output.Type != def.Output.Type {
					t.Errorf("only/output = %q/%q, want defaults %q/%q", got.Only, got.Output.Type, def.Only, def.Output.Type)
				}
				if len(got.DirIgnore) != len(def.DirIgnore) {
					t.Errorf("dirIgnore has %d entries, want %d defaults", len(got.DirIgnore), len(def.DirIgnore))
				}
				if got.MaxWorkers != def.MaxWorkers || !got.ShowVCSColumn() {
					t.Errorf("maxWorkers=%d showVCS=%v, want defaults", got.MaxWorkers, got.ShowVCSColumn())
				}
			},
		},
		{
			name: "explicit empty dirIgnore clears defaults",
			toml: "dirIgnore = []",
			check: func(t *testing.T, got Config) {
				if len(got.DirIgnore) != 0 {
					t.Errorf("dirIgnore = %v, want empty", got.DirIgnore)
				}
			},
		},
		{
			name: "output table with only type keeps other fields",
			toml: "[output]\ntype = \"json\"\n[tui]\nshowVCS = false",
			check: func(t *testing.T, got Config) {
				if got.Output.Type != OutputJson || got.ShowVCSColumn() {
					t.Errorf("output=%q showVCS=%v, want json/false", got.Output.Type, got.ShowVCSColumn())
				}
				if got.Only != Defaults().Only {
					t.Errorf("only = %q, want default", got.Only)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(cfgPath, []byte(tt.toml), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := CreateOrReadConfigs(cfgPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, got)
		})
	}
}
