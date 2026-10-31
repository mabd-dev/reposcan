package reposcan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/mabd-dev/reposcan/pkg/report"
)

func TestRootCmdRunE(t *testing.T) {
	tests := []struct {
		name string
		// setup prepares the isolated home and returns the CLI args to parse.
		setup   func(t *testing.T, home string) []string
		wantErr string
	}{
		{
			name: "rejects an invalid output flag",
			setup: func(t *testing.T, home string) []string {
				return []string{"-r", t.TempDir(), "-o", "yaml"}
			},
			wantErr: "yaml",
		},
		{
			name: "rejects a root that does not exist",
			setup: func(t *testing.T, home string) []string {
				return []string{"-r", filepath.Join(home, "missing")}
			},
			wantErr: "invalid configuration",
		},
		{
			name: "rejects an invalid config value no flag overrides",
			setup: func(t *testing.T, home string) []string {
				writeConfig(t, home, fmt.Sprintf("roots = [%q]\nonly = \"dirty\"\n\n[output]\ntype = \"json\"\n", filepath.Join(home, "missing")))
				return nil
			},
			wantErr: "invalid configuration",
		},
		{
			name: "fails when the config path cannot be read",
			setup: func(t *testing.T, home string) []string {
				if runtime.GOOS == "windows" {
					t.Skip("Windows reports a path through a file as not-exist")
				}
				// A file where the config dir should be makes os.Stat fail with ENOTDIR.
				if err := os.WriteFile(filepath.Join(home, ".config"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
				return []string{"-r", t.TempDir()}
			},
			wantErr: "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := isolateHome(t)
			cmd := newRootFlagsCmd()
			parseFlags(t, cmd, tt.setup(t, home)...)

			var err error
			captureStdout(t, func() { err = RootCmd.RunE(cmd, nil) })

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRootCmdRunE_WritesJSONToStdoutAndOutputDir(t *testing.T) {
	isolateHome(t)
	root := t.TempDir()
	outDir := t.TempDir()

	cmd := newRootFlagsCmd()
	parseFlags(t, cmd, "-r", root, "-o", "json", "--json-output-path", outDir, "--no-telemetry")

	var err error
	out := captureStdout(t, func() { err = RootCmd.RunE(cmd, nil) })
	if err != nil {
		t.Fatalf("RunE: %v", err)
	}

	// json.Unmarshal accepts "null", so also require a populated field.
	var stdoutReport report.ScanReport
	if err := json.Unmarshal([]byte(out), &stdoutReport); err != nil || stdoutReport.GeneratedAt.IsZero() {
		t.Fatalf("stdout is not a JSON scan report (err=%v):\n%s", err, out)
	}

	files, err := filepath.Glob(filepath.Join(outDir, "ScanReport *.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("expected one report file in %s, got %v (err=%v)", outDir, files, err)
	}
	saved, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(saved) != strings.TrimSuffix(out, "\n") {
		t.Fatalf("saved report differs from stdout:\n%s\nvs\n%s", saved, out)
	}
}

func TestRootCmdRunE_FlagOverridesInvalidConfigValue(t *testing.T) {
	home := isolateHome(t)
	writeConfig(t, home, fmt.Sprintf("roots = [%q]\nonly = \"dirty\"\n\n[output]\ntype = \"json\"\n", filepath.Join(home, "missing")))

	cmd := newRootFlagsCmd()
	parseFlags(t, cmd, "-r", t.TempDir(), "--no-telemetry")

	var err error
	out := captureStdout(t, func() { err = RootCmd.RunE(cmd, nil) })
	if err != nil {
		t.Fatalf("expected -r to replace the missing config root, got %v", err)
	}
	var scanReport report.ScanReport
	if err := json.Unmarshal([]byte(out), &scanReport); err != nil || scanReport.GeneratedAt.IsZero() {
		t.Fatalf("expected a JSON scan report on stdout (err=%v):\n%s", err, out)
	}
}

// Flags are registered with zero values so --help shows no defaults that
// could contradict the user's config.toml.
func TestRootCmd_HelpShowsNoDefaults(t *testing.T) {
	if usage := RootCmd.PersistentFlags().FlagUsages(); strings.Contains(usage, "(default") {
		t.Fatalf("help output shows a default value:\n%s", usage)
	}
}

func TestRun_Telemetry(t *testing.T) {
	// Clear any token injected via -ldflags so telemetry prints to stdout
	// instead of sending real events to Mixpanel.
	oldToken := mixpanelToken
	mixpanelToken = ""
	t.Cleanup(func() { mixpanelToken = oldToken })

	tests := []struct {
		name        string
		noTelemetry bool
		wantEvent   bool
	}{
		{name: "sends a usage event by default", noTelemetry: false, wantEvent: true},
		{name: "sends nothing with --no-telemetry", noTelemetry: true, wantEvent: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateHome(t)
			cfg := config.Defaults()
			cfg.Roots = []string{t.TempDir()}
			cfg.Output.Type = config.OutputNone
			cfg.NoTelemetry = tt.noTelemetry

			var err error
			out := captureStdout(t, func() { err = run(cfg) })
			if err != nil {
				t.Fatalf("run: %v", err)
			}

			if got := strings.Contains(out, `event="usage"`); got != tt.wantEvent {
				t.Fatalf("usage event sent = %v, want %v; stdout:\n%s", got, tt.wantEvent, out)
			}
		})
	}
}

func TestRun_JSONOutputPathError(t *testing.T) {
	isolateHome(t)
	// A regular file cannot be used as the report directory.
	notADir := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(notADir, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Defaults()
	cfg.Roots = []string{t.TempDir()}
	cfg.Output.Type = config.OutputNone
	cfg.Output.JSONPath = notADir
	cfg.NoTelemetry = true

	if err := run(cfg); err == nil {
		t.Fatal("expected an error writing the report under a regular file")
	}
}
