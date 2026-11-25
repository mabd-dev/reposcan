package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestUpdateConfigsAndLoad_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reposcan", "config.toml")
	want := Defaults()
	want.Roots = []string{"/tmp/a", "/tmp/b"}
	want.DirIgnore = []string{"**/node_modules/**", "**/.git/**"}
	want.Only = OnlyUnpulled
	want.Output = Output{
		Type:            OutputJson,
		JSONPath:        "/tmp/out",
		ColorSchemeName: "catppuccin-latte",
	}
	want.CountStashAsDirty = true
	want.MaxWorkers = 16
	want.Debug = true
	want.NoTelemetry = true
	want.Version = 2

	if err := UpdateConfigs(want, path); err != nil {
		t.Fatalf("update configs: %v", err)
	}

	var got Config
	if err := Load(&got, path); err != nil {
		t.Fatalf("load configs: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config mismatch:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestLoad_WhenFileIsMissing_ReturnsError(t *testing.T) {
	var cfg Config
	if err := Load(&cfg, filepath.Join(t.TempDir(), "missing.toml")); err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestUpdateConfigs_WhenMarshalFails_ReturnsError(t *testing.T) {
	marshalErr := errors.New("marshal failed")
	originalMarshal := tomlMarshal
	tomlMarshal = func(any) ([]byte, error) {
		return nil, marshalErr
	}
	t.Cleanup(func() {
		tomlMarshal = originalMarshal
	})

	err := UpdateConfigs(Defaults(), filepath.Join(t.TempDir(), "config.toml"))
	if !errors.Is(err, marshalErr) {
		t.Fatalf("expected marshal error, got %v", err)
	}
}

func TestSetColorScheme_KeepsOtherKeysAsInFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := "roots = ['/home/me']\nonly = 'dirty'\n\n[output]\ntype = 'json'\ncolorscheme = 'catppuccin-mocha'\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := SetColorScheme(path, "catppuccin-latte"); err != nil {
		t.Fatalf("set colorscheme: %v", err)
	}

	want := map[string]any{
		"roots":  []any{"/home/me"},
		"only":   "dirty",
		"output": map[string]any{"type": "json", "colorscheme": "catppuccin-latte"},
	}
	got := map[string]any{}
	if err := Load(&got, path); err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config mismatch:\n got: %v\nwant: %v", got, want)
	}
}

func TestSetColorScheme_WhenFileIsMissing_WritesOnlyColorScheme(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reposcan", "config.toml")

	if err := SetColorScheme(path, "catppuccin-latte"); err != nil {
		t.Fatalf("set colorscheme: %v", err)
	}

	want := map[string]any{"output": map[string]any{"colorscheme": "catppuccin-latte"}}
	got := map[string]any{}
	if err := Load(&got, path); err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config mismatch:\n got: %v\nwant: %v", got, want)
	}
}

func TestSetColorScheme_WhenFileIsInvalid_ReturnsErrorAndKeepsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := "roots = ["
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := SetColorScheme(path, "catppuccin-latte"); err == nil {
		t.Fatal("expected error for invalid config file")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(b) != content {
		t.Fatalf("config file changed to %q", b)
	}
}

func TestCreateOrReadConfigs_WhenPathStatFails_ReturnsError(t *testing.T) {
	// NUL is rejected by filesystem APIs on all supported platforms.
	if _, err := CreateOrReadConfigs("\x00"); err == nil {
		t.Fatal("expected error for invalid config path")
	}
}
