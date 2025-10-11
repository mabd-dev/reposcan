package reposcan

import (
	"testing"

	"github.com/mabd-dev/reposcan/internal/config"
	"github.com/spf13/cobra"
)

func TestReadFlags_AppliesAllFlags(t *testing.T) {
	// base config defaults
	cfg := config.Defaults()

	// build a throwaway command with the same flags as RootCmd
	cmd := &cobra.Command{Use: "reposcan"}
	cmd.Flags().StringArrayP("root", "r", nil, "")
	cmd.Flags().StringArrayP("dirIgnore", "d", nil, "")
	cmd.Flags().StringP("output", "o", "table", "")
	cmd.Flags().StringP("filter", "f", "dirty", "")
	cmd.Flags().String("json-output-path", "", "")
	cmd.Flags().IntP("max-workers", "w", 8, "")
	cmd.Flags().BoolP("debug", "", false, "")
	// cmd.Flags().StringP("colorscheme", "", "", "")

	args := []string{
		"-r", "/tmp/root1",
		"-r", "/tmp/root2",
		"-d", "**/node_modules/**",
		"-d", "**/.cache/**",
		"-o", "json",
		"-f", "all",
		"--json-output-path", "/tmp/out",
		"-w", "16",
		"--debug", "true",
		// "--colorscheme", "something",
	}
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected execute error: %v", err)
	}

	if err := readFlags(cmd, &cfg); err != nil {
		t.Fatalf("readFlags error: %v", err)
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
		t.Fatalf("debugnot applied: %t", cfg.Debug)
	}
}
