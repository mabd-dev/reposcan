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

	args := []string{
		"-r", "/tmp/root1",
		"-r", "/tmp/root2",
		"-d", "**/node_modules/**",
		"-d", "**/.cache/**",
		"-o", "json",
		"-f", "all",
		"--json-output-path", "/tmp/out",
		"-w", "16",
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
	if cfg.Output != config.OutputJson {
		t.Fatalf("output not applied: %v", cfg.Output)
	}
	if cfg.Only != config.OnlyAll {
		t.Fatalf("only filter not applied: %v", cfg.Only)
	}
	if cfg.JsonOutputPath != "/tmp/out" {
		t.Fatalf("json output path not applied: %v", cfg.JsonOutputPath)
	}
	if cfg.MaxWorkers != 16 {
		t.Fatalf("max workers not applied: %d", cfg.MaxWorkers)
	}
}
