package reposcan

import (
	"reflect"
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
)

func TestReadFlags_AppliesAllFlags(t *testing.T) {
	// base config defaults
	cfg := config.Defaults()
	cfg.NoTelemetry = true // so --no-telemetry=false must override it

	cmd := newRootFlagsCmd()

	args := []string{
		"-r", "/tmp/root1",
		"-r", "/tmp/root2",
		"-d", "**/node_modules/**",
		"-d", "**/.cache/**",
		"-o", "json",
		"-f", "all",
		"--json-output-path", "/tmp/out",
		"--stale-days", "7",
		"-w", "16",
		"--debug=true",
		"--no-telemetry=false",
		// "--colorscheme", "something",
	}
	parseFlags(t, cmd, args...)

	if err := readFlags(cmd, &cfg); err != nil {
		t.Fatalf("readFlags error: %v", err)
	}

	if cfg.StaleDays != 7 {
		t.Fatalf("stale days not applied: %d", cfg.StaleDays)
	}
	if len(cfg.Roots) != 2 || cfg.Roots[0] != "/tmp/root1" || cfg.Roots[1] != "/tmp/root2" {
		t.Fatalf("roots not applied: %#v", cfg.Roots)
	}
	if len(cfg.DirIgnore) != 2 {
		t.Fatalf("dirIgnore not applied: %#v", cfg.DirIgnore)
	}
	if cfg.Only != config.OnlyAll {
		t.Fatalf("only filter not applied: %v", cfg.Only)
	}
	if cfg.Output.Type != config.OutputJson {
		t.Fatalf("output not applied: %v", cfg.Output.Type)
	}
	if cfg.Output.JSONPath != "/tmp/out" {
		t.Fatalf("json output path not applied: %v", cfg.Output.JSONPath)
	}
	// if cfg.Output.ColorSchemeName != "something" {
	// 	t.Fatalf("colorscheme not applied: %s", cfg.Output.ColorSchemeName)
	// }
	if cfg.MaxWorkers != 16 {
		t.Fatalf("max workers not applied: %d", cfg.MaxWorkers)
	}
	if cfg.Debug != true {
		t.Fatalf("debug not applied: %t", cfg.Debug)
	}
	if cfg.NoTelemetry != false {
		t.Fatalf("telemetry not applied: %t", cfg.NoTelemetry)
	}
}

func TestReadFlags_UnsetFlagsKeepConfigValues(t *testing.T) {
	cfg := config.Defaults()
	cfg.Roots = []string{"/from/config"}
	cfg.DirIgnore = []string{"**/from-config/**"}
	cfg.Only = config.OnlyUnpushed
	cfg.Output.Type = config.OutputJson
	cfg.Output.JSONPath = "/config/out"
	cfg.StaleDays = 3
	cfg.Debug = true
	cfg.NoTelemetry = true
	want := cfg
	want.MaxWorkers = 2

	cmd := newRootFlagsCmd()
	parseFlags(t, cmd, "-w", "2")

	if err := readFlags(cmd, &cfg); err != nil {
		t.Fatalf("readFlags error: %v", err)
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("unset flags changed the config:\ngot  %#v\nwant %#v", cfg, want)
	}
}

func TestReadTableOutput_DefaultInteractiveOutput(t *testing.T) {
	// base config defaults
	cfg := config.Defaults()

	if cfg.Output.Type != config.OutputInteractive {
		t.Fatalf("output not applied: %v", cfg.Output.Type)
	}

}

func TestReadTableOutput_SwitchToInteractiveOutput(t *testing.T) {
	cfg := config.Defaults()
	cmd := newRootFlagsCmd()
	parseFlags(t, cmd, "-o", "table")

	if err := readFlags(cmd, &cfg); err != nil {
		t.Fatalf("readFlags error: %v", err)
	}

	if cfg.Output.Type != config.OutputInteractive {
		t.Fatalf("output not applied: %v", cfg.Output.Type)
	}
}
